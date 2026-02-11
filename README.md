# MCP Gateway (聚合网关)

🚀 **一处聚合，全方位连接**。MCP Gateway 是一个面向 **Model Context Protocol (MCP)** 的企业级代理聚合方案。

本项目旨在解决多 MCP Server 接入复杂、入口分散的问题，通过统一的 HTTP/SSE 网关提供标准化服务。

- **[规范指南 (Rules)](./readme-rules.md)**: 涵盖安全、质量、技术栈锁定及 P0/P1/P2 风险分级开发标准。
- **[自动化工作流 (Workflows)](./readme-workflow.md)**: 提供 API 变更、契约同步、测试与部署等标准化任务。
- **[AI 技能库 (Skills)](./.agent/skills/)**: 包含数据 Mock (backend)、UI 设计 (frontend)、GUI 自动化 (backend/gui-automation)、数据可视化 (backend/data-visualization) 等专项技能。

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