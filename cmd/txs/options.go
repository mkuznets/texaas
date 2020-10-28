package main

import (
	"mkuznets.com/go/texaas/internal/txs"
	"mkuznets.com/go/texaas/internal/txs/api"
)

type Options struct {
	Common *txs.Options `group:"Common Options"`
	API    *api.Command `command:"api" description:""`
}
