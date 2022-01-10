GOBIN=$(shell pwd)/bin

install-lint:
	@if [ ! -f "$(GOBIN)/golangci-lint" ]; then \
		curl -sSfL -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.42.0; \
	fi

lint: install-lint
	@echo "golangci-lint run ./..."
	@$(GOBIN)/golangci-lint run ./...

test:
	go test -race -failfast ./...
