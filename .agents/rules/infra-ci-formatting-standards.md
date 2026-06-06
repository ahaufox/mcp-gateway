# CI 与代码格式化规范

为了保证代码质量、提高协作效率，并确保 CI (GitHub Actions) 自动化检查顺利通过，降低因为格式不统一导致的代码冲突和 Review 成本，请所有开发人员和 AI Agent 严格遵循以下规范：

## 1. 后端规范 (Python)
后端统一使用 **Ruff** 作为 Lint 和 Format 工具，**MyPy** 作为静态类型检查工具。

- **格式化标准**: 遵循 Ruff 默认规则（类似于 Black 风格，但更快速）。
- **行宽限制**: `120` 字符（请在 IDE 中配置对应的 ruler）。
- **强制要求**:
    - 提交代码前，请在 `backend` 目录下执行 `ruff check . --fix` 自动修复常见格式与导入错误。
    - 提交代码前，请执行 `mypy .` 确保无阻塞性的静态类型错误（当前配置已在 `mypy.ini` 中忽略了部分非关键的动态类型警告，但基础结构必须正确）。
- **IDE 推荐配置 (VS Code)**:
    - 安装 `Ruff` 官方扩展。
    - 设置 `"editor.formatOnSave": true`。
    - 设置 `"editor.defaultFormatter": "charliermarsh.ruff"`。
    - 设置 `"editor.codeActionsOnSave": { "source.fixAll": "explicit", "source.organizeImports": "explicit" }`。

## 2. 前端规范 (React/TypeScript)
前端根据子项目不同，分别采用以下工具链，并且 CI 会进行严格拦截。

### 2.1 管理端 (admin-frontend)
- **工具链**: [Biome](https://biomejs.dev/)
- **操作规范**: 提交前必须确保 `npm run lint` 没有 Error 级别的警告。
- **IDE 推荐配置 (VS Code)**:
    - 安装 `Biome` 官方扩展。
    - 设置 `"editor.defaultFormatter": "biomejs.biome"`。
    - 设置 `"editor.formatOnSave": true`。

### 2.2 用户端 (frontend-v2)
- **工具链**: **ESLint** (结合 TypeScript 和 React 插件)
- **操作规范**: 遵循 `.eslintrc.cjs` 中的配置，重点修复未使用变量、未定义类型等语法层面的错误。
- **IDE 推荐配置 (VS Code)**:
    - 安装 `ESLint` 扩展。
    - 设置 `"editor.codeActionsOnSave": { "source.fixAll.eslint": "explicit" }`。

## 3. Git 提交规范 (Conventional Commits)
遵循本项目全局规则的约定式提交规范。格式如下：
`<type>(<scope>): <description>`

- `feat`: 新功能
- `fix`: 修补 Bug
- `docs`: 文档变更
- `style`: 代码格式（不影响逻辑）
- `refactor`: 重构
- `perf`: 性能优化
- `chore`: 构建过程或辅助工具的变动

**注意**：如果有专门针对代码格式的自动化变更，请务必使用 `style` 或 `chore` 作为 type，**禁止**将其与 `feat` 或 `fix` 混合在同一个 Commit 中。

## 4. CI 拦截机制
GitHub Actions 已配置对上述两套规范进行强制校验。如果由于本地未格式化或类型不兼容导致 CI 失败：
1. 请先查看 CI 报错日志（例如 Ruff 或 Biome 的输出）。
2. 在本地对应目录执行 `lint` 和 `build` 脚本进行修复。
3. 重新提交。
