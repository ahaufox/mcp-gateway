# 多代理共识评审规则 (Agentic Consensus Rule)

在处理复杂任务或高风险决策时，AI 助手必须通过内部多角度交叉验证，以降低幻觉（Hallucinated Suggestion）和单一视角的噪音。

## 1. 角色模拟协议 (Role Simulation)

当触发“高复杂度 (High Complexity)”或“P0/P1 重构”任务时，内部应模拟以下三个核心角色的对抗性讨论：
1.  **实现者 (Implementer)**: 关注功能实现的速度和简洁性。
2.  **评审员 (Reviewer - Bug Hunter)**: 极其挑剔，专注于边缘案例、逻辑死循环、性能瓶颈。
3.  **架构师 (Architect)**: 关注对既有模式的破坏性、未来维护成本及扩展性。

## 2. 共识机制 (Consensus Mechanism)

- **输出要求**: 最终方案必须包含 [Consensus] 区块，总结三方讨论的结论。
- **冲突处理**: 
  - 如果角色间出现分歧，必须显式向用户展示讨论详情并标记 **[DECISION REQUIRED]**。
  - 只有达成 2/3 以上共识的方案才应自动执行。

## 3. 适用场景 (Applicability)

- **大型重构**: 涉及 5 个以上文件的改动。
- **安全变更**: 涉及鉴权逻辑、加密算法或核心中间件。
- **数据库 Schema 变更**: 涉及大表迁移或迁移脚本。

## 4. 输出示例 (Internal CoT Example)

> **[Internal Discussion Summary]**
> - **Reviewer**: 担心在异步环境下 `global_lock` 会导致死锁。
> - **Architect**: 建议使用 `Redis Distributed Lock` 替代。
> - **Implementer**: 同意，虽然成本略高但更安全。
> **[Consensus]**: 采用 Redis 锁方案。

---
> [!IMPORTANT]
> 慢就是快。通过增加内部博弈时间，减少回滚和修复 Bug 的成本。
