# contractfault — build, test and demo orchestration.
#
# The Makefile drives both halves of the project: the Go analyzer CLI and the
# TypeScript seismic viewer. Targets are deliberately small and composable so
# CI and humans run the same commands.

GO        ?= go
NPM       ?= npm
BIN       ?= bin/contractfault
VIEWER    ?= viewer
EXAMPLES  ?= examples
OLD       ?= $(EXAMPLES)/contracts/orders-v1.json
NEW       ?= $(EXAMPLES)/contracts/orders-v2.json
CONSUMERS ?= $(EXAMPLES)/consumers/*.json
REPORT    ?= $(EXAMPLES)/report.json

.PHONY: all build build-go build-viewer test test-go test-viewer \
        report demo svg fmt vet clean tidy ci

all: build test

## build: compile the Go CLI and the TypeScript viewer.
build: build-go build-viewer

build-go:
	$(GO) build -o $(BIN) ./cmd/contractfault

build-viewer:
	cd $(VIEWER) && $(NPM) install --no-audit --no-fund && $(NPM) run build

## test: run all Go and TypeScript tests.
test: test-go test-viewer

test-go:
	$(GO) test ./...

test-viewer:
	cd $(VIEWER) && $(NPM) test

## report: analyze the example contracts and emit deterministic JSON.
report: build-go
	$(BIN) -old $(OLD) -new $(NEW) -consumers "$(CONSUMERS)" -format json -out $(REPORT)

## demo: full pipeline — analyze then render the seismic map + SVG.
