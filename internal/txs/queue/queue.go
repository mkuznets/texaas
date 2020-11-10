package queue

import (
	"net"

	"github.com/rs/zerolog/log"
	"mkuznets.com/go/ocher"
	"mkuznets.com/go/ocher/log/zerologadapter"
	"mkuznets.com/go/texaas/internal/opts"
	"mkuznets.com/go/texaas/internal/txs"
)

type Command struct {
	DB *opts.DB `group:"PostgreSQL" namespace:"db" env-namespace:"DB"`
	txs.Command
}

func (cmd *Command) Execute([]string) error {
	srv := ocher.NewServer(cmd.DB.DSN(), ocher.ServerLogger(zerologadapter.NewLogger(log.Logger)))

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}

	return nil
}
