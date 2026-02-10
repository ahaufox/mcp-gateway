# Python 项目结构与通用规范 (Python Projects)

## 项目结构
- **目录分离**: 源码 (`src/` 或 `app/`)、测试 (`tests/`)、文档 (`docs/`) 和配置 (`config/`) 明确分离。
- **模块化设计**: 为 Model, Service, Controller 和 Utility 创建独立的文件。

## 配置与依赖
- **环境变量**: 始终通过环境变量管理配置，禁止硬编码敏感信息。
- **依赖管理**: 推荐使用 `Rye` 或 `Poetry` 结合虚拟环境进行管理。
- **代码一致性**: 使用 `Ruff` 进行 Linting 和格式化。

## AI 友好编码实践
- **描述性名称**: 使用详尽的变量和函数名，方便 AI 理解意图。
- **强制类型提示**: 所有代码必须包含 Type Hints。
- **详尽文档**: 复杂逻辑需配有 Docstrings 和 README 说明。
- **丰富错误上下文**: 抛出异常时提供充足的业务上下文信息，便于调试。

## 测试与 CI/CD
- **Pytest 优先**: 使用 `pytest` 编写全面的单元和集成测试。
- **CI 加固**: 通过 GitHub Actions 或 GitLab CI 自动化运行测试和代码质量检查。
