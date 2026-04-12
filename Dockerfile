# syntax=docker/dockerfile:1.4

# === Stage 1: Build Go backend ===
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache git bash
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/app ./cmd/url-shortener

# === Stage 2: Build Frontend ===
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
# Копируем только package-файлы для кэширования зависимостей
COPY package*.json ./
RUN npm ci
# Копируем остальное и билдим
COPY . .
RUN npx start-hexlet-url-shortener-frontend --build --outDir /frontend/dist

# === Stage 3: Runtime with Caddy ===
FROM caddy:2.8-alpine

# Устанавливаем postgres-client для goose и ca-certificates
RUN apk --no-cache add postgresql-client ca-certificates

# Копируем конфиг Caddy
COPY Caddyfile /etc/caddy/Caddyfile

# Копируем фронтенд
COPY --from=frontend-builder /frontend/dist /app/frontend/dist

# Копируем бэкенд
COPY --from=backend-builder /app/bin/app /app/bin/app

# Копируем миграции
COPY --from=backend-builder /src/db/migrations /app/db/migrations

# 🔹 Устанавливаем goose прямо в финальном образе (проще и надёжнее)
RUN apk add --no-cache git && \
    go install github.com/pressly/goose/v3/cmd/goose@latest && \
    mv /go/bin/goose /usr/local/bin/goose && \
    apk del git

# Скрипт запуска
COPY bin/run.sh /app/bin/run.sh
RUN chmod +x /app/bin/run.sh

EXPOSE 8080
CMD ["/app/bin/run.sh"]
