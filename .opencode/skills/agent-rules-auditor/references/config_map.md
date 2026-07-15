# Agent 配置映射表 (Agent Configuration Map)

本文件定义了项目中各种 AI 驱动的编辑器与工具所对应的配置文件及规则目录映射关系。

## 编辑器 / 工具映射

| 工具 | 规则目录 | 文件扩展名 | 特殊要求 |
| :--- | :--- | :--- | :--- |
| **权威源 (SSOT)** | `.agents/rules/` | `.md` | 唯一的规则事实来源 |
| **Cursor** | `.cursor/rules/` | `.mdc` | 需要包含 `globs` 和 `alwaysApply` 的 frontmatter。 |
| **Windsurf** | `.windsurf/rules/` | `.md` | 与权威源结构相似。 |
| **Claude Dev / Desktop** | `.claude/rules/` | `.md` | 与权威源结构相似。 |
| **Opencode** | `.agents/rules/` | `.md` | 通过根目录的 `opencode.json` 直接配置引用。 |
| **Opencode Skills** | `.opencode/skills/` | `SKILL.md` | 每个技能文件夹必须包含一个 `SKILL.md`。 |
| **Gemini (项目)** | `.` | `GEMINI.md` | 项目共享规则，**需提交 Git**。 |
| **Gemini (个人项目)** | `~/.gemini/tmp/<proj>/memory/` | `MEMORY.md` | 私有项目记忆，记录本地环境/私有笔记。 |
| **Gemini (全局个人)** | `~/.gemini/` | `GEMINI.md` | 跨项目个人偏好，不随项目变动。 |
| **Trae** | `.trae/rules/` | `.md` | (如果适用) |
| **Cline / Roo Code** | 根目录 | `.clinerules*` | 单个规则文件或模式。 |
| **Aider** | 根目录 | `.aider.instructions.md` | 单个指令文件。 |

## 同步逻辑

1. **存在性验证**: `.agents/rules/` 中除 `README.md` 外的每个 `.md` 文件，都应在目标目录 (Cursor, Windsurf, Claude) 中有对应的文件。
   - **Opencode 特例**: Opencode 的指令直接指向 `.agents/rules/*.md`，因此不需要执行物理同步。
1. **内容一致性**: 核心 Markdown 内容（标题、列表、文本）必须保持完全一致。
1. **格式转换**:
   - 对于 `.mdc` (Cursor)，必须保留 YAML frontmatter，并根据需要增强 `globs` 字段。
   - 对于其他 `.md` 目标，应保留其原有的 frontmatter 描述信息。

## 配置文件

- **Opencode**: `opencode.json` (根目录) - 用于配置指令路径 (`instructions`) 和 MCP 服务。
- **Gemini (项目设置)**: `.gemini/settings.json` - 存储项目特定的模型、编辑器和工具审批偏好。
- **Gemini (全局设置)**: `~/.gemini/settings.json` - 跨工作区的全局技术参数配置。
- **Gemini (系统级)**: `/etc/gemini-cli/settings.json` - 企业级强制约束（Linux）。
