---
description: README 更新技能 / README Update Skill。根据代码/配置/架构变更类型，精准定位需要更新的 README 文件和具体章节，避免无差别全量扫描。适用场景：启动方式变更、依赖/技术栈变更、目录结构调整、模块架构重构、API 同步流程变更、Agent 配置变更、测试配置变更、新增/废弃核心模块、README 内容与代码事实不一致。关键词：README更新/文档同步/启动命令更新/架构文档同步/README同步/readme update/documentation sync/readme sync/architecture docs/startup command update。
---

# README 更新技能 (README Update Skill)

## 范围

### 覆盖的 README（项目自有）

| 层级 | 文件 | 内容类型 |
|------|------|---------|
| 入口配置 | `AGENTS.md` | AI 助手速查手册 |
| 入口配置 | `CLAUDE.md` | Claude Code 工作指引 |
| 规则索引 | `.agents/rules/README.md` | 规则文件索引与分类 |
| 项目概览 | `README.md` | 项目根入口 |
| 后端指南 | `backend/README.md` | 后端架构与开发者指南 |
| 前端指南 | `frontend-v2/README.md` | 前端架构与开发者指南 |
| 模块文档 | `backend/providers/llm/README.md` | LLM Provider 模块 |
| 模块文档 | `backend/agents/doc_ir/README.md` | DocIR 模块 |
| 模块文档 | `frontend-v2/src/hooks/README.md` | Hooks 索引与 API 文档 |
| 模块文档 | `frontend-v2/tests/README.md` | 测试配置与运行指南 |

### 不覆盖
- `langgraph/`、`langgraphjs/` 下所有 README（仅作为上游示例与文档的本地副本，**不再作为子模块**，禁止业务修改；内容陈旧时通过从上游仓库重新拉取 `examples/` 与 `docs/` 更新）
- `docs/` 下 README（分别由 `doc-sync-planning`、`docs-writer` 等技能维护）
- 自动生成文件中的注释式 README

## 变更 → README 映射表

这是技能的核心。根据变更类型定位需要更新的文件和章节。

| 变更类型 | 需要更新的文件 | 受影响章节 |
|----------|--------------|-----------|
| 后端启动/部署方式变更（Docker 入口、`main.py` → `uvicorn` 等） | `AGENTS.md`、`CLAUDE.md`、`backend/README.md` | 最常用命令·Backend、常用命令、Local 验证步骤、注意事项 |
| 后端端口/中间件注册顺序变化 | `AGENTS.md`、`CLAUDE.md` | 仓库边界、高优先级工程约束 |
| 后端工作流引擎/ScopedState 架构重构 | `backend/README.md` | 核心数据流转、工作流引擎、架构图 |
| 后端 LLM Provider 新增/签名变化 | `backend/providers/llm/README.md` | 关联文件、工厂入口、调用方式、统计日志 |
| 后端 DocIR 模块新增/修改 | `backend/agents/doc_ir/README.md` | 模块详解、数据流、使用示例 |
| 前端技术栈变更（Next.js、shadcn、Tailwind 等版本/配置） | `frontend-v2/README.md` | 技术栈对照表、目录结构、开发流程 |
| 前端目录结构约定变更 | `frontend-v2/README.md` | 目录结构说明、关键约束 |
| 前端 Design Token/Tailwind 配置变化 | `frontend-v2/README.md` | 设计规范、Design Tokens 对照表 |
| 前端路由/页面新增或废弃 | `frontend-v2/README.md` | 页面实现进度、路由清单 |
| 前端 hooks API 签名或类型变化 | `frontend-v2/src/hooks/README.md` | 对应 hook 的 API 文档、示例 |
| 前端测试配置/Playwright 参数变化 | `frontend-v2/tests/README.md` | 安装步骤、运行测试、注意事项 |
| OpenAPI 同步流程变更 | `AGENTS.md`、`CLAUDE.md` | OpenAPI/代码生成陷阱、同步方式 |
| Agent 规则/技能同步机制变更 | `AGENTS.md`、`CLAUDE.md` | 规则与技能同步、事实来源优先级 |
| 新增/废弃核心子项目 | `README.md`、`AGENTS.md`、`CLAUDE.md` | 快速导航、核心模块、仓库结构 |
| 测试框架/验证策略变更 | `AGENTS.md`、`CLAUDE.md` | 验证策略、测试要点、本地验证顺序 |
| 数据表/迁移相关变更 | `backend/README.md` | 注意事项（数据库迁移） |
| 包管理器/运行环境变更（uv/npm 等） | `AGENTS.md`、`CLAUDE.md`、`backend/README.md`、`frontend-v2/README.md` | 最常用命令、常用命令、安装步骤 |
| 新增/废弃/重命名 `.agents/rules/` 规则文件 | `.agents/rules/README.md` | 对应分类章节（Frontend/Backend/Infra/General） |
| `.agents/rules/` 规则分类调整 | `.agents/rules/README.md` | Categories 章节 |
| `.agents/rules/` 规则内容重大变更（适用范围、核心约束） | `.agents/rules/README.md` | 对应规则条目说明（`—` 后的描述） |

