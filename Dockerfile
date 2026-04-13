# Multi-stage Dockerfile for the URL shortener application.
# Stages: frontend-builder → backend-builder → runtime (alpine + caddy).

# =============================================================================
# Stage 1: Build frontend static assets
# =============================================================================
FROM node:24-alpine AS frontend-builder
WORKDIR /build/frontend

# Copy package manifests first to leverage Docker layer caching.
COPY package*.json ./

# Install production dependencies with npm cache mount for faster rebuilds.
# --prefer-offline: use cached packages when available
# --no-audit: skip security audit for faster CI/CD
RUN --mount=type=cache,target=/root/.npm \
  npm ci --prefer-offline --no-audit

# =============================================================================
# Stage 2: Build Go backend binary
# =============================================================================
FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache git
WORKDIR /build/code

# Copy Go module files first for dependency caching.
COPY go.mod go.sum ./

# Download module dependencies with cache mount.
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download

# Install goose migration tool for runtime database migrations.
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy application source code.
COPY . .

# Build a statically linked, production-ready binary for Linux amd64.
# CGO_ENABLED=0 ensures no dynamic C dependencies; cache mount speeds up compilation.
RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/app .

# =============================================================================
# Stage 3: Runtime image (minimal Alpine with Caddy reverse proxy)
# =============================================================================
FROM alpine:3.22

# Install runtime dependencies:
# - ca-certificates: for HTTPS requests to external services
# - tzdata: for correct time zone handling in logs
# - bash: for run.sh script execution
# - caddy: reverse proxy to serve static files and proxy API requests
RUN apk add --no-cache ca-certificates tzdata bash caddy

WORKDIR /app

# Copy the compiled Go backend binary.
COPY --from=backend-builder /build/app /app/bin/app

# Copy pre-built frontend static assets from the npm package.
# These are served by Caddy for the web UI.
COPY --from=frontend-builder \
  /build/frontend/node_modules/@hexlet/project-url-shortener-frontend/dist \
  /app/public

# Copy database migration files for goose to apply at startup.
COPY --from=backend-builder /build/code/db/migrations /app/db/migrations

# Copy the goose binary installed in the backend-builder stage.
COPY --from=backend-builder /go/bin/goose /usr/local/bin/goose

# Copy and make executable the entrypoint script that orchestrates startup.
COPY bin/run.sh /app/bin/run.sh
RUN chmod +x /app/bin/run.sh

# Copy Caddy configuration for reverse proxy and static file serving.
COPY Caddyfile /etc/caddy/Caddyfile

# Expose port 80 for HTTP traffic (Caddy listens here).
EXPOSE 80

# Entrypoint: run.sh handles migrations, starts backend, then Caddy.
CMD ["/app/bin/run.sh"]
