---
trigger: model_decision
description: 用于 Python、FastAPI 后端开发及数据库迁移场景，包含代码风格、鉴权校验、异常处理结构、异步并发安全及测试覆盖率要求。
---

# Python & FastAPI 后端开发规范

## 1. 项目结构与核心风格
- **目录分离**: 源码 (`src/` 或 `app/`)、测试 (`tests/`)、文档 (`docs/`) 和配置 (`config/`) 明确分离。
- **模块化设计**: 为 Model, Service, Controller 和 Utility 创建独立的文件。
- **异步优先**: 涉及 I/O 或异步操作使用 `async def`。
- **类型提示**: 所有函数签名必须包含 Type Hints，优先使用 Pydantic 模型进行验证。
- **命名规范**: 目录和文件使用小写下划线 (如 `routers/user_routes.py`)。
- **直接访问**: 优先直接访问属性（如 `obj.base_date`），而非使用 getter。必要时使用 `@property`。
- **负向约束**: 严禁捕获基础 `Exception` 后无处理（`pass`），严禁使用全局依赖引发未受控的本地模型或数据库连接初始化阻塞。
- **思维链引导 (CoT)**: 触发条件：处理多系统交互（如 ASR 或 Orchestrator 博弈多模型调度）、复杂的异步任务及事务安全时。必须先使用 `<thinking>分析死锁风险、事务边界及防幻觉校验...</thinking>` 输出推理过程。

## 2. 异常处理与响应结构
成功响应：
```json
{ "code": 200, "data": {...}, "message": "success" }
```
失败响应：
```json
{ "code": 400, "detail": "详细错误原因", "message": "error_id_or_summary" }
```
- **错误码**: `400` (参数错误), `401` (认证失败), `403` (权限不足), `404` (不存在), `422` (逻辑冲突), `500` (内部错误)。
- **防御性编程**: 使用 Early Returns 和 Guard Clauses 避免深层嵌套。
- **日志**: `500` 错误必须包含完整 Stack Trace。禁止向前端暴露敏感信息或 SQL。

## 3. 鉴权与安全
- **强制注入**: 所有写操作及敏感读接口必须使用 `Depends` 注入鉴权（如 `current_user`）。
- **权限校验**: 涉及私有数据必须校验 `user_id`。管理接口必须包含角色校验。
- **禁止项**:
  - 禁止手动解析 `Authorization` Header。
  - 严禁硬编码敏感信息（密钥、PII）。

## 4. 异步与并发安全规范

### 4.1 触发场景
- 编写或重构使用 `asyncio` 的核心逻辑（如大批量并发请求、并发检索）。
- 处理后台队列任务或长耗时推断任务。

### 4.2 异步基础规则
- **严禁使用同步阻塞调用** (如 `time.sleep`, `requests.get`) 在 `async def` 路由中。必须使用 `asyncio.sleep` 或 `httpx.AsyncClient` 等异步替代方案。
- 禁止在 async 函数中使用 `time.sleep`，必须使用 `asyncio.sleep`。
- 禁止在同步函数中使用 `asyncio.run` 创建嵌套事件循环。

### 4.3 并发防范与循环异常
- **避免循环内捕获异常失效**: 在 `asyncio.gather` 或其他并发执行器中，必须明确配置 `return_exceptions=True`，或在每一个独立的 `Task` 内部捕获并处理所有未预期异常，防止单点崩溃导致整个并发池雪崩。
- **防止并发作用域变量污染**: 在 `for` 循环启动协程或线程时，严格防范 `NameError` 或闭包导致的变量逃逸。确保每个协程接收的参数都是闭包绑定的正确副本。

### 4.4 内存与资源防范
- **连接池管理**: 任何数据库连接、HTTP 会话（如 `httpx.AsyncClient`）或 Redis 连接，必须通过依赖注入或生命周期管理器进行全局复用，严禁在每次请求内频繁实例化与销毁 Client。
- **大对象与流处理**: 读取或生成巨大文件、向量流时，严禁一次性载入内存 (`.read()`)。必须使用 Streaming 模式并配置合理的 chunk size。

