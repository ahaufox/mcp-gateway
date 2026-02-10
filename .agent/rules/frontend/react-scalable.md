# 可扩展 React/TS 开发规范 (Scalable React)

## 技术路径
- **核心栈**: TypeScript + React + Vite + Tailwind CSS + Shadcn UI。
- **TypeScript**: 强制使用 TS，优先使用 `interface` 而非 `type`，禁止使用 `enum`（使用 Map 替代）。

## 组件设计
- **函数式组件**: 所有组件均使用函数式声明。
- **Shadcn UI**: 充分利用 Shadcn UI 组件库构建一致的交互体验。
- **响应式**: 采用移动优先 (Mobile-First) 的设计思路。

## 性能与渲染
- **最小化 Hook**: 谨慎使用 `useEffect` 和 `setState`，尽量保持组件纯粹。
- **服务端渲染**: 优先考虑 SSR/SSG，减少客户端渲染负担。
- **并发渲染**: 利用 `Suspense` 为大数据量或异步加载的组件提供 Fallback。

## 项目协作
- **Docker 容器化**: 确保开发环境与生产环境高度一致。
- **单元测试**: 核心组件和工具函数必须有单元测试覆盖。
- **Web Vitals**: 持续关注并优化 LCP, CLS, FID 等关键性能指标。
