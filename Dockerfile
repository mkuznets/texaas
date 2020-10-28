FROM ghcr.io/mkuznets/build-go:1.15.2-2020.09.20 as build
WORKDIR /build
# Cache go modules
ADD go.sum go.mod /build/
RUN go mod download

# Build [and lint] the thing
ADD . /build
#RUN golangci-lint run --out-format=tab --tests=false ./...

RUN go build -ldflags="-s -w" mkuznets.com/go/texaas/cmd/txs

FROM ghcr.io/mkuznets/alpine:3.12-2020.09.20 as txs
COPY --from=build /build/txs /srv/txs
WORKDIR /srv
