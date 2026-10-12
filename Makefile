.PHONY: run build tidy fmt vet test

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
