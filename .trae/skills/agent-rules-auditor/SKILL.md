---
name: agent-rules-auditor
description: 审计并同步面向不同编辑器(Cursor, Windsurf, Claude, Cline/Roo, OpenCode, Lingma, Trae)的规则、配置及 subagents / Multi-editor rules & config consistency auditor。适用场景：修改了 .agents/rules/ 中的 SSOT 规则后需要同步到各编辑器平台、验证多编辑器环境下的配置一致性、检查规则镜像是否冲突或过期。关键词：规则同步/多编辑器/Cursor规则/Claude规则/OpenCode/配置一致性审计/rule sync/multi-editor/config audit/cross-platform rules/sync verification。
---

# Agent 规则审计员 (Agent Rules Auditor)

本技能用于维护项目中多个 AI 编辑器规则文件的一致性。

## 核心职责

1. **同步验证**: 确保 `.agents/rules/` (SSOT) 中的所有规则已正确同步到 `.cursor/rules/`, `.windsurf/rules/`, `.claude/rules/` 等目录。
1. **格式转换**: 检查 Cursor 特有的 `.mdc` 格式是否包含必要的 `globs` 字段，并确保 body 内容与 master 一致。
1. **SSOT 维护**: 提醒开发者优先修改 `.agents/rules/` 目录，而非直接修改镜像规则。

## 使用流程

### 1. 执行审计脚本
使用配套脚本检查当前项目的规则同步状态：

```bash
python .agents/skills/agent-rules-auditor/scripts/audit_rules.py
```text

### 2. 处理同步差异
如果审计报告显示有 `Desync` 或 `Missing rule`：
- **缺失规则**: 将 Master 规则复制到目标目录，并根据目标编辑器调整后缀和 frontmatter。
- **内容差异**: 以 Master 内容为准覆盖镜像文件，保留目标编辑器特有的 frontmatter 字段（如 Cursor 的 `globs`）。

## 参考资料
- 查看 [config_map.md](references/config_map.md) 了解编辑器与规则目录的详细映射关系。
- 审计逻辑详情请见 [audit_rules.py](scripts/audit_rules.py)。

## 常见问题
- **为什么 SSOT 是 .agents/rules/?** 因为 `AGENTS.md` 明确规定了这是唯一的权威来源。
- **什么时候触发审计?**
  - 在修改任何 `.agents/rules/*.md` 文件后。
  - 在创建新的规则后。
  - 在准备提交 PR 之前。

## 核心约束与执行标准

### 严格边界与禁止项
- **安全 > 质量 > 速度**：任何全栈决策和代码修改必须遵循此总原则。
- **禁止项**：严禁在未经用户确认的情况下删除核心业务规则。
- **禁止项**：严禁输出虚假或未经测试的配置。

## 思维链引导

在执行复杂步骤、架构设计或遇到歧义时，必须通过 `<thought>` 标签显式输出你的思考过程和逻辑推导，然后再输出最终行动或代码方案。
