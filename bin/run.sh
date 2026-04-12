#!/bin/sh
set -e

echo "[run.sh] Starting service"

# Миграции
if [ -d "/app/db/migrations" ] && [ -n "${DATABASE_URL:-}" ]; then
    echo "[run.sh] Running DB migrations"
    goose -dir /app/db/migrations postgres "${DATABASE_URL}" up || true
fi

echo "[run.sh] Starting Go app on port 8080"
exec /app/bin/app