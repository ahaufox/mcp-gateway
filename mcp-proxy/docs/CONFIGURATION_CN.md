# 配置说明

本项目支持 v2 JSON 配置。v1 版本的配置在加载时会自动迁移。

配置可以通过两种方式提供：

- **单文件模式** (`--config <文件>` 或 `<URL>`)：经典 `config.json` 形态，保持向后兼容。
- **拆分目录模式** (`--config-dir <目录>`)：扫描目录下的所有 `*.json` 文件并在启动时合并。MCP 服务较多时推荐使用，详见 [拆分配置](#拆分配置)。

未显式指定时，程序会自动检测：若 `./configs` 目录存在则使用拆分模式，否则回退到 `./config.json`。

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

## 拆分配置

当 MCP 服务数量较多时，单一 `config.json` 越来越难维护。`--config-dir` 模式让你按传输类型（也可以按你习惯的任意维度）把配置拆到多个文件中。

### 目录结构

```
configs/
├── base.json                 # 代理服务自身配置 (mcpProxy 块)
├── categories/
│   ├── stdio.json            # stdio (子进程) 服务
│   ├── sse.json              # SSE 服务
│   ├── streamable-http.json  # streamable-http 服务
│   └── websocket.json        # websocket 服务 (空占位)
└── overrides/                # 本地/个人覆盖 (已加入 .gitignore)
    └── *.json
```

每个文件都是一份**部分配置** (`PartialConfig`)：所有顶层字段 (`mcpProxy`、`mcpServers`) 均为可选。加载顺序为文件相对路径的字典序：

1. `base.json` 优先加载 (提供 `mcpProxy`)。
2. 然后是 `categories/*` (提供 `mcpServers`)。
3. 最后是 `overrides/*`，冲突时后者获胜。

由于 `base` < `categories` < `overrides` 的字典序天然成立，无需额外配置就能得到期望的优先级。允许新增子目录，新子目录会自然落到合并顺序的某一档。

### 单文件形态

`base.json` 通常只设置 `mcpProxy`：

```json
{
  "mcpProxy": {
    "baseURL": "${MCP_BASE_URL}",
    "addr": ":9090",
    "name": "MCP Proxy",
    "version": "1.0.0",
    "type": "streamable-http",
    "options": {
      "panicIfInvalid": false,
      "logEnabled": true,
      "authTokens": ["${AUTH_TOKENS}"]
    }
  }
}
```

`categories/stdio.json` 之类的文件只设置 `mcpServers`：

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": { "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PERSONAL_ACCESS_TOKEN}" },
      "options": {
        "toolFilter": {
          "mode": "block",
          "list": ["create_or_update_file", "create_repository"]
        }
      }
    }
  }
}
```

### 覆盖

`configs/overrides/*.json` 已被 `.gitignore` 排除。把 `overrides/local.json` 这样的文件丢进去即可在某个具体部署上覆盖个别服务，而无需把密钥或个人偏好提交到仓库。后加载的覆盖文件获胜；`overrides/` 的字典序在 `categories/` 之后，所以覆盖始终生效。

### 合并规则

- `mcpProxy`：被最后一个设置它的文件整体替换。
- `mcpServers`：按服务名合并，同名条目后者替换前者。
- 允许出现 `"mcpServers": {}` 这种空对象，视为空操作。
- 仅有 `_comment` 等被忽略字段的文件视为空操作。
- 跳过以 `.` 开头的隐藏文件 (例如 `.DS_Store`) 和非 `.json` 文件。
- 目录必须至少包含一个 `*.json` 文件，否则加载失败。
- 必须有文件提供非空的 `mcpProxy` 块，否则报错并提示应在 `base.json` 中设置。

合并完成后，得到的配置会走与单文件模式完全相同的后处理流程：Trae 兼容、逗号分隔 Token 拆分、Options 继承、toolFilter 应用等行为保持一致。

### 在拆分模式下运行

```bash
# 显式指定目录
mcp-proxy --config-dir ./configs

# 自动检测 (./configs 存在则用拆分模式，否则用 ./config.json)
mcp-proxy

# Docker (默认 CMD 即拆分模式)
docker run -v $(pwd)/mcp-proxy/configs:/config/configs \
  -v $(pwd)/mcp-proxy/.env:/config/.env \
  -p 9090:9090 ghcr.io/tbxark/mcp-proxy:latest
```

`--config-dir` 模式下，前端的 `/api/config` 接口返回内存中的合并视图，能看到生效的最终配置，而不是某一个分片。
