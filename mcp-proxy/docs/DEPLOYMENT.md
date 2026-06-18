# 部署指南 (Deployment)

本项目提供三种部署方式：Docker Compose 全栈部署（推荐）、独立 Docker 容器部署、本地源码构建运行。

## 方式一：Docker Compose 全栈部署（推荐）

项目根目录 [docker-compose.yaml](file:///mnt/bd757a96-8cd5-4430-aba4-bbeeb031a354/mcp-gateway/docker-compose.yaml) 编排了 `mcp-proxy` 及下游 MCP 服务，可一键启动完整网关。

### 1. 准备环境变量

```bash
cd mcp-gateway
cp mcp-proxy/.env.example mcp-proxy/.env
# 按需修改 AUTH_TOKENS / STITCH_API_KEY / GITHUB_PERSONAL_ACCESS_TOKEN 等
```

### 2. 启动服务

```bash
docker compose build && docker compose up -d
```

### 3. 包含的服务

| 服务 | 外部端口 | 说明 |
|------|----------|------|
| `app` (mcp-proxy) | 9090 | MCP 代理网关，聚合所有下游 MCP 服务 |
| `douyin-mcp` | 8100 | 抖音视频下载 / 解析 MCP 服务 |
| `jules-mcp-server` | 8002 | Jules AI 编码助手 MCP 服务 |

启动完成后访问 [http://localhost:9090](http://localhost:9090) 即可使用 Dashboard。

> [!IMPORTANT]
> - 首次启动前必须配置 [mcp-proxy/.env](file:///mnt/bd757a96-8cd5-4430-aba4-bbeeb031a354/mcp-gateway/mcp-proxy/.env.example) 中的 `AUTH_TOKENS`，否则下游服务的鉴权头会因变量未替换而失败。
> - 修改 `mcp-proxy/config.json` 后需执行 `docker compose restart app` 才会生效。

## 方式二：独立 Docker 容器部署

仅运行 `mcp-proxy` 容器，对接已有的下游 MCP 服务（远程 URL 或宿主机 stdio）。

### 使用本地配置文件

```bash
docker run -d \
  --name mcp-proxy \
  --restart unless-stopped \
  -p 9090:9090 \
  -v $(pwd)/mcp-proxy/config.json:/config/config.json \
  --env-file mcp-proxy/.env \
  mcp-gateway/proxy:latest
```

### 使用远程配置

```bash
docker run -d \
  --name mcp-proxy \
  --restart unless-stopped \
  -p 9090:9090 \
  mcp-gateway/proxy:latest \
  --config https://example.com/your-config.json
```

镜像构建请参见 [mcp-proxy/Dockerfile](file:///mnt/bd757a96-8cd5-4430-aba4-bbeeb031a354/mcp-gateway/mcp-proxy/Dockerfile)，默认入口已包含 `npx` 与 `uvx`，可直接启动 stdio 类型的下游服务器。

## 方式三：本地源码构建运行

适用于二次开发或调试场景。

### 1. 构建前端资源

```bash
cd web
npm ci && npm run build
# 构建产物输出到 mcp-proxy/internal/server/frontend/dist
```

### 2. 编译二进制

```bash
cd mcp-proxy
make build           # 当前平台
# 或
make buildLinuxX86   # 交叉编译 linux/amd64
```

产物位于 `mcp-proxy/build/mcp-proxy`。

### 3. 启动服务

```bash
./build/mcp-proxy --config ./config.json
```

命令行参数与配置文件约定见 [CONFIGURATION.md](file:///mnt/bd757a96-8cd5-4430-aba4-bbeeb031a354/mcp-gateway/mcp-proxy/docs/CONFIGURATION.md) / [CONFIGURATION_CN.md](file:///mnt/bd757a96-8cd5-4430-aba4-bbeeb031a354/mcp-gateway/mcp-proxy/docs/CONFIGURATION_CN.md)。

## 健康检查

容器启动后可通过以下方式验证：

```bash
# 检查网关 HTTP 端口
curl -fsS http://localhost:9090/healthz

# 查看容器日志
docker compose logs -f app
```

## 部署安全建议

- **必填 `authTokens`**：始终为 `mcpProxy.options.authTokens` 显式设置 Token，避免空集合导致鉴权失效。
- **下游隔离**：对不可信的下游服务，在其 `options.authTokens` 中覆盖独立的 Token，避免共享根 Token。
- **失败快速暴露**：对关键服务（如 fetch、github）设置 `options.panicIfInvalid: true`，配置错误时立即启动失败而不是静默降级。
- **凭据注入**：所有敏感信息（API Key、Token）应通过环境变量 `${VAR}` 注入，不要硬编码到 `config.json` 中。
- **最小权限**：使用 `options.toolFilter` 屏蔽危险工具（如 `create_repository`、`fork_repository`），参考 [config.json](file:///mnt/bd757a96-8cd5-4430-aba4-bbeeb031a354/mcp-gateway/mcp-proxy/config.json) 中 `github` 的 `mode: "block"` 示例。
- **公网暴露**：若将 9090 暴露到公网，请额外配置反向代理（HTTPS + 速率限制），并保留 `authTokens` 鉴权。
