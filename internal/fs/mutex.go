package fs

import (
	"errors"
	"os"
)

type Mutex struct {
	Path string
}

func (m *Mutex) Lock() (ok bool, err error) {
	f, err := os.OpenFile(m.Path, os.O_CREATE|os.O_EXCL, os.FileMode(600))
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			err = nil
		}
		return false, err
	}
	_ = f.Close()
	return true, nil
}

func (m *Mutex) Unlock() error {
	return os.Remove(m.Path)
}
