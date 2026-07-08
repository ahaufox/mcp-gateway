---
alwaysApply: false
description: 项目规则: backend-workflow-and-test-safeguards
globs: mcp-proxy/**/*, douyin-mcp/**/*.py, jules-mcp-server/**/*.py
---

# 后端测试健壮性防范规则 (Backend Testing Safeguards)

## 1. 异步流式接口单元测试死锁防范
在涉及流式响应（SSE）与异步队列的测试场景中，如果将异步队列的写入端或事件源 Mock 成无动作的空白 `Mock`，消费协程可能会永久阻塞在 `await queue.get()` 上，导致测试进程永久死锁并造成 CI/CD 构建挂起。

- **强制 Mock side_effect 模拟退出**：当 Mock 异步队列写入或事件推流时，必须配置 `side_effect`。在调用时要么发送结束符号（如写入 `None`），要么抛出结束异常以使消费协程正常退出。
- **强制超时机制**：所有涉及异步流消费、等待 Goroutine/线程 或事件通知的测试用例，必须配置显式超时机制（如 pytest 的 timeout 装饰器，或 Go 的 `select` 配合 `time.After`），防止发生死锁时进程阻塞。

## 2. 测试桩 (Mock) 与环境状态的隔离与清理
测试中临时修改全局配置、环境变量、依赖注入覆写或 Mock 全局单例，如果测试结束后没有彻底清理，会导致常驻内存并污染后续其他测试用例（如连接池抢占、脏 Mock、越权检测失效）。

- **严禁模块导入级覆盖**：严禁在模块导入时直接修改全局状态以进行测试。
- **生命周期动态覆盖与彻底恢复**：所有在测试中进行的全局覆盖、配置修改和注入重载，必须在测试的 Setup 阶段进行，且**必须**在 TearDown 阶段显式将其还原或清空，确保测试用例之间的状态完全隔离。
- **Fixture 最佳实践**：使用带有自动清理机制的 Fixture。在 `yield` 之后恢复所有被 Mock 的全局状态。