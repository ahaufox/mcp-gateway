---
trigger: always_on
description: 测试覆盖率与质量门槛规范
---

# 测试覆盖率规范 (Test Coverage Threshold)

## 1. 核心要求
- **P0 模块**: 关键业务逻辑（支付、鉴权、核心算法）必须 **100%** 单元测试覆盖。
- **P1 模块**: 核心 API 路径必须包含集成测试，代码覆盖率建议 ≥ **80%**。
- **P2 模块**: 建议覆盖率 ≥ **60%**，重点验证 Happy Path。

## 2. 测试类型
- **Unit Test (单元测试)**: 针对独立函数、工具类。
- **Integration Test (集成测试)**: 模拟 API 请求，验证后端各层（Route -> Service -> CRUD -> DB）协作。
- **E2E Test (端到端测试)**: 模拟用户真实操作流。

## 3. CI 阻断
- 任何 PR 若导致总覆盖率大幅下降（>5%）或 P0 覆盖率下降，应触发警告。
- 严禁通过 `@pytest.mark.skip` 绕过失败的 P0 测试。
