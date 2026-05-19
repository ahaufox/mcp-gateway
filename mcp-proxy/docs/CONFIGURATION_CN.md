# 配置说明

本项目支持 v2 JSON 配置。v1 版本的配置在加载时会自动迁移。

## 完整示例

```jsonc
{
  "mcpProxy": {
    "baseURL": "https://mcp.example.com",
    "addr": ":9090",
    "name": "MCP Proxy",
    "version": "1.0.0",
    "type": "streamable-http", // 或 "sse" (默认)
    "options": {
      "panicIfInvalid": false,
      "logEnabled": true,
      "authTokens": ["DefaultToken"]
    }
  },
  "mcpServers": {
    "github": {
      // stdio 客户端
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "<YOUR_TOKEN>" },
      "options": {
        "toolFilter": {
          "mode": "block",
          "list": ["create_or_update_file"]
        }
      }
    },
    "fetch": {
      // stdio 客户端
      "command": "uvx",
      "args": ["mcp-server-fetch"],
      "options": {
        "panicIfInvalid": true,
        "logEnabled": false,
        "authTokens": ["SpecificToken"]
      }
    },
    "amap": {
      // SSE 客户端
      "url": "https://mcp.amap.com/sse?key=<YOUR_TOKEN>",
      "options": {
        "disabled": true
      }
    }
  }
}
```

## mcpProxy (代理配置)

- `baseURL`: 用于构建客户端端点的公共 URL 基础路径。
- `addr`: 绑定地址（例如 `:9090`）。
- `name`, `version`: 用于 MCP 握手的服务器标识。
- `type`: `sse` (默认) 或 `streamable-http`。
- `options`: 被 `mcpServers.*.options` 继承的默认选项（可以针对每个服务器进行覆盖）。

## mcpServers (MCP 服务器配置)

每一项定义了一个下游 MCP 服务器。支持的客户端类型：

- `stdio` (当设置了 `command` 时隐式使用，或指定 `transportType: "stdio"`): 通过标准输入输出运行子进程。
- `sse` (当设置了 `url` 且 `transportType` ≠ `streamable-http` 时隐式使用，或指定 `transportType: "sse"`): 通过服务器发送事件 (Server‑Sent Events) 连接。
- `streamable-http` (需要 `transportType: "streamable-http"`): 通过 HTTP 流连接。

常用字段：

- `transportType` — 显式指定客户端传输类型（`"stdio"`, `"sse"`, 或 `"streamable-http"`）。省略时根据 `command`（stdio）或 `url`（默认 SSE）自动推断。
- `description` — 服务器的可读描述，显示在仪表板上。
- `command`, `args`, `env` — 用于 `stdio` 客户端。
- `url`, `headers` — 用于 `sse` 和 `streamable-http` 客户端。
- `timeout` — `streamable-http` 的请求超时时间。
- `options` — 每个服务器的覆盖选项和过滤器（见下文）。

## options (选项)

- `panicIfInvalid` (bool): 如果为 true，当客户端无法初始化时，程序启动将失败。
- `logEnabled` (bool): 为此客户端记录请求和事件日志。
- `authTokens` ([]string): 有效的 Bearer Token；请求必须包含 `Authorization: <token>`。
- `toolFilter` (object): 选择性地向代理暴露工具：
  - `mode`: `allow` (允许列表) 或 `block` (黑名单)。
  - `list`: 工具名称列表。
- `disabled` (bool): 启用或禁用此服务器。禁用的服务器在启动时会被跳过。
- `disablePing` (bool): 禁用 SSE/streamable-http 客户端的定期 Ping 健康检查。适用于不支持 Ping 的服务器。
- `maintenanceInterval` (duration): Ping/重连尝试的间隔时间（默认 `30s`）。支持 Go 持续时间格式（例如 `15s`, `1m`）。

注意：

- 如果服务器省略了 `options.authTokens`，则 `mcpProxy.options.authTokens` 将作为默认的 Token 集合。
- 若要发现用于过滤的工具名称，可以先在不带过滤器的情况下启动，并在日志中查看类似 `<server> Adding tool <name>` 的行。

## 环境变量 (Environment Variables)

配置文件中的值支持使用 `${VAR_NAME}` 格式引用环境变量。

- **字符串字段**: 直接替换为环境变量的值。例如 `"baseURL": "${MCP_BASE_URL}"`。
- **数组字段 (如 authTokens)**: 如果环境变量包含逗号分隔的字符串（例如 `TOKEN1,TOKEN2`），会自动分割为数组。
  ```json
  "authTokens": ["${AUTH_TOKENS}"]
  ```