### 4.5 思维链引导 (CoT)
在涉及高并发锁（Locks）、信号量（Semaphores）或多协程竞态条件的代码修改前，必须先在 `<thinking>...</thinking>` 标签中输出资源流转图与潜在的死锁风险评估。

## 5. 数据库迁移 (Alembic)
- **禁止手动改库**: 所有 Schema 变更必须通过 Alembic。
- **原子性与可回滚**: 每个迁移脚本应包含有效的 `upgrade` 和 `downgrade`。
- **迁移流程**: 修改 `models/*.py` -> 生成脚本 -> 检查脚本 -> 预览 SQL -> 提交 PR。
- **大表变更**: 严禁执行会导致锁表的 `ALTER TABLE`。推荐：新增 NULL 列 -> 异步刷数据 -> 加约束。

## 6. 代码质量与测试
- **Linting (Ruff)**: 必须符合 Ruff 默认规则。最大行宽 120。
- **测试覆盖率**:
  - 尽最大努力保持核心业务模块的单元测试和集成测试覆盖。
- **自动化**: 提交前在 `backend` 目录下执行 `ruff check . --fix` 和 `mypy .`。

## 7. AI 友好与最佳实践
- **AI 理解力**: 使用详尽的变量名和 Docstrings。
- **依赖注入**: 充分利用 FastAPI 的 DI 系统。
- **生命周期**: 优先使用 `lifespan` 上下文管理器。
- **时间写库规范（强制）**:
  - 所有写入数据库的业务时间（例如 `created_at`、`updated_at`、`last_updated`）统一使用 `utils.timezone.now_cst_naive()`。
  - 禁止直接使用 `datetime.now()`、`datetime.utcnow()` 或 `datetime.now(timezone.utc)` 写入数据库字段。
  - 外部传入的带时区时间入库前必须先转换到东八区（可使用 `utils.timezone.to_cst_naive()`）。
  - 推荐写法：`default_factory=lambda: now_cst_naive()`。
  - 例外（JWT / 安全令牌）:
    - `exp` / `iat` / `nbf` 表示绝对时间点，必须使用 UTC timezone-aware 时间（例如 `utils.timezone.now_utc_aware()`），禁止使用 `now_cst_naive()`。

  ## 8. 提示词与工作流规范 (Prompt & Workflow)

  - **提示词管理**: 严禁在 Python 代码中硬编码长文本提示词。所有提示词必须通过 `backend/configs/atomic_snippets/` 下的 JSON 片段管理。
  - **LLM 调用**: 统一使用 `BaseLogic.run_atomic` (或其子类封装的方法) 托管 LLM 调用。禁止直接实例化 OpenAI Client 或 LlamaIndex Predictor 进行业务逻辑处理。
  ## 9. 导入完整性与依赖管理 (Import Integrity)
  - **显式检查**: 在添加新接口、函数或复杂逻辑后，必须二次检查所有引用的外部符号（如 `get_session`, `AsyncSession`, `select`, `UUID`, `Field`, `Column` 等）是否已在文件头部正确导入。
  - **避免 Undefined Name**: 严禁在代码中直接使用未导入的依赖项。
  - **循环依赖防范**:
    - 核心 Schema (`ai.py`) 或通用 Service 中，若引用了可能产生循环依赖的模块，应使用 **函数内局部导入** (Deferred Import)。
    - 严禁在模块级别 (Top-level) 相互引用 Service。
  - **验证工具**: 若对导入链有疑虑，应执行一次简单的加载测试（如 `python3 -c "from your_module import your_func"`）确保模块可解析。
  - **标准库与三方库分离**: 遵循 PEP 8，导入顺序应为：标准库 -> 三方库 -> 本地模块，各组之间空一行。