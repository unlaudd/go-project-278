# === Stage 1: Build Go backend ===
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache git bash
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/app ./cmd/url-shortener

# === Stage 2: Runtime (Alpine + Go app) ===
FROM alpine:3.22
RUN apk --no-cache add ca-certificates postgresql-client

# Копируем бэкенд и миграции
COPY --from=backend-builder /app/bin/app /app/bin/app
COPY --from=backend-builder /src/db/migrations /app/db/migrations

# Скачиваем goose (готовый бинарник)
RUN wget -q -O /usr/local/bin/goose \
    https://github.com/pressly/goose/releases/download/v3.20.0/goose_linux_x86_64 && \
    chmod +x /usr/local/bin/goose

# Скрипт запуска
COPY bin/run.sh /app/bin/run.sh
RUN chmod +x /app/bin/run.sh

EXPOSE 8080
CMD ["/app/bin/run.sh"]