---
trigger: always_on
description: 数据库迁移规范 (Alembic/SQLModel)
---

# 数据库迁移规范 (Database Migration Policy)

## 1. 核心原则
- **禁止手动修改数据库**: 所有 Schema 变更必须通过 Alembic 迁移脚本。
- **原子性**: 一个 PR 对应一个迁移脚本，迁移脚本应只包含该 PR 所需的变更。
- **可回滚**: 每个迁移脚本必须包含有效的 `upgrade` 和 `downgrade` 函数。

## 2. 迁移流程
1. 修改 `models/*.py` 定义模型变更。
2. 运行 `alembic revision --autogenerate -m "description"` 生成脚本。
3. **关键：** 检查自动生成的脚本，确保没有误判。
4. 运行 `alembic upgrade head --sql` 预览生成的 SQL，确保安全。
5. 提交 PR 时附带迁移脚本。

## 3. 大表变更 (P1)
- 严禁在大表上执行 `ALTER TABLE ... ADD COLUMN ... DEFAULT ...`（会导致锁表）。
- 推荐：先新增允许 NULL 的列 -> 异步刷数据 -> 添加 NOT NULL 约束。
