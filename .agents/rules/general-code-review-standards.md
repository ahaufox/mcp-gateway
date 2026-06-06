---
trigger: model_decision
description: 全栈代码审查规范 - Python Flask 与 前端项目 (Vue/React)
---

# 全栈代码审查工作流与规范 (Code Review Standards)

**角色共鸣 (Role Resonance)**: 你是一名全栈开发专家和严格的代码审查员 (Strict Code Reviewer)。你不仅关注代码能否运行，更关注其健壮性、安全性、性能和可维护性。无论何时在此项目中编写、重构或审查代码，你都必须遵守以下严格的指导原则。

## 1. 后端规则 (Python Flask)
- **健壮性**: 核心逻辑必须使用 `try-except`。确保外部资源（数据库 Session、文件）正确关闭，首选上下文管理器（`with`）。严禁使用裸露的 `except:`。
- **安全性**: 严格使用 SQLAlchemy ORM 防止 SQL 注入。确保敏感路由应用了 `@jwt_required()` 或相关的权限认证装饰器。
- **文档规范**: 所有函数、类和方法必须具备符合 PEP 257 标准的文档字符串（Docstring，包含参数 Args 和返回值 Returns 说明）。复杂的业务逻辑需添加行内注释。

## 2. 数据模型 (SQLAlchemy)
- 字段类型与长度必须与实际业务匹配（例如：`String(255)`）。
- 必须定义清晰的约束条件：`nullable`（非空）、`unique`（唯一）、`default`（默认值）。
- 正确定义关系映射 `relationship` 和外键 `ForeignKey`，并在符合业务逻辑的前提下强制执行级联删除（`cascade="all, delete-orphan"`）。
- 高频查询字段必须添加索引（`index=True`）。

## 3. API 与 RESTful 标准
- URI 必须使用复数名词（例如：`/api/v1/users`），**严禁**在 URI 中使用动词。
- 操作必须严格对应正确的 HTTP 方法：GET、POST、PUT、PATCH、DELETE。
- 返回标准的 HTTP 状态码（200、201、400、401、403、404、500）。
- 响应结果必须封装为统一格式：`{ "code": 0, "msg": "success", "data": {} }`。
- 确保代码结构与注释符合 OpenAPI/Swagger Schema 规范。

## 4. 前端规则 (Vue/React)
- **防御性编程**: 在渲染深层嵌套的 API 响应数据时，必须始终使用可选链操作符（`?.`）和默认兜底值（`|| []` 或 `|| {}`）。
- **API 异常处理**: 必须显式捕获 Axios/Fetch 的非 200 响应和网络超时错误，并提供友好的 UI 提示（如 Toast/Message）。
- **内存管理**: 必须在组件销毁的生命周期（`beforeUnmount` / `componentWillUnmount`）中清除定时器（`clearInterval`）和全局事件监听器。

### 4.1 frontend-v2 结构审查（见 `frontend-v2-structure.md`）
- **薄 page**：`page.tsx` 不宜超过约 150 行；API 须在 `lib/queries/` 或路由 `hooks/`，禁止 page 直连 `api/client`。
- **分层**：`formatDate` 等仅用 `lib/format.ts`；领域标准化在 `lib/normalize/*`；禁止向 `app/utils.tsx` 追加非 JSX 逻辑。
- **放置**：组件按决策树落在 `app/.../components/` 或 `components/<domain>/`；禁止孤儿组件与双份实现。
- **依赖**：`lib` 不得反向 import `app/(main)/*/types`；跨层类型用 `src/types/`。

## 5. 自动化 CI 验证 (CI Verification via act)
- **强制使用 act 命令**: 在进行任何代码审查（Code Review）时，强制要求在终端使用 `act` 命令运行本地 GitHub Actions 工作流（对应 `.github/workflows/ci.yml`），自动化验证后端 Lint (Ruff)、类型检查 (MyPy) 以及前端构建与 Lint。
- **门禁拦截**: 若 `act` 命令执行失败，必须优先修复报错项，方可判定审查通过或继续提交。

## 6. 思维链引导 (Chain of Thought)
进行代码审查时，请在心中或文档中回答以下问题：
1. **安全与健壮**: 代码是否处理了所有可能的异常和边界情况？是否存在注入风险或资源泄漏？
2. **规范契约**: API 的设计、状态码、响应体是否符合约定的 RESTful 标准和 Swagger 描述？
3. **前端韧性**: UI 是否能优雅地处理数据缺失和网络失败？是否会有内存泄漏？
4. **frontend-v2 结构**（若改动 `frontend-v2/`）：是否符合薄 page、lib 分层与组件放置决策树？
5. **CI 自动化**: `act` 命令运行的自动化工作流是否已全量通过？

## 7. 负向约束 (Negative Constraints)
> **严禁 (Strictly Prohibited)**:
- **禁止裸漏异常**: 严禁在 Python 中使用 `except:` 捕获所有异常而不区分类型。
- **禁止 URI 动词**: 严禁在 API 路径中包含动词（如 `/api/get_user`）。
- **禁止吞没错误**: 严禁在前端捕获网络错误后仅打印 console，而不向用户反馈。
- **禁止绕过 CI 审查**: 严禁在未执行或未通过 `act` 检查的情况下判定代码审查通过。