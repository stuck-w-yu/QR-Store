#!/bin/sh
set -e

# Port assignments
export CADDY_PORT="${PORT:-80}"
export APP_PORT="${BACKEND_PORT:-8080}"
export HOST="${FRONTEND_HOST_BIND:-127.0.0.1}"
export FRONTEND_PORT="${INTERNAL_FRONTEND_PORT:-3000}"

# SSR Backend URL for SvelteKit Node adapter
export BACKEND_URL="http://127.0.0.1:${APP_PORT}"

echo "============================================="
echo " Starting QR-Store All-in-One Container"
echo " Public Port (Caddy) : ${CADDY_PORT}"
echo " Backend Port (Go)   : ${APP_PORT}"
echo " Frontend Port (Node): ${FRONTEND_PORT}"
echo "============================================="

# Graceful termination handler
cleanup() {
    echo "Stopping all processes..."
    kill -TERM "$BACKEND_PID" 2>/dev/null || true
    kill -TERM "$FRONTEND_PID" 2>/dev/null || true
    kill -TERM "$CADDY_PID" 2>/dev/null || true
    wait
    exit 0
}

trap cleanup INT TERM

# 1. Start Go Backend
echo "Starting Backend service..."
cd /app/backend
PORT="$APP_PORT" ./server &
BACKEND_PID=$!

# 2. Start SvelteKit Frontend
echo "Starting Frontend service..."
cd /app/frontend
PORT="$FRONTEND_PORT" HOST="$HOST" node build &
FRONTEND_PID=$!

# 3. Start Caddy Reverse Proxy
echo "Starting Caddy reverse proxy on port ${CADDY_PORT}..."
caddy run --config /etc/caddy/Caddyfile --adapter caddyfile &
CADDY_PID=$!

# Health monitor loop: terminate container if any service crashes
while kill -0 "$BACKEND_PID" 2>/dev/null && \
      kill -0 "$FRONTEND_PID" 2>/dev/null && \
      kill -0 "$CADDY_PID" 2>/dev/null; do
    sleep 2
done

echo "A service process stopped unexpectedly. Initiating shutdown..."
cleanup
