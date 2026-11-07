BINARY := clean-my-cache
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build run test lint fmt check snapshot

build:
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/$(BINARY) ./cmd/$(BINARY)

run:
	go run ./cmd/$(BINARY) $(ARGS)

test:
	go test -race ./...

lint:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && echo "run: make fmt" && exit 1)
	go vet ./...

fmt:
	gofmt -w .

check: lint test

snapshot:
	goreleaser release --snapshot --clean --skip=publish
