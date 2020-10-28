package repo

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"mkuznets.com/go/texaas/internal/fs"
)

type Input struct {
	AbsPath  string
	RepoPath string `json:"path"`
	Hash     string `json:"hash"`
}

type Makefile struct {
	Inputs []*Input
}

type makeRaw struct {
	ID       string
	TexLive  string
	Compiler string
	Inputs   []string
	Targets  []string
}

type InputsNotExistError struct {
	Missing []string
}

func (e *InputsNotExistError) Error() string {
	return fmt.Sprintf("Missing inputs:\n%s", strings.Join(e.Missing, "\n"))
}

func (repo *Repo) Makefile(mfPath string) (*Makefile, error) {

	mfPath, err := filepath.Abs(filepath.Clean(mfPath))
	if err != nil {
		return nil, err
	}

	makefile := new(Makefile)

	mk, err := parseYaml(mfPath)
	if err != nil {
		return nil, err
	}

	missing := make([]string, 0)
	inputs := make([]*Input, 0, len(mk.Inputs))

	for _, path := range mk.Inputs {
		var fullPath string
		if filepath.IsAbs(path) {
			fullPath = filepath.Join(repo.Root, path)
		} else {
			fullPath = filepath.Join(repo.WorkDir, path)
		}
		fullPath = filepath.Clean(fullPath)

		fi, err := os.Lstat(fullPath)
		switch {
		case err != nil:
			if os.IsNotExist(err) {
				missing = append(missing, path)
				continue
			} else {
				return nil, err
			}
		case !fi.Mode().IsRegular():
			return nil, fmt.Errorf("input is not a regular file: %s", path)
		}

		if isPrefix, err := fs.HasFilepathPrefix(filepath.Dir(fullPath), repo.Root); err != nil {
			return nil, err
		} else if !isPrefix {
			return nil, fmt.Errorf("input is not in the repository: %s", path)
		}

		repoPath, err := filepath.Rel(repo.Root, fullPath)
		if err != nil {
			return nil, err
		}

		inputs = append(inputs, &Input{
			AbsPath:  fullPath,
			RepoPath: repoPath,
		})
	}

	if len(missing) > 0 {
		return nil, &InputsNotExistError{Missing: missing}
	}

	for _, input := range inputs {
		h, err := hash(input.AbsPath)
		if err != nil {
			return nil, err
		}
		input.Hash = h
	}

	makefile.Inputs = inputs

	return makefile, nil
}

func parseYaml(path string) (mkr *makeRaw, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := f.Close(); e != nil {
			err = e
		}
	}()

	mkr = &makeRaw{}
	if err := yaml.NewDecoder(f).Decode(&mkr); err != nil {
		return nil, err
	}
	return
}

func hash(src string) (_ string, err error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer func() {
		if e := in.Close(); e != nil {
			err = e
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, in); err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}
