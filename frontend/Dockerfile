# syntax=docker/dockerfile:1

# ---------------------------------------------------------
# Build Stage
# ---------------------------------------------------------
FROM node:20-alpine AS builder

WORKDIR /app

# Build arguments for Vite client-side environment
ARG VITE_API_URL
ARG VITE_WS_URL
ARG VITE_GDRIVE_UPLOAD_URL

ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_WS_URL=${VITE_WS_URL}
ENV VITE_GDRIVE_UPLOAD_URL=${VITE_GDRIVE_UPLOAD_URL}

COPY package*.json ./
RUN npm ci

COPY . .

RUN npm run build \
    && npm prune --omit=dev

# ---------------------------------------------------------
# Production Runner Stage
# ---------------------------------------------------------
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

COPY --from=builder --chown=node:node /app/package*.json ./
COPY --from=builder --chown=node:node /app/node_modules ./node_modules
COPY --from=builder --chown=node:node /app/build ./build

USER node

EXPOSE 3000

CMD ["node", "build"]
