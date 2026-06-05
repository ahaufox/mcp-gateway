---
trigger: always_on
description: frontend-v2 目录结构、薄 page、lib/format、lib/normalize、lib/queries 解耦与聚合开发规范。在 frontend-v2 内新增或重构代码时必须遵循。
---

# frontend-v2 架构解耦与聚合规范

> 详细说明见仓库文档：`frontend-v2/docs/frontend-structure.md`
> 本规则与 `.cursor/rules/frontend-v2-structure.mdc`、`.lingma/rules/frontend-v2-structure.md` 内容一致。

## 1. 依赖方向（硬性）

```text
src/types/<domain>.ts          ← 跨层共享领域类型
src/lib/format.ts              ← 纯格式化，零业务类型依赖
src/lib/normalize/*.ts         ← 领域数据标准化，可依赖 src/types
src/lib/queries/*.ts           ← Server State（React Query），唯一走 lib/api/client
src/lib/store/*.ts             ← Zustand 客户端 UI/会话状态
src/hooks/*.ts                 ← 跨路由 Hook（工作流、Record 等）
src/components/<domain>/       ← 跨 ≥2 路由复用的 UI
src/app/(main)/<route>/        ← 薄 page、路由内 hooks/components/types
```

- **允许**：`app` → `components` / `hooks` / `lib` / `types`
- **禁止**：`lib/*` 或 `components/*` 反向 `import` `@/app/(main)/.../types`（类型须上迁 `src/types/`）
- **禁止**：在 `page.tsx` 内直接 `import { api } from '@/lib/api/client'` 发起请求（须经 `lib/queries` 或路由 `hooks/`）

## 2. 目录放置决策树

| 条件 | 放置位置 |
|------|----------|
| 仅单一路由、与 URL/recordId 强绑定 | `app/(main)/<route>/components/` |
| 被 ≥2 路由或 layout 使用 | `components/<domain>/` |
| 跨域 UI（Loading、PageHeader） | `components/shared/` |
| 纯数据、无 JSX | `app/.../hooks/` 或 `lib/queries/` |
| 日期/文件大小等纯函数 | `lib/format.ts`（禁止在业务文件内重复实现 `formatDate`） |
| 领域 normalize（parseDisputes 等） | `lib/normalize/<domain>.ts` |
| 类型仅单路由 | `app/.../types.ts`；跨层共享 → `src/types/<domain>.ts` |

## 3. 薄 page + 聚合 Hook

- **`page.tsx` 目标**：< 150 行，仅负责布局编排与组合子组件，不写大段 JSX 与 API 样板。
- **数据与副作用**：抽到 `app/(main)/<route>/hooks/use*.ts` 或 `lib/queries/*.ts`。
- **Server State**：新页面/新接口**必须**使用 `@tanstack/react-query`，封装在 `lib/queries/`（`queryKeys` 工厂 + `useXxx`）。
- **研判类页面**：优先复用 `usePageRecord`（`app/(main)/usePageRecord.ts`），勿重复 `useUrlRecordId` + `useRecord` 两步样板。
- **领域只读 API**：可仿 `app/(main)/dashboard/lib/dashboard.ts`，逐步迁移为 `lib/queries/`。

## 4. 工具与兼容层

| 模块 | 职责 | 新代码应 |
|------|------|----------|
| `lib/format.ts` | `formatDate`、`stripMarkdown`、`formatFileSize` 等 | **直接 import** |
| `lib/normalize/*` | `parseDisputes`、`extractXxxFromRecord` 等 | **直接 import** |
| `app/utils.tsx` | 兼容 re-export + 含 JSX 的证据审计展示辅助 | **勿追加**领域逻辑 |
| `lib/utils.ts` | `cn`、`formatStatuteTitle` 等 shadcn 工具 | 样式类名合并 |

## 5. 反模式（严禁）

| 反模式 | 说明 |
|--------|------|
| 上帝 `app/utils.tsx` | 继续堆领域函数；应拆入 `lib/normalize` |
| 孤儿组件 | 在 `components/<domain>/` 建文件但 page 未 import |
| 双份实现 | page 内联 UI 与已抽离组件并存 |
| 局部 `formatDate` | 在 cases/customers 等文件内复制日期格式化 |
| 页面内直连 API | `page.tsx` 中 `useSWR` + `api.xxx()` 混写 |
| 第二套 API 生成物 | `src/api/generated/`、 `lib/api/modular_test/` |
| Store 分裂 | ASR 等全局状态须放在 `lib/store/`，禁止 `src/store/` 孤立目录 |

## 6. 代码审查 Checklist（frontend-v2）

- [ ] `page.tsx` 是否薄编排（< 150 行为佳）？
- [ ] 新 API 是否经 `lib/queries` 或路由 `hooks/`，而非 page 直连 `api/client`？
- [ ] 格式化是否复用 `lib/format.ts`？
- [ ] 领域标准化是否在 `lib/normalize`，且未反向依赖 `app` 路由 types？
- [ ] 新组件是否按决策树放置，且无孤儿文件？
- [ ] 是否未向 `app/utils.tsx` 追加非 JSX 逻辑？

## 7. 正向样板（复制而非另起炉灶）

- `settings/page.tsx` + `settings/hooks/` + `components/settings/*`
- `dispute-risk/page.tsx` + `app/.../components/`
- `usePageRecord`、`hooks/useWorkflow`
- `lib/queries/profile.ts`、`lib/normalize/dispute-risk.ts`
