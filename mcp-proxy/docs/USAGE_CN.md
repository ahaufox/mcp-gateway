# mcp-proxy 使用指南

`mcp-proxy` 是一个高性能的 MCP (Model Context Protocol) 代理服务器，旨在聚合多个下游 MCP 服务器并提供统一的 HTTP/SSE 访问入口。

## 🚀 快速启动

### 镜像运行 (推荐)

```bash
docker run -d -p 9090:9090 \
  -v /path/to/config.json:/config/config.json \
  ghcr.io/tbxark/mcp-proxy:latest
```

### 源码编译

需要 Go 1.24+ 环境。

```bash
git clone https://github.com/ahaufox/mcp-proxy.git
cd mcp-proxy
make build
./build/mcp-proxy --config path/to/config.json
```

## 🛠️ 命令行参数

- `-config`: 配置文件路径或远程 URL (默认 "config.json")。
- `-expand-env`: 是否展开配置文件中的环境变量 (默认 true)。
- `-http-headers`: 获取远程配置时的可选 HTTP 头。
- `-insecure`: 跳过远程配置的 TLS 验证。
- `-version`: 打印版本并退出。

## 📡 接口端点 (Endpoints)

假设 `mcpProxy.baseURL = https://mcp.example.com` 且服务器 Key 为 `github`：

- **SSE 模式**: `https://mcp.example.com/github/sse`
- **流式 HTTP 模式**: `https://mcp.example.com/github/mcp`

## 🔐 身份验证

在配置中设置 `authTokens` 后，请求需包含：
`Authorization: Bearer <token>`

如果客户端不支持自定义 Header，可将 Token 嵌入 URL：
`https://mcp.example.com/github/<token>/sse`

---
详细配置示例请参考 [docs/CONFIGURATION.md](CONFIGURATION.md)。
