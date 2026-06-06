---
trigger: always_on
description: 用于本仓库前端（frontend-v2 + admin-frontend）的 React/TypeScript 开发架构与工程规范。frontend-v2 解耦见 frontend-v2-structure.md；视觉见 frontend-ui-design-system.md。
---

# React & 前端工程开发规范

> **视觉规范提示**：本文件主要约束前端架构、类型与状态流转。所有的 UI 组件排版、Design Token、颜色/阴影定义以及响应式约束，**必须严格遵循 `.agents/rules/frontend-ui-design-system.md`**。

## 1. 适用范围与技术栈

- **客户端（`frontend-v2/`）**：Next.js 16+ (App Router) + React 19+ + shadcn/ui + Tailwind v4 + Zustand + OpenAPI Client。
- **管理端（`admin-frontend/`）**：Ant Design Pro（Umi Max）+ antd v6。
  - **API 协作规范**：禁止手动定义后端接口服务文件。必须通过执行 `npm run openapi`（底层调用 `max openapi`）自动生成 `src/services/ant-design-pro/` 下的定义。
  - **真相源同步**：后端接口变更后，需先在 backend 生成最新的 `openapi.json` 并同步至 `admin-frontend/config/oneapi.json`，随后执行同步命令。
  - **调用习惯**：始终从 `@/services/ant-design-pro/` 导入生成的函数，入参需严格匹配生成的 `API.*` 类型。
- **TypeScript**：强制使用 `interface` 而非 `type`。禁止使用 `enum`（改用 Const Map 或 Literal Types）。
- **类型校验忽略 (Linter Ignore)**：
  - **允许使用 `any`**：在处理复杂的后端 API 响应（如 `UnifiedResponse`）、第三方库类型定义不全或临时原型开发场景下，**允许使用 `any`**。
  - **忽略建议**：Agent 在执行任务时，应**忽略** "Unexpected any. Specify a different type." 的 Linter 警告，不应为了消除此警告而引入过度复杂的泛型或类型体操，除非该位置对业务安全至关重要。
- **原生开发 (Vanilla)**：除非明确要求，优先使用原生 JS 替代外部库。HTML 必须使用语义化标签。
- **生成代码保护 (Generated Code Protection)**：
  - 严禁手动修改 `frontend-v2/src/lib/api/Api.ts` 和 `data-contracts.ts`。这些文件由 `swagger-typescript-api` 自动生成。
  - 任何接口变更应在后端修改 OpenAPI 定义后，通过运行生成脚本同步。手动修改将被后续的自动化同步覆盖。
...
## 8. 用户反馈与 Toast 规范 (Feedback & Toast)
> 旨在提供清晰、克制且不干扰操作的即时反馈。主要针对 `sonner` / `antd.message`。

### 8.1 Toast 减法规则 (Toast Reduction)
- **禁止单一操作弹出多个 Toast**：严禁在“保存并跳转”或“触发并执行”的单一链路中，连续弹出如“保存成功”、“正在触发分析”、“正在跳转”等多个气泡。
- **合并语义**：
  - **不正确**：`toast.success('已保存'); toast.info('正在分析...');`
  - **正确**：直接展示终态或最重要的中间态。若随后有页面跳转，通常只需一个“进入下一步：正在...”的提示。
- **静默跳转**：当操作结果通过页面状态（如进度条、新页面内容）能被用户直观感知时，应**取消成功类 Toast**（如：移除“研判结果已保存”等干扰性提示）。

### 8.2 工作流推进准则 (Workflow Auto-Advance)
- **即时性响应**：在工作流步骤切换时，前端应优先执行**路由跳转**。
- **正确做法**：点击“下一步” → 发起 `resume-from` 请求（非阻塞） → **瞬间完成页面切换** → 在新页面通过 `WorkflowExecutor` 展示全局进度。
- **错误做法**：在当前页面显示 Loading 或 Toast 等待 AI 分析完成后再跳转（会导致页面“死掉”或操作阻断）。

### 8.3 错误反馈
- **持久化错误**：严重的 API 错误（如 500）使用 Toast。
- **验证反馈**：表单验证、数据缺失等逻辑建议直接在 UI 元素旁显示文字提示，减少 Toast 干扰。

## 9. 开发流程与质量

- **命名**: 组件 `PascalCase.tsx`, Hooks `useCamelCase.ts`, 工具函数 `camelCase.ts`, 常量 `SCREAMING_SNAKE_CASE`。
- **组件模式**: 每个组件必须有 `interface Props` 定义。逻辑较重时抽离自定义 Hook，保持视图层纯粹。
- **组件规模限制**: `frontend-v2` 的 `page.tsx` 目标 < 150 行；其他页面/组件超过 300 行时必须拆分，业务逻辑抽离为自定义 Hook，UI 区块拆分为独立子组件。
- **响应式架构**: 采用移动优先 (Mobile-First) 布局架构。

### 2.1 frontend-v2 解耦与聚合（硬性）

> 完整条文见 **`.agents/rules/frontend-v2-structure.md`**（与 `.cursor/rules/frontend-v2-structure.mdc`、`.lingma/rules/frontend-v2-structure.md` 同步）。

- **薄 page**：只编排布局；禁止在 `page.tsx` 内 `import { api }` 发起请求。
- **数据层**：`lib/queries/`（React Query）+ `app/.../hooks/`；格式化 `lib/format.ts`；标准化 `lib/normalize/*`。
- **类型**：跨层共享放 `src/types/<domain>.ts`，禁止 `lib` 反向依赖 `app/(main)/*/types`。
- **禁止**：上帝 `app/utils.tsx`、孤儿组件、局部重复 `formatDate`、`src/api/generated` 第二套客户端。

