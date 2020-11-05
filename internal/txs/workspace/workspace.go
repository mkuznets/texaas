package workspace

import (
	"io"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"mkuznets.com/go/texaas/internal/tasks/latexmk"
	"mkuznets.com/go/texaas/internal/txs"
)

const (
	UID = 2000
	GID = 2000
)

type Command struct {
	CacheDir string `long:"cache-dir" env:"CACHE_DIR" description:"input cache path" required:"true"`
	txs.Command
}

func (cmd *Command) Execute([]string) error {

	g := grpc.NewServer(grpc.ConnectionTimeout(10 * time.Second))

	server := &Service{
		cacheDir: cmd.CacheDir,
	}
	latexmk.RegisterWorkspaceServer(g, server)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		return err
	}

	if err := g.Serve(lis); err != nil {
		return err
	}

	return nil
}

func mkDir(path string) error {
	if err := os.MkdirAll(path, os.FileMode(0755)); err != nil {
		return err
	}
	if err := os.Chown(path, UID, GID); err != nil {
		return err
	}
	return nil
}

func copyFile(src, dst string) error {
	fout, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0400))
	if err != nil {
		return err
	}
	defer func() {
		if e := fout.Close(); e != nil {
			err = e
		}
	}()

	if err := fout.Chown(UID, GID); err != nil {
		return err
	}

	fin, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if e := fin.Close(); e != nil {
			err = e
		}
	}()

	if _, err = io.Copy(fout, fin); err != nil {
		return err
	}

	return nil
}
