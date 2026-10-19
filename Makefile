.PHONY: build test lint fmt docker clean

BINARY := prometheus-ovh-exporter
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/prometheus-ovh-exporter/

test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./internal/...
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}' | \
		awk -F'%' '{if ($$1 < 100) { print "FAIL: coverage is " $$1 "%, need 100%"; exit 1 } else { print "OK: coverage is " $$1 "%" }}'

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

docker:
	docker build -t $(BINARY):$(VERSION) .

clean:
	rm -rf bin/ coverage.out