## 3. 状态管理与性能优化

- **状态分层**:
  - **Server State**: 必须用 `React Query`，封装在 `frontend-v2/src/lib/queries/`（`queryKeys` + `useXxx`）。旧代码中的 `useSWR` 随页面重构渐进迁移，**新接口禁止**在 page 内混写 SWR。
  - **UI State**: 优先用 `frontend-v2/src/lib/store/` 下 Zustand 或本地 `useState`。禁止在顶层滥用 `Context` 导致全应用渲染穿透。
- **不可变数据 (Immutability)**: 严禁直接修改状态（Mutations），必须使用展开运算符或 `immer`。
- **性能红线**: 最小化 `useEffect` 副作用依赖，充分利用 `Suspense` 处理异步渲染边界，密切关注核心页面的 Web Vitals 指标。
- **负向约束**: 严禁在 React 状态更新函数内部（例如 `setDemands(prev => ...)`）直接放置副作用或发起 API 调用。
- **思维链引导 (CoT)**: 触发条件为在处理复杂的 UI 状态机（如：多步表单、拖拽交互）或设计组件复合模式时。必须先显式输出思考逻辑 `<thinking>分析状态流转和副作用触发条件...</thinking>`，随后再输出代码。

## 4. 错误处理规范

- **禁止使用 `err: any`**: 所有 `catch` 块中的错误参数必须使用 `err: unknown` 而非 `err: any`。
- **类型断言安全**: 使用类型断言安全地访问错误属性，如 `(err as { response?: { data?: { detail?: string } }, message?: string })`。
- **错误消息提取**: 使用统一的错误消息提取模式：优先获取 `response.data.detail`，其次获取 `response.data.message`，最后获取 `message`，并提供默认错误消息。
- **推荐实践**: 可使用 `src/lib/api/error.ts` 中定义的 `ApiError` 接口和 `extractErrorMessage` 工具函数来统一处理错误（如适用）。

## 5. 开发流程与质量

- **无占位符**: 始终实现完整功能，杜绝 TODO。
- **测试覆盖**: 核心业务组件和工具函数必须有单元测试覆盖（如适用）。

## 6. 内联编辑与状态模式

- **复合键编辑标识**: 多字段内联编辑场景下，编辑标识必须使用复合键格式（如 `${id}-fieldName` 字符串），禁止使用单一数字 ID 作为全局编辑状态，确保单字段独立编辑隔离。
- **编辑隔离**: 每个可编辑字段应拥有独立的编辑状态，避免一个字段的编辑操作影响其他字段。使用 `autoFocus` 和 `onBlur` 管理编辑进出。

## 7. 后台前端技术栈避坑指南 (Ant Design Pro / Umi)

- **国际化 (I18n) 与路由**: 严禁在 `routes.ts` 的 `name` 属性中直接硬编码中文。必须使用标准英文 Key（如 `name: 'users'`），并在 `src/locales/zh-CN/menu.ts` 中补充对应的中文翻译，以避免 `[React Intl] Missing message` 报错。
- **长文本与日志渲染**:
  - 渲染大段带有换行符的文本时，不要盲目套用 `JSON.stringify`，应先判断类型 `{typeof text === 'string' ? text : JSON.stringify(text, null, 2)}`，否则原生字符串中的换行符会被错误转义为字面量 `\n`。
  - 对于包裹长文本的 `<pre>` 标签，必须设置 `whiteSpace: 'pre-wrap'` 和 `wordBreak: 'break-word'` 以支持自动换行；并在外层容器添加 `maxHeight` 和 `overflow: 'auto'` 限制高度。
- **组件 API 迭代**: 时刻关注 Ant Design 组件库的 Deprecation 警告（例如使用 `Drawer` 时，避免使用废弃的 `width`，应改用 `size="large"`）。
- **API 响应结构 (UnifiedResponse)**: 后端接口已全面升级为 `UnifiedResponse` 标准结构。前端调用生成服务后获取到的 `res` 结构为 `{ success: boolean, data: any, message?: string }`。**必须先判断 `res.success`，然后再从 `res.data` 中解包实际数据**，严禁直接使用 `res.status === 'ok'` 判定。

## 8. 变量初始化安全与代码顺序 (Initialization Safety)

- **强约束**: 严禁在变量（特别是通过 Store Hook 获取的 `token`, `sessionId`, `recordId` 等）声明之前引用它们。严禁触发 JavaScript 的“暂时性死区”(TDZ) 导致 `ReferenceError`。
- **组件内部 Hook 排序准则 (必须遵守)**:
  1. **基础环境 Hooks**: `useRouter`, `useParams`, `usePathname`。
  2. **全局/业务 Store**: `useAuthStore`, `useSessionStore`, `usePageRecord` (获取基础上下文)。
  3. **基础计算变量**: 基于上述 Store 计算的常量（如 `apiBase = ...`, `effectiveId = ...`）。
  4. **复杂业务 Hooks**: `useWorkflow`, `useAgent`, `useWorkflowAutoAdvance` (这些 Hook 通常依赖上述所有变量)。
  5. **本地组件状态**: `useState`, `useRef`, `useMemo`, `useCallback`。
  6. **副作用**: `useEffect`。
- **验证要求**: 在调用任何自定义 Hook 之前，必须确保其构造参数中使用的所有本地变量已在上方完成初始化。