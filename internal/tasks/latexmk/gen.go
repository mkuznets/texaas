//go:generate protoc -I. --go_out=paths=source_relative:. latexmk.proto
//go:generate gofmt -w .
//go:generate goimports -w .

package latexmk
