# MCP Gateway (聚合网关)

🚀 **一处聚合，全方位连接**。MCP Gateway 是一个面向 **Model Context Protocol (MCP)** 的企业级代理聚合方案。

本项目旨在解决多 MCP Server 接入复杂、入口分散的问题，通过统一的 HTTP/SSE 网关提供标准化服务。

- **[规范指南 (Rules)](./docs/readme-rules.md)**: 涵盖安全、质量、技术栈锁定及 P0/P1/P2 风险分级开发标准。
- **[自动化工作流 (Workflows)](./docs/readme-workflow.md)**: 提供 API 变更、契约同步、测试与部署等标准化任务。
- **[开发路线图 (Roadmap)](./docs/ROADMAP_CN.md)**: 项目后续开发里程碑与关键任务规划。
- **[AI 技能库 (Skills)](./.agent/skills/)**: 包含数据 Mock (backend)、UI 设计 (frontend)、GUI 自动化 (backend/gui-automation)、数据可视化 (backend/data-visualization) 等专项技能。

## 📁 项目结构

```text
.
├── .agent/              # AI 助手配置 (Rules, Skills, Workflows)
├── docs/                # 详细说明文档 (中/英)
│   ├── DEVELOPMENT_CN.md # 二次开发指南
│   ├── USAGE_CN.md       # 使用指南
│   ├── ROADMAP_CN.md     # 开发路线图
│   ├── readme-rules.md   # 开发规范索引
│   └── readme-workflow.md # 自动化工作流索引
├── mcp-proxy/           # 核心聚合网关（主要逻辑）
├── douyin-mcp/          # 抖音视频解析 MCP 服务
├── jules-mcp-server/    # [子模块] Jules AI 代理服务
├── PyMCPAutoGUI/        # [子模块] GUI 自动化 MCP 服务
├── mcp-server-chart/    # [子模块] 图表生成 MCP 服务
└── README.md            # 项目主入口 (当前文件)
```

## ✨ 核心特性

- **多 Server 聚合**: 自动聚合各服务器的 Tools、Prompts 和 Resources。
- **标准协议**: 完全兼容 Model Context Protocol（支持 Streamable HTTP 与 SSE 双模式）。
- **Docker 一键部署**: 通过 `docker-compose` 统一编排所有服务。
- **自动化驱动**: 内置 Git 自动化提交、安全审计及发布准备工作流。
- **双语支持**: 核心文档提供中英双语版本。
- **全新 UI**: 现代化的 Dashboard，全站汉化，集成 [Changelog](./mcp-proxy/templates/changelog.html) 页面。
- **多格式转换**: 内置配置转换器，支持 Claude 及 Antigravity 格式一键生成。

## 🔌 已集成 MCP Server

| 名称 | 类型 | 说明 |
|------|------|------|
| **stitch** | Streamable HTTP | Google Stitch UI 设计与代码生成 |
| **github** | Stdio (npx) | GitHub 仓库操作（PR、Issue 等） |
| **chart** | Stdio (npx) | AntV 图表生成 |
| **fetch** | Stdio (uvx) | 网页内容抓取 |
| **notion** | Stdio (npx) | Notion 笔记管理 |
| **jules** | SSE (Docker) | Jules AI 代理服务 |
| **douyin** | SSE (Docker) | 抖音无水印下载与文案提取 |

---
> [!TIP]
> 如果你是第一次使用本项目，建议从 **[使用指南](./docs/USAGE_CN.md)** 开始。