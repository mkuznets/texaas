package api

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/go-chi/chi"
	"mkuznets.com/go/texaas/internal/tx"
)

type Command struct {
	tx.Command
}

func (cmd *Command) Execute([]string) error {

	router := chi.NewRouter()

	router.Route("/", func(r chi.Router) {
		r.Post("/upload", handler)
	})

	server := &http.Server{
		Addr:    ":7777",
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Warn().Err(err).Msg("server has terminated")
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(32 * 1024 * 1024)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	mf := r.MultipartForm

	fmt.Println(r.Proto)

	for i, file := range mf.File["files"] {
		fmt.Println(i, file.Filename)
	}

	w.WriteHeader(204)
}
