package worker

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"mkuznets.com/go/texaas/internal/tasks/latexmk"

	"mkuznets.com/go/ocher"
	"mkuznets.com/go/ocher/log/zerologadapter"
	"mkuznets.com/go/texaas/internal/txs"
)

type Command struct {
	CacheDir string `long:"cache-dir" env:"CACHE_DIR" description:"input cache path" required:"true"`
	txs.Command
}

func LatexMk(_ context.Context, task ocher.Task) ([]byte, error) {

	args := latexmk.Args{}
	if err := proto.Unmarshal(task.Args(), &args); err != nil {
		return nil, err
	}

	err := task.Report(fmt.Sprintf("Got %d files", len(args.Files)))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (cmd *Command) Execute([]string) error {
	q := ocher.NewWorker(
		"127.0.0.1:50051",
		ocher.WorkerID("lapworth"),
		ocher.WorkerLogger(zerologadapter.NewLogger(log.Logger)),
	)

	q.RegisterTask("latexmk", LatexMk)
	if err := q.Serve(); err != nil {
		return err
	}

	return nil
}
