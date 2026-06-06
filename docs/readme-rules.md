# 项目开发规范 (Rules)

## 规则文件索引

所有规则文件存放在 `.agents/rules/` 目录下。

### 后端 (Backend)
- **[backend-python-standards.md](../.agents/rules/backend-python-standards.md)** — Python 代码风格、鉴权、异常处理、异步安全
- **[backend-db-migration-standards.md](../.agents/rules/backend-db-migration-standards.md)** — Alembic 迁移规范
- **[backend-exception-interception-rule.md](../.agents/rules/backend-exception-interception-rule.md)** — 全局异常拦截
- **[backend-startup-memory-safeguard.md](../.agents/rules/backend-startup-memory-safeguard.md)** — 启动挂起与异步死锁防范
- **[backend-workflow-and-test-safeguards.md](../.agents/rules/backend-workflow-and-test-safeguards.md)** — 工作流与测试防护
- **[backend-api-collaboration-rule.md](../.agents/rules/backend-api-collaboration-rule.md)** — 前后端接口协作规范

### 前端 (Frontend)
- **[frontend-standards.md](../.agents/rules/frontend-standards.md)** — React/TypeScript 工程规范
- **[frontend-v2-structure.md](../.agents/rules/frontend-v2-structure.md)** — 薄 page、lib/queries 解耦
- **[frontend-ui-design-system.md](../.agents/rules/frontend-ui-design-system.md)** — 视觉设计标准与品牌规范
- **[frontend-ui-state-fallback-rule.md](../.agents/rules/frontend-ui-state-fallback-rule.md)** — UI 状态兜底
- **[frontend-quote-escape-rule.md](../.agents/rules/frontend-quote-escape-rule.md)** — JSX 引号转义
- **[frontend-header-comment-standards.md](../.agents/rules/frontend-header-comment-standards.md)** — 文件头部注释规范

### 通用 (General)
- **[general-global-standards.md](../.agents/rules/general-global-standards.md)** — 全栈核心准则
- **[general-code-review-standards.md](../.agents/rules/general-code-review-standards.md)** — 代码审查规范
- **[general-ai-response-parsing-rule.md](../.agents/rules/general-ai-response-parsing-rule.md)** — AI 响应解析规范
- **[general-git-commit-message.md](../.agents/rules/general-git-commit-message.md)** — Git 提交信息规范

### 基础设施 (Infrastructure)
- **[infra-ai-assistant-rules-standards.md](../.agents/rules/infra-ai-assistant-rules-standards.md)** — AI 辅助配置规范
- **[infra-ci-formatting-standards.md](../.agents/rules/infra-ci-formatting-standards.md)** — CI 格式化规范
- **[infra-dependency-compat-rule.md](../.agents/rules/infra-dependency-compat-rule.md)** — 依赖兼容性规则
