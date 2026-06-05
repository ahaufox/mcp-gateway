# MCP Gateway 使用指南

MCP Gateway 是一个企业级 MCP 协议聚合网关，将多个下游 MCP 服务统一聚合在单一 HTTP/SSE 入口点之后。

## 快速开始

### 使用 Docker Compose（推荐）

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 填入必要的 API Key

# 2. 启动所有服务
docker compose build && docker compose up -d

# 3. 访问 Dashboard
open http://localhost:9090
```

### 源码编译（仅网关）

```bash
cd mcp-proxy
make build-all    # 先构建前端，再编译 Go 二进制
./build/mcp-proxy --config config.json
```

## 配置指南

网关通过 `mcp-proxy/config.json` 进行配置，支持 JSON 格式，且自动展开 `${VAR}` 环境变量引用。

### 网关自身配置

```json
{
  "mcpProxy": {
    "baseURL": "http://localhost:9090",
    "addr": ":9090",
    "name": "MCP Gateway",
    "version": "1.0.0",
    "type": "streamable-http",
    "options": {
      "authTokens": ["${AUTH_TOKENS}"],
      "logEnabled": true,
      "panicIfInvalid": false,
      "callTimeout": "30s",
      "maintenanceInterval": "30s",
      "toolFilter": {
        "mode": "allow",
        "list": ["*"]
      }
    }
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `baseURL` | string | 网关对外暴露的基础地址 |
| `addr` | string | 监听地址（`:9090`） |
| `type` | `sse` / `streamable-http` | 服务端暴露的传输协议 |
| `authTokens` | []string | 全局 Bearer Token 列表 |
| `callTimeout` | duration | 工具调用超时（如 `"30s"`） |
| `initializeTimeout` | duration | 初始化握手超时 |
| `listToolsTimeout` | duration | 工具列表查询超时 |
| `maxRetries` | int | 最大重试次数 |
| `retryDelay` | duration | 重试初始延迟 |
| `retryBackoff` | float | 重试退避因子 |
| `disablePing` | bool | 禁用健康检查 Ping |
| `maintenanceInterval` | duration | 维护检查间隔 |
| `toolFilter` | object | 全局工具过滤策略 |

### 注册 MCP 服务

#### Stdio 类型（本地子进程）

通过 `npx` / `uvx` 启动的本地 MCP 服务：

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PERSONAL_ACCESS_TOKEN}"
      },
      "options": {
        "toolFilter": {
          "mode": "block",
          "list": ["create_repository", "fork_repository"]
        }
      }
    },
    "fetch": {
      "command": "uvx",
      "args": ["mcp-server-fetch"]
    }
  }
}
```

#### SSE 类型（远程服务）

通过 SSE 协议连接远程服务：

```json
{
  "mcpServers": {
    "douyin": {
      "transportType": "sse",
      "url": "http://douyin-mcp:8000/sse",
      "options": {
        "disablePing": true
      }
    }
  }
}
```

#### Streamable HTTP 类型

```json
{
  "mcpServers": {
    "stitch": {
      "url": "https://stitch.googleapis.com/mcp",
      "transportType": "streamable-http",
      "headers": {
        "X-Goog-Api-Key": "${STITCH_API_KEY}"
      }
    }
  }
}
```

### 服务级选项

每个服务可独立配置以下选项（未设置时继承全局配置）：

| 选项 | 说明 |
|------|------|
| `disabled` | 是否禁用该服务 |
| `authTokens` | 服务级独立鉴权令牌 |
| `callTimeout` | 该服务的工具调用超时 |
| `maxRetries` / `retryDelay` / `retryBackoff` | 重试策略 |
| `toolFilter` | 该服务的工具过滤（allow/block 列表） |
| `circuitBreaker` | 熔断器配置 |
| `disablePing` | 禁用 Ping 健康检查 |
| `logEnabled` | 启用请求日志 |
| `panicIfInvalid` | 初始化失败是否直接退出 |

## API 端点

网关启动后，每个注册的服务暴露以下 MCP 端点：

| 路径 | 协议 | 说明 |
|------|------|------|
| `/{service_name}/` | SSE / Streamable HTTP | MCP 服务端点 |
| `/` | HTTP | Dashboard 主页 |
| `/api/servers` | JSON | 所有服务状态、工具列表、提示词、资源 |
| `/api/config` | JSON | 当前生效配置 |
| `/api/platform-config` | JSON | 平台配置转换 |
| `/docs/` | HTML | 配置转换器工具 |
| `/changelog/` | HTML | 变更日志 |
| `/login` | HTML | 登录页面 |

### 客户端连接

LLM 客户端通过 SSE 或 Streamable HTTP 连接网关：

```
# SSE 端点
http://localhost:9090/github/sse

# Streamable HTTP 端点
http://localhost:9090/github/messages
```

## Dashboard 使用

访问 `http://localhost:9090` 可查看 Dashboard：

- **服务状态总览**：每个注册服务的连接状态、工具数量、提示词和资源计数
- **实时健康**：绿色/黄色/红色指示器反映服务健康状态
- **配置转换器**：`/docs/` 页面可将 Claude Desktop 配置格式转换为网关配置格式

## 认证配置

支持 Bearer Token 认证。客户端在请求头中携带令牌：

```
Authorization: Bearer <your_token>
```

配置方式：

```json
{
  "mcpProxy": {
    "options": {
      "authTokens": ["token1", "token2"]
    }
  },
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "options": {
        "authTokens": ["service_specific_token"]
      }
    }
  }
}
```

- 全局令牌对所有服务生效
- 服务级令牌仅对该服务生效
- 如果某服务未设置 `authTokens`，则继承全局令牌

## 工具过滤

支持两种过滤模式：

- `allow` 模式：仅列出清单中的工具
- `block` 模式：排除清单中的工具，其余全部暴露

```json
{
  "options": {
    "toolFilter": {
      "mode": "block",
      "list": ["create_repository", "delete_repository"]
    }
  }
}
```

## 环境变量一览

| 变量 | 说明 |
|------|------|
| `MCP_BASE_URL` | 网关基础地址，默认 `http://localhost:9090` |
| `AUTH_TOKENS` | 全局认证令牌 |
| `GITHUB_PERSONAL_ACCESS_TOKEN` | GitHub API 令牌 |
| `STITCH_API_KEY` | Stitch API 密钥 |
| `NOTION_API_KEY` | Notion API 令牌 |
| `JULES_MCP_URL` | Jules 服务地址 |
| `DOUYIN_MCP_URL` | 抖音服务地址 |
| `DOUYIN_MCP_API_KEY` | 抖音认证密钥 |

## 故障排查

### 查看服务状态

```bash
curl http://localhost:9090/api/servers | jq
```

返回每个服务的名称、状态、工具列表、错误信息。

### 检查容器日志

```bash
docker compose logs -f app
```

### 常见问题

| 问题 | 原因 | 解决 |
|------|------|------|
| 服务状态 `error` | 下游服务未启动或网络不通 | 检查 `depends_on` 和容器健康检查 |
| 工具调用超时 | 下游响应慢 | 增大 `callTimeout` 配置 |
| 401 未授权 | Token 缺失或不匹配 | 检查 `authTokens` 配置和请求头 |
| 工具找不到 | 工具名被过滤 | 检查 `toolFilter` 配置 |
| npx 服务启动失败 | 依赖未安装 | 检查 Docker 镜像内 Node.js 环境 |

## 更多文档

- [二次开发指南](./DEVELOPMENT_CN.md)
- [功能迭代计划](./FUTURE_ITERATION_PLAN.md)
- [项目路线图](./ROADMAP_CN.md)
