
# 自动化工作流 (Workflows)

本项目提供了一系列自动化工作流，用于标准化常见的开发任务，确保代码质量和文档同步。

## 核心工作流分类

### ⚙️ 后端/Python (Backend)
- **[api-change-backend.md](../.agent/workflows/backend/api-change-backend.md)**: 安全实施后端 API 变更。
- **[db-schema-sync.md](../.agent/workflows/backend/db-schema-sync.md)**: 自动同步模型、Schema 与数据库。
- **[contract-test-sync.md](../.agent/workflows/backend/contract-test-sync.md)**: 对齐契约源与实现。
- **[perf-profile-py.md](../.agent/workflows/python/perf-profile-py.md)**: Python 性能分析与优化。

### 🎨 前端 (Frontend)
- **[ui-design.md](../.agent/workflows/frontend/ui-design.md)**: 顶级视觉设计与审美优化流程。
- **[frontend-page-extension.md](../.agent/workflows/frontend/frontend-page-extension.md)**: 扩展页面、路由与模块。

### 🛠️ 通用/运维 (General)
- **[bug-finder.md](../.agent/workflows/general/bug-finder.md)**: 自动化发现潜在 bug。
- **[bug-fix.md](../.agent/workflows/general/bug-fix.md)**: 标准化 bug 修复与测试。
- **[security-audit.md](../.agent/workflows/general/security-audit.md)**: 定期安全审计。
- **[docker-build-verify.md](../.agent/workflows/general/docker-build-verify.md)**: Docker 构建验证。
- **[e2e-suite-run.md](../.agent/workflows/general/e2e-suite-run.md)**: E2E 自动化测试。
- **[readme-update.md](../.agent/workflows/general/readme-update.md)**: 自动更新文档。
- **[release-prep.md](../.agent/workflows/general/release-prep.md)**: 发布准备工作流。
- **[agent-evolution.md](../.agent/workflows/agent-evolution.md)**: AI 助手每日进化与维护协议。
- **[git-auto-suite.md](../.agent/workflows/general/git-auto-suite.md)**: Git 自动化提交与 PR 套件。
- **[feature-dev-lifecycle.md](../.agent/workflows/general/feature-dev-lifecycle.md)**: 7 阶段功能开发生命周期。

---
> [!NOTE]
> 运行工作流时，请优先通过 AI IDE 的斜杠命令触发，以获得最佳引导体验。
