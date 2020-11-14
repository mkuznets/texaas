package main

import (
	"mkuznets.com/go/texaas/internal/txs"
	"mkuznets.com/go/texaas/internal/txs/api"
	"mkuznets.com/go/texaas/internal/txs/fileserver"
	"mkuznets.com/go/texaas/internal/txs/queue"
	"mkuznets.com/go/texaas/internal/txs/worker"
	"mkuznets.com/go/texaas/internal/txs/workspace"
)

type Options struct {
	Common     *txs.Options        `group:"Common Options" env-namespace:"TXS"`
	API        *api.Command        `command:"api" description:""`
	Queue      *queue.Command      `command:"queue" description:""`
	Worker     *worker.Command     `command:"worker" description:""`
	Workspace  *workspace.Command  `command:"workspace" description:""`
	Fileserver *fileserver.Command `command:"fileserver" description:""`
}
