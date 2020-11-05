package workspace

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"mkuznets.com/go/texaas/internal/cache"
	"mkuznets.com/go/texaas/internal/tasks/latexmk"
)

type Service struct {
	cacheDir string
	latexmk.UnimplementedWorkspaceServer
}

func (s *Service) New(_ context.Context, args *latexmk.Args) (*latexmk.WS, error) {
	basePath, err := ioutil.TempDir("", "texaas_*")
	if err != nil {
		return nil, err
	}

	for _, name := range []string{"lower/texaas", "upper", "work", "merged"} {
		path := filepath.Join(basePath, name)
		if err := mkDir(path); err != nil {
			return nil, err
		}
	}

	repoDir := filepath.Join(basePath, "lower", "texaas")

	for _, file := range args.Files {
		item, err := cache.NewItem(file.Key)
		if err != nil {
			return nil, err
		}
		src := item.Path(s.cacheDir)
		dst := filepath.Join(repoDir, file.Path)

		if err := mkDir(filepath.Dir(dst)); err != nil {
			return nil, err
		}

		if err := copyFile(src, dst); err != nil {
			return nil, err
		}
	}

	merged := filepath.Join(basePath, "merged")
	opts := fmt.Sprintf(
		"lowerdir=%s,upperdir=%s,workdir=%s",
		filepath.Join(basePath, "lower"),
		filepath.Join(basePath, "upper"),
		filepath.Join(basePath, "work"),
	)

	if err := syscall.Mount("overlay", merged, "overlay", 0, opts); err != nil {
		return nil, err
	}

	result := &latexmk.WS{
		Path:   basePath,
		Volume: merged,
	}

	return result, nil
}

func (s *Service) Close(stream latexmk.Workspace_CloseServer) error {
	ws, err := stream.Recv()
	if err != nil {
		return err
	}

	if err := syscall.Unmount(ws.Volume, 0); err != nil {
		return err
	}
	if err := os.RemoveAll(ws.Path); err != nil {
		return err
	}

	if err := stream.SendAndClose(&latexmk.Empty{}); err != nil {
		return err
	}

	return nil
}
