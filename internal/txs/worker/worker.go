package worker

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"mkuznets.com/go/texaas/internal/workspace/pb"

	"mkuznets.com/go/ocher"
	"mkuznets.com/go/ocher/log/zerologadapter"
	"mkuznets.com/go/texaas/internal/txs"
)

type Command struct {
	CacheDir string `long:"cache-dir" env:"CACHE_DIR" description:"input cache path" required:"true"`
	txs.Command
}

func LatexMk(ctx context.Context, task ocher.Task) ([]byte, error) {

	args := pb.Args{}
	if err := proto.Unmarshal(task.Args(), &args); err != nil {
		return nil, err
	}

	conn, err := grpc.DialContext(ctx, "127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	workspace := pb.NewWorkspaceClient(conn)

	ws, err := workspace.New(ctx, &args)
	if err != nil {
		return nil, err
	}

	closeStream, err := workspace.Close(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = closeStream.Send(ws)
		_ = closeStream.CloseSend()
	}()

	fmt.Println(ws.Path)
	fmt.Println(ws.Volume)

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
