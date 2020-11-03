VERSION := $(shell cat VERSION)
REVISION := "" #$(shell scripts/git-rev.sh)
BUILD_TIME := $(shell TZ=Etc/UTC date +'%Y-%m-%dT%H:%M:%SZ')

VERSION_FLAG := -X mkuznets.com/go/texaas/internal/version.version=${VERSION}
REVISION_FLAG := -X mkuznets.com/go/texaas/internal/version.revision=${REVISION}
BUILD_TIME_FLAG := -X mkuznets.com/go/texaas/internal/version.buildTime=${BUILD_TIME}
LDFLAGS := "-s -w ${VERSION_FLAG} ${REVISION_FLAG} ${BUILD_TIME_FLAG}"

all: texaas

texaas:
	export CGO_ENABLED=0
	go generate ./...
	go build -ldflags=${LDFLAGS} mkuznets.com/go/texaas/cmd/tx
	go build -ldflags=${LDFLAGS} mkuznets.com/go/texaas/cmd/txs

.PHONY: texaas
