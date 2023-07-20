SHELL := /bin/bash

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# LINT is the path to the golangci-lint binary
LINT = $(shell which golangci-lint)

.PHONY: golangci-lint
golangci-lint:
ifeq (, $(LINT))
  ifeq (, $(shell which golangci-lint))
	@{ \
	set -e ;\
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest ;\
	}
  override LINT=$(GOBIN)/golangci-lint
  else
  override LINT=$(shell which golangci-lint)
  endif
endif

.PHONY: verify
verify: golangci-lint
	$(LINT) run

.PHONY: test
test:
	go test -race --coverprofile=coverage.coverprofile --covermode=atomic -v ./...
