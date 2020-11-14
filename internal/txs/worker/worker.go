package worker

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"mkuznets.com/go/texaas/internal/docker"
	"mkuznets.com/go/texaas/internal/docker/run"
	"mkuznets.com/go/texaas/internal/opts"
	"mkuznets.com/go/texaas/internal/workspace"
	"mkuznets.com/go/texaas/internal/workspace/pb"

	"mkuznets.com/go/ocher"
	"mkuznets.com/go/ocher/log/zerologadapter"
	"mkuznets.com/go/texaas/internal/txs"
)

var (
	texLive = map[string]string{
		"texlive:2013": "2013-ubuntu-2020.11.13",
		"texlive:2019": "2019-ubuntu-2020.11.13",
	}
)

type Command struct {
	Docker   *opts.Docker `group:"Docker" namespace:"docker" env-namespace:"DOCKER"`
	CacheDir string       `long:"cache-dir" env:"CACHE_DIR" description:"input cache path" required:"true"`
	txs.Command
}

func (cmd *Command) LatexMk(ctx context.Context, task ocher.Task) ([]byte, error) {

	var (
		result     []byte
		buildError error
	)

	err := func() error {
		mf := &pb.Makefile{}
		if err := proto.Unmarshal(task.Args(), mf); err != nil {
			return err
		}

		imageVersion, ok := texLive[mf.Latex]
		if !ok {
			return fmt.Errorf("unknown latex version")
		}

		ws := workspace.New()
		if err := ws.Connect(ctx); err != nil {
			return err
		}
		defer ws.Close()

		output, err := ws.Run(ctx, mf, func(ws *workspace.Workspace) error {
			dc, err := docker.New()
			if err != nil {
				return err
			}
			options := []run.Option{
				run.Image("ghcr.io/mkuznets/texlive", imageVersion),
				run.Cmd("/tools/build.sh", mf.BasePath, mf.MainSource),
				run.Mount(ws.Latex, "/latex", ""),
				run.Mount(ws.Tools, "/tools", "ro"),
				run.EnableNetwork(false),
				run.Autoremove(true),
				run.UID(cmd.Docker.UID, cmd.Docker.GID),
				run.Memory(512 * 1024 * 1024),
			}

			buildError = dc.Run(ctx, options...)
			return nil
		})
		if err != nil {
			return err
		}

		result, err = proto.Marshal(output)
		if err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		log.Err(err).Msg("internal task error")
		return nil, fmt.Errorf("internal system error")
	}

	if errors.Is(buildError, docker.ExitError) {
		buildError = fmt.Errorf("build failed, see output.log for details")
	}

	return result, buildError
}

func (cmd *Command) Execute([]string) error {
	q := ocher.NewWorker(
		"127.0.0.1:50051",
		ocher.WorkerID("lapworth"),
		ocher.WorkerLogger(zerologadapter.NewLogger(log.Logger)),
	)

	q.RegisterTask("latexmk", cmd.LatexMk)
	if err := q.Serve(); err != nil {
		return err
	}

	return nil
}
