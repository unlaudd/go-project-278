#!/usr/bin/env bash
# Exit immediately on error, undefined variable, or pipe failure.
set -euo pipefail

echo "[run.sh] Starting service"

# Apply pending database migrations.
echo "[run.sh] Running DB migrations"
goose -dir ./db/migrations postgres "${DATABASE_URL}" up

# Start Caddy as a reverse proxy in the background.
echo "[run.sh] Starting Caddy"
caddy run --config /etc/caddy/Caddyfile &

# Replace the shell process with the Go application.
# Ensures the app runs as PID 1 and receives OS signals directly.
echo "[run.sh] Starting Go app"
exec /app/bin/app