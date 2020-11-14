package workspace

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rakyll/statik/fs"
	ufs "mkuznets.com/go/texaas/internal/fs"
	_ "mkuznets.com/go/texaas/internal/workspace/tools"
)

func initTools(uid, gid int) (string, error) {

	tmpDir := filepath.Join(os.TempDir(), "texaas")
	basePath, err := ioutil.TempDir(tmpDir, "tools_*")
	if err != nil {
		return "", err
	}

	toolsFs, err := fs.NewWithNamespace("tools")
	if err != nil {
		return "", err
	}

	err = fs.Walk(toolsFs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		dst := filepath.Join(basePath, path)

		if info.IsDir() {
			if err := os.MkdirAll(dst, os.FileMode(0755)); err != nil {
				return err
			}
			return nil
		}

		fout, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		fin, err := toolsFs.Open(path)
		if err != nil {
			return err
		}
		if err := copyIO(fin, fout); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	if err := ufs.Chown(basePath, uid, gid); err != nil {
		return "", err
	}

	return basePath, nil
}

func copyIO(fin io.ReadCloser, fout io.WriteCloser) (err error) {
	defer func() {
		if e := fout.Close(); e != nil {
			err = e
		}
	}()

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
