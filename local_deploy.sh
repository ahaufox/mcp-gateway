#!/bin/bash

# ==============================================================================
# MCP Gateway 本地部署脚本
# 适用系统: 本地 Linux/macOS
# 使用方法: ./local_deploy.sh
# ==============================================================================

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}开始本地部署 MCP Gateway...${NC}"

# 0. 检查并补全本地环境映射文件 (Docker Compose 强依赖)
echo -e "${YELLOW}[0/3] 检查环境变量文件...${NC}"
ENV_FILES=(
    "mcp-proxy/.env"
    "jules-mcp-server/.env"
)
for f in "${ENV_FILES[@]}"; do
    if [ ! -f "$f" ]; then
        echo -e "${YELLOW}警告: $f 不存在，正在创建空文件以避免 Docker 报错...${NC}"
        mkdir -p "$(dirname "$f")"
        touch "$f"
    fi
done

# 0.1 解决端口 9090 冲突 (自动停止占用端口的容器)
echo -e "${YELLOW}[0.1/3] 检查端口 9090 占用情况...${NC}"
CONF_CONTAINER=$(docker ps --format "{{.Names}}" -a --filter "publish=9090")
if [ ! -z "$CONF_CONTAINER" ]; then
    # 检查冲突容器是否属于当前 docker-compose 项目
    # docker-compose ps -q 返回当前项目管理的容器 ID
    IS_OWN=$(docker compose ps -q | grep -w "$(docker inspect -f '{{.Id}}' "$CONF_CONTAINER")" || true)
    
    if [ -z "$IS_OWN" ]; then
        echo -e "${YELLOW}警告: 发现外部容器 '$CONF_CONTAINER' 占用 9090 端口，正在强制停止...${NC}"
        docker rm -f "$CONF_CONTAINER"
    else
        echo -e "${BLUE}端口 9090 由本项目容器占有，将通过 docker compose down 处理。${NC}"
    fi
fi

# 1. 本地构建
echo -e "${YELLOW}[1/3] 构建 Docker 镜像...${NC}"
docker compose build

# 2. 停止旧容器
echo -e "${YELLOW}[2/3] 停止并移除现有容器...${NC}"
docker compose down --remove-orphans

# 3. 启动新容器
echo -e "${YELLOW}[3/3] 启动 Docker 容器...${NC}"
docker compose up -d

# 完成
echo -e "${GREEN}[DONE] 本地部署完成！${NC}"
echo -e "${BLUE}管理后台地址: http://localhost:9090${NC}"
echo -e "${YELLOW}服务状态:${NC}"
docker compose ps
