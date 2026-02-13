---
trigger: always_on
description: **FastAPI（Python 3.10+）+ React/Next.js（TS）+ PostgreSQL 14+ + Redis 6+** 全栈开发核心准则
---

# 🏗️ 全栈开发核心准则 (Full-Stack Architecture Standards)

本规则体现了“**首席系统架构师 (Chief Systems Architect)**”对安全与质量的极致追求。在任何全栈决策中，必须遵循：**安全 > 质量 > 速度**。

---

## ⚖️ 风险分级与决策权重 (Risk Levels)

- **P0（阻断级）**：密钥/PII 泄露、SQL 注入、鉴权绕过、数据不可逆丢失、生产危险命令。
- **P1（必验级）**：登录/鉴权逻辑、核心业务主流程、外部 API 集成、数据库 Migration、高 QPS 路径。
- **P2（建议级）**：UI/文案微调、内部管理工具、监控指标。

---

## 🚫 硬性禁令与负向约束 (Negative Constraints)

### 1. 安全红线 (P0)
- **严禁** 明文存储或提交任何敏感信息（密钥、密码、Token、PII 数据）。
- **严禁** 使用字符串拼接构造 SQL：必须使用参数化查询或 ORM。
- **严禁** 暴露未鉴权的敏感写操作接口或批量导出接口。
- **严禁** 在生产环境执行任何具有阻塞性的 Redis 命令（如 `KEYS *`）。

### 2. 质量红线 (P1)
- **严禁** 将大规模重构与功能逻辑变更混在同一个提交/PR 中（单一职责原则）。
- **严禁** 硬编码环境变量或配置：必须通过 `.env` 或配置中心管理。
- **严禁** 在核心逻辑（P0/P1）中缺乏异常捕获与追踪点。

---

## 📂 最小工程规范 (Project Structure)

### 1. 后端 (FastAPI) 结构规范
```text
backend/
  app/
    main.py
    core/        # 核心配置/安全/DB初始化
    api/v1/      # 路由分发与依赖注入
    models/      # SQLModel/ORM 模型定义
    schemas/     # Pydantic 数据交换协议 (IO)
    services/    # 纯净业务逻辑 (Domain Logic)
    crud/        # 原子化数据库操作
    utils/       # 通用工具
  alembic/       # 数据库迁移脚本
  tests/         # 自动化测试套件
```

### 2. 前端 (React/Next.js) 结构规范
```text
frontend/
  src/
    app/ or pages/
    components/  # UI 组件 (PascalCase)
    services/    # API 客户端封装
    types/       # TS 类型定义
    utils/
    hooks/       # 自定义逻辑复用
```

---

## 📡 API 与数据规范 (API & Data)

- **RESTful 路径**: 统一使用 `/api/v1/...` 前缀。路径中**禁止**包含动词（如 `/get_user`）。
- **响应体协议**: 必须保持结构稳定。
  - 成功: `{ "code": 200, "data": {...}, "message": "success" }`
  - 失败: `{ "code": 400, "detail": "reason", "message": "error" }`
- **鉴权隔离**: 使用 Bearer Token (JWT)。所有写操作接口必须强制注入 `current_user` 依赖。

---

## 💾 数据库与缓存 (DB & Cache)

- **Migration**: 严禁手动改库。在大表同步变更时，优先采用“增量添加 -> 异步刷数 -> 约束加固”的平滑方案。
- **Redis 约束**: 必须为所有 Key 设置 **TTL**。Key 命名规范：`项目:模块:业务:ID`。单 Value 严禁超过 **1MB**。

---

## 🧪 验证与观测 (Validation & Monitoring)

- **日志**: 采用结构化日志。P0/P1 变更必须在关键分支埋点（含 `request_id`）。
- **测试**: P0 逻辑必须 100% 覆盖关键路径；P1 逻辑至少包含 1 条成功的集成测试。

---
> [!IMPORTANT]
> **架构师提示**: “过早的优化是万恶之源”，但“缺失的安全是万劫不复”。宁可代码冗长，不可逻辑存疑。

