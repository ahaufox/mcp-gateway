---
description: 7 阶段功能开发生命周期工作流 (7-Phase Feature Dev Lifecycle)
---

# 🚀 功能开发生命周期 (Feature Dev Lifecycle)

本工作流通过结构化的 7 个阶段，确保大型功能的开发从设计到落地都符合最高工程标准，避免破坏性变更。

---

## 📅 阶段定义

### 🔍 阶段 1: 调研 (Explorer)
- 深入理解需求，定位受影响的模块。
- 检查是否存在冲突的既有业务逻辑。

### 📐 阶段 2: 架构 (Architect)
- 产出 **`implementation_plan.md`**。
- 定义新的 API 契约、数据库 Schema 或核心 Service 逻辑。

### 🛡️ 阶段 3: 预评审 (Pre-Review)
- 触发 **`agentic-consensus`** 规则，进行内部角色的对抗演练。
- 向用户提交方案并获取 Approval。

### 💻 阶段 4: 编码 (Implementation)
- 分模块、小步快跑地进行编码。
- 实时应用 `LINTING_FLAKE8` 和 `naming-convention-react` 规则。

### 🧪 阶段 5: 验证 (Verification)
- 编写并运行单元测试与集成测试。
- 确保覆盖率满足 `test-coverage-threshold` 标准。

### 📖 阶段 6: 记录 (Walkthrough)
- 生成 **`walkthrough.md`**，记录变更点及测试结果。
- 包含截图或录屏（如有）。

### 🚢 阶段 7: 交付 (Delivery)
- 运行 `/git-auto` 进行提交并创建 PR。
- 运行 `/readme-update` 同步项目主文档。

---

## ⚠️ 强制准则
- 严禁跳过“架构”阶段直接进入“编码”。
- 每个 PR 必须对应一个完整的生命周期记录。

---
> [!IMPORTANT]
> **适用范围**: 建议对 P0 或 P1 级的新功能开发强制使用此工作流。
