---
trigger: model_decision
description: 当数据结构发送变更时，需要生成对应的数据库迁移脚本。
---
# 数据库迁移规范 (Alembic)

## 总则
- 必须使用 **Alembic** 管理所有生产环境和开发环境的数据库 Schema 变更。
- 严禁直接在数据库中执行 DDL 语句进行结构变更。
- 严禁仅修改 `SQLModel` 模型而不生成迁移脚本。

## 迁移流程
1. **修改模型**：在 `backend/models/` 目录下修改相应的 `SQLModel` 类。
2. **生成脚本**：在 `backend` 目录下执行：
   ```bash
   export PYTHONPATH=.
   alembic revision --autogenerate -m "描述变更内容"
   ```
3. **检查脚本**：手动检查生成的 `backend/alembic/versions/xxxx.py` 文件，确保：
   - 包含必要的 `import sqlmodel`。
   - `upgrade` 和 `downgrade` 函数逻辑正确且对称。
   - 没有误删其他库（如 Llama-Index）自动生成的表。
4. **应用变更**：
   ```bash
   alembic upgrade head
   ```

## 代码与提交要求
- **SQLModel 兼容性**：在迁移文件中，如果使用了 `AutoString` 等 SQLModel 特有类型，必须确保文件顶部有 `import sqlmodel`。
- **原子性与幂等性**: 迁移脚本必须是幂等的。在 `upgrade` 中执行 DDL（如 `add_column`）时，应尽可能添加存在性检查（或使用 Alembic 的 `if_not_exists` 逻辑），以防在自动启动（Auto-Migration）场景下由于版本记录不一致导致的重复执行冲突。
- **Git 提交**：迁移脚本必须随模型变更一同提交。提交信息格式：`feat(db): 增加用户表 level 字段`。
- **原子性**：一个迁移脚本应仅包含一个逻辑上的 Schema 变更。

## 注意事项
- **环境隔离**：Alembic 配置会自动从 `backend/configs/config.py` 读取 `POSTGRES_URL`，请确保 `.env` 文件配置正确。
- **表过滤**：`env.py` 已配置忽略 `data_` 和 `index_` 开头的表（由 Llama-Index 管理），如需添加其他忽略规则，请修改 `env.py` 中的 `include_object` 函数。
- **数据迁移**：复杂的变更（如字段拆分、类型转换且需保留旧数据）需在 `upgrade` 中编写自定义的 `op.execute` 或数据处理逻辑。