## 执行流程

### Step 1: 确认变更范围

接收或询问用户最近的变更描述。如果用户未提供，自动执行：

```bash
git diff --name-only HEAD~1..HEAD
```text

提取变更文件列表，与映射表进行匹配。

### Step 2: 定位受影响 README 和章节

根据映射表，输出一个**最小化更新计划**，格式如下：

```text
受影响文件：
  - AGENTS.md → §最常用命令、§仓库边界
  - CLAUDE.md → §常用命令

不受影响（排除在外的 README）：
  - frontend-v2/README.md（变更不涉及前端）
  - backend/README.md（变更不涉及后端架构）
```text

### Step 3: 验证文档是否确实过期

对每个可能受影响的 README，**只读需要更新的章节**（不读全文）。通过检查关键信息（如命令字符串、端口号、目录路径）是否与代码事实一致来判断是否需要更新。

### Step 4: 确认更新并执行

输出 diff，等待用户确认是否应用更新。

## 负向约束（不得违反）

1. **不碰上游副本**：`langgraph/`、`langgraphjs/` 下 README 一律不更新（仅作为上游仓库的本地参考副本）
1. **不碰 `docs/`**：`docs/00-7待办/`、`docs/04-技术/` 等 docs 目录由其他技能覆盖
1. **不碰未变更的章节**：即使 README 需要更新，也只更新映射表定位到的章节，不全文翻新
1. **不格式化**：不改 README 排版、缩进、换行风格，只改内容
1. **不预先更新**：确认 README 内容与实际代码事实不一致后才修改，不凭空猜测
1. **不无差别全量扫描**：不得读取全部 README 文件后再判断，必须先由映射表缩小范围再选择性读取

## 使用示例

### 示例 1：启动方式变更

**输入**：后端从 `python main.py` 改为 `uvicorn main:app`

**输出**：
```text
受影响文件：
  - AGENTS.md → §最常用命令：Backend（更新启动命令）
  - CLAUDE.md → §常用命令（更新 Backend 启动命令）
  - backend/README.md → §常用命令（更新启动命令）
```text

### 示例 2：新增前端路由

**输入**：在 `frontend-v2` 新增 `cases/[id]/workbench` 路由

**输出**：
```text
受影响文件：
  - frontend-v2/README.md → §页面实现进度（新增路由条目）
```text

### 示例 3：纯后端内部重构（无对外接口变化）

**输入**：重构 `workflow/node_base.py` 内部实现

**输出**：
```text
无需更新任何 README（内部重构不改变对外接口、启动命令或架构约定）
```text

## 输出格式

```text
受影响文件：
  - <file1> → <章节1>、<章节2>
  - <file2> → <章节3>

不受影响：
  - <file3>（原因：<原因>）
  - <file4>（原因：<原因>）

变更说明：
  - 变更类型：<变更类型>
  - 需要检查的要点：<要点列表>
```text
