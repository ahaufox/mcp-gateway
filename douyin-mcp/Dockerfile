# Stage 1: Build
FROM python:3.12-slim-bookworm AS builder

# 安装 uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /bin/uv

# 使用国内镜像源加速
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources
ENV UV_INDEX_URL=https://mirrors.aliyun.com/pypi/simple/

WORKDIR /app

# 复制依赖文件
COPY pyproject.toml uv.lock ./

# 安装依赖到 .venv
RUN uv sync --frozen --no-cache --no-install-project

# Stage 2: Runtime
FROM python:3.12-slim-bookworm

WORKDIR /app

# 安装运行时的系统依赖 (ffmpeg 和 curl 用于健康检查)
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# 从构建阶段复制 .venv
COPY --from=builder /app/.venv /app/.venv

# 复制项目源代码
COPY . .

# 设置环境变量
ENV PATH="/app/.venv/bin:$PATH"
ENV MCP_TRANSPORT=sse

# 暴露端口
EXPOSE 8000

# 默认启动命令
ENTRYPOINT ["python", "server.py"]
