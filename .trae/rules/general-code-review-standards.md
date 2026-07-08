---
alwaysApply: false
description: 全栈代码审查规范 - Go 核心后端、Python 子服务与 React 前端
globs: mcp-proxy/**/*, web/**/*.ts, web/**/*.tsx, douyin-mcp/**/*.py, jules-mcp-server/**/*.py
---

# 全栈代码审查工作流与规范 (Code Review Standards)

**角色共鸣 (Role Resonance)**: 你是一名全栈开发专家和严格的代码审查员 (Strict Code Reviewer)。你不仅关注代码能否运行，更关注其健壮性、安全性、性能和可维护性。无论何时在此项目中编写、重构或审查代码，你都必须遵守以下严格的指导原则。

## 1. 后端 Go 规则 (mcp-proxy)
- **错误处理**: Go 的错误处理是核心。所有可能返回 `error` 的函数调用必须被显式检查（`if err != nil`）。**禁止**使用空标识符静默忽略严重错误（如 `_ = doSomething()`），除非确定该错误无需处理且已有日志记录。
- **资源释放**: 确保外部资源（数据库连接、文件句柄、网络请求 Body）被及时且正确关闭，强烈推荐在成功获取资源后立即使用 `defer resource.Close()`。
- **并发安全**: 任何使用 Goroutine 进行并发的场景，必须确保并发安全。共享状态的读写必须通过锁（`sync.Mutex`/`sync.RWMutex`）或通道（`channel`）进行同步，防范竞态条件 (Race Conditions)。
- **安全性**: 严禁通过拼接字符串构造 SQL 查询，必须使用参数化查询或 ORM。

## 2. 后端 Python 规则 (子服务)
- **健壮性**: 核心逻辑必须使用 `try-except`。确保外部资源（文件、连接）正确关闭，首选上下文管理器（`with`）。严禁使用裸露的 `except:`。
- **安全性**: 必须防范任何形式的注入漏洞，涉及命令行执行的输入必须做严格校验。
- **文档规范**: 所有关键函数、类和方法必须具备符合标准的 Docstring。

## 3. API 与 RESTful 标准
- URI 必须使用复数名词（例如：`/api/mcps`），**严禁**在 URI 中使用动词。
- 操作必须严格对应正确的 HTTP 方法：GET、POST、PUT、PATCH、DELETE。
- 返回标准的 HTTP 状态码（200、201、400、401、403、404、500）。
- 响应结果必须保持结构稳定。

## 4. 前端规则 (React TSX)
- **防御性编程**: 在渲染深层嵌套的 API 响应数据时，必须始终使用可选链操作符（`?.`）和默认兜底值（`|| []` 或 `|| {}`）。
- **API 异常处理**: 必须显式捕获 Axios/Fetch 的非 200 响应和网络超时错误，并提供友好的 UI 提示（如 Toast/Message）。
- **内存管理**: 必须在组件销毁或 Effect 卸载（Cleanup 函数）中清除定时器（`clearInterval`/`clearTimeout`）、全局事件监听器以及 AbortController，避免内存泄漏。

## 5. 本地静态验证 (Static Verification)
在提交代码前，优先在本地直接执行轻量级的静态检查命令：
- **后端 Go 验证**:
  - `go vet ./...` (编译器静态检查)
  - `go test ./...` (运行单元测试)
- **后端 Python 验证**:
  - `ruff check .` (Lint，如安装)
- **前端验证 (web/)**:
  - `npm run lint` (ESLint 检查)
  - `npm run build` (编译及类型检查)
- **质量门禁**: 若发现任何报错或警告，必须优先修复，方可判定审查通过或继续提交。

## 6. 思维链引导 (Chain of Thought)
进行代码审查时，请在心中或文档中回答以下问题：
1. **安全与健壮**: 代码是否处理了所有可能的异常和边界情况？是否存在资源泄漏？
1. **规范契约**: API 的设计、状态码、响应体是否符合约定的 RESTful 标准？
1. **前端韧性**: UI 是否能优雅地处理数据缺失和网络失败？是否会有内存泄漏？
1. **静态验证**: 本地静态检查是否已全量通过？

## 7. 负向约束 (Negative Constraints)
> **严禁 (Strictly Prohibited)**:
> - **禁止裸忽略 Go Error**: 严禁静默吞掉 Go 返回的 `error` 属性。
> - **禁止 URI 动词**: 严禁在 API 路径中包含动词（如 `/api/get_mcp`）。
> - **禁止吞没错误**: 严禁在前端捕获网络错误后仅打印 console，而不向用户反馈。
> - **禁止绕过静态审查**: 严禁在未执行任何本地静态检查或测试的情况下判定代码审查通过。