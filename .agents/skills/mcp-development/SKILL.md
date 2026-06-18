---
name: mcp-development
description: 专门用于构建、验证和分发 MCP (Model Context Protocol) 服务的高级技能。
---

# 🔌 MCP 开发者技能规范 (MCP Development)

本技能由“**资深 MCP 架构师 (Senior MCP Architect)**”主导，旨在确保所有新增的 MCP 服务具备强健的安全性、清晰的接口定义以及高效的资源管理。

## 🎯 触发场景
- 发现项目中新增了 `mcp-servers` 或类似目录。
- 正在编写 `mcp` 协议相关的 JSON-RPC 接口或工具。

## 🛠️ 核心能力
### 1. 协议合规性 (Protocol Compliance)
- 强制遵循 MCP 核心规范（Resources, Prompts, Tools）。
- 确保所有 Tool 定义具备详尽的 `description` 和合法的 `inputSchema`。

### 2. 安全隔离 (Security Sandbox)
- **输入校验**: 所有来自模型端的工具参数必须进行防注入校验。
- **权限最小化**: 确保 MCP 服务器仅具备执行其任务所需的最小文件系统/网络权限。

### 3. 连接与生命周期
- 实现稳健的错误处理和优雅降级。
- 确保上下文窗口的有效利用，避免向模型发送冗余的资源数据。

## 🚫 负向约束 (Negative Constraints)
- **禁止** 在工具响应中透传敏感的系统环境变量。
- **禁止** 开发无超时控制的长链接工具。
- **禁止** 使用不具备 JSON Schema 校验的直接参数映射。

## 💡 最佳实践 (CoT Checklist)
1. **定义阶段**: 我的工具名是否唯一？参数描述是否能让模型准确理解其副作用？
2. **实现阶段**: 异常捕获是否能返回给模型有意义的提示，而非原始堆栈？
3. **分发阶段**: 是否配置了 `mcp_config.json` 及必要的启动脚本？

## 📦 Python MCP SDK 选型规范

### 强制要求
- **新建 Python MCP 服务必须使用官方 `mcp` SDK**（`from mcp.server.fastmcp import FastMCP`）。
- **禁止**在新项目中使用第三方 `fastmcp` 包（`from fastmcp import FastMCP`）。
- 已有外部引入项目（如 `jules-mcp-server`）可保留第三方 `fastmcp`，不强制迁移。

### 选型理由

| 维度 | 官方 `mcp` SDK | 第三方 `fastmcp` 包 |
|------|---------------|---------------------|
| 维护方 | Anthropic 官方 | 社区 |
| 协议跟进 | 零延迟 | 有延迟风险 |
| 安全特性 | 内置 DNS rebinding 保护 | 无内置保护 |
| ASGI app | `streamable_http_app()` | `http_app(transport=...)` |

### Docker 环境配置
官方 SDK 默认启用 DNS rebinding protection，Docker 容器名不在默认 `allowed_hosts` 中，需配置：

```python
from mcp.server.fastmcp import FastMCP
from mcp.server.transport_security import TransportSecuritySettings

mcp = FastMCP("My MCP Server",
              transport_security=TransportSecuritySettings(
                  enable_dns_rebinding_protection=False  # Docker 内网安全
              ))
```

---
> [!TIP]
> **Proactive Growth**: 发现新工具需求时，应优先考虑将其封装为 MCP Tool 以实现智能助手的能力平替。
