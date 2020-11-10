package workspace

import (
	"net"
	"time"

	"google.golang.org/grpc"
	"mkuznets.com/go/texaas/internal/txs"
	"mkuznets.com/go/texaas/internal/workspace"
	"mkuznets.com/go/texaas/internal/workspace/pb"
)

type Command struct {
	CacheDir string `long:"cache-dir" env:"CACHE_DIR" description:"input cache path" required:"true"`
	txs.Command
}

func (cmd *Command) Execute([]string) error {

	g := grpc.NewServer(grpc.ConnectionTimeout(10 * time.Second))

	server := workspace.NewService(cmd.CacheDir)
	pb.RegisterWorkspaceServer(g, server)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		return err
	}

	if err := g.Serve(lis); err != nil {
		return err
	}

	return nil
}
