#!/bin/sh
set -e

echo "[run.sh] Starting service"

# Миграции
if [ -d "/app/db/migrations" ] && [ -n "${DATABASE_URL:-}" ]; then
    echo "[run.sh] Running DB migrations"
    goose -dir /app/db/migrations postgres "${DATABASE_URL}" up || true
fi

echo "[run.sh] Starting backend on port ${BACKEND_PORT:-9000}"
# Запускаем бэкенд в фоне
/app/bin/app &
BACKEND_PID=$!

# Ждём, пока бэкенд станет доступен (простая проверка)
echo "[run.sh] Waiting for backend to be ready..."
for i in $(seq 1 30); do
    if wget -q -O /dev/null "http://localhost:${BACKEND_PORT:-9000}/ping" 2>/dev/null; then
        echo "[run.sh] Backend is ready"
        break
    fi
    sleep 1
done

echo "[run.sh] Starting Caddy on port ${PORT:-8080}"
# 🔹 Запускаем Caddy через его стандартный механизм (без exec, без явного пути)
# Образ caddy:alpine уже имеет правильный ENTRYPOINT
caddy