package api

import (
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
	"mkuznets.com/go/texaas/internal/opts"
	"mkuznets.com/go/texaas/internal/txs"

	"github.com/go-chi/chi"
)

type Command struct {
	DB       *opts.DB `group:"PostgreSQL" namespace:"db" env-namespace:"DB"`
	Addr     string   `long:"addr" env:"ADDR" description:"HTTP service address" default:"127.0.0.1:7777"`
	CacheDir string   `long:"cache-dir" env:"CACHE_DIR" description:"input cache path" required:"true"`
	txs.Command
}

func (cmd *Command) Execute([]string) error {

	router := chi.NewRouter()

	pool, err := cmd.DB.GetPool()
	if err != nil {
		return err
	}

	api := &API{
		DB:       pool,
		CacheDir: cmd.CacheDir,
	}

	router.Route("/", func(r chi.Router) {
		r.Post("/builds", api.CreateBuild)
		r.Get("/builds/{buildID:[0-9]+}", api.GetBuild)
		r.Post("/builds/{buildID:[0-9]+}/start", api.StartBuild)
		r.Post("/upload", api.Upload)
	})

	server := &http.Server{
		Addr:    cmd.Addr,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Warn().Err(err).Msg("server has terminated")
	}

	return nil
}

type API struct {
	DB       *pgxpool.Pool
	CacheDir string
}
