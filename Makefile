BINARY_NAME=shortener
CMD_PATH=./cmd/url-shortener
BIN_PATH=./bin/$(BINARY_NAME)

.PHONY: run build clean test lint test-race deps dev

run:
	go run $(CMD_PATH)

build:
	go build -o $(BIN_PATH) $(CMD_PATH)

clean:
	rm -f $(BIN_PATH)

test:
	go test -v ./...

test-race:
	go test -race -v ./...

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

deps:
	go mod tidy
	go get -u github.com/gin-gonic/gin
	go get github.com/stretchr/testify/assert

dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air --build.cmd "go build -o $(BIN_PATH) $(CMD_PATH)" --build.bin "./$(BIN_PATH)"
