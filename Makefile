# apitpl package Makefile

SHELL       = /bin/bash
CFG         = .env
GO         ?= go
SOURCES    ?= *.go */*.go */*/*.go

# ------------------------------------------------------------------------------

.PHONY: help

##
## Available make targets
##

# default: show target list
all: help

# ------------------------------------------------------------------------------
## Sources

## Run linters
lint:
	golint ./...
	golangci-lint run ./... ./ginapitpl/...

## Run tests and fill coverage.out
cov: coverage.out

# internal target
coverage.out: $(SOURCES)
	$(GO) test -race -coverprofile=$@ -covermode=atomic ./...

## Open coverage report in browser
cov-html: cov
	$(GO) tool cover -html=coverage.out

## Clean coverage report
cov-clean:
	rm -f coverage.*

# ------------------------------------------------------------------------------
## Misc

## Count lines of code (including tests) and update LOC.md
cloc: LOC.md

LOC.md: $(SOURCES)
	cloc --by-file --not-match-f='(_mock_test.go|.sql|ml|.md|file|resource.go)$$' --md . > $@

## List Makefile targets
help:  Makefile
	@grep -A1 "^##" $< | grep -vE '^--$$' | sed -E '/^##/{N;s/^## (.+)\n(.+):(.*)/\t\2:\1/}' | column -t -s ':'
