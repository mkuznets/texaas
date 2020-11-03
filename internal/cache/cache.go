package cache

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"path/filepath"
	"strings"
)

var hashFuncs = map[string]func() hash.Hash{
	"sha256": sha256.New,
}

type Item struct {
	Alg string
	Sum string
}

func NewItem(filename string) (*Item, error) {
	ps := strings.Split(filename, ":")
	if len(ps) != 2 {
		return nil, fmt.Errorf("invalid format: %s", filename)
	}

	alg, sum := ps[0], ps[1]

	if _, ok := hashFuncs[alg]; !ok {
		return nil, fmt.Errorf("unknown hash algorithm: %s", alg)
	}

	return &Item{Alg: alg, Sum: sum}, nil
}

func (item *Item) Path(base string) string {
	return filepath.Join(base, item.Alg, item.Sum[:2], item.Sum[2:4], item.Sum)
}

func (item *Item) Validate(f io.Reader) error {
	hashFunc, ok := hashFuncs[item.Alg]
	if !ok {
		panic("unknown hash algorithm")
	}

	h := hashFunc()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	sum := fmt.Sprintf("%x", h.Sum(nil))

	if sum != item.Sum {
		return fmt.Errorf("hash mismatch")
	}

	return nil
}
