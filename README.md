# MCP Gateway (聚合网关)

一处聚合，全方位连接。**MCP Gateway** 是一个面向 **Model Context Protocol (MCP)** 的企业级网关聚合方案，提供统一的 HTTP/SSE 标准化服务。

## 项目组成

本项目由 **一个主项目 + 五个子项目** 组成：

| 项目 | 类型 | 语言 | 说明 |
|------|------|------|------|
| **mcp-proxy** | 🏠 主项目（核心网关） | Go | MCP 协议聚合网关引擎，负责统一接入、认证、路由、熔断、工具聚合 |
| **web** | 🏠 主项目（前端面板） | TypeScript/React | 现代化 Dashboard，提供服务监控、配置转换、变更日志等管理界面 |
| **douyin-mcp** | 🔌 子服务 | Python | 抖音视频解析 MCP 服务，支持无水印下载、音频提取、语音转文字 |
| **jules-mcp-server** | 🔌 子服务 [子模块] | Python | Jules AI 代理服务，提供 AI 驱动的自动化能力 |
| **PyMCPAutoGUI** | 🔌 子服务 [子模块] | Python | GUI 自动化 MCP 服务，支持屏幕控制与 OCR 识别 |
| **mcp-server-chart** | 🔌 子服务 [子模块] | TypeScript | 图表生成 MCP 服务，基于 AntV 的数据可视化 |

主项目 `mcp-proxy` 作为中心网关，通过 Stdio、SSE、Streamable HTTP、WebSocket 等多种协议连接各子服务，对外暴露统一的 MCP 端点。

```
┌──────────────┐     ┌─────────────────────────────────────┐     ┌──────────────────────┐
│  客户端/LLM   │────▶│         MCP Gateway (mcp-proxy)      │────▶│   douyin-mcp          │
│  (HTTP/SSE)  │◀────│  认证 → 路由 → 聚合 → 钩子 → 熔断    │◀────│   jules-mcp-server    │
└──────────────┘     └─────────────────────────────────────┘     │   PyMCPAutoGUI        │
                                                                    │   mcp-server-chart    │
                                                                    │   (及外部 npx/uvx 服务) │
                                                                    └──────────────────────┘
```

## 技术栈

