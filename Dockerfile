# === Stage 1: Build ===
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git bash

WORKDIR /src

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Устанавливаем goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Копируем исходники
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/app ./cmd/url-shortener

# === Stage 2: Runtime ===
FROM alpine:3.22

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /app/bin/app /app/bin/app

# Копируем миграции (папка должна существовать!)
COPY db/migrations /app/db/migrations

# Копируем goose и скрипт запуска
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY bin/run.sh /app/bin/run.sh

RUN chmod +x /app/bin/run.sh

EXPOSE 8080

CMD ["/app/bin/run.sh"]