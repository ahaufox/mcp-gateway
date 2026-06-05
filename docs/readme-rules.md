---
trigger: model_decision
description: 文档目录；全部rules的描述、介绍
---

# 项目开发规范 (Rules)

本项目遵循 **FastAPI + React/Next.js** 技术栈的最佳实践，核心原则是：**安全 > 质量 > 速度**。

## 核心规则分类

### 🌍 全局规则 (Global)
- **[global-full-stack-standards.md](../.agents/rules/global-full-stack-standards.md)**: 全栈开发红线与规范。
- **[global-error-handling-standard.md](../.agents/rules/global-error-handling-standard.md)**: 统一异常处理与响应结构。
- **[global-test-coverage-threshold.md](../.agents/rules/global-test-coverage-threshold.md)**: 测试覆盖率与质量门槛。
- **[global-agentic-consensus.md](../.agents/rules/global-agentic-consensus.md)**: 多代理共识评审规则。
- **[global-chinese-language.md](../.agents/rules/global-chinese-language.md)**: 强制中文回复规范。
- **[global-tech-writing.md](../.agents/rules/global-tech-writing.md)**: 技术文档与文案写作规范。

### 🐍 后端 (Backend)
- **[backend-db-migration-policy.md](../.agents/rules/backend-db-migration-policy.md)**: 数据库迁移与 Alembic 规范。
- **[backend-auth-check-strict.md](../.agents/rules/backend-auth-check-strict.md)**: 接口鉴权强制校验。
- **[backend-py-fastapi-scalable.md](../.agents/rules/backend-py-fastapi-scalable.md)**: FastAPI 高可扩展 API 开发规范。
- **[backend-django-standards.md](../.agents/rules/backend-django-standards.md)**: Django 开发最佳实践。
- **[backend-python-modern-stack.md](../.agents/rules/backend-python-modern-stack.md)**: 现代 Python 技术栈规范。
- **[backend-typescript-modern-stack.md](../.agents/rules/backend-typescript-modern-stack.md)**: 现代 TypeScript 后端技术栈规范。
- **[backend-mcp-gateway-safety.md](../.agents/rules/backend-mcp-gateway-safety.md)**: MCP 网关安全规范。

### 🎨 前端 (Frontend)
- **[frontend-naming-convention-react.md](../.agents/rules/frontend-naming-convention-react.md)**: React/TS 命名与开发规范。
- **[frontend-vanilla-frontend.md](../.agents/rules/frontend-vanilla-frontend.md)**: 原生前端开发规范 (HTML/Tailwind/JS)。
- **[frontend-react-scalable.md](../.agents/rules/frontend-react-scalable.md)**: 可扩展 React/TS 开发规范。
- **[frontend-aesthetic-standards.md](../.agents/rules/frontend-aesthetic-standards.md)**: 视觉审美与 UI 细节标准。

### 🐳 基础设施 (Infrastructure)
- **[docker-best-practices.md](../.agents/rules/docker-best-practices.md)**: Docker 与容器化开发最佳实践。

## 规则存放目录

所有规则文件均扁平化存放在 `[.agents/rules/](../.agents/rules)` 目录下，以便于统一管理和识别。

---
> [!TIP]
> 在进行任何 P0 或 P1 级别的变更时，请务必参考对应的规则文档。
