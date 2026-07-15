# Stage 1: Build frontend
FROM node:22.14.0-bookworm-slim AS node-builder
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN mkdir -p /mcp-proxy/internal/server/frontend/dist && npm run build

# Stage 2: Build Go binary
FROM golang:1.24.1-bookworm AS builder
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

# 配置 npm 和 uv 使用国内镜像源，以解决远程服务器拉取超时的问题
ENV UV_INDEX_URL="https://pypi.tuna.tsinghua.edu.cn/simple"
RUN npm config set registry https://registry.npmmirror.com

# 预全局安装常用的 MCP npm 依赖，以秒级启动 npx
RUN npm install -g @modelcontextprotocol/server-github \
                   @antv/mcp-server-chart \
                   @notionhq/notion-mcp-server \
                   @playwright/mcp@latest \
                   @modelcontextprotocol/server-sequential-thinking \
                   @upstash/context7-mcp \
                   reactbits-dev-mcp-server

# 预安装 python 相关的 mcp 服务的工具，并将命令路径加入环境变量
RUN uv tool install mcp-server-fetch
ENV PATH="/root/.local/bin:${PATH}"

COPY --from=builder /app/build/mcp-proxy /main
COPY --from=builder /app/docs /docs
COPY --from=builder /app/platform-config.json /config/platform-config.json
WORKDIR /config
ENTRYPOINT ["/main"]
# 默认从拆分后的 configs/ 目录加载；可通过 --config 参数回退到单文件模式。
CMD ["--config-dir", "/config/configs"]