---
trigger: always_on
description: Git 提交信息规范，包括 Conventional Commits 格式和中文要求。
---
- `feat`: 新功能
- `fix`: 修补 Bug
- `docs`: 文档变更
- `style`: 代码格式（不影响逻辑）
- `refactor`: 重构
- `perf`: 性能优化
- `chore`: 构建过程或辅助工具的变动

简短概况，不要超过100个字符
如果有专门针对代码格式的自动化变更，请务必使用 `style` 或 `chore` 作为 type，**禁止**将其与 `feat` 或 `fix` 混合在同一个 Commit 中。