# MCP Proxy Server

一个高性能的 MCP 代理服务器，可将多个 MCP 服务器聚合在单个 HTTP 入口点之后。

[English Version](./README_EN.md)

## ✨ 核心特性

- **多 Server 聚合**: 自动聚合工具 (Tools)、提示词 (Prompts) 和资源 (Resources)。
- **多样化传输**: 支持 SSE 以及可流式 HTTP 模式。
- **灵活配置**: 支持 `stdio`、`sse` 和 `streamable-http` 客户端。
- **内置支持**: Docker 镜像原生支持 `npx` 和 `uvx` 运行 downstream 服务器。

## 📖 指南导航

- 🚀 **[使用指南 (Guide)](./docs/USAGE_CN.md)**: 部署步骤、参数说明与接口端点。
- 🛠️ **[配置指南 (Config)](./docs/CONFIGURATION.md)** ([中文](./docs/CONFIGURATION_CN.md)): 详细的 JSON 配置项说明。
- 💻 **[内部开发 (Dev)](./docs/DEVELOPMENT_CN.md)**: 源码结构、编译命令与扩展逻辑。

## ⚡ 快速开始

```bash
docker run -d -p 9090:9090 \
  -v $(pwd)/config.json:/config/config.json \
  ghcr.io/tbxark/mcp-proxy:latest
```

在线 Claude 配置转换器: [https://tbxark.github.io/mcp-proxy](https://tbxark.github.io/mcp-proxy)

## 🤝 鸣谢

- 本项目由 [adamwattis/mcp-proxy-server](https://github.com/adamwattis/mcp-proxy-server) 启发。
- 特别感谢 [@ccbikai](https://github.com/ccbikai) 在 [Docker 沙箱运行 MCP](https://miantiao.me/posts/guide-to-running-mcp-server-in-a-sandbox/) 方面的分享。

## 📄 开源协议

MIT License. 详见 [LICENSE](LICENSE)。
