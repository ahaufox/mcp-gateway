# MCP Proxy Server (中文版)

一个 MCP 代理服务器，可将多个 MCP 服务器聚合在单个 HTTP 入口点之后。

[English Version](./README.md)

## 特性

- **聚合多个 MCP 客户端**：从多个服务器聚合工具 (Tools)、提示词 (Prompts) 和资源 (Resources)。
- **SSE 与流式 HTTP**：支持通过 Server‑Sent Events (SSE) 或可流式 HTTP 进行服务。
- **灵活配置**：支持 `stdio`、`sse` 和 `streamable-http` 类型的客户端。

## 文档

- **配置指南**：[docs/CONFIGURATION.md](docs/CONFIGURATION.md)
- **使用说明**：[docs/USAGE.md](docs/USAGE.md)
- **部署指南**：[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)
- **Claude 配置转换器**：https://tbxark.github.io/mcp-proxy

## 快速开始

### 从源码编译

```bash
git clone https://github.com/tbxark/mcp-proxy.git
cd mcp-proxy
make build
./build/mcp-proxy --config path/to/config.json
```

### 通过 Go 安装

```bash
go install github.com/tbxark/mcp-proxy@latest
```

### Docker 运行

该镜像内置了对通过 `npx` 和 `uvx` 启动 MCP 服务器的支持。

```bash
# 使用本地挂载的配置运行
docker run -d -p 9090:9090 -v /path/to/config.json:/config/config.json ghcr.io/tbxark/mcp-proxy:latest

# 或者使用远程配置文件运行
docker run -d -p 9090:9090 ghcr.io/tbxark/mcp-proxy:latest --config https://example.com/config.json
```

更多部署选项（包括 docker‑compose）请参阅 [docs/deployment.md](docs/DEPLOYMENT.md)。

## 配置

详细的配置参考和示例请参阅 [docs/configuration.md](docs/CONFIGURATION.md)。
在线 Claude 配置转换器：https://tbxark.github.io/mcp-proxy

## 使用方法

命令行参数、接口端点和认证示例请参阅 [docs/usage.md](docs/USAGE.md)。

## 感谢

- 本项目深受 [adamwattis/mcp-proxy-server](https://github.com/adamwattis/mcp-proxy-server) 启发。
- 如果您对部署有任何疑问，可以参考 [@ccbikai](https://github.com/ccbikai) 的文章 [《在 Docker 沙箱中运行 MCP Server》](https://miantiao.me/posts/guide-to-running-mcp-server-in-a-sandbox/)。

## 开源协议

本项目采用 MIT 协议。详情请参阅 [LICENSE](LICENSE) 文件。
