# orchestratr Makefile
#
# Targets:
#   build     — compile with version injection from git
#   install   — go install with version injection
#   test      — run all tests
#   lint      — run golangci-lint
#   fmt       — format all Go files
#   clean     — remove build artifacts

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
LDFLAGS  = -X main.Version=$(VERSION)
BINARY   = orchestratr

.PHONY: build build-windows install test lint fmt clean

build:
	CGO_ENABLED=1 go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/orchestratr

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY).exe ./cmd/orchestratr

install:
	CGO_ENABLED=1 go install -ldflags "$(LDFLAGS)" ./cmd/orchestratr

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY) $(BINARY).exe
