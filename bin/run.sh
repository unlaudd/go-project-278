#!/usr/bin/env bash
set -euo pipefail

echo "[run.sh] Starting service"

# Запускаем миграции БД (если есть)
# Если миграций пока нет — эта строка просто вернёт успех
if [ -d "/app/db/migrations" ] && [ -n "${DATABASE_URL:-}" ]; then
    echo "[run.sh] Running DB migrations"
    goose -dir /app/db/migrations postgres "${DATABASE_URL}" up
else
    echo "[run.sh] No migrations to run"
fi

echo "[run.sh] Starting Go app"
# exec заменяет процесс bash на приложение — это важно для корректной обработки сигналов
exec /app/bin/app
