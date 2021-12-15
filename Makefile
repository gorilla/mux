GOBIN=$(shell pwd)/bin

install-lint:
	@curl -sSfL -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.42.0

lint: install-lint
	@echo "golangci-lint run ../..."
	@$(GOBIN)/golangci-lint run ./...
