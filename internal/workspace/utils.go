package workspace

import (
	"io"
	"os"
)

const (
	UID = 2000
	GID = 2000
)

func mkDir(path string) error {
	if err := os.MkdirAll(path, os.FileMode(0755)); err != nil {
		return err
	}
	if err := os.Chown(path, UID, GID); err != nil {
		return err
	}
	return nil
}

func copyFile(src, dst string) (err error) {
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
