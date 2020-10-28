package main

import (
	"mkuznets.com/go/texaas/internal/tx"
	"mkuznets.com/go/texaas/internal/tx/build"
)

type Options struct {
	Common  *tx.Options        `group:"Common Options"`
	Build   *build.Command     `command:"build" description:""`
	Version *tx.VersionCommand `command:"version" description:""`
}
