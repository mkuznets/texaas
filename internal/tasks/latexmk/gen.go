//go:generate protoc -I. --go_out=paths=source_relative:. --go-grpc_out=. --go-grpc_opt=paths=source_relative latexmk.proto
//go:generate gofmt -w .
//go:generate goimports -w .

package latexmk
