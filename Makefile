# Makefile for the URL shortener project.
# Provides common development, testing, and build targets.
# Run `make help` to list all available commands.

BINARY_NAME=shortener
CMD_PATH=./cmd/url-shortener
BIN_PATH=./bin/$(BINARY_NAME)

# Mark all targets as phony to avoid conflicts with files of the same name.
.PHONY: run build clean test test-race lint lint-fix fmt deps dev help

# help: Display this help message (default target).
# Parses comments marked with '##' to show available targets and descriptions.
help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# run: Execute the application directly with `go run` for quick local testing.
run:
	go run $(CMD_PATH)

# build: Compile the application to a binary at $(BIN_PATH).
build:
	go build -o $(BIN_PATH) $(CMD_PATH)

# clean: Remove the compiled binary to ensure a fresh build.
clean:
	rm -f $(BIN_PATH)

# test: Run all unit tests with verbose output.
test:
	go test -v ./...

# test-race: Run tests with Go's race detector enabled.
test-race:
	go test -race -v ./...

# lint: Run golangci-lint to enforce code quality and style guidelines.
lint:
	golangci-lint run

# lint-fix: Run golangci-lint with --fix to auto-correct fixable issues.
lint-fix:
	golangci-lint run --fix

# fmt: Format all Go source files using goimports.
fmt:
	goimports -w .

# deps: Tidy go.mod and fetch key dependencies.
deps:
	go mod tidy
	go get -u github.com/gin-gonic/gin
	go get github.com/stretchr/testify/assert

# dev: Start the application with hot-reload using air.
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air --build.cmd "go build -o $(BIN_PATH) $(CMD_PATH)" --build.bin "./$(BIN_PATH)"
