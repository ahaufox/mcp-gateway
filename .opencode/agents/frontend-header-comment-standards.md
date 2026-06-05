---
trigger: always_on
description: 前端文件头部注释规范，要求所有前端源文件必须在顶部声明职责注释块，并在代码逻辑变更时同步更新注释内容。
---

# 前端文件头部注释规范

## 1. 适用范围

本规范适用于仓库内所有前端工程的核心业务源文件：

- **客户端（`frontend-v2/`）**：`src/` 下的 Hooks（`hooks/`、`app/**/hooks/`）、组件（`components/`、`app/**/components/`）、`lib/queries/`、`lib/normalize/`、`lib/format.ts`、`lib/store/`（含 `use*Store.ts`）。`app/utils.tsx` 仅作兼容 re-export，新逻辑注释应写在实际实现文件。
- **管理端（`admin-frontend/`）**：`src/` 下的同等类别文件（Hooks、Pages、Services、Utils、Models）。

**豁免范围**：纯配置文件（如 `vite.config.ts`、`tsconfig.json`）、自动生成文件、入口文件（`main.tsx`、`index.tsx`）、类型声明文件（`*.d.ts`）。

## 2. 强制性要求

**每个符合条件的源文件，必须在文件顶部（import 语句之前）放置一个注释块，说明该文件的核心信息。**

该注释块是文件不可分割的一部分，必须在生成或修改代码时同步维护。

## 3. 注释块结构

注释块应包含以下层级：

```typescript
/**
 * 文件名 / Hook 名 / 模块名 — 一句话中文简述
 *
 * 职责：用 1~3 句话描述该文件/模块的核心职责与对外暴露的能力。
 *
 * 变更说明（可选，但强烈建议）：
 *   - 当内部实现发生重大变更时（如状态管理方案切换、API 重构），
 *     在此处用简洁的条目记录变更内容，保持历史可追溯性。
 *   - 格式参考：
 *     - 原内部 useState (facts / reportData) → 全部委托给 useAnalysisStore
 *     - facts 类型由 any 升级为 CaseFacts | null
 */
```

### 3.1 必填字段

| 字段 | 说明 | 示例 |
|------|------|------|
| **标题行** | 文件名或导出名 + 破折号 + 一句话简述 | `useAnalysisFlow — 分析流程核心 Hook` |
| **职责** | 描述该文件负责什么、对外暴露什么 | `职责：封装 startAnalysis / runDiagnosis 等流式 API 调用逻辑，以及 cancel / regenerate 等辅助操作。` |

### 3.2 选填字段

| 字段 | 说明 | 触发条件 |
|------|------|----------|
| **变更说明** | 记录内部实现的重大变更 | 状态管理方案切换、API 调用方式重构、类型签名升级、核心算法替换 |
| **依赖说明** | 记录强依赖的外部模块或全局状态 | 依赖特定 Store、特定 Context、特定 API 版本 |

## 4. 示例

### 4.1 Hook 文件

```typescript
/**
 * useAnalysisFlow — 分析流程核心 Hook
 *
 * 职责：封装 startAnalysis / runDiagnosis / confirmAudit / generateReport 等
 * 流式 API 调用逻辑，以及 cancel / regenerate / saveSupplement 等辅助操作。
 *
 * 变更说明（Zustand 迁移）：
 *   - 原内部 useState (facts / reportData / isGenerating / recordId / analysisProgress)
 *     → 全部委托给 useAnalysisStore
 *   - 对外 return 接口保持不变（组件无感知）
 *   - facts 类型由 any 升级为 CaseFacts | null
 */
```

### 4.2 组件文件

```typescript
/**
 * CaseDetailPage — 案件详情主页面
 *
 * 职责：聚合案件基本信息、证据树、分析结果面板，并提供导航至子模块（取证、报告、审计）的入口。
 * 依赖 useAnalysisFlow Hook 驱动分析状态，通过 useParams 获取案件 ID。
 */
```

### 4.3 API 服务文件

```typescript
/**
 * caseApi — 案件相关 API 封装
 *
 * 职责：提供 createCase / fetchCases / deleteCase / updateCaseMeta 等 RESTful 请求方法。
 * 所有方法均返回 TypedResponse<T>，并在网络异常时统一抛出 ApiError。
 *
 * 变更说明（API 版本升级）：
 *   - 响应结构由 { status, data } 升级为 { success, data, message }（UnifiedResponse）
 *   - 新增批量删除接口 batchDeleteCases
 */
```

### 4.4 Store 文件

```typescript
/**
 * useAnalysisStore — 分析流程全局状态管理
 *
 * 职责：集中管理分析任务的生命周期状态（facts / reportData / isGenerating / recordId / progress），
 * 并提供 setFacts / setReport / resetAnalysis 等 dispatch 方法，供 useAnalysisFlow 及各消费组件调用。
 * 基于 Zustand 实现，支持 devtools 调试。
 */
```

## 5. 维护规则

### 5.1 新建文件时

- 必须在创建文件的同时编写顶部注释块。
- 标题行与职责描述必须与实际功能相符，不得使用模糊描述（如"这是一个工具文件"）。

### 5.2 代码逻辑变更时

**当发生以下类型的变更时，必须同步更新顶部注释：**

| 变更类型 | 是否必须更新 | 更新内容 |
|----------|-------------|----------|
| 职责范围扩大/缩小 | ✅ 必须 | 修改"职责"段落 |
| 状态管理方案变更 | ✅ 必须 | 在"变更说明"中追加条目 |
| API 接口/调用方式重构 | ✅ 必须 | 在"变更说明"中追加条目 |
| 核心类型签名升级 | ✅ 必须 | 在"变更说明"中追加条目 |
| 新增/删除对外暴露的方法 | ✅ 必须 | 修改"职责"段落或追加变更说明 |
| 纯 UI 样式/文案调整 | ❌ 不必须 | 无需更新 |
| 内部重构（对外接口不变） | 建议 | 如改动较大，建议记录 |

### 5.3 更新格式

- 在"变更说明"段落末尾追加新条目，保留历史记录。
- 每条变更记录应包含"原方案 → 新方案"的对比，或简明描述变更内容。
- 可在括号中注明变更标签，如 `(Zustand 迁移)`、`(API v2 升级)`。

## 6. 质量检查（Checklist）

在代码审查或生成代码时，应检查：

- [ ] 文件顶部是否存在注释块（import 语句之前）？
- [ ] 标题行是否清晰表达了文件名/导出名与核心职责？
- [ ] "职责"段落是否准确描述了当前功能（而非过时描述）？
- [ ] 如近期有重大变更，是否在"变更说明"中记录？
- [ ] 注释内容是否与实际代码一致（无过时信息）？