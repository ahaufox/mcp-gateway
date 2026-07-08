---
trigger: always_on
---

# Git 提交规范 (Conventional Commits)

- `feat`: 新功能
- `fix`: 修补 Bug
- `docs`: 文档变更
- `style`: 代码格式（不影响逻辑）
- `refactor`: 重构
- `perf`: 性能优化
- `chore`: 构建过程或辅助工具的变动

提交信息必须使用中文撰写。
简短概况，不要超过 100 个字符。
如果有专门针对代码格式的自动化变更，请务必使用 `style` 或 `chore` 作为 type，**禁止**将其与 `feat` 或 `fix` 混合在同一个 Commit 中。

## 相关技能 (Related Skills)

执行本规则前，应加载以下技能：
- **`infra-git-auto-suite`**：Git 自动化套件（commit→PR 全流程）。本规则定义的 Conventional Commits 格式可由该技能在选择性提交与 PR 创建时自动校验与分类。