| 模块 | 技术 |
|------|------|
| **核心网关** | Go 1.24 + [mcp-go SDK](https://github.com/mark3labs/mcp-go) |
| **前端面板** | React 19 + TypeScript + Vite 8 + Tailwind CSS 4 |
| **子服务** | Python 3.10+ (FastMCP), Node.js 22 |
| **部署** | Docker + Docker Compose |

## 核心特性

- **多 Server 聚合**：同时连接多个下游 MCP 服务器，自动聚合 Tools、Prompts 和 Resources
- **多传输协议**：stdio（子进程）、SSE、Streamable HTTP、WebSocket
- **企业级保障**：认证鉴权（Bearer Token）、工具过滤（allow/block）、熔断器、重试机制、健康检查与自动重连
- **钩子系统**：请求/响应钩子链，支持自定义扩展
- **批量调用**：支持顺序/并行批量工具调用
- **工具重命名**：命名空间、前缀、重命名映射
- **现代化 UI**：内置 Dashboard，实时监控服务状态
- **环境变量注入**：配置中 `${VAR}` 格式自动引用环境变量

## 快速开始

### 本地启动

```bash
docker compose build && docker compose up -d
```

或使用一键脚本：

```bash
./init.sh && ./local_deploy.sh
```

### 远程部署

```bash
# 配置服务器 SSH 地址和项目路径
vim scripts/remote_deploy.sh

# 运行远程部署
chmod +x scripts/remote_deploy.sh && ./scripts/remote_deploy.sh
```

脚本将自动完成：`git pull` → `docker compose build` → `docker compose up -d`。

## 已集成服务

| 名称 | 传输类型 | 启动方式 | 说明 |
|------|---------|---------|------|
| **stitch** | streamable-http | 远程 API | UI 设计与代码生成 |
| **github** | stdio | npx | 仓库操作 (PR/Issue) |
| **chart** | stdio | npx | 图表生成 (AntV) |
| **fetch** | stdio | uvx | 网页爬取 |
| **notion** | stdio | npx | Notion API |
| **playwright** | stdio | npx | 浏览器自动化 |
| **sequential-thinking** | stdio | npx | 序列化思维 |
| **context7** | stdio | npx | Upstash Context7 |
| **reactbits** | stdio | npx | React Bits 组件 |
| **langchain-docs** | streamable-http | 远程 API | LangChain 文档 |
| **jules** | sse | 容器 | AI 代理服务 |
| **douyin** | sse | 容器 | 抖音下载与文案提取 |

## 项目结构

```
├── mcp-proxy/              # 核心聚合网关（Go）
│   ├── cmd/mcp-proxy/      #   程序入口
│   ├── internal/
│   │   ├── config/         #   配置加载与解析
│   │   ├── core/           #   MCP 客户端核心（连接/心跳/重连/熔断）
│   │   ├── server/         #   HTTP 服务器 + 嵌入前端
│   │   ├── transport/      #   统一传输层接口
│   │   ├── tools/          #   工具名称重写
│   │   ├── batch/          #   批量工具调用
│   │   ├── hook/           #   请求/响应钩子链
│   │   ├── cache/          #   工具缓存
│   │   ├── retry/          #   重试中间件
│   │   ├── circuitbreaker/ #   熔断器
│   │   ├── process/        #   进程管理器
│   │   └── errors/         #   错误码与包装
│   ├── config.json         #   服务注册配置
│   └── Dockerfile          #   多阶段构建
├── web/                    # React 前端 Dashboard
│   └── src/
│       ├── pages/          #   仪表盘 / 配置转换 / 变更日志 / 登录
│       ├── components/     #   导航栏 / 页脚
│       └── context/        #   主题上下文
├── douyin-mcp/             # 抖音视频解析服务（Python）
├── jules-mcp-server/       # Jules AI 代理 [子模块]
├── PyMCPAutoGUI/           # GUI 自动化 [子模块]
├── mcp-server-chart/       # 图表生成 [子模块]
├── docs/                   # 文档
├── scripts/                # 部署脚本
├── .agents/                # AI 代理规则
├── docker-compose.yaml     # Docker Compose 编排
├── init.sh                 # 初始化脚本
├── local_deploy.sh         # 本地部署脚本
└── remote_deploy.sh        # 远程部署脚本
```

## 环境变量配置

详见 [.env.example](./.env.example)。各服务关键变量：

| 变量 | 适用服务 | 说明 |
|------|---------|------|
| `API_KEY` | douyin-mcp | 语音转文字 API 密钥 |
| `DOUYIN_MCP_API_KEY` | 网关 | 抖音 MCP 认证密钥 |
| `GITHUB_PERSONAL_ACCESS_TOKEN` | 网关 | GitHub API 令牌 |
| `JULES_API_KEY` / `JULES_MCP_URL` | 网关 | Jules 认证与地址 |
| `STITCH_API_KEY` | 网关 | Stitch API 密钥 |
| `NOTION_API_KEY` | 网关 | Notion API 令牌 |
| `AUTH_TOKENS` | 网关 | 全局 Bearer 认证令牌 |
| `MCP_BASE_URL` | 网关 | 服务基础地址 |

## API 端点

| 路径 | 说明 |
|------|------|
| `/sse` | SSE 端点（MCP over SSE） |
| `/messages` | Streamable HTTP 消息端点 |
| `/api/servers` | 服务列表与状态 |
| `/api/config` | 当前配置 |
| `/api/platform-config` | 平台配置转换 |

## 🚀 未来规划与社区贡献 (Future & Contributing)

本项目正在积极迭代中，许多核心特性已经完成了底层代码的设计，急需将它们集成到网关的主流程中。我们非常欢迎并期待社区提交 PR 来共同完善！

详细的开发计划、集成步骤和架构设计详见 📋 **[未来迭代优化方案 (FUTURE_ITERATION_PLAN.md)](./docs/FUTURE_ITERATION_PLAN.md)**。

### 🛠️ 核心认领看板

| 模块名称 | 任务类别 | 核心工作描述 | 推荐认领 |
| :--- | :--- | :--- | :--- |
| **重试中间件集成** | ⏳ 模块集成 | 将 `retry/middleware.go` 的指数退避重试接入 `core/client.go` 的工具调用链中。 | 💡 推荐新手 |
| **命名空间与重写集成** | ⏳ 模块集成 | 将 `tools/rewrite.go` 引入到工具聚合注册和反向路由映射中。 | 🔥 核心工作 |
| **缓存集成** | ⏳ 模块集成 | 将 `cache/toolcache.go` 接入 `client.CallTool` 调用流以实现结果缓存。 | ⚡ 性能优化 |
| **批量调用路由挂载** | ⏳ 模块集成 | 将 `batch/batch.go` 挂载为服务端 `/api/batch` 路由端点。 | 🌐 接口开发 |
| **WebSocket 客户端集成** | ⏳ 模块集成 | 启用 `transport/websocket_client.go` 替换现有 SSE 回退逻辑。 | 🔌 传输优化 |
| **进程管理器替换** | ⏳ 模块集成 | 用 `process/manager.go` 替代 Stdio 客户端原生的 stdio 进程启动/重启逻辑。 | ⚙️ 进程管理 |
| **WebSocket 服务端** | 🛠️ 新增开发 | 实现 `server/websocket_server.go` 并接入统一 Transport 循环。 | 🔥 核心工作 |
| **高级限流器 (RateLimiter)** | 🛠️ 新增开发 | 从零开发令牌桶限流器并集成到网关中间件。 | 🛡️ 安全策略 |
| **请求队列与连接池** | 🛠️ 新增开发 | 从零开发 `queue/queue.go` 与 `pool/pool.go` 提供并发背压保护与连接复用。 | 🚀 性能优化 |

欢迎参考 [二次开发指南](./docs/DEVELOPMENT_CN.md) 并在认领任务前，先在 Issue 中进行沟通以避免重复开发！

## 详细文档

- [使用指南](./docs/USAGE_CN.md)
- [二次开发](./docs/DEVELOPMENT_CN.md)
- [路线图](./docs/ROADMAP_CN.md)
