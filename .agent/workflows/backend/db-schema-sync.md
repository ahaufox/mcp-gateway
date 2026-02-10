---
description: 自动同步后端模型 (SQLModel)、Pydantic Schema 与数据库状态。
---

# DB 架构同步工作流 (DB Schema Sync)

## 场景
当你修改了 `models/*.py` 中的数据库模型定义后。

## 执行步骤

### 1. 校验模型
- **动作**: 运行 `mypy` 或类型检查，确保 `SQLModel` 定义无语误。

### 2. 生成迁移
- **命令**: `alembic revision --autogenerate -m "Sync schema for ..."`
- **动作**: 检查生成的脚本。确保新增字段有注释，删除字段是预期的。

### 3. 同步 Schema
- **动作**: 如果项目使用了导出 OpenAPI 或 JSON Schema 到前端的操作，运行 `python scripts/export_schemas.py`。
- **目的**: 确保前端 TypeScript 类型定义与后端模型一致。

### 4. 验证
- **动作**: 运行 `pytest` 中的 DB 相关测试，确保迁移后数据增删改查正常。
