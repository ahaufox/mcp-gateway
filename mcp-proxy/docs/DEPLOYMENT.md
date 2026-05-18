# Deployment

## Docker Compose (推荐)

项目根目录下包含完整的 `docker-compose.yaml`，一键启动所有服务：

```bash
cd mcp-gateway
docker compose build && docker compose up -d
```

包含以下服务：

| 服务 | 端口 | 说明 |
|------|------|------|
| `app` (mcp-proxy) | 9090 | MCP 代理网关，聚合所有下游 MCP 服务 |
| `douyin-mcp` | 8100 | 抖音视频下载/解析 MCP 服务 |
| `jules-mcp-server` | 8002 | Jules AI 编码助手 MCP 服务 |

> [!IMPORTANT]
> 首次启动前，请确保 `mcp-proxy/.env` 中配置了所需的环境变量（如 `AUTH_TOKENS`, `STITCH_API_KEY` 等）。

## Docker (单独运行 mcp-proxy)

Run with a local config file mounted into the container:

```bash
docker run -d \
  -p 9090:9090 \
  -v /path/to/config.json:/config/config.json \
  ghcr.io/tbxark/mcp-proxy:latest
```

Or reference a remote config URL:

```bash
docker run -d -p 9090:9090 \
  ghcr.io/tbxark/mcp-proxy:latest \
  --config https://example.com/config.json
```

The image supports launching MCP servers via `npx` and `uvx` out of the box.

## Security Notes

- Prefer `authTokens` per downstream server; only use the `mcpProxy` default when appropriate.
- If a downstream server cannot set headers, you can embed a token in the route key (e.g. `fetch/<token>`) and route via that path.
- Set `options.panicIfInvalid: true` for critical servers to fail fast on misconfiguration.
