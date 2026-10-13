.PHONY: run build tidy fmt vet test test-integration test-external

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

test-integration:
	go test -p 1 -tags integration -count=1 ./internal/infrastructure/database/... ./internal/repository/...

test-external:
	go test -tags external -count=1 -v ./internal/infrastructure/emsifa/...
