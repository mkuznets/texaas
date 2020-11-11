package fs

import (
	"os"
	"path/filepath"
)

func Chown(path string, uid, gid int) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Lchown(path, uid, gid)
	})
}
