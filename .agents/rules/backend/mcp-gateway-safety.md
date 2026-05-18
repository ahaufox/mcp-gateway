---
description: Rules for maintaining and configuring the MCP Gateway securely
---

# MCP Gateway Safety Rules

## 1. Security (P0)
- **NO Plain Text Secrets**: Never commit `config.json` containing real API keys.
  - Use environment variables expansion: `${ENV_VAR}`.
  - Or use a separate `secrets.json` if supported (currently `config.json` supports env expansion).
- **Gitignore**: Ensure `config.json` is gitignored if it must contain secrets, or use a template `config.json.example`.

## 2. Configuration Integrity (P1)
- **Validation**: Always validate `config.json` syntax (JSONC) before restarting the service.
- **Backup**: Create a backup (e.g., `cp config.json config.json.bak`) before any manual edit.

## 3. Deployment
- **Restart Required**: Configuration changes require a process restart (`systemctl restart mcp-gateway` or `docker restart mcp-gateway`).
