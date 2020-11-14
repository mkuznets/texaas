package workspace

import (
	"fmt"
	"net"
	"os/user"
	"time"

	"google.golang.org/grpc"
	"mkuznets.com/go/texaas/internal/opts"
	"mkuznets.com/go/texaas/internal/txs"
	"mkuznets.com/go/texaas/internal/workspace"
	"mkuznets.com/go/texaas/internal/workspace/pb"
)

type Command struct {
	Docker   *opts.Docker `group:"Docker" namespace:"docker" env-namespace:"DOCKER"`
	Output   *opts.Output `group:"Output" namespace:"output" env-namespace:"OUTPUT"`
	CacheDir string       `long:"cache-dir" env:"CACHE_DIR" description:"input cache path" required:"true"`
	txs.Command
}

func (cmd *Command) Execute([]string) error {

	u, err := user.Current()
	if err != nil {
		return err
	}
	if u.Uid != "0" {
		return fmt.Errorf("workspace service must be run as root")
	}

	toolsDir, err := initTools(cmd.Docker.UID, cmd.Docker.GID)
	if err != nil {
		return err
	}

	g := grpc.NewServer(grpc.ConnectionTimeout(10 * time.Second))

	server := workspace.NewService(cmd.CacheDir, toolsDir, cmd.Docker, cmd.Output)
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
