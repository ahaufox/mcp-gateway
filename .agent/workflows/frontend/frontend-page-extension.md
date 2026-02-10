---
description: 在既有架构约束下扩展页面/路由/模块（新增功能或增强交互），确保类型、样式规范、可回归验证
---

inputs:
  - id: feature_spec
    type: text
    required: true
    description: "功能规格说明"
  - id: design_refs
    type: text
    required: false
    description: "设计参考 (URL或描述)"
  - id: routes_or_pages
    type: list
    required: false
    description: "涉及的路由/页面"
  - id: files_hint
    type: list
    required: false
    description: "相关文件提示"

steps:
  - id: read_context
    action: read_files
    args:
      hints: "{{inputs.files_hint}}"
      max_files: 12
      include:
        - "路由配置"
        - "目标页面组件"
        - "状态管理/数据请求层"
        - "设计系统组件"
        - "样式规范"
        - "相关测试"

  - id: propose_plan
    action: write_plan
    args:
      must_include:
        - user_flows
        - architecture_fit
        - state_and_data
        - accessibility_i18n
        - risk_assessment
        - verification_commands

  - id: apply_patch
    action: edit_code
    args:
      constraints:
        - minimal_diff
        - no_new_dependencies
        - no_cross_layer_refactor

  - id: add_tests
    action: edit_code
    args:
      goal: add_regression_tests
      constraints:
        - cover_core_interactions
        - cover_render_branches
        - cover_route_permissions

  - id: run_checks
    action: run_commands
    args:
      commands:
        - "npm test"
        - "npm run lint"
        - "npm run typecheck"
        - "npm run build"
      allow_fail: false

  - id: generate_commit
    action: generate_commit_message
    args:
      convention: conventional_commits
      rules:
        - type: feat  # 或 fix
        - scope: pages  # 或 ui
        - max_subject_len: 72
      include_body:
        - feature_description
        - ui_changes
        - test_coverage

outputs:
  - id: patch_summary
  - id: ui_change_notes
