# MCP Gateway (聚合网关)

🚀 **一处聚合，全方位连接**。MCP Gateway 是一个面向 **Model Context Protocol (MCP)** 的企业级代理聚合方案。

本项目旨在解决多 MCP Server 接入复杂、入口分散的问题，通过统一的 HTTP/SSE 网关提供标准化服务。

---

## 📖 核心指南

- 📖 **[使用指南 (Guide)](./docs/USAGE_CN.md)**: 快速部署、配置服务器并开始使用。
- 🛠️ **[二次开发 (Development)](./docs/DEVELOPMENT_CN.md)**: 了解如何添加新服务器、使用自动化工作流及遵守开发规范。

## 📁 项目结构

```text
.
├── .agent/              # AI 助手配置 (Rules, Skills, Workflows)
├── docs/                # 详细说明文档 (中/英)
├── mcp-proxy/           # 核心聚合网关（主要逻辑）
├── jules-mcp-server/    # [子模块] 预置服务器 1
├── PyMCPAutoGUI/        # [子模块] 预置服务器 2
├── mcp-server-chart/    # [子模块] 预置服务器 3
├── readme-rules.md      # 开发规范索引
├── readme-workflow.md   # 自动化工作流索引
└── README.md            # 项目主入口 (当前文件)
```

## ✨ 核心特性

- **多 Server 聚合**: 自动聚合各服务器的 Tools、Prompts 和 Resources。
- **标准协议**: 完全兼容 Model Context Protocol。
- **自动化驱动**: 内置 Git 自动化提交、安全审计及发布准备工作流。
- **双语支持**: 核心文档提供中英双语版本。

---
> [!TIP]
> 如果你是第一次使用本项目，建议从 **[使用指南](./docs/USAGE_CN.md)** 开始。