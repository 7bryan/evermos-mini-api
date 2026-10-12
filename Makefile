.PHONY: run build tidy fmt vet test test-integration

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
	go test -tags integration -count=1 ./internal/infrastructure/database/...
