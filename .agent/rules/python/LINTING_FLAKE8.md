# Python Linting (Flake8) 规范

此规则强制执行项目中的代码风格审查，确保所有 Python 代码符合 PEP 8 规范及项目特定的质量标准。

## 1. 核心要求 (Core Requirements)

- **强制校验**: 所有提交的代码必须通过 Flake8 静态检查。
- **配置一致性**: 必须遵循项目根目录下的 `.flake8` 或 `tox.ini` 配置文件。
- **零容忍**: 严禁在未经特殊说明的情况下使用 `# noqa` 忽略错误。

## 2. 关键检查项 (Key Checks)

| 规则编码 | 说明 (Description) | 强制程度 |
| :--- | :--- | :--- |
| **E/W** | 常见的 PEP 8 错误和警告（如：缩进、空格、行长） | P1 (Must Fix) |
| **F** | 逻辑错误（如：未使用的导入、未定义的变量） | P0 (Critical) |
| **C90** | 代码复杂度 (McCabe complexity) | P2 (Recommended) |

## 3. 具体约束 (Specific Constraints)

- **行长度 (Line Length)**: 最大行宽限制为 **88** 字符（与 Black 兼容）或 **79** 字符（严格 PEP 8）。
- **导入排序**: 建议配合 `isort` 使用，导入必须按标准库、第三方库、本地模块分块排序。
- **DOCSTRINGS**: 函数和类必须包含符合 Google 或 NumPy 风格的 Docstrings (特别是对于公共 API)。

## 4. 自动化集成 (Automation)

- **Pre-commit**: 推荐在本地安装并运行 `pre-commit install`，以便在提交前自动执行 `flake8`。
- **CI 拦截**: 在持续集成流水线中，任何 Flake8 报错都将阻断合入。

---
> [!WARNING]
> 严禁硬编码敏感信息。即使 Flake8 未报错，涉及 PII 或密钥的代码也将被 Security Audit 工作流拦截。
