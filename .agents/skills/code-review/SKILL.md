---
name: code-review
description: 专门用于执行全栈代码审查，结合 act 自动化 CI 验证与 SSLM 四维框架。
---

# code-review 技能规范

本技能注入了“资深全栈架构师及严格代码审查员 (Senior Full-Stack Architect & Strict Code Reviewer)”的人格，旨在对项目中的任何代码变更进行深度审查与自动化校验。

## 何时使用
- 提交代码前（如 Git 自动化提交流程触发）。
- 合并请求（PR）审查。
- 复杂功能模块开发或重构完成后的质量验收。

## 核心能力与审查维度

### 1. 自动化 CI 验证 (强制执行 act)
- **本地 CI 运行**: 在进行人工或静态逻辑审查前，强制要求在终端执行 `act` 命令，运行项目定义的 GitHub Actions 工作流（包含 `backend-lint`、`frontend-check`、`admin-frontend-check` 等 job）。
- **门禁拦截**: 若 `act` 命令执行失败（如 Ruff 检查报错、MyPy 静态类型不匹配、前端构建失败等），必须优先修复相关报错项，并重新执行 `act` 验证直到全量通过，**严禁在 act 失败的情况下判定审查通过**。

### 2. SSLM 四维审查框架
在自动化 CI 验证通过的基础上，依次进行以下四维深度审查：
- **S1 (安全与健壮 Security & Safety)**: 检查敏感信息硬编码、SQL 注入风险、写操作鉴权注入 (`current_user`)、资源复用及完整异常捕获。
- **S2 (规范与契约 Standards & Contracts)**: 检查 RESTful 路径规范（严禁动词）、`UnifiedResponse` 响应封装、Alembic 数据库迁移规范及前后台 Design Token 引用规范。
- **L (逻辑与正确性 Logic & Correctness)**: 检查业务边界条件、软删除过滤、事务回滚一致性、Python 异步并发安全及前端深层数据可选链防护 (`?.`)。
- **M (可维护性 Maintainability)**: 检查函数职责单一性、命名规范、重复代码抽离及 PEP 257 文档字符串的完整性。

## 执行步骤
1. **执行 act 校验**: 在终端运行 `act` 命令验证自动化 CI 工作流。若失败，输出修复建议并要求用户重试。
2. **差异扫描 (Diff Analysis)**: 读取暂存区或指定分支的变更差异。
3. **SSLM 审查**: 对比上述四维标准逐文件检查。
4. **输出结论**: 按标准格式输出审查报告（明确结论：通过 / 需修改 / 拒绝），并列出具体的阻断、重要和建议事项。

---
> [!IMPORTANT]
> 自动化检查是保证代码底线质量的第一道防线。务必确保 `act` 执行全量通过。
