---
description: 用于安全实施后端 API 变更（新增/修改字段、版本化），确保兼容性并通过测试固化契约。
---
## 输入参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `change_request` | text | ✅ | API 变更请求说明 |
| `compatibility_mode` | text | ❌ | 兼容模式：`backward` / `dual_write` / `versioned` |
| `clients_impact` | text | ❌ | 客户端影响评估 |
| `files_hint` | list | ❌ | 相关文件路径提示 |

---

## 执行步骤

### 1. 读取上下文 `read_context`

- **动作：** `read_files`
- **范围：** 最多 12 个文件
- **包含：** 路由/控制器、DTO/序列化、校验、鉴权、中间件、OpenAPI/Schema、客户端调用处

### 2. 生成变更计划 `propose_plan`

- **动作：** `write_plan`
- **必须包含：**
  - `current_contract` — 当前契约
  - `new_contract` — 新契约
  - `compatibility_strategy` — 兼容策略
  - `migration_steps` — 迁移步骤
  - `deprecation_plan` — 废弃计划
  - `risk_assessment` — 风险评估
  - `verification_commands` — 验证命令

### 3. 应用补丁 `apply_patch`

- **动作：** `edit_code`
- **约束：**
  - 最小化 diff
  - 不引入新依赖
  - 禁止顺手重构

### 4. 添加测试 `add_tests`

- **动作：** `edit_code`
- **目标：** `add_contract_and_compat_tests`
- **覆盖：**
  - 响应结构验证
  - 老客户端兼容性
  - 校验与鉴权

### 5. 运行检查 `run_checks`

- **动作：** `run_commands`
- **命令：** `pytest` → `tox` → `lint` → `typecheck` → `openapi 校验`
- **要求：** 全部通过（`allow_fail: false`）

### 6. 生成提交 `generate_commit`

- **动作：** `generate_commit_message`
- **规范：** Conventional Commits
- **格式：**
  - type: `feat` 或 `fix`
  - scope: `api`
  - subject: ≤ 72 字符
- **正文包含：** 迁移步骤、兼容说明、废弃时间表、测试结果

---

## 输出

| 输出 | 说明 |
|------|------|
| `patch_summary` | 代码变更摘要 |
| `migration_notes` | 迁移指南 |

---

## 流程图

```
输入参数
    ↓
[1] 读取上下文 → 扫描相关文件
    ↓
[2] 生成计划 → 输出完整变更方案
    ↓
[3] 应用补丁 → 最小化代码修改
    ↓
[4] 添加测试 → 契约 + 兼容性测试
    ↓
[5] 运行检查 → pytest/tox/lint/typecheck/openapi
    ↓
[6] 生成提交 → Conventional Commits 格式
    ↓
输出结果
```
