# syntax=docker/dockerfile:1

# ---------------------------------------------------------
# Build Stage
# ---------------------------------------------------------
FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# ---------------------------------------------------------
# Production Runner Stage
# ---------------------------------------------------------
FROM alpine:3.20 AS runner

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrations /app/migrations

USER appuser

ENV APP_ENV=production
ENV APP_PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/server"]
