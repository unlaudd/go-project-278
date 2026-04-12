# === Stage 1: Build Go backend ===
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache git bash
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/app ./cmd/url-shortener

# === Stage 2: Prepare Frontend ===
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY package*.json ./
RUN npm ci --omit=dev
RUN cp -r /frontend/node_modules/@hexlet/project-url-shortener-frontend/dist /frontend/dist

# === Stage 3: Runtime with Caddy ===
FROM caddy:2.8-alpine

# Устанавливаем postgres-client для goose и ca-certificates
RUN apk --no-cache add postgresql-client ca-certificates wget

# Копируем конфиг Caddy
COPY Caddyfile /etc/caddy/Caddyfile

# Копируем фронтенд
COPY --from=frontend-builder /frontend/dist /app/frontend/dist

# Копируем бэкенд
COPY --from=backend-builder /app/bin/app /app/bin/app

# Копируем миграции
COPY --from=backend-builder /src/db/migrations /app/db/migrations

# Скачиваем готовый бинарник goose
RUN wget -q -O /usr/local/bin/goose \
    https://github.com/pressly/goose/releases/download/v3.20.0/goose_linux_x86_64 && \
    chmod +x /usr/local/bin/goose

# Скрипт запуска
COPY bin/run.sh /app/bin/run.sh
RUN chmod +x /app/bin/run.sh

# 🔹 EXPOSE принимает только литерал
EXPOSE 8080

# Запуск через run.sh
CMD ["/app/bin/run.sh"]
