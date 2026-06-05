---
description: 全局异常拦截处理规范（如 LLM 及 API 层的异常处理与兜底方案）。
---

# 异常拦截处理规范 (Exception Interception Rule)

## 角色设定
作为资深的后端系统架构师，负责保障系统在高并发及异常情况下的稳定性与高可用性。

## 核心要求

1. **统一的异常处理捕获**：在 FastAPI (或所使用的对应框架) 中，所有的外部请求失败、LLM API 调用异常（例如 `APITimeoutError`, `RateLimitError`, `AuthenticationError`, `APIStatusError`）必须在全局 `exception_handler` 层被拦截，以防止服务端产生 `500 Internal Server Error` 崩溃。
2. **兜底方案 (Fallback)**：发生异常时，接口必须有明确的兜底返回（如友好的错误消息或降级的数据处理），并记录详细的错误上下文至监控日志中。
3. **严格禁止**：禁止在业务逻辑的每个分散函数里硬编码异常抛出而不做捕获。所有的 `try-except` 如果在业务层处理，必须有针对性的恢复策略或向上抛出自定义的 HTTP 异常类供全局捕获。
