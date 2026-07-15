---
name: infra-rules-skills-sync
description: 规则与技能多端同步 / Rules & Skills Sync across AI assistants (Cursor/Lingma/OpenCode/Claude Code/Trae)。自动发现并同步 .agents/rules/ 和 .agents/skills/ 中的变更到各 AI 编辑器的对应配置目录。适用场景：新增/修改了 .agents/rules/ 中的规则后需要同步到 Cursor/Lingma/Claude/Trae、新增/修改了 .agents/skills/ 中的技能后需要同步到 OpenCode/Claude Subagents、跨平台配置一致性审计、规则文件整理与清理。关键词：规则同步/技能同步/多端同步/智能体配置/Cursor/Lingma/Claude/Trae/OpenCode/Subagent/rule sync/skill sync/cross-platform/config consistency/agent config。
---

# 规则与技能多端同步技能 (Rules & Skills Sync)

## 场景

当你在 `.agents/rules/` 中新增/修改了规则文档，或者在 `.agents/skills/` 中新增/修改了技能，需要将这些变更同步对齐到各个 AI 辅助工具（包括 Cursor、通义灵码、OpenCode、Claude Code 等）的对应配置中。

## 同步机制

### 1. 规则同步 (Rules)

- **源文件**：`.agents/rules/*.md`（Single Source of Truth 事实唯一来源）
- **同步终端**：
  - **Cursor**：转换为 `.cursor/rules/*.mdc` (MDC 格式)。依据规则名分类分配 `globs` 文件类型。源 `trigger: always_on` → `alwaysApply: true`，`trigger: model_decision` → `alwaysApply: false`。
  - **通义灵码**：以原有 Markdown 格式直接同步到 `.lingma/rules/*.md`。
  - **Claude Code**：以原有 Markdown 格式直接同步到 `.claude/rules/*.md`。
  - **Trae**：转换为 `.trae/rules/*.md`（Trae 前端格式）。源 `trigger: always_on` → `alwaysApply: true`，其他 → `alwaysApply: false`。仅当 `alwaysApply: false` 时才设置 `description` 和 `globs` 字段（`globs` 根据规则名前缀自动匹配）。

### 2. 技能同步 (Skills)

- **源文件夹**：`.agents/skills/<name>/`
- **同步终端**：
  - **OpenCode Skills**：直接复制完整技能目录到 `.opencode/skills/<name>/`。
  - **Claude Code Skills**：直接复制完整技能目录到 `.claude/skills/<name>/`。
  - **Claude Code Subagents**：将技能合并为 14 个分组子代理（backend-architect, backend-asr, frontend-ui, code-reviewer, issue-manager, task-manager, infra-ops, agent-config, doc-exporter, workflow-planner, research-analyst, qa-engineer, speckit-manager, performance-ops），生成 `.claude/subagents/<group>.md` 和 `.claude/subagents/<group>.json`。
  - **Mimocode Skills**：直接复制完整技能目录到 `.mimocode/skills/<name>/`。
  - **Trae Skills**：直接复制完整技能目录到 `.trae/skills/<name>/`。

### 3. MCP 服务器配置同步 (MCP Servers)

- **SSOT**: `~/.gemini/config/mcp_config.json`（用户级全局唯一事实源）
- **同步终端**：
  - `.claude/settings.json` — 顶层 `mcpServers` 段，`serverUrl` → `url` 转换
  - `.gemini/settings.json` — 顶层 `mcpServers` 段，原样写入
  - `.cursor/mcp.json` — 顶层 `mcpServers` 段，`serverUrl` → `url` 转换
  - `.mcp.json` — 顶层 `mcpServers` 段，`serverUrl` → `url` 转换
  - `opencode.json` — 顶层 `mcp` 段，远程 → `{"type": "remote", "url": ...}`，本地 → `{"type": "local", "command": [...]}`
- **同步策略**：仅替换 MCP 配置段，保留目标文件其他字段不变。
- **格式适配**：
  - `mcpServers` 格式（Gemini/Claude/Cursor/.mcp.json）：`serverUrl` → `url`，保留 `command`/`args` 等 stdio 字段
  - `mcp` 格式（OpenCode）：远程服务器 → `type: "remote"` + `url`，本地服务器 → `type: "local"` + `command[]`

### 4. 智能体同步 (Agents)

- **源文件夹**：`.agents/agents/`（JSON 格式，Gemini/Antigravity 智能体定义）
- **同步终端**：
  - **Gemini Agents**：直接复制到 `.gemini/agents/`。支持自动清理已失效的智能体文件。
- **格式要求**：源文件必须是带 `name`、`displayName`、`description`、`customAgentSpec` 等字段的标准 JSON 格式，详见 `.agents/rules/infra-ai-assistant-rules-standards.md §6`。

---

## Execution Script (执行脚本)

此技能的执行由以下自动化同步脚本驱动：

- **入口**: `.agents/skills/infra-rules-skills-sync/execute.py`
- **调用示例**:

  ```bash
  python3 .agents/skills/infra-rules-skills-sync/execute.py --input '{"force_refresh": true}'
  ```text

- **输出格式**: 标准化的 JSON 结果，包含 `status`, `message`, `data` 字段。

### 输入参数

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `force_refresh` | bool | — | 兼容旧参数，保留但暂未使用 |
| `auto_stage` | bool | `true` | 同步后是否自动 `git add` / `git rm` 变更的目标文件（不含源文件） |
| `run_audit_after` | bool | `true` | 同步后是否调用审计脚本验证规则同步一致性 |

### 同步后自动 git 暂存 (Auto Stage)

为减少手工 `git add` 的繁琐，同步成功后脚本会**自动将变更的目标文件加入暂存区**：

- **目标范围**（仅限以下目录，避免污染源文件）：
  - `.cursor/rules/`、`.lingma/rules/`、`.claude/rules/`、`.trae/rules/`
  - `.opencode/skills/`、`.opencode/commands/`、`.mimocode/skills/`、`.mimocode/commands/`
  - `.claude/skills/`、`.claude/subagents/`、`.trae/skills/`
  - `.gemini/agents/`
  - `.claude/settings.json`、`.gemini/settings.json`、`.cursor/mcp.json`、`.mcp.json`、`opencode.json`（MCP 同步目标）
- **不会自动暂存**：`.agents/rules/`、`.agents/skills/`、`.agents/agents/`（源文件由开发者手工管理）
- **删除处理**：若目标文件在同步过程中被清理（如失效规则），使用 `git rm` 暂存删除
- **非 git 环境**：自动跳过 auto-stage（不影响同步本身）
- **禁用方式**：传参 `{"auto_stage": false}` 即可禁用

### 同步后审计 (Audit)

审计脚本路径已修正为：

```text
.agents/skills/agent-rules-auditor/scripts/audit_rules.py
```text
（早期版本引用的是 `agent-rules-auditor/...`，与技能重命名后的真实路径不一致；现已对齐。）
该脚本以 JSON 格式输出 `pass` / `fail` 状态与差异列表，集成在 `run_audit()` 中。
可通过 `{"run_audit_after": false}` 禁用。

## 核心约束与执行标准

### 严格边界与禁止项
- **安全 > 质量 > 速度**：任何全栈决策和代码修改必须遵循此总原则。
- **禁止项**：严禁在未经用户确认的情况下删除核心业务规则。
- **禁止项**：严禁输出虚假或未经测试的配置。

## 思维链引导

在执行复杂步骤、架构设计或遇到歧义时，必须通过 `<thought>` 标签显式输出你的思考过程和逻辑推导，然后再输出最终行动或代码方案。
