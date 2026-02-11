---
trigger: model_decision
description: 文档目录；全部rules的描述、介绍
---

# 项目开发规范 (Rules)

本项目遵循 **FastAPI + React/Next.js** 技术栈的最佳实践，核心原则是：**安全 > 质量 > 速度**。

## 核心规则分类

### 🌍 全局规则 (Global)
- **[back_frond.md](../.agent/rules/global/back_frond.md)**: 全栈开发红线与规范。
- **[error-handling-standard.md](../.agent/rules/global/error-handling-standard.md)**: 统一异常处理与响应结构。
- **[test-coverage-threshold.md](../.agent/rules/global/test-coverage-threshold.md)**: 测试覆盖率与质量门槛。
- **[agentic-consensus.md](../.agent/rules/global/agentic-consensus.md)**: 多代理共识评审规则。
- **[chinese-language.md](../.agent/rules/global/chinese-language.md)**: 强制中文回复规范。
- **[tech-writing.md](../.agent/rules/global/tech-writing.md)**: 技术文档与文案写作规范。

### 🐍 Python/后端 (Backend)
- **[db-migration-policy.md](../.agent/rules/backend/db-migration-policy.md)**: 数据库迁移与 Alembic 规范。
- **[auth-check-strict.md](../.agent/rules/backend/auth-check-strict.md)**: 接口鉴权强制校验。
- **[py-fastapi-scalable.md](../.agent/rules/backend/py-fastapi-scalable.md)**: FastAPI 高可扩展 API 开发规范。
- **[django-standards.md](../.agent/rules/backend/django-standards.md)**: Django 开发最佳实践。
- **[CODING_STANDARDS.md](../.agent/rules/python/CODING_STANDARDS.md)**: Pythonic 编码规范。
- **[LINTING_FLAKE8.md](../.agent/rules/python/LINTING_FLAKE8.md)**: Python 静态检查 (Flake8) 规范。
- **[python-projects.md](../.agent/rules/python/python-projects.md)**: Python 项目结构与通用规范。

### 🎨 前端 (Frontend)
- **[naming-convention-react.md](../.agent/rules/frontend/naming-convention-react.md)**: React/TS 命名与开发规范。
- **[vanilla-frontend.md](../.agent/rules/frontend/vanilla-frontend.md)**: 原生前端开发规范 (HTML/Tailwind/JS)。
- **[react-scalable.md](../.agent/rules/frontend/react-scalable.md)**: 可扩展 React/TS 开发规范。
- **[aesthetic-standards.md](../.agent/rules/frontend/aesthetic-standards.md)**: 视觉审美与 UI 细节标准。

## 目录结构

- `backend/`: 后端相关的特定规则（预留）。
- `frontend/`: 前端相关的特定规则（预留）。

---
> [!TIP]
> 在进行任何 P0 或 P1 级别的变更时，请务必参考对应的规则文档。
