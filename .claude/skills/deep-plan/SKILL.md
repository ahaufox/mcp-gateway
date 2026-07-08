---
name: "deep-plan"
description: "深度方案规划工作流 / Deep Planning with Implementation Plan template (同 speckit-plan 但更注重深度分析)。使用实施方案模板执行深度方案规划，生成详细设计文档与架构制品。适用场景：需要比标准 Plan 更深入的架构分析、复杂功能需要多维度方案评估、需要输出完整实施方案文档供评审、架构决策记录(ADR)创建。关键词：深度规划/方案规划/架构设计/实施方案/设计文档/deep plan/solution planning/architecture design/implementation plan/plan template/ADR。"
argument-hint: "规划阶段的可选指导"
compatibility: "需要 spec-kit 项目结构，包含 .specify/ 目录"
metadata:
  author: "github-spec-kit"
  source: "templates/commands/plan.md"
user-invocable: true
disable-model-invocation: false
---

## 用户输入

```text
$ARGUMENTS
```text

你**必须**在继续之前考虑用户输入（如果非空）。

## 前置检查

**检查扩展钩子（规划前）**：
- 检查项目根目录下是否存在 `.specify/extensions.yml`
- 如果存在，读取它并查找 `hooks.before_plan` 键下的条目
- 如果 YAML 无法解析或无效，静默跳过钩子检查并正常继续
- 过滤掉 `enabled` 显式为 `false` 的钩子。将没有 `enabled` 字段的钩子视为默认启用
- 对于每个剩余的钩子，**不要**尝试解释或评估钩子的 `condition` 表达式：
  - 如果钩子没有 `condition` 字段，或为 null/空，则视该钩子为可执行
  - 如果钩子定义了非空的 `condition`，跳过该钩子，将条件评估留给 HookExecutor 实现
- 从钩子命令名构造斜杠命令时，将点号（`.`）替换为连字符（`-`）。例如 `speckit.git.commit` → `/speckit-git-commit`
- 对于每个可执行的钩子，根据其 `optional` 标志输出以下内容：
  - **可选钩子**（`optional: true`）：
    ```text
    ## 扩展钩子

    **可选前置钩子**：{extension}
    命令：`/{command}`
    描述：{description}

    提示：{prompt}
    执行：`/{command}`
    ```
  - **强制钩子**（`optional: false`）：
    ```text
    ## 扩展钩子

    **自动前置钩子**：{extension}
    执行：`/{command}`
    EXECUTE_COMMAND：{command}

    在继续到大纲之前等待钩子命令的结果。
    ```
- 如果没有注册任何钩子或 `.specify/extensions.yml` 不存在，则静默跳过

## 大纲

1. **设置**：从仓库根目录运行 `.specify/scripts/bash/setup-plan.sh --json`，解析 JSON 获取 FEATURE_SPEC、IMPL_PLAN、SPECS_DIR、BRANCH。对于参数中的单引号（如 "I'm Groot"），使用转义语法：例如 'I'\''m Groot'（或尽可能使用双引号："I'm Groot"）。

1. **加载上下文**：读取 FEATURE_SPEC 和 `.specify/memory/constitution.md`。加载 IMPL_PLAN 模板（已复制）。

1. **执行规划工作流**：遵循 IMPL_PLAN 模板中的结构：
   - 填写技术上下文（将未知项标记为"需要澄清"）
   - 从章程填写章程检查章节
   - 评估关卡（如果违规无正当理由则报错）
   - 阶段 0：生成 research.md（解决所有需要澄清项）
   - 阶段 1：生成 data-model.md、contracts/、quickstart.md
   - 阶段 1：通过运行智能体脚本更新智能体上下文
   - 设计后重新评估章程检查

## 强制后置执行钩子

**你必须在向用户报告完成之前完成本节。**

