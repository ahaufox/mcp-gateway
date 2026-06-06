---
trigger: model_decision
description: 前端 JSX/TSX 中引号及特殊字符的转义规范，防止编译与 ESLint 报错。
---

# 前端引号转义规范 (Frontend Quote Escape Rule)

## 1. JSX 文本中的引号与特殊字符转义
- 在 JSX/TSX 模板的直接文本内容中，**禁止**直接书写 `'` (单引号)、`"` (双引号)、`<`、`>`、`{`、`}` 等字符，避免触发 ESLint (`react/no-unescaped-entities`) 或编译错误。
- 必须使用以下两种方式之一进行处理：
  1. **HTML 实体转义**：
     - 单引号 `'` -> `&apos;`
     - 双引号 `"` -> `&quot;`
     - 左尖括号 `<` -> `&lt;`
     - 右尖括号 `>` -> `&gt;`
  2. **JSX 表达式包裹**：将整段或含引号的字符串放入 `{"..."}` 中。

- **示例**：
  ```tsx
  // ❌ 错误做法
  <div>It's a error.</div>

  //  正确做法 (使用 HTML 实体)
  <div>It&apos;s a correct way.</div>

  //  正确做法 (使用 JSX 表达式)
  <div>{"It's a correct way."}</div>
  ```

## 2. 属性值中的引号
- 在 JSX 属性（如 `title="..."`）中，使用双引号包裹属性值时，内部可直接书写单引号，无需转义；若包含双引号，则需使用 JSX 表达式。
- **示例**：
  ```tsx
  //  正确做法
  <button title="Click 'OK' to proceed" />

  //  正确做法
  <button title={'Click "OK" to proceed'} />
  ```
