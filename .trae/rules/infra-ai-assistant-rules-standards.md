---
alwaysApply: true
---

# AI 助手规则与配置目录规范

## 1. 事实来源与同步顺序
- **唯一主规则库**：`.agents/rules/` 是智能体规则的主要事实来源。新增或更新全局规范、前后端规范、智能体能力治理规范时，必须先更新这里。
- **镜像目录原则**：`.cursor/`、`.lingma/`、`.trae/`、`.opencode/`、`.claude/` 只做必要格式适配，禁止创建与 `.agents/rules/` 相互矛盾的第二套规则。
- **任务前置阅读**：开发或修复前，必须先遵循全局标准的查阅要求。

## 2. Cursor 规则 (光标规则)
- **目录结构**：项目根目录下的 `.cursor/rules/` 文件夹。
- **规则与格式**：采用 MDC (Markdown Cursor) 模式。
  - 文件必须以 `.mdc` 作为后缀。
  - 文件顶部必须包含 YAML Frontmatter 元数据块，用于定义触发条件，其下方为 Markdown 格式的系统指令。支持通过 `globs` 进行精准的路径级按需加载。
  - **Frontmatter 字段要求**：
    - `description`: 简短明了的规则作用描述。
    - `globs`: 触发匹配的文件路径通配符数组。
    - `alwaysApply`: 布尔值，是否在所有对话中强制挂载。

---

## 3. Claude Code
- **目录结构**：项目根目录下的 `.claude/`。
- **规则与格式**：
  - **根文件 (`CLAUDE.md`)**：每次会话必读的核心项目指令、常驻内存。用于简述项目基础技术栈及核心标准引用。
  - **规则文件目录 (`.claude/rules/`)**：存放 `.md` 后缀的局部规则文件，基于路径或对话意图条件触发加载。
  - **工具调用子代理目录 (`.claude/subagents/`)**：存放通过 Agent tool 程序化调用的子代理定义。
  - **技能目录 (`.claude/skills/`)**：存放技能定义，每个技能一个子目录，包含 `SKILL.md`。

---

## 4. 通义灵码
- **目录结构**：项目根目录下的 `.lingma/rules/` 文件夹。
- **规则与格式**：基于文本的编码规范集合，存放 `.md` 或纯文本文件：
  - **Global (全局生效)**：适用于全项目的注释规约、通用安全底线等。
  - **Conditional (条件生效)**：通过匹配语言、框架或文件后缀自动决策挂载。

---

## 5. 开源代码代理 (OpenCode)
- **目录结构**：项目根目录下的 `.opencode/agents/` 或 `.opencode/workflows/` 文件夹。
- **规则与格式**：基于“角色-能力-工作流”的多智能体系统定义。
  - **智能体描述 (`.opencode/agents/*.md`)**：配置智能体的人设与可用工具。Frontmatter 必须包含 `description`、`mode: subagent`，并显式声明 `tools` 参数。
  - **权限约束**：审查类代理默认只读；确需写入、Bash 或网络能力时必须显式设置。

---

## 6. Gemini 智能体 (Antigravity/Gemini CLI)
- **目录结构**：
  - 源头配置目录：`.agents/agents/` 文件夹（如果存在）。
  - 分发目标目录：项目根目录下的 `.gemini/agents/` 文件夹。
- **规则与格式**：适用于 Gemini / Antigravity 编译器的专业智能体配置。
  - 必须使用 `.json` 格式定义。
  - **JSON 字段结构要求**：包含 `name`、`displayName`、`description`、`hidden`、`customAgentSpec`（内部包含 `systemPromptSections` 以及 `toolNames` 数组）。
  - **工具列表要求**：必须且仅可使用当前编辑器支持的原生工具名：
    - `grep_search`, `list_dir`, `view_file`, `write_to_file`, `replace_file_content`, `multi_replace_file_content`, `run_command`, `manage_task`, `search_web`, `read_url_content`, `invoke_subagent`, `send_message`。
  - **同步要求**：严禁直接在目标目录 `.gemini/agents/` 下手工修改，所有变更必须在源头进行并进行同步分发。