检查项目根目录下是否存在 `.specify/extensions.yml`。
- 如果不存在，或没有在 `hooks.after_plan` 键下注册任何钩子，跳到完成报告
- 如果存在，读取它并查找 `hooks.after_plan` 键下的条目
- 如果 YAML 无法解析或无效，静默跳过钩子检查并继续到完成报告
- 过滤掉 `enabled` 显式为 `false` 的钩子。将没有 `enabled` 字段的钩子视为默认启用
- 对于每个剩余的钩子，**不要**尝试解释或评估钩子的 `condition` 表达式：
  - 如果钩子没有 `condition` 字段，或为 null/空，则视该钩子为可执行
  - 如果钩子定义了非空的 `condition`，跳过该钩子，将条件评估留给 HookExecutor 实现
- 从钩子命令名构造斜杠命令时，将点号（`.`）替换为连字符（`-`）。例如 `speckit.git.commit` → `/speckit-git-commit`
- 对于每个可执行的钩子，根据其 `optional` 标志输出以下内容：
  - **强制钩子**（`optional: false`）——**你必须为每个强制钩子输出 `EXECUTE_COMMAND:`**：
    ```text
    ## 扩展钩子

    **自动钩子**：{extension}
    执行：`/{command}`
    EXECUTE_COMMAND：{command}
    ```
  - **可选钩子**（`optional: true`）：
    ```text
    ## 扩展钩子

    **可选钩子**：{extension}
    命令：`/{command}`
    描述：{description}

    提示：{prompt}
    执行：`/{command}`
    ```

## 完成报告

命令在阶段 2 规划后结束。报告分支、IMPL_PLAN 路径和生成的制品。

## 阶段

### 阶段 0：大纲与研究

1. **从上述技术上下文中提取未知项**：
   - 对每个需要澄清项 → 研究任务
   - 对每个依赖项 → 最佳实践任务
   - 对每个集成项 → 模式任务

1. **生成并派发研究代理**：

   ```text
   对技术上下文中的每个未知项：
     任务："为 {功能上下文} 研究 {未知项}"
   对每个技术选择：
     任务："寻找 {领域} 中 {技术} 的最佳实践"
   ```text

1. **在 `research.md` 中整合发现**，使用以下格式：
   - 决策：[选择什么]
   - 理由：[为什么选择]
   - 考虑的替代方案：[还评估了什么]

**输出**：包含所有需要澄清项已解决的 research.md

### 阶段 1：设计与契约

**先决条件：** `research.md` 完成

1. **从功能 spec 提取实体** → `data-model.md`：
   - 实体名称、字段、关系
   - 需求的验证规则
   - 如果适用，状态转换

1. **定义接口契约**（如果项目有外部接口） → `/contracts/`：
   - 识别项目向用户或其他系统暴露的接口
   - 记录适合项目类型的契约格式
   - 示例：库的公共 API、CLI 工具的命令 Schema、Web 服务的端点、解析器的语法、应用程序的 UI 契约
   - 如果项目纯粹是内部的（构建脚本、一次性工具等），跳过

1. **创建快速启动验证指南** → `quickstart.md`：
   - 记录可运行的验证场景，证明功能端到端工作
   - 包括先决条件、设置命令、测试/运行命令和预期结果
   - 使用指向契约和数据模型细节的链接或引用，而非重复
   - 不包含完整实现代码、模型/服务/控制器主体、迁移或完整测试套件
   - 保持此制品作为验证/运行指南；实现细节属于 `tasks.md` 和实施阶段

1. **智能体上下文更新**：
   - 更新 `CLAUDE.md` 中 `<!-- SPECKIT START -->` 和 `<!-- SPECKIT END -->` 标记之间的计划引用，指向步骤 1 中创建的计划文件（IMPL_PLAN 路径）

**输出**：data-model.md、/contracts/*、quickstart.md、更新后的智能体上下文文件

## 关键规则

- 文件系统操作使用绝对路径；文档和智能体上下文文件中的引用使用项目相对路径
- 在关卡失败或未解决的澄清时报错

## 完成条件

- [ ] 规划工作流已执行，设计制品已生成
- [ ] 扩展钩子已根据上述强制后置执行钩子中的规则分发或跳过
- [ ] 完成报告已向用户呈现分支、计划路径和生成的制品
