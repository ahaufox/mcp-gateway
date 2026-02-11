# MCP Gateway 二次开发指南

本指南面向希望扩展本项目功能或添加新服务器的开发者。

## 🏗️ 核心开发流程

### 1. 添加新的 MCP 服务器

我们推荐使用 Git 子模块来管理子服务器：

```bash
git submodule add <REPOSITORY_URL>
```

添加后，请在根目录的 `README.md` 中同步更新子模块列表，并告知 `git-auto-suite` 进行提交。

### 2. 开发规范与自动化

本项目有一套严谨的开发规范，请务必遵守：

- **[规范指南 (Rules)](./readme-rules.md)**: 涵盖安全、质量、技术栈锁定及 P0/P1/P2 风险分级开发标准。
- **[自动化工作流 (Workflows)](./readme-workflow.md)**: 提供 API 变更、契约同步、测试与部署等标准化任务。

## 🤖 AI 辅助开发 (Skills)

如果你使用的是支持 AI 助手的 IDE（如 Cursor 或 Gemini Code Assist），可以利用 `.agent/skills/` 下的专项技能：

- **项目优化**: 自动扫描并提出结构化改进建议。
- **安全检查**: 在提交前自动执行安全扫描。

## 🧪 调试与测试

- **控制台输出**: 启动 `mcp-proxy` 时通过 `--log-enabled` 查看实时请求日志。
- **Mock 服务**: 使用后端数据 Mock 技能快速验证前端交互。

---
> [!IMPORTANT]
> 所有的写操作（提交、推送）建议优先通过 `/git-auto-suite` 工作流完成，以确保提交信息符合 Conventional Commits 规范。
