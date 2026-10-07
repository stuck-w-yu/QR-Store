# syntax=docker/dockerfile:1

# =========================================================
# Stage 1: Build Backend (Golang)
# =========================================================
FROM golang:1.27-alpine AS backend-builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# =========================================================
# Stage 2: Build Frontend (SvelteKit / Node.js)
# =========================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Client-side environment variables available during build
ARG VITE_API_URL
ARG VITE_WS_URL
ARG VITE_GDRIVE_UPLOAD_URL

ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_WS_URL=${VITE_WS_URL}
ENV VITE_GDRIVE_UPLOAD_URL=${VITE_GDRIVE_UPLOAD_URL}

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build \
    && npm prune --omit=dev

# =========================================================
# Stage 3: Production Runner (All-in-One: Node, Go, Caddy)
# =========================================================
FROM node:20-alpine AS runner

WORKDIR /app

# Install ca-certificates and tzdata for secure HTTPS and timezone accuracy
RUN apk add --no-cache ca-certificates tzdata

# Copy Caddy binary from official Caddy image
COPY --from=caddy:2-alpine /usr/bin/caddy /usr/bin/caddy

# Set up Backend binaries and migrations
WORKDIR /app/backend
COPY --from=backend-builder /app/server ./server
COPY --from=backend-builder /app/migrations ./migrations

# Set up Frontend application
WORKDIR /app/frontend
COPY --from=frontend-builder /app/package*.json ./
COPY --from=frontend-builder /app/node_modules ./node_modules
COPY --from=frontend-builder /app/build ./build

# Set up Caddy configuration and container supervisor script
WORKDIR /app
COPY Caddyfile /etc/caddy/Caddyfile
COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh

# Default environment configurations
ENV NODE_ENV=production
ENV APP_ENV=production
ENV PORT=80

# Expose Caddy public HTTP port
EXPOSE 80

ENTRYPOINT ["/app/start.sh"]
