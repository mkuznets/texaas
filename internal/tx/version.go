package tx

import (
	"fmt"

	"mkuznets.com/go/texaas/internal/version"
)

type VersionCommand struct {
	Command
}

func (cmd *VersionCommand) Execute([]string) error {
	fmt.Print(version.Version())
	return nil
}
