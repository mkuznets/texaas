package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/proto"
	"mkuznets.com/go/texaas/internal/db"
	"mkuznets.com/go/texaas/internal/repo"
	E "mkuznets.com/go/texaas/internal/txs/api/errors"
	"mkuznets.com/go/texaas/internal/workspace/pb"
)

func (api *API) CreateBuild(w http.ResponseWriter, r *http.Request) {
	var data repo.Makefile

	if err := render.DecodeJSON(r.Body, &data); err != nil {
		E.SendError(w, r, nil, http.StatusBadRequest, "invalid input")
		return
	}

	ctx := r.Context()

	uncached := make([]string, 0)

	var buildID uint64

	err := db.Tx(ctx, api.DB, func(tx pgx.Tx) error {
		var taskID uint64
		err := tx.QueryRow(ctx, `INSERT INTO ocher_tasks (name) VALUES ('latexmk') RETURNING id;`).Scan(&taskID)
		if err != nil {
			return err
		}

		err = tx.QueryRow(ctx, `
		INSERT INTO texaas_builds (task_id, base_path, main_source, compiler, latex)
		VALUES ($1, $2, $3, $4, $5) RETURNING id;
		`, taskID, data.RepoPath, data.MainSource, data.Compiler, data.Latex).Scan(&buildID)
		if err != nil {
			return err
		}

		var (
			cacheReady bool
			cacheID    uint64
		)

		for _, input := range data.Inputs {
			err := tx.QueryRow(ctx, `
			INSERT INTO texaas_cache (hash) VALUES ($1)
			ON CONFLICT (hash) DO UPDATE
			SET used_at=statement_timestamp()
			RETURNING id, is_ready;
			`, input.Hash).Scan(&cacheID, &cacheReady)
			if err != nil {
				return err
			}

			if !cacheReady {
				uncached = append(uncached, input.Hash)
			}

			_, err = tx.Exec(ctx, `
			INSERT INTO texaas_inputs (build_id, cache_id, path)
			VALUES ($1, $2, $3);`, buildID, cacheID, input.RepoPath)
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

	result := struct {
		BuildID  uint64 `json:"build_id"`
		Uncashed []string
	}{
		BuildID:  buildID,
		Uncashed: uncached,
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, result)
}

func (api *API) GetBuild(w http.ResponseWriter, r *http.Request) {
	buildID, err := strconv.ParseUint(chi.URLParam(r, "buildID"), 10, 64)
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	var result struct {
		ID        uint64    `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		Status    struct {
			Name      string    `json:"name"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"status"`
	}

	ctx := r.Context()

	err = db.Tx(ctx, api.DB, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
		SELECT b.id, b.created_at, st.status, st.changed_at
		FROM texaas_builds as b JOIN ocher_statuses as st
		ON (st.task_id = b.task_id)
		WHERE b.id=$1 ORDER BY st.id DESC LIMIT 1;
		`, buildID).Scan(&result.ID, &result.CreatedAt, &result.Status.Name, &result.Status.UpdatedAt)
		if err != nil {
			if err == pgx.ErrNoRows {
				return E.New(nil, http.StatusNotFound, "build not found")
			}
			return err
		}

		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.JSON(w, r, &result)
}

func (api *API) StartBuild(w http.ResponseWriter, r *http.Request) {
	buildID, err := strconv.ParseUint(chi.URLParam(r, "buildID"), 10, 64)
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	ctx := r.Context()

	args := &pb.Args{}

	err = db.Tx(ctx, api.DB, func(tx pgx.Tx) error {
		var taskID uint64

		err := tx.QueryRow(ctx, `
		SELECT task_id, base_path, main_source, compiler, latex FROM texaas_builds WHERE id=$1;
		`, buildID).Scan(&taskID, &args.BasePath, &args.MainSource, &args.Compiler, &args.Latex)
		if err != nil {
			if err == pgx.ErrNoRows {
				return E.New(nil, http.StatusNotFound, "invalid build")
			}
			return err
		}

		rows, err := tx.Query(ctx, `
		SELECT inp.path, c.hash, c.is_ready
		FROM texaas_inputs as inp JOIN texaas_cache as c
		ON (inp.cache_id = c.id)
		WHERE inp.build_id=$1
		`, buildID)
		if err != nil {
			return err
		}

		err = db.IterRows(rows, func(rows pgx.Rows) error {
			isReady := false
			file := &pb.File{}
			if err := rows.Scan(&file.Path, &file.Key, &isReady); err != nil {
				return err
			}
			if !isReady {
				return E.New(nil, http.StatusBadRequest, "build is not ready")
			}
			args.Files = append(args.Files, file)
			return nil
		})
		if err != nil {
			return err
		}

		args, err := proto.Marshal(args)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
		UPDATE ocher_tasks SET status='ENQUEUED', args=$2 WHERE id=$1
		`, taskID, args)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(`{}`)); err != nil {
		E.Handle(w, r, err)
	}
}
