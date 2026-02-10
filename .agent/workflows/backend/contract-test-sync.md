---
description: 对齐契约源（OpenAPI/JSON Schema/类型定义）与真实实现，补充契约回归测试，避免前后端/调用方脱节。
---

## 输入参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `contract_source` | text | ✅ | 契约来源：`openapi` / `schema` / `types` |
| `endpoints_or_entities` | list | ❌ | 涉及的端点或实体 |
| `files_hint` | list | ❌ | 相关文件路径提示 |

---

## 执行步骤

### 1. 读取上下文 `read_context`

- **动作：** `read_files`
- **范围：** 最多 12 个文件
- **包含：** 契约文件（OpenAPI/Schema）、服务端 handler、序列化/校验层、客户端类型生成、现有契约测试

### 2. 生成同步计划 `propose_plan`

- **动作：** `write_plan`
- **必须包含：**
  - `current_mismatch` — 当前不一致项
  - `source_of_truth_decision` — 以谁为准（契约 or 实现）
  - `sync_actions` — 同步动作
  - `test_strategy` — 测试策略
  - `risk_assessment` — 风险评估
  - `verification_commands` — 验证命令

### 3. 应用补丁 `apply_patch`

- **动作：** `edit_code`
- **约束：**
  - 最小化 diff
  - 不引入新依赖
  - 复用现有测试框架

### 4. 添加测试 `add_tests`

- **动作：** `edit_code`
- **目标：** `add_contract_tests`
- **覆盖：**
  - 状态码
  - 字段必填性
  - 字段类型
  - 错误响应

### 5. 运行检查 `run_checks`

- **动作：** `run_commands`
- **命令：** `服务端测试` → `类型生成验证` → `schema 校验` → `lint/typecheck`
- **要求：** 全部通过（`allow_fail: false`）

### 6. 生成提交 `generate_commit`

- **动作：** `generate_commit_message`
- **规范：** Conventional Commits
- **格式：**
  - type: `test` 或 `fix`
  - scope: `contract`
  - subject: ≤ 72 字符
- **正文包含：** 不一致摘要、同步动作、测试覆盖

---

## 输出

| 输出 | 说明 |
|------|------|
| `patch_summary` | 代码变更摘要 |
| `contract_diff_summary` | 契约差异摘要 |

---

## 流程图

```
输入参数
    ↓
[1] 读取上下文 → 扫描契约文件与实现代码
    ↓
[2] 生成计划 → 识别不一致项，确定同步方向
    ↓
[3] 应用补丁 → 最小化修改，对齐契约与实现
    ↓
[4] 添加测试 → 状态码/字段/类型/错误响应
    ↓
[5] 运行检查 → 服务端测试/类型生成/schema校验/lint
    ↓
[6] 生成提交 → Conventional Commits 格式
    ↓
输出结果
```