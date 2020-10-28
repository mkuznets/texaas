package build

import (
	"mkuznets.com/go/texaas/internal/repo"
	"mkuznets.com/go/texaas/internal/tx"
)

type Command struct {
	tx.Command
}

func (cmd *Command) Execute([]string) error {

	r, er := repo.New("./")
	if er != nil {
		panic(er)
	}

	_, er = r.Makefile("./tx.make.yml")
	if er != nil {
		panic(er)
	}

	return nil
}
