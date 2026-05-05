.PHONY: test run fmt

GOCACHE ?= /tmp/rag-service-go-build

test:
	GOCACHE=$(GOCACHE) go test ./...

run:
	GOCACHE=$(GOCACHE) go run ./cmd/query-api

fmt:
	gofmt -w ./cmd ./internal
