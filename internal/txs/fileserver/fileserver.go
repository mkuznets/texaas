package fileserver

import (
	"net/http"

	"mkuznets.com/go/texaas/internal/opts"
	"mkuznets.com/go/texaas/internal/txs"
)

type Command struct {
	Output *opts.OutputFileServer `group:"Output" namespace:"output" env-namespace:"OUTPUT"`
	txs.Command
}

func (cmd *Command) Execute([]string) error {
	fs := http.FileServer(http.Dir(cmd.Output.Dir))
	if err := http.ListenAndServe(cmd.Output.Addr, fs); err != nil {
		return err
	}
	return nil
}
