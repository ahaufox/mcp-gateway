---
description: Git 自动化套件 (Commit, Push, PR Automation)
---

# 🔄 Git 自动化套件 (Git Auto Suite)

本工作流旨在通过单条指令完成从代码变更预览到 PR 创建的全过程，极大减少手动 Git 命令的使用。

## 🛠️ 自动化步骤

1.  **准备阶段 (Stage)**: 检查当前暂存区 (Staging Area)。如果暂存区为空，必须提示用户先手动使用 `git add` 暂存需要提交的文件，本工作流**严禁进行全量自动暂存**。
2.  **变更分析 (Analyze)**: AI 自动扫描暂存区的 `diff`，提取核心变更点。
3.  **生成提交信息 (Commit Message)**: 
    - **强制要求**: 使用**中文**编写提交信息，并遵循 **Conventional Commits** 规范。
    - 示例:
      - `feat(backend): 实现安全哨兵元技能`
      - `fix(frontend): 修复登录页面在移动端的布局重叠问题`
      - `docs(global): 更新 AI 每日进化协议文档`
      - `refactor(python): 优化数据模拟逻辑并增加缓存机制`
      - `chore(infra): 升级 Docker 基础镜像版本`
4.  **本地提交 (Commit)**: 执行 `git commit -m "[Message]"`.
5.  **推送与 PR (Push & Draft PR)**:
    - 推送当前分支到远程仓库。
    - **[可选]**: 在支持的平台上（如 GitHub）自动生成 Draft PR 链接或调用 API 创建 PR。

---
> [!TIP]
> **质量检查**: 在 commit 前，自动触发 `security-guard` 技能进行扫描，如果发现安全风险将中断提交。
