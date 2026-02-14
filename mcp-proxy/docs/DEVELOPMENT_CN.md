# mcp-proxy 二次开发指南

本指南面向希望修改 `mcp-proxy` 核心逻辑或增加新协议支持的开发者。

## 💻 环境准备

- **语言**: Go 1.25+
- **工具**: Makefile, Docker (可选)

## 📁 核心结构

- `main.go`: 程序入口，处理命令行参数和服务器启动。
- `config.go`: 配置加载与 V1/V2 迁移逻辑。
- `client.go`: 下游 MCP 客户端解析与初始化逻辑。
- `http.go`: HTTP Server 路由与 SSE/Stream 处理。

## 🛠️ 常用命令 (Makefile)

- `make build`: 在 `./build/` 目录下生成可执行文件。
- `make format`: 执行代码格式化 (gofmt, tidy)。
- `make buildImage`: 构建 Docker 镜像。

> [!TIP]
> Docker 构建环境中没有 `.git` 目录，`BUILD` 变量会自动回退为 `unknown`。也可通过 `make build BUILD=<自定义版本号>` 手动指定。

## 🧪 核心逻辑扩展

### 添加新客户端类型
如果需要支持新的下游传输协议，请在 `client.go` 中扩展 `McpClientConfig` 解析逻辑，并实现相应的 MCP Client 接口。

### 修改 HTTP 聚合逻辑
如需自定义 Tools/Prompts 的聚合或过滤规则，请关注 `http.go` 中的 Handler 实现。

---

## 🐛 常见故障排查 (Troubleshooting)

### 1. `405 Method Not Allowed` — 传输协议不匹配

**症状**：
```
<service_name> Failed to add client to server: transport error: request failed with status 405: Method Not Allowed
```

**根因**：`config.json` 中下游 MCP 服务的 `transportType` 与服务端实际提供的传输协议不一致。

例如，`config.json` 配置了 `transportType: "streamable-http"`，但 MCP 服务端仅支持 SSE 协议。mcp-proxy 以 HTTP POST 方式连接 SSE 端点，SSE 端点不接受 POST 请求，返回 405。

**修复方案**：

| 修复方向 | 操作 |
|---------|------|
| 修改 proxy 配置（快速） | 将 `config.json` 中的 `transportType` 去掉或改为匹配协议 |
| 升级 MCP 服务（推荐） | 让下游服务支持 `streamable-http` 传输 |

> [!TIP]
> `streamable-http` 是 MCP 标准推荐的新协议。建议参照 `jules-mcp-server` 的做法，在下游服务中优先支持 `streamable-http`。

**关键点**：

- SSE 协议的标准端点路径为 `/sse`
- streamable-http 协议的标准端点路径为 `/mcp`

如果切换了协议，**必须同时修改 `config.json` 中的 `url` 路径**。

---

### 2. `421 Invalid Host header` — MCP SDK DNS rebinding protection

**症状**：
```
<service_name> Failed to add client to server: transport error: request failed with status 421: Invalid Host header
```

**根因**：**MCP Python SDK** 的 `TransportSecuritySettings` 默认启用 DNS rebinding protection，`allowed_hosts` 仅包含：

```python
['127.0.0.1:*', 'localhost:*', '[::1]:*']
```

当 mcp-proxy 通过 Docker 网络连接下游 Python MCP 服务时，HTTP 请求的 Host 头为 Docker 容器名（如 `douyin-mcp:8000`），不在允许列表中，MCP SDK 返回 421。

> [!WARNING]
> 这个 421 **不是** Uvicorn 或 Starlette 抛出的，而是 MCP SDK 内部的 `transport_security.py` 模块。因此 Starlette 的 `TrustedHostMiddleware` 无法解决此问题。

**修复方案**：在 `FastMCP` 构造时配置 `transport_security` 参数禁用 DNS rebinding protection：

```python
from mcp.server.transport_security import TransportSecuritySettings

mcp = FastMCP("My MCP Server",
              transport_security=TransportSecuritySettings(
                  enable_dns_rebinding_protection=False
              ))
```

或者，仅允许特定 Host：

```python
mcp = FastMCP("My MCP Server",
              transport_security=TransportSecuritySettings(
                  allowed_hosts=["127.0.0.1:*", "localhost:*", "my-service:*"]
              ))
```

> [!IMPORTANT]
> `enable_dns_rebinding_protection=False` 仅适用于 Docker 内部通信或受信任网络。对外暴露的服务应保持启用并配置具体的 `allowed_hosts`。

---

### 3. Python MCP SDK 选型对比：官方 `mcp` vs 第三方 `fastmcp`

本项目中存在两种 Python MCP 实现方式，以下是完整对比：

| 维度 | 官方 `mcp` SDK (`mcp.server.fastmcp`) | 第三方 `fastmcp` 包 |
|------|---------------------------------------|---------------------|
| **维护方** | Anthropic 官方 | 社区（[gofastmcp.com](https://gofastmcp.com)） |
| **协议跟进** | 协议更新零延迟 | 社区跟进，有延迟风险 |
| **性能** | 底层都是 Starlette + Uvicorn，无差异 | 同左 |
| **安全特性** | 内置 DNS rebinding 保护（Docker 需配置） | 无内置保护，需自行保障 |
| **ASGI app** | `streamable_http_app()` | `http_app(transport="streamable-http")` |
| **本项目使用** | `douyin-mcp`（自研服务） | `jules-mcp-server`（外部引入） |

**项目决策**：

- **新建 Python MCP 服务统一使用官方 `mcp` SDK**（`from mcp.server.fastmcp import FastMCP`）
- `jules-mcp-server` 作为外部引入项目，保留第三方 `fastmcp`，不做迁移
- Docker 环境下需配置 `transport_security`（参见上方 421 章节）

> [!NOTE]
> 官方 SDK 作为 Anthropic 第一方实现，在 MCP 协议快速演进阶段具备最佳的兼容性保证。第三方 `fastmcp` 虽然 API 更简洁，但引入了额外的依赖维护风险。

---

### 4. `invalid character '\x1f'` — Gzip 压缩响应未解码

**症状**：
```
transport error: failed to decode response: invalid character '\x1f' looking for beginning of value
```

**根因**：上游 MCP 服务（如 Google Stitch API）返回 Gzip 压缩的 HTTP 响应，而 `mcp-go` 的 `StreamableHttpClient` 未自动解压，导致 JSON 解析器尝试解析二进制 Gzip 数据（`\x1f` 是 Gzip 魔数的首字节）。

**修复方案**（已内置）：

`client.go` 中实现了 `gzipDecompressor`（自定义 `http.RoundTripper`），在 HTTP 传输层透明解压 Gzip 响应。该解压器通过 `transport.WithHTTPBasicClient` 注入到 `streamable-http` 客户端中。

如果您在新增的 `streamable-http` 下游服务中仍遇到此问题，请检查自定义 `http.Client` 是否正确注入。

---

> [!NOTE]
> 提交代码前，请确保运行 `make format` 以保持代码风格一致。
