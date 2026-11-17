# semscope development gates. Plain GNU Make syntax, so GNU Make and zmake can run this file.
# The ZMakefile holds the same tasks in modern mode. Run `zmake -f Makefile check` for this file.

GO ?= go
PYTHON ?= python3
GOLANGCI_LINT ?= $(shell $(GO) env GOPATH)/bin/golangci-lint
GOFMT ?= $(shell $(GO) env GOROOT)/bin/gofmt
GO_DIRS := cmd internal rules
PARITY_CORPUS ?= $(shell cd ../zmake 2>/dev/null && pwd)
VERSION_PKG := github.com/infrashift/semscope/internal/version
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X $(VERSION_PKG).Version=$(VERSION) -X $(VERSION_PKG).Commit=$(COMMIT) -X $(VERSION_PKG).BuildDate=$(BUILD_DATE)

.PHONY: all build test lint fmt docs-style wiki selfscan sarif parity check

all: check

## build: compile the semscope binary into bin/
build:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o bin/semscope ./cmd/semscope

## test: run the unit tests with the race detector
test:
	$(GO) test -race ./...

## lint: check the formatting, the Power of 10 rules and the static analyzers
lint:
	@out=$$($(GOFMT) -l $(GO_DIRS)); test -z "$$out" || { echo "gofmt: $$out"; exit 1; }
	$(PYTHON) -I scripts/nasa-lint.py
	$(PYTHON) -I scripts/nasa-lint-go.py
	$(PYTHON) -I scripts/tests/nasa_lint_go_test.py
	$(GO) vet ./...
	$(GOLANGCI_LINT) run

## fmt: format the Go sources in place
fmt:
	$(GOFMT) -w $(GO_DIRS)

## docs-style: check the documentation against ASD-STE100 rules
docs-style:
	$(PYTHON) -I scripts/ste-check.py --strict

## wiki: validate the OKF wiki
wiki:
	okfw validate --all
	okfw validate --okf

## selfscan: scan the semscope sources with semscope and the overlay .semscope.cue
selfscan:
	$(GO) run ./cmd/semscope scan -r nasa-p10 -r .semscope.cue --fail-level warning .

## sarif: write the SARIF 2.1.0 and 2.2 reports of the semscope repository
sarif: build
	./bin/semscope scan -r nasa-p10 -r .semscope.cue --format sarif --sarif-version 2.1.0 -o semscope-self.sarif .
	./bin/semscope scan -r nasa-p10 -r .semscope.cue --format sarif --sarif-version 2.2 -o semscope-self-v22.sarif .

## parity: compare semscope with nasa-lint.py on the src/ tree of PARITY_CORPUS (ADR-0005)
parity:
	SEMSCOPE_PARITY_CORPUS=$(PARITY_CORPUS) $(GO) test -count=1 -run NasaParityCorpus -v ./internal/engine/

## check: run every gate
check: lint test wiki docs-style selfscan
