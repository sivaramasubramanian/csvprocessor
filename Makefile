GOLANGCI_LINT_VERSION := "v1.48.0" # Optional configuration to pinpoint golangci-lint version.

# The head of Makefile determines location of dev-go to include standard targets.
GO ?= go
export GO111MODULE = on

ifneq "$(GOFLAGS)" ""
  $(info GOFLAGS: ${GOFLAGS})
endif


PWD = $(shell pwd)

# Detecting GOPATH and removing trailing "/" if any
GOPATH = $(realpath $(shell $(GO) env GOPATH))

install:  .git/hooks/pre-commit lint test bench

.git/hooks/pre-commit:
	@echo "setting githooks path"
	@git config core.hooksPath .githooks/
	@echo "checking githooks path"
	@git config --get core.hooksPath
	@echo "installing errcheck"
	@$(GO) install github.com/kisielk/errcheck@latest
	@echo "installing golangci-lint"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

## Run linter
lint:
	@golangci-lint run

## Run tests
test: lint test-unit

## Run unit tests
test-unit:
	@echo "Running unit tests."
	@CGO_ENABLED=1 $(GO) test -short -coverprofile=unit.coverprofile -covermode=atomic -race ./...

BENCH_COUNT ?= 5
MASTER_BRANCH ?= main
REF_NAME ?= $(shell git symbolic-ref HEAD --short | tr / - 2>/dev/null)
SHELL := /bin/bash

## Run benchmark and show result stats, iterations count controlled by BENCH_COUNT, default 5.
bench: bench-run bench-stat-diff bench-stat

bench-stat-cli:
	@test -s $(GOPATH)/bin/benchstat || GOFLAGS= GOBIN=$(GOPATH)/bin $(GO) install golang.org/x/perf/cmd/benchstat@latest

## Run benchmark, iterations count controlled by BENCH_COUNT, default 5.
bench-run:
	@echo $(BENCH_COUNT) $(REF_NAME)
	@set -o pipefail && $(GO) test -bench=. -count=$(BENCH_COUNT) -run=^a  ./... | tee bench-$(REF_NAME).txt

## Show benchmark comparison with base branch.
bench-stat-diff: bench-stat-cli
	@test ! -e bench-$(MASTER_BRANCH).txt || benchstat bench-$(MASTER_BRANCH).txt bench-$(REF_NAME).txt

## Show result of benchmark.
bench-stat: bench-stat-cli
	@$(GOPATH)/bin/benchstat bench-$(REF_NAME).txt
