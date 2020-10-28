package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"mkuznets.com/go/texaas/internal/fs"
)

type Repo struct {
	Root    string
	WorkDir string
}

func New(workDir string) (*Repo, error) {

	workDir, err := filepath.Abs(filepath.Clean(workDir))
	if err != nil {
		return nil, err
	}

	isDir, err := fs.IsDir(workDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if !isDir {
		return nil, fmt.Errorf("not a directory: %s", workDir)
	}

	repo := new(Repo)
	repo.WorkDir = workDir

	path := workDir
For:
	for {
		_, err := os.Lstat(filepath.Join(path, ".txroot"))
		switch {
		case err == nil:
			repo.Root = path
			break For
		case os.IsNotExist(err):
			parent := filepath.Dir(path)
			if parent == path {
				return nil, fmt.Errorf("could not find repository root")
			}
			path = parent
		default:
			return nil, err
		}
	}

	return repo, nil
}
