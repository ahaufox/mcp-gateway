---
trigger: always_on
description: 针对 AI 响应解析的规范，包括安全 JSON 解析和标题文本归一化。
---

# AI 响应解析规范 (AI Response Parsing Rule)

## 1. 安全 JSON 解析 (Safe JSON Parsing)
- **宽容解析**: 由于 AI 返回的 JSON 结构可能存在细微瑕疵（例如使用 Python 风格的单引号字典字符串，或者在 JSON 外包裹了 Markdown 代码块），必须使用安全的解析函数，而不是直接调用 `json.loads`。
- **降级策略**: 如果 JSON 解析完全失败，系统应当有合理的默认值降级或抛出明确的解析异常，不能导致服务 Crash 或进入死胡同。
- **禁用直接评估**: 严禁使用 `eval()` 解析不规范的 JSON 字符串。

## 2. 文本归一化 (Text Normalization)
- **清理 AI 伪影**: 提取标题或文本内容时，必须实现归一化逻辑，移除 AI 生成时可能残留的占位符（如 `[Insert Title Here]`）、首尾多余的引号，以及 JSON 结构的碎片残留。
- **前端展示**: 前端组件在接收到这些文本时，也应进行标准化的显示处理，防止脏数据破坏 UI 布局。

## 3. 类型定义与错误处理 (Types & Error Handling)
- 增强相关的类型定义，确保从解析层到业务层的数据类型安全。
- 在发生解析错误时，日志应当记录原始的异常返回文本，以备后续分析和优化 Prompt。
