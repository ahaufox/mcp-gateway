---
description: 修复 bug，并补齐测试与约定式提交信息（内测期/生产前通用）
---

# /bugfix 工作流（中文增强版）

## 适用范围
- 目标：以**最小改动**修复可复现 bug，补齐**回归测试**，并生成符合 **Conventional Commits** 的提交信息
- 风险级别：默认 **P1**；触及鉴权/支付/PII/注入 → **P0**（阻断合入，必须严格验证）

---

## inputs

- id: symptom  
  type: text  
  required: true  
  description: 症状描述（报错信息/截图文字/接口路径/期望 vs 实际/影响范围/发生频率）

- id: files_hint  
  type: list  
  required: false  
  description: 相关文件提示（路由、组件、服务层、CRUD、测试、日志、迁移等路径）

---

## steps

### 1) read_context
action: read_files

args:
- hints: "{{inputs.files_hint}}"
- max_files: 8

阅读范围建议（按优先级）：
- 直接相关入口（API 路由 / 前端页面 / 触发按钮）
- 调用链核心（service / crud / utils）
- 数据模型与校验（schemas / models）
- 现有测试与 fixtures
- 错误处理与日志（统一异常、拦截器）

输出要求（必须写在后续 plan 中）：
- 初步复现路径（最短步骤）
- 可能受影响的模块边界（前端/后端/DB/缓存）
- 是否涉及 P0 红线（鉴权/PII/注入等）

---

### 2) propose_plan
action: write_plan

args.must_include:
- root_cause_hypotheses
- chosen_fix
- risk_assessment
- verification_commands

Plan 编写规范（必须）：
- 全中文
- 先“**复现与定位**”，再“**修复**”，最后“**加测试与验证**”
- root cause 至少 2 个假设，并说明如何排除
- chosen_fix 必须是**最小可行修复**（minimal diff），避免顺手重构

#### root_cause_hypotheses（根因假设）
至少包含两类视角：
- **逻辑错误**：边界条件、空值处理、时区/格式、状态机缺口
- **集成错误**：接口契约不一致、序列化字段差异、DB 约束/索引缺失、缓存 TTL/Key 冲突
- **并发/时序**：幂等性不足、重复提交、事务边界、竞态条件
- **配置/环境**：环境变量缺失、构建产物差异、CORS/代理头、超时设置

每个假设必须写：
- 证据来源（日志/堆栈/代码路径）
- 验证方式（加日志、断点、最小复现、单测构造输入）

#### chosen_fix（选定修复）
必须说明：
- 修复点位置（文件/函数/接口）
- 修复方式（校验、容错、事务、返回码、类型收紧等）
- 为什么不会引入副作用（语义保持 / 兼容性）
- 是否需要数据修复/迁移（如需要，升级为 P1+ 并补回滚）

#### risk_assessment（风险评估）
至少包含：
- 风险等级：P0 / P1 / P2
- 影响面：哪些接口/页面/用户路径
- 安全检查：是否触及鉴权/权限/PII/SQL 注入
- 回滚方案：
  - 代码：`git revert <hash>`
  - DB：如有 migration，`alembic downgrade -1`（或等价）
  - 配置：如引入新开关/变量，给出恢复旧值

#### verification_commands（验证命令）
必须列出：
- 单元/集成测试命令（项目实际命令）
- lint / typecheck
- 复现步骤（人工或自动）
- 若涉及后端：建议加 `pytest -k <case>`（或等价）

---

### 3) apply_patch
action: edit_code

args.constraints:
- no_new_dependencies（不新增依赖）
- no_secrets_in_code（不写入密钥/Token/PII）
- minimal_diff（最小改动）

修复实现要求（必须）：
- 不做无关重构；保持 diff 聚焦
- 对关键边界条件补充中文注释（说明“为什么这样处理”）
- 不吞异常：错误要么显式处理，要么记录结构化日志并抛出合理错误码
- 若修复涉及输入校验：前后端契约必须一致（字段名、可空性、默认值）

---

### 4) add_tests
action: edit_code

args.goal: add_or_update_tests  
args.constraints:
- tests_cover_regression（测试必须覆盖本次回归点）

测试要求（必须满足）：
- **至少 1 个回归测试**：能在修复前失败、修复后通过
- 覆盖“症状触发条件”的最小输入
- 尽量避免时间敏感/随机性（固定时间、固定 seed、固定数据规模）

建议策略（按优先级）：
1. 后端 bug：优先写 `tests/integration/`（覆盖路由到 DB/Redis 的关键链路）
2. 纯函数/工具 bug：写 `tests/unit/`
3. 前端 bug：补组件测试或服务层测试（若项目已有测试框架）；否则至少补 TypeScript 类型与最小行为测试

---

### 5) run_checks
action: run_commands

args.commands:
- "npm test"
- "npm run lint"

args.allow_fail: false

说明（可按项目补充/替换，建议在 workflow 落地时参数化）：
- 如果是后端修复，应增加：
  - `pytest`
  - `ruff check .`（或项目 lint 命令）
  - `mypy .`（或项目 typecheck 命令）
- 如果是全栈改动，应同时跑前后端检查，避免“前端通过但后端挂”的尴尬

---

### 6) generate_commit
action: generate_commit_message

args:
- convention: conventional_commits
- rules:
  - type: fix
  - scope_from_paths: true
  - max_subject_len: 72
- include_body:
  - reproduction_steps
  - what_changed
  - test_results

提交信息要求（必须）：
- subject 用中文动宾结构，说明修复对象与结果
- body 必须可审计：能让 reviewer 快速知道“怎么复现、改了什么、怎么验证”
- 如涉及用户可见行为变化：在 body 标注兼容性说明

示例：
```
fix(frontend): 修复登录后回跳地址丢失导致的循环重定向

- reproduction:
  1. ...
  2. ...
- what changed:
  - ...
- tests:
  - npm test
  - npm run lint
```

---

## outputs

### patch_summary
必须包含：
- 根因结论（对应被验证/排除的假设）
- 改动文件清单与关键点（每个文件 1 行）
- 回归测试说明（新增/更新了哪些测试，覆盖什么场景）
- 风险等级与回滚方式（代码/DB/配置）

### commit_message
输出最终可直接使用的 Conventional Commit 文本：
- `<type>(<scope>): <subject>`
- body 含 reproduction_steps / what_changed / test_results

---
