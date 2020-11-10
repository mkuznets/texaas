package workspace

import (
	"io"
)



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
