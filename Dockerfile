# syntax=docker/dockerfile:1.4

# === Build stage ===
FROM golang:1.25-alpine AS backend-builde

RUN apk add --no-cache git bash

WORKDIR /build/code

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Устанавливаем goose для миграций
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Копируем исходники
COPY . .

# Собираем статический бинарник
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/app ./cmd/url-shortener

# === Runtime stage ===
FROM alpine:3.22

# Устанавливаем ca-certificates для HTTPS-запросов (в т.ч. в Sentry)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник
COPY --from=backend-builder /build/app /app/bin/app

# Копируем миграции (если есть)
COPY --from=backend-builder /build/code/db/migrations /app/db/migrations 2>/dev/null || true

# Копируем goose
COPY --from=backend-builder /go/bin/goose /usr/local/bin/goose

# Копируем и делаем исполняемым run.sh
COPY bin/run.sh /app/bin/run.sh
RUN chmod +x /app/bin/run.sh

EXPOSE 8080

# Запускаем через run.sh
CMD ["/app/bin/run.sh"]