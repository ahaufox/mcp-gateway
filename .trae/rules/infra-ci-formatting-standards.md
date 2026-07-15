---
alwaysApply: false
description: CI 与代码格式化规范，包含 Go/Python 后端及 React 前端在 CI 中的强制校验与本地工具要求。
globs: web/**/*.ts, web/**/*.tsx, mcp-proxy/**/*
---

# CI 与代码格式化规范

为了保证代码质量、提高协作效率，并确保本地及 CI 自动化检查顺利通过，降低因为格式不统一导致的代码冲突和 Review 成本，请所有开发人员和 AI Agent 严格遵循以下规范：

## 1. 后端规范 (Go & Python)
### 1.1 Go 后端 (`mcp-proxy`)
- **格式化标准**: 遵循 Go 官方的标准格式化工具 `go fmt`。
- **强制要求**:
  - 提交代码前，请在本地执行 `go fmt ./...`，确保所有 Go 文件格式化正确。
  - 提交代码前，执行 `go vet ./...` 进行编译器静态分析，确保没有低级语法与类型错误。

### 1.2 Python 子服务
- **格式化标准**: 遵循 PEP 8 规范。如有 Ruff，使用 Ruff 进行检查和格式化。
- **强制要求**:
  - 确保 Python 代码缩进统一，没有未使用的导入和变量。

## 2. 前端规范 (React/TypeScript)
- **前端项目 (`web/`)**:
  - **工具链**: **ESLint** (结合 TypeScript 和 React 插件)。
  - **操作规范**: 提交前确保执行 `npm run lint` 没有错误，重点修复未使用变量、未定义类型等语法错误。

## 3. Git 提交规范
- 提交前必须保证 commit 信息符合 `general-git-commit-message.md` 的规范，包含语义化前缀且为中文。

## 4. 本地静态验证与拦截机制
如果由于本地未格式化或类型不兼容导致 CI 失败：
1. 请先查看 CI 报错日志。
1. 在本地对应目录（如 `mcp-proxy` 或 `web`）执行本地 `fmt`/`lint` 和 `build` 脚本进行修复。
1. 重新提交。