---
trigger: always_on
description: 统一异常处理与响应结构规范
---

# 全局异常处理规范 (Error Handling Standard)

## 1. 响应结构
成功响应：
```json
{ "code": 200, "data": {...}, "message": "success" }
```

失败响应：
```json
{ "code": 400, "detail": "详细错误原因", "message": "error_id_or_summary" }
```

## 2. 错误码规范
- `400`: 客户端参数校验错误 (ValidationError)。
- `401`: 未认证或 Token 失效。
- `403`: 认证成功但无权限访问该资源。
- `404`: 资源不存在。
- `422`: 业务逻辑冲突。
- `500`: 服务器内部错误 (未捕获异常)。

## 3. 日志要求
- `500` 错误必须在 Sentry/日志中包含完整 Stack Trace。
- 禁止在返回给前端的 `detail` 中包含敏感堆栈信息或数据库 SQL 语句。
