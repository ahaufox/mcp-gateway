---
description: git发布准备工作流：自动更新版本号、生成变更日志、制定发布计划并执行构建验证
---

inputs:
  - id: version
    type: text
    required: true
    description: "目标发布的版本号 (e.g. 1.2.0)"
  - id: files_hint
    type: list
    required: false
    description: "需要检查或修改的文件列表 (默认为 pyproject.toml, package.json, CHANGELOG.md)"

steps:
  - id: check_git_clean
    action: run_commands
    args:
      commands:
        - "git status --porcelain"
      expect_empty: true
      # 确保当前工作区干净，避免未提交的更改混入发布

  - id: read_context
    action: read_files
    args:
      hints: "{{inputs.files_hint}}"
      max_files: 12
      include:
        - "pyproject.toml"
        - "package.json"
        - "CHANGELOG.md"
        - "README.md"
        - ".env.example"

  - id: propose_plan
    action: write_plan
    args:
      goal: "Prepare release for version {{inputs.version}}"
      must_include:
        - release_contents_summary # 发布内容摘要：本次发布包含哪些主要功能或修复
        - breaking_changes         # 破坏性变更：是否有不兼容的接口或数据变更
        - rollout_plan             # 发布计划：灰度策略、回滚预案
        - verification_commands    # 验证命令：如何验证发布是否成功

  - id: apply_patch
    action: edit_code
    args:
      constraints:
        - minimal_diff
        - "Update version fields to {{inputs.version}} in package.json/pyproject.toml"
        - "Update CHANGELOG.md with release notes"

  - id: run_checks
    action: run_commands
    args:
      commands:
        # 前端检查
        - "npm run lint"
        - "npm run type-check"
        # 后端检查 (根据项目实际情况)
        - "pytest"
      allow_fail: false

  - id: verify_build
    action: run_commands
    args:
      commands:
        - "npm run build"
        - "docker build ." 
      allow_fail: false

outputs:
  - id: patch_summary
  - id: release_tag_command