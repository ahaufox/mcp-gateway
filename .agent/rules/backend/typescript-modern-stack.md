# 🟦 TypeScript Modern Stack

本规则旨在标准化 TypeScript 项目开发工具链与实践。

## 1. 依赖管理 (Dependency Management)
- **推荐**: 使用 `npm` 或 `pnpm`。
- **锁定文件**: `package-lock.json` 或 `pnpm-lock.yaml`。

## 2. 代码风格 (Linting & Formatting)
- **推荐**: 使用 `biome` 进行代码风格检查与格式化 (Formatting/Linting)。
- **配置**: `biome.json`
- **示例**:
  ```json
  {
    "formatter": {
      "indentStyle": "space",
      "indentWidth": 2
    }
  }
  ```

## 3. 测试 (Testing)
- **推荐**: 使用 `vitest`。
- **运行**: `npm run test` 或 `vitest`.
- **覆盖率**: `c8` 或 `v8`.

## 4. 类型检查 (Type Checking)
- **必选**: 启用 `strict: true` 在 `tsconfig.json` 中。
- **避免**: 尽量避免使用 `any`，使用 `unknown` 或具体的接口。

## 5. 项目结构 (Project Structure)
- 使用 `src/` 布局。
- 必须包含 `package.json`, `tsconfig.json`, `README.md`.
