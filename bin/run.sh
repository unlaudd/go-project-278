#!/bin/sh
set -eu

echo "[run.sh] Starting service"

# Миграции
if [ -d "/app/db/migrations" ] && [ -n "${DATABASE_URL:-}" ]; then
    echo "[run.sh] Running DB migrations"
    goose -dir /app/db/migrations postgres "${DATABASE_URL}" up || true
fi

echo "[run.sh] Starting backend on port 9000"
# Запускаем бэкенд в фоне
/app/bin/app &
BACKEND_PID=$!

echo "[run.sh] Starting Caddy on port 8080"
# Запускаем Caddy (он будет проксировать на :9000)
exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
