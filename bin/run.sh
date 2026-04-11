#!/bin/sh
set -eu

echo "[run.sh] Starting service"

# Запускаем миграции БД (если есть)
# || true — игнорируем ошибку, если миграций нет
if [ -d "/app/db/migrations" ] && [ -n "${DATABASE_URL:-}" ]; then
    echo "[run.sh] Running DB migrations"
    goose -dir /app/db/migrations postgres "${DATABASE_URL}" up || true
else
    echo "[run.sh] No migrations to run"
fi

echo "[run.sh] Starting Go app"
exec /app/bin/app