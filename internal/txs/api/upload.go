package api

import (
	"compress/gzip"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"
	"mkuznets.com/go/texaas/internal/cache"
	"mkuznets.com/go/texaas/internal/db"
	E "mkuznets.com/go/texaas/internal/txs/api/errors"
)

func (api *API) Upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 * 1024 * 1024)
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	mf := r.MultipartForm
	defer func() {
		err := mf.RemoveAll()
		if err != nil {
			log.Err(err).Msg("could not close multipart form")
		}
	}()

	for _, file := range mf.File["files"] {
		if err := handleFile(file, api.CacheDir); err != nil {
			E.Handle(w, r, err)
			return
		}
	}

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		for _, file := range mf.File["files"] {
			_, err := tx.Exec(r.Context(), `
			UPDATE texaas_cache SET is_ready='t' WHERE hash=$1
			`, file.Filename)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		E.Handle(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, map[string]string{"result": "ok"})
}

func handleFile(file *multipart.FileHeader, cacheDir string) (err error) {
	f, err := file.Open()
	if err != nil {
		return err
	}

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	item, err := cache.NewItem(file.Filename)
	if err != nil {
		return err
	}
	if err := item.Validate(gz); err != nil {
		return E.New(err, http.StatusBadRequest, "validation error")
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := gz.Reset(f); err != nil {
		return err
	}

	path := item.Path(cacheDir)

	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0755)); err != nil {
		return err
	}

	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		return err
	}
	defer func() {
		if e := dst.Close(); e != nil {
			err = e
		}
	}()

	if _, err := io.Copy(dst, gz); err != nil {
		return err
	}

	return nil
}
