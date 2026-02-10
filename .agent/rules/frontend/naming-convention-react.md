---
trigger: always_on
description: React/Typescript 命名与开发规范
---

# React 开发规范 (Naming Convention & Standards)

## 1. 命名规范
- **组件**: `PascalCase.tsx` (如 `AccountTable.tsx`)。
- **Hook**: `useCamelCase` (如 `useAuth.ts`)。
- **工具函数**: `camelCase` (如 `formatDate.ts`)。
- **常量**: `SCREAMING_SNAKE_CASE` (如 `MAX_RETRY_COUNT`)。

## 2. 状态管理
- **Server State**: 优先使用 `React Query`。
- **UI State**: 优先使用 `Zustand` (复查逻辑) 或 `useState` (局部逻辑)。
- **禁止**: 在组件顶层滥用 `Context` 导致全量重绘。

## 3. 样式
- **Tailwind CSS**: 优先使用语义化类名。
- **cn()**: 使用 `clsx` + `tailwind-merge` 合并动态类名。

## 4. 最佳实践
- 每个组件必须有 `Props` 类型定义。
- 逻辑较重的组件建议抽离自定义 Hook，保持视图层简洁。
