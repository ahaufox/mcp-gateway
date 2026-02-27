# MCP Gateway (聚合网关)

🚀 **一处聚合，全方位连接**。MCP Gateway 是一个面向 **Model Context Protocol (MCP)** 的企业级网关聚合方案，提供统一的 HTTP/SSE 标准化服务。

## 🚀 快速开始

- **一键部署**：根目录下执行
```bash
docker compose build && docker compose up -d
```

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