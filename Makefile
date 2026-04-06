BINARY_NAME=shortener
CMD_PATH=./cmd/url-shortener
BIN_PATH=./bin/$(BINARY_NAME)

.PHONY: run build clean test deps

run:
	go run $(CMD_PATH)

build:
	go build -o $(BIN_PATH) $(CMD_PATH)

clean:
	rm -f $(BIN_PATH)

test:
	@echo "Testing /ping endpoint..."
	curl -s http://localhost:8080/ping && echo "" || echo "❌ Server not running"

deps:
	go mod tidy
	go get -u github.com/gin-gonic/gin

# Для разработки с авто-релоадом (требуется air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air --build.cmd "go build -o $(BIN_PATH) $(CMD_PATH)" --build.bin "./$(BIN_PATH)"
