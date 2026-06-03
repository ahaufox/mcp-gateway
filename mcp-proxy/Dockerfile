# Stage 1: Build frontend
FROM node:lts-bookworm-slim AS node-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24.0-bookworm AS builder
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY mcp-proxy/go.mod mcp-proxy/go.sum ./
RUN go mod download
COPY mcp-proxy/ ./
# Copy pre-built frontend assets for Go embed
COPY --from=node-builder /mcp-proxy/internal/server/frontend/dist /app/internal/server/frontend/dist
ARG BUILD_VERSION=docker
RUN make build BUILD=${BUILD_VERSION}

# Stage 3: Copy node binaries for runtime
FROM node:lts-bookworm-slim AS node

# Stage 4: Final runtime image
FROM ghcr.io/astral-sh/uv:python3.13-bookworm-slim

COPY --from=node /usr/local/bin/node /usr/local/bin/node
COPY --from=node /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s /usr/local/lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm && \
    ln -s /usr/local/lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx && \
    ln -s /usr/local/bin/node /usr/local/bin/nodejs

RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/build/mcp-proxy /main
COPY --from=builder /app/docs /docs
ENTRYPOINT ["/main"]
CMD ["--config", "/config/config.json"]