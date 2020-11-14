package workspace

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"mkuznets.com/go/texaas/internal/workspace/pb"
)

type Workspace struct {
	Latex, Tools string
}

type Client struct {
	conn    *grpc.ClientConn
	service pb.WorkspaceClient
	target  string
}

func New() *Client {
	client := &Client{
		target: "127.0.0.1:50052",
	}
	return client
}

func (client *Client) Connect(ctx context.Context) error {
	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Debug().Str("target", client.target).Msg("workspace: connecting...")

	conn, err := grpc.DialContext(connCtx, client.target, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		if connCtx.Err() == context.DeadlineExceeded {
			err = errors.Errorf("workspace: timed out: %s", client.target)
		}
		return err
	}

	log.Debug().Msg("workspace: ok")
	client.conn = conn
	client.service = pb.NewWorkspaceClient(conn)

	return nil
}

func (client *Client) Close() {
	log.Debug().Msg("workspace: closing")
	err := client.conn.Close()
	if err != nil {
		log.Err(err).Msg("workspace: closing error")
	}
}

func (client *Client) Run(ctx context.Context, mf *pb.Makefile, fn func(ws *Workspace) error) (*pb.WSOutput, error) {
	stream, err := client.service.Get(ctx)
	if err != nil {
		return nil, err
	}

	log.Debug().Msg("workspace: sending request")
	if err := stream.Send(&pb.WSReq{Makefile: mf}); err != nil {
		return nil, err
	}

	workspace, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	if workspace.Closed {
		return nil, fmt.Errorf("could not get workspace")
	}

	log.Debug().Str("id", workspace.Id.Id).Msg("workspace: locked successfully")

	defer func() {
		if err := stream.CloseSend(); err != nil {
			log.Err(err).Msg("CloseSend")
		}
		// wait for the workspace to cleanup
		if _, err = stream.Recv(); err != nil {
			log.Err(err).Str("id", workspace.Id.Id).Msg("workspace: cleanup error")
		}
	}()

	ws := &Workspace{
		Latex: filepath.Join(workspace.Path, "merged"),
		Tools: workspace.Tools,
	}

	log.Debug().Str("id", workspace.Id.Id).Msg("workspace: running the function")

	if err := fn(ws); err != nil {
		log.Err(err).Str("id", workspace.Id.Id).Msg("workspace: task error")
		return nil, err
	}

	log.Debug().Str("id", workspace.Id.Id).Msg("workspace: fetching output")

	out, err := client.service.Output(ctx, workspace.Id)
	if err != nil {
		log.Err(err).Str("id", workspace.Id.Id).Msg("workspace: output error")
		return nil, err
	}

	return out, nil
}
