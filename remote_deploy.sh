#!/bin/bash

# ==============================================================================
# MCP Gateway 远程部署脚本 (本地构建模式)
# 适用系统: 本地 macOS/Linux 构建, 远程 Ubuntu 运行
# 使用方法: ./remote_deploy.sh [user@host]
# ==============================================================================

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 参数检查
REMOTE_TARGET=$1
if [ -z "$REMOTE_TARGET" ]; then
    echo -e "${RED}错误: 未指定远程目标。${NC}"
    echo -e "用法: $0 [user@host]"
    exit 1
fi

# SSH 连接复用配置
SOCKET_PATH="/tmp/mcp_deploy.sock"
SSH_OPTS="-o ControlMaster=auto -o ControlPath=$SOCKET_PATH -o ControlPersist=600"

# 清理函数
cleanup() {
    if [ -S "$SOCKET_PATH" ]; then
        echo -e "\n${BLUE}正在关闭 SSH 控制连接...${NC}"
        ssh -S "$SOCKET_PATH" -O exit "$REMOTE_TARGET" 2>/dev/null || true
    fi
}
trap cleanup EXIT

TARGET_DIR="/media/ahaufox/data/mcp-gateway/images"
IMAGES=("mcp-gateway/proxy:latest" "mcp-gateway/douyin-mcp:latest" "mcp-gateway/jules-mcp-server:latest")

echo -e "${BLUE}开始部署 MCP Gateway 到 ${REMOTE_TARGET}...${NC}"

# 0. 建立主连接 (仅需输入一次密码)
echo -e "${YELLOW}[0/4] 建立 SSH 控制连接 (如需密码请在此输入)...${NC}"
ssh $SSH_OPTS -fN "$REMOTE_TARGET"

# 1. 本地构建
echo -e "${YELLOW}[1/4] 在本地构建 Docker 镜像...${NC}"
docker compose build

# 2. 同步配置文件
echo -e "${YELLOW}[2/4] 同步配置文件到远程服务器...${NC}"

# 安全清理：如果远程存在同名的目录（通常是 Docker 自动创建的），先将其删除，否则挂载会失败
ssh -S "$SOCKET_PATH" "$REMOTE_TARGET" "mkdir -p $TARGET_DIR && rm -rf $TARGET_DIR/mcp-proxy/config.json"

# 使用 --relative 保持目录结构，且仅同步特定文件，避免扫描 .venv 等无关目录
RSYNC_FILES=(
    "docker-compose.yaml"
    "mcp-proxy/config.json"
    "mcp-proxy/.env"
    "jules-mcp-server/.env"
)

rsync -avz -e "ssh -S $SOCKET_PATH" --relative "${RSYNC_FILES[@]}" "$REMOTE_TARGET:$TARGET_DIR/"

# 3. 传输镜像
echo -e "${YELLOW}[3/4] 传输并加载 Docker 镜像 (可能较慢)...${NC}"
docker save "${IMAGES[@]}" | gzip | ssh -S "$SOCKET_PATH" -C "$REMOTE_TARGET" "gzip -d | docker load"

# 4. 远程启动
echo -e "${YELLOW}[4/4] 在远程启动 Docker 容器...${NC}"
ssh -S "$SOCKET_PATH" -t "$REMOTE_TARGET" "cd $TARGET_DIR && docker compose down && docker compose up -d"

# 完成
echo -e "${GREEN}[DONE] 部署完成！${NC}"
echo -e "${BLUE}服务地址: http://$(echo $REMOTE_TARGET | cut -d@ -f2):9090${NC}"
echo -e "${YELLOW}提示: 镜像已在本地构建并传输，远程服务器无需保留源码。${NC}"
