# MCP Gateway 使用指南

欢迎使用 MCP Gateway！本项目旨在提供一个统一的入口点，将多个 MCP (Model Context Protocol) 服务器聚合在一起。

## 🚀 快速开始

### 1. 运行核心代理 (mcp-proxy)

本项目使用 `mcp-proxy` 作为核心网关。

#### 使用 Docker (推荐)

确保你已经准备好了 `config.json` 配置文件。

```bash
docker run -d -p 9090:9090 \
  -v ./mcp-proxy/config.json:/config/config.json \
  ghcr.io/tbxark/mcp-proxy:latest
```

#### 源码编译

```bash
cd mcp-proxy
make build
./build/mcp-proxy --config config.json
```

### 2. 配置 MCP 服务器

在 `config.json` 中配置你想要聚合的服务器：

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "<YOUR_TOKEN>" }
    },
    "chart": {
       "url": "http://localhost:8080/sse"
    }
  }
}
```

## 📦 核心组件说明

- **mcp-proxy**: 核心聚合网关，支持 SSE 和流式 HTTP。
- **jules-mcp-server**: 预置的 MCP 服务器子模块。
- **PyMCPAutoGUI**: 预置的自动化控制服务器。
- **mcp-server-chart**: 预置的图表/数据可视化服务器。

## 🛠️ 常见问题

关于认证、端点地址以及更多高级配置，请参阅各子模块内的 `README` 或 `docs` 目录。
