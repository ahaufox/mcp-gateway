# MCP Proxy Server

一个高性能的 MCP 代理服务器，可将多个 MCP 服务器聚合在单个 HTTP 入口点之后。

[English Version](./README_EN.md)

## ✨ 核心特性

- **多 Server 聚合**: 自动聚合工具 (Tools)、提示词 (Prompts) 和资源 (Resources)。
- **多样化传输**: 支持 SSE 以及可流式 HTTP 模式。
- **灵活配置**: 支持 `stdio`、`sse` 和 `streamable-http` 客户端。
- **内置支持**: Docker 镜像原生支持 `npx` 和 `uvx` 运行 downstream 服务器。
- **安全配置**: 支持通过环境变量注入敏感凭据 (`AUTH_TOKENS`) 和基础配置 (`MCP_BASE_URL`)。
- **状态监控**: Dashboard 实时显示 MCP 服务的连接状态 (Connected/Failed) 与错误信息。
- **标准架构**: 采用标准的 Go 项目目录结构 (`cmd`, `internal`)，提升可维护性。

## 📖 指南导航

- 🚀 **[使用指南 (Guide)](./docs/USAGE_CN.md)**: 部署步骤、参数说明与接口端点。
- 🛠️ **[配置指南 (Config)](./docs/CONFIGURATION.md)** ([中文](./docs/CONFIGURATION_CN.md)): 详细的 JSON 配置项说明。
- 📦 **[部署指南 (Deploy)](./docs/DEPLOYMENT.md)**: Docker Compose 多服务部署与安全配置。
- 💻 **[内部开发 (Dev)](./docs/DEVELOPMENT_CN.md)**: 源码结构、编译命令与故障排查。

## ⚡ 快速开始

在**项目根目录** (`mcp-gateway/`) 下执行：

```bash
docker compose build && docker compose up -d
```

在线 Claude 配置转换器: [config-converter](http://localhost:9090/docs/)
