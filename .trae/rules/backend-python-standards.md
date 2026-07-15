---
alwaysApply: false
description: 用于 Python 子服务开发，包含代码风格、类型提示、异常处理、异步并发安全、防御性解析。
globs: douyin-mcp/**/*.py, jules-mcp-server/**/*.py
---

# Python 开发规范 (子服务)

本规范适用于本项目中所有使用 Python 编写的子服务（如 `douyin-mcp`、`jules-mcp-server` 等 MCP 服务）。

## 1. 核心代码风格与类型提示
- **类型提示 (Type Hints)**: 所有新增的函数和方法签名必须包含完整的 Type Hints。
- **显式变量注解**: 容器类型（如字典、列表）在初始化时若 mypy 无法自动推断其泛型，必须显式标注（如 `results: dict[str, Any] = {}`）。
- **防御性命名**: 变量与文件名均遵循小写下划线 (`snake_case`)。
- **PEP 8**: 严格遵循标准 Python 代码规范，使用合适的格式化工具（如 Ruff）。

## 2. 异常处理与防御性编程
- **错误捕获**: 核心逻辑及可能抛出异常的外部调用（如网络 IO、文件读取）必须使用 `try-except`。
- **禁止吞没异常**: 严禁在捕获 `Exception` 后直接 `pass` 或仅打印 console，必须妥善降级处理并记录日志。
- **资源清理**: 任何打开的文件句柄、网络连接等，必须在 `finally` 块中或通过上下文管理器（`with` 语句）显式释放，防范资源泄露。

## 3. 异步与并发安全
- **禁止在异步函数中同步阻塞**: 严禁在 `async def` 协程中使用同步阻塞调用（如 `time.sleep`、`requests.get`），必须使用 `asyncio.sleep`、`httpx.AsyncClient` 或将其委托给线程池。
- **防止并发作用域变量污染**: 在 `for` 循环中并发启动协程或后台任务时，确保被引用的闭包变量已被绑定为正确副本，防范变量逃逸导致的计算逻辑错乱。

## 4. 外部输入与边界数据解析防范 (Data Parsing Safeguard)
- **宽容解析**: 解析外部 LLM 返回或客户端传入的 JSON 字符串时，严禁直接强转，必须在 `try-except json.JSONDecodeError` 块中进行防御性解析。
- **默认降级**: 当 JSON 解析失败或格式不符时，必须配备合理的默认降级值（Fallback），确保服务不崩溃。

## 5. mypy 类型注解与闭包变量规范
- **闭包变量必须显式标注类型**: 当变量需要被闭包到后台任务或协程时，为了防止 mypy 隐式推断出不正确的类型，必须在赋值时显式声明最终类型。
- **类型忽略 (`type: ignore`) 的规范**:
  - **禁止使用裸 `# type: ignore`**: 必须使用带具体错误码的形式，如 `# type: ignore[arg-type]`。
  - **必须写明中文注释**: 每一处使用 `type: ignore` 的代码行尾必须加中文注释，说明为什么需要抑制，严禁无注释忽略类型检查。