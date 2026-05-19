# MCP Gateway (聚合网关)

🚀 **一处聚合，全方位连接**。MCP Gateway 是一个面向 **Model Context Protocol (MCP)** 的企业级网关聚合方案，提供统一的 HTTP/SSE 标准化服务。

## 🚀 快速开始

### 本地启动

根目录下执行：
```bash
docker compose build && docker compose up -d
```

### 远程自动化部署

项目内置了远程部署脚本，适用于服务器自动更新场景：

```bash
# 1. 配置脚本中的服务器 SSH 地址和项目路径
vim scripts/remote_deploy.sh

# 2. 给予执行权限并运行
chmod +x scripts/remote_deploy.sh && ./scripts/remote_deploy.sh
```

脚本将自动完成：`git pull` → `docker compose build` → `docker compose up -d`。

## ⚙️ 环境变量配置

详见根目录 [.env.example](./.env.example)。以下为各子服务使用的环境变量说明：

| 变量 | 适用服务 | 必填 | 默认值 | 说明 |
|------|---------|------|--------|------|
| `API_KEY` | douyin-mcp | 是 | - | 硅基流动 / 阿里云百炼 API 密钥（用于语音转文字） |
| `LOG_LEVEL` | douyin-mcp | 否 | `INFO` | 日志级别：`DEBUG`, `INFO`, `WARNING`, `ERROR` |
| `MCP_TRANSPORT` | douyin-mcp | 否 | `stdio` | 传输模式：`stdio`, `sse`, `streamable-http` |
| `MCP_PORT` | douyin-mcp | 否 | `8000` | SSE/HTTP 模式下的监听端口 |

## 📁 项目结构

- **mcp-proxy/**: 核心聚合网关（主要逻辑）
- **douyin-mcp/**: 抖音视频解析服务
- **jules-mcp-server/**: Jules AI 代理 [子模块]
- **PyMCPAutoGUI/**: GUI 自动化 [子模块]
- **mcp-server-chart/**: 图表生成 [子模块]
- **docs/**: [使用指南](./docs/USAGE_CN.md) | [二次开发](./docs/DEVELOPMENT_CN.md) | [路线图](./docs/ROADMAP_CN.md)

## ✨ 核心特性

- **多 Server 聚合**: 自动聚合 Tools、Prompts 和 Resources。
- **全平台兼容**: 支持 Streamable HTTP 与 SSE 模式。
- **现代化 UI**: 内置全站汉化 Dashboard。
- **自动化驱动**: 内置提交、安全审计及发布工作流。

## 🔌 已集成服务

| 名称 | 类型 | 说明 |
|------|------|------|
| **stitch** | HTTP | UI 设计与代码生成 |
| **github** | Stdio | 仓库操作 (PR/Issue) |
| **chart** | Stdio | 图表生成 |
| **fetch** | Stdio | 网页爬取 |
| **douyin** | SSE | 抖音下载与文案提取 |
| **jules** | SSE | AI 代理服务 |

---
> [!TIP]
> 详细文档请参考 **[使用指南](./docs/USAGE_CN.md)**。