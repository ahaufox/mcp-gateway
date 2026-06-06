---
trigger: always_on
description: Cursor, Claude, Lingma, OpenCode AI编码助手的配置文件目录结构与配置格式规范。
---

# AI 助手规则与配置目录规范 (AI Assistant Rules & Configurations Specification)

## 1. Cursor
- **目录结构**：项目根目录下的 `.cursor/rules/` 文件夹。
- **规则与格式**：采用 MDC (Markdown Cursor) 模式。
  - 文件必须以 `.mdc` 作为后缀。
  - 文件顶部必须包含 YAML Frontmatter 元数据块，用于定义触发条件，其下方为 Markdown 格式的系统指令。支持通过 `globs` 进行精准的路径级按需加载。
  - **Frontmatter 字段要求**：
    - `description`: 简短明了的规则作用描述。
    - `globs`: 触发匹配的文件路径通配符数组（例如 `["backend/**/*"]` 或 `["frontend-v2/**/*", "admin-frontend/**/*"]`）。
    - `alwaysApply`: 布尔值，是否在所有对话中强制挂载（建议全栈基础规范设为 `true`）。
- **示例**：
  ```markdown
  ---
  alwaysApply: true
  description: 针对后端 Python/FastAPI 开发、设计与测试规范
  globs: ["backend/**/*"]
  ---
  (此处为规则具体 Markdown 内容)
  ```

---

## 2. Claude Code (Anthropic CLI)
- **目录结构**：项目根目录下的 `.claude/`。
- **规则与格式**：采用模块化分层配置，按目录自动识别并注入不同的能力域：
  - **根文件 (`CLAUDE.md`)**：每次会话必读的核心项目指令、常驻内存。用于简述项目基础技术栈及核心标准引用。
  - **规则文件目录 (`.claude/rules/`)**：存放 `.md` 后缀的局部规则文件，基于路径或对话意图条件触发加载，避免上下文浪费。
  - **智能体文件目录 (`.claude/subagents/`)**：存放具备独立 System Prompt 和工具权限的子智能体定义。
    - 存放文件类型包括 `.md`（编写 System Prompt 及人设指令）和 `.json`（智能体独立工具链权限与配置）。
  - **主配置文件 (`.claude/settings.json`)**：严格管理全局工具调用权限（如文件读写、Bash 命令的 `allow`/`ask`/`deny` 机制）。

---

## 3. 通义灵码 (Lingma)
- **目录结构**：项目根目录下的 `.lingma/rules/` 文件夹。
- **规则与格式**：基于文本的编码规范集合，存放 `.md` 或纯文本文件：
  - **Global (全局生效)**：适用于全项目的注释规约、通用安全底线等。
  - **Conditional (条件生效)**：通过匹配语言、框架或文件后缀自动决策挂载。

---

## 4. OpenCode / 开源多智能体框架
- **目录结构**：项目根目录下的 `.opencode/agents/` 或 `.opencode/workflows/` 文件夹。
- **规则与格式**：基于“角色-能力-工作流”的 MAS（多智能体系统）定义，侧重任务调度和角色分工。存放 `.yaml`、`.json` 及配合的 `.md` 规则：
  - **智能体描述 (`.opencode/agents/*.yaml`)**：配置智能体的人设与可用工具。必须包含如下字段：
    - `role` / `description`: 智能体名称与人设定位（如 Architect, Reviewer）。
    - `system_prompt`: 核心逻辑链指令，指示其需遵循的规则路径。
    - `tools_allowed`: 数组，显式声明可用的系统级工具（如 `read_file`, `git_diff`, `write_file` 等）。

---