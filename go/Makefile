GO := $(shell which go)

.PHONY: all
all: lint test

.PHONY: clean
clean:
	@echo "Cleaning files in ./bin..."
	@[ -d ./bin ] && rm -rf ./bin/* || echo "No bin directory to clean."
	@echo "Cleaning testing artifacts..."
	@[ -f coverage.txt ] && rm -f coverage.txt || echo "No coverage.txt to remove"

.PHONY: fmt
fmt:
	@echo "Formatting code with gofumpt..."
	@gofumpt -w ./ # format everything, not just pkg/cmd

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	@CGO_LDFLAGS="" $(GO) test -coverprofile=coverage.txt -v ./...