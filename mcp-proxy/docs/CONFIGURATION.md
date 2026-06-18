# Configuration

This project supports a v2 JSON configuration. v1 configs are automatically migrated at load time.

The configuration can be supplied in two ways:

- **Single file** (`--config <file>` or `<url>`): the classic `config.json` shape. Backwards compatible.
- **Split directory** (`--config-dir <dir>`): scan a directory of `*.json` files, merge them at startup. Recommended for installations with many MCP services — see [Split configuration](#split-configuration) below.

When neither flag is given, the binary auto-detects: if `./configs` exists it is used as the split directory, otherwise `./config.json` is loaded.

- Online converter (build Claude config from your proxy): https://tbxark.github.io/mcp-proxy

## Full Example

```jsonc
{
  "mcpProxy": {
    "baseURL": "https://mcp.example.com",
    "addr": ":9090",
    "name": "MCP Proxy",
    "version": "1.0.0",
    "type": "streamable-http", // or "sse" (default)
    "options": {
      "panicIfInvalid": false,
      "logEnabled": true,
      "authTokens": ["DefaultToken"]
    }
  },
  "mcpServers": {
    "github": {
      // stdio client
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
      // stdio client
      "command": "uvx",
      "args": ["mcp-server-fetch"],
      "options": {
        "panicIfInvalid": true,
        "logEnabled": false,
        "authTokens": ["SpecificToken"]
      }
    },
    "amap": {
      // SSE client
      "url": "https://mcp.amap.com/sse?key=<YOUR_TOKEN>",
      "options": {
        "disabled": true
      }
    }
  }
}
```

## mcpProxy

- `baseURL`: Public URL base used to build client endpoints.
- `addr`: Bind address (e.g. `:9090`).
- `name`, `version`: Server identity for MCP handshake.
- `type`: `sse` (default) or `streamable-http`.
- `options`: Defaults inherited by `mcpServers.*.options` (can be overridden per server).

## mcpServers

Each entry defines a downstream MCP server. Supported client types:

- `stdio` (implicit when `command` is set, or `transportType: "stdio"`): run a subprocess via stdio.
- `sse` (implicit when `url` is set and `transportType` ≠ `streamable-http`, or `transportType: "sse"`): connect via Server‑Sent Events.
- `streamable-http` (requires `transportType: "streamable-http"`): connect via HTTP streaming.

Common fields:

- `transportType` — explicitly specify the client transport type (`"stdio"`, `"sse"`, or `"streamable-http"`). When omitted, it is inferred from `command` (stdio) or `url` (SSE by default).
- `description` — human-readable description of the server, displayed on the dashboard.
- `command`, `args`, `env` — for `stdio` clients.
- `url`, `headers` — for `sse` and `streamable-http` clients.
- `timeout` — request timeout for `streamable-http`.
- `options` — per‑server overrides and filters (see below).

## options

- `panicIfInvalid` (bool): If true, startup fails when a client cannot initialize.
- `logEnabled` (bool): Log requests and events for this client.
- `authTokens` ([]string): Valid bearer tokens; requests must include `Authorization: <token>`.
- `toolFilter` (object): Selectively expose tools to the proxy:
  - `mode`: `allow` or `block`.
  - `list`: List of tool names.
- `disabled` (bool): Enable or disable this server. Disabled servers are skipped at startup.
- `disablePing` (bool): Disable periodic ping health checks for SSE/streamable-http clients. Useful for servers that do not support ping.
- `maintenanceInterval` (duration): Interval between ping/reconnect attempts (default `30s`). Accepts Go duration format (e.g. `15s`, `1m`).

Notes:

- `mcpProxy.options.authTokens` serves as the default token set if a server omits `options.authTokens`.
- To discover tool names for filtering, start without a filter and check logs for lines like `<server> Adding tool <name>`.

## Environment Variables

Values in the config file support `${VAR_NAME}` syntax to reference environment variables.

- **String fields**: Directly replaced with the environment variable value. E.g. `"baseURL": "${MCP_BASE_URL}"`.
- **Array fields (e.g. authTokens)**: If the environment variable contains comma-separated values (e.g. `TOKEN1,TOKEN2`), it will be automatically split into an array.
  ```json
  "authTokens": ["${AUTH_TOKENS}"]
  ```

## Split configuration

When you have more than a handful of MCP services it gets painful to manage a single `config.json`. The `--config-dir` mode lets you split the configuration across multiple files organized by transport type (or any other grouping you prefer).

### Layout

```
configs/
├── base.json                 # proxy server settings (mcpProxy block)
├── categories/
│   ├── stdio.json            # stdio (subprocess) services
│   ├── sse.json              # SSE services
│   ├── streamable-http.json  # streamable-http services
│   └── websocket.json        # websocket services (empty placeholder)
└── overrides/                # local-only / personal overrides (gitignored)
    └── *.json
```

Each file is a *partial* configuration: every top-level field (`mcpProxy`, `mcpServers`) is optional. Files are merged in lexical order of their relative path, so:

1. `base.json` is loaded first (provides `mcpProxy`).
2. Files in `categories/*` are loaded next (provide `mcpServers`).
3. Files in `overrides/*` are loaded last and win on conflicts.

Because `base` < `categories` < `overrides` lexically, you get the conventional precedence for free. Adding a new subdirectory is allowed and gets its own slot in the merge order.

### Per-file shape

`base.json` typically only sets `mcpProxy`:

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

`categories/stdio.json` and friends only set `mcpServers`:

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

### Overrides

`configs/overrides/*.json` is git-ignored. Drop a file like `overrides/local.json` there to override individual services for a particular checkout without committing secrets or personal preferences. Later files win, and `overrides/` sorts after `categories/`, so any override always beats a category file.

### Merging rules

- `mcpProxy` — replaced by whichever file sets it last.
- `mcpServers` — merged by server name; later entries replace earlier ones with the same name.
- An empty category file (`"mcpServers": {}`) is allowed and acts as a no-op.
- A file that contains only comments (`_comment`) is treated as a no-op.
- Files whose basename starts with `.` (e.g. `.DS_Store`) and non-`.json` files are skipped.
- The directory must contain at least one `*.json` file; otherwise loading fails.
- Exactly one file must provide a non-nil `mcpProxy` block; otherwise loading fails with a clear error pointing to `base.json`.

After merging, the resulting config goes through the same post-processing as the single-file path: Trae compatibility, comma-separated token splitting, option inheritance, and tool-filter logic are all applied unchanged.

### Running in split mode

```bash
# explicit path
mcp-proxy --config-dir ./configs

# auto-detect (uses ./configs if it exists, else ./config.json)
mcp-proxy

# Docker (default in the published image)
docker run -v $(pwd)/mcp-proxy/configs:/config/configs \
  -v $(pwd)/mcp-proxy/.env:/config/.env \
  -p 9090:9090 ghcr.io/tbxark/mcp-proxy:latest
```

The dashboard's `/api/config` endpoint serves the in-memory merged view when `--config-dir` is used, so the UI shows the effective configuration rather than a single file fragment.

