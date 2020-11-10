package pb

//go:generate protoc -I../ --go_out=paths=source_relative:. --go-grpc_out=. --go-grpc_opt=paths=source_relative service.proto
//go:generate gofmt -w .
//go:generate goimports -w .
