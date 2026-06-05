---
trigger: model_decision
description: 前端依赖兼容性规则，包含 antd v6 + UmiJS utoopack 冲突及 rehype-raw ESM 兼容问题的已知修复方案。
---

# 依赖兼容性规则 (dependency-compat-rule.md)

> **适用范围**：所有前端项目（`frontend-v2/`、`admin-frontend/`）在引入或升级第三方依赖时，必须遵守本规则。

---

## 规则 1：admin-frontend 禁止开启 `utoopack`

### 问题描述

`admin-frontend/config/config.ts` 中若启用 `utoopack: {}` 配置（UmiJS 内置的模块打包加速器），会将 `node_modules` 拆分为若干 `node_modules_xxxx.async.js` 产物。

**antd v6+** 内部大量使用 Web Components（`connectedCallback`、`uv.append` 等）和现代 ESM 语法，与 `utoopack` 的产物切分方式存在引用链断裂，表现为：

```
Uncaught (in promise) ReferenceError: __TURBOPACK__imported__module__xxxxx__ is not defined
    at e.connectedCallback (node_modules_bcfa1be0.1023fbc0.async.js:...)
```

### 强制规则

```ts
// ❌ 禁止在 antd v6+ 环境中启用
utoopack: {},

// ✅ 正确做法：注释掉或完全删除
// utoopack: {},
```

### 配套缓解措施

移除 `utoopack` 后，需同步配置 `extraBabelIncludes`，强制让 antd v6 及其周边依赖经过 Babel 转译，确保 ESM/CJS 混用不出错：

```ts
// admin-frontend/config/config.ts
extraBabelIncludes: [
  /node_modules[\\/]antd/,
  /node_modules[\\/]@ant-design[\\/](?!icons)/,  // icons 本身已 CJS 友好，排除
  /node_modules[\\/]@rc-component/,
],
```

### 排查清单

若出现 `__TURBOPACK__imported__module__` 相关 ReferenceError，按以下顺序排查：

1. 检查 `config/config.ts` 中是否存在 `utoopack: {}`
2. 清理 UmiJS 缓存：`rm -rf src/.umi src/.umi-production .turbopack`
3. 确认 `extraBabelIncludes` 已覆盖出问题的依赖包
4. 重新执行 `npm run build`

---

## 规则 2：`rehype-raw` 禁止在客户端前端中使用

### 问题描述

`rehype-raw`（及其依赖 `hast-util-raw`）为纯 ESM 包（`"type": "module"`），在 Next.js 或相关打包工具处理时其内部动态模块引用链会断裂，导致运行时报错：

```
Uncaught ReferenceError: rehypeRaw is not defined
```

或构建时报 TypeScript 错误：

```
error TS2304: Cannot find name 'rehypeRaw'.
```

### 强制规则

```tsx
// ❌ 禁止在 frontend-v2/ (Next.js)项目中使用
import rehypeRaw from 'rehype-raw';
<ReactMarkdown rehypePlugins={[rehypeRaw]} />

// ✅ 正确做法：仅使用 remark 插件，不使用 rehype-raw
<ReactMarkdown remarkPlugins={[remarkGfm]} />
```

### 理由

本项目的法律文档 Markdown 内容不含原始 HTML，`rehype-raw` 的功能完全不需要。即便将来需要，也应等待 Next.js 对纯 ESM 包的处理方式稳定后再引入，或改用 `rehype-sanitize` 等 CJS 友好的替代方案。

### 受影响文件（历史修复记录）

- `frontend-v2/src/components/Research/FoldableInput.tsx` — 已于 2026-05-02 移除
- `frontend-v2/src/components/Intake/CaseInputPanel/EditorSection.tsx` — 已于 2026-05-02 移除

---

## 通用原则

| 场景 | 检查点 |
|------|--------|
| 引入新的 npm 包 | 检查 `package.json` 中 `"type": "module"` 是否为纯 ESM |
| 升级 antd 大版本 | 确认 UmiJS 版本是否官方声明支持该 antd 版本 |
| 构建报 `__TURBOPACK__` | 先查 `utoopack`/`mfsu` 配置，再排查 ESM-only 依赖 |
| 构建报模块找不到 | 检查是否遗漏了 import 但删除了实际使用的变量 |

---

## 规则 3：后端 (Python) 依赖与类型检查兼容性

### 问题描述
随着项目依赖增加（如 `dashscope`、`pypdf`），部分第三方库可能缺少类型定义或在特定环境下产生冲突。

### 强制规则
1. **MyPy 忽略非关键报错**：对于缺少类型存根（Type Stubs）的第三方库（如 `dashscope`），应在 `backend/mypy.ini` 中通过 `[mypy-libname.*] ignore_missing_imports = True` 进行配置，严禁在代码中滥用 `# type: ignore`（除非针对单行特定的不可避免报错）。
2. **环境隔离保护**：任何涉及全局环境变量修改（如 `TOKENIZERS_PARALLELISM`）或代理清理的代码，必须置于 `lifespan` 或应用最顶层初始化阶段，严禁在业务逻辑深处修改全局 Runtime 配置。
3. **版本锁定**：升级核心库（如 `langgraph`, `langchain-core`）后，必须立即同步运行 `pytest` 核心链路测试，确保 API 兼容性未发生 Breaking Change。
