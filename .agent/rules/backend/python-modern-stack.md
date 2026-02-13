# 🐍 Python Modern Stack

本规则旨在标准化 Python 项目开发工具链与实践。

## 1. 依赖管理 (Dependency Management)
- **推荐**: 使用 `uv`。
- **配置**: `pyproject.toml`
- **示例**:
  ```toml
  [tool.uv]
  # ...
  ```
- **命令**: `uv sync`, `uv run`, `uv add`.

## 2. 代码风格 (Linting & Formatting)
- **推荐**: 使用 `flake8` 进行代码风格检查，`black` 或 `ruff` 进行格式化。
- **配置**:
  ```toml
  [tool.ruff]
  line-length = 88
  select = ["E", "F", "B"]
  ```

## 3. 测试 (Testing)
- **必选**: 使用 `pytest`。
- **覆盖率**: 推荐 `pytest-cov`。
- **运行**: `uv run pytest`.

## 4. 类型检查 (Type Checking)
- **推荐**: 使用 Python 3.10+ 类型提示 (Type Hints)。
- **工具**: `mypy` 或 `pyright`.
- **示例**:
  ```python
  def process_data(data: dict[str, Any]) -> list[int]:
      ...
  ```

## 5. 项目结构 (Project Structure)
- 使用 `src/` 布局 (Source Layout) 或扁平布局 (Flat Layout)。
- 必须包含 `pyproject.toml`, `README.md`.
