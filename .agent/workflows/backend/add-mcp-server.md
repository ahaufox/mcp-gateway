---
description: How to safely add a new MCP Server to the Gateway
---

# Add MCP Server Workflow

## 1. Preparation
1. Identify the MCP Server type (`stdio`, `sse`, `streamable-http`).
2. Get the connection details (Command+Args for stdio, URL for SSE).
3. **Check for Secrets**: If the server requires API keys, ensure they are in `.env` or exported variables.

## 2. Configuration
1. Backup existing config:
   ```bash
   cp mcp-proxy/config.json mcp-proxy/config.json.bak
   ```
2. Edit `mcp-proxy/config.json`:
   - Add new entry under `mcpServers`.
   - Use `${VAR}` for secrets.

   **Example (Stdio):**
   ```json
   "my-server": {
     "command": "npx",
     "args": ["-y", "@modelcontextprotocol/server-xyz"],
     "env": { "API_KEY": "${XYZ_API_KEY}" }
   }
   ```

   **Example (SSE):**
   ```json
   "remote-server": {
     "url": "http://localhost:3000/sse"
   }
   ```

## 3. Validation
1. Verify JSON syntax (use an editor or `jq`).
   ```bash
   # If config has comments (JSONC), use a tool that supports it, or just visual check.
   ```

## 4. Application
1. Restart the Gateway:
   - If binary: `pkill mcp-gateway && ./mcp-proxy/mcp-gateway &`
   - If Docker: `docker restart mcp-gateway`

## 5. Verification
1. Access the Web UI: `http://localhost:8080` (or configured port).
2. Check if the new server appears in the list and is "Connected".
3. Check logs for errors.
