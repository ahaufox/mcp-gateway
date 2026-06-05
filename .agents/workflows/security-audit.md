---
description: 定期安全合规检查，扫描依赖漏洞与硬编码敏感信息。
---

# 安全审计工作流 (Security Audit)

## 执行步骤

### 1. 扫描依赖漏洞
- **Python**: 运行 `safety check`。
- **Node.js**: 运行 `npm audit` 或 `pnpm audit`。

### 2. 检查硬编码密钥
- **动作**: 使用 `grep` 或专项工具搜索代码库中的 `sk-`, `password=`, `token:` 等模式。
- **规则**: 严禁将 `.env` 提交到仓库。

### 3. 校验权限
- **动作**: 随机抽查新增 API，确保均有 `Depends(get_current_user)` 保护。

### 4. 依赖更新建议
- **输出**: 若发现 P0 级严重漏洞，必须立即发起 `chore: security fix` 分支进行更新。
