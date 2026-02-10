---
description: 自动发现潜在/新增 bug：从变更与信号入手，复现、最小化、定位并产出可执行报告（不改代码）
---

inputs:
  - id: context
    type: text
    required: false
    description: "线索：报错日志/监控告警/用户反馈/issue 链接/复现片段"
  - id: changed_files
    type: list
    required: false
    description: "可选：本次变更涉及的文件列表（用于聚焦扫描）"
  - id: baseline_ref
    type: text
    required: false
    default: "origin/main"
    description: "对比基线（用于定位回归）"
  - id: test_commands
    type: list
    required: false
    default:
      - "npm test"
      - "npm run lint"
    description: "项目默认校验命令（按仓库替换）"
  - id: time_budget_minutes
    type: number
    required: false
    default: 20
    description: "最多执行时长（避免无限跑）"

triggers:
  - id: on_file_change
    event: file_change
  - id: on_build_success
    event: build_success

gates:
  - id: secrets_guard
    type: block_on_match
    match:
      - "(?i)api[_-]?key\\b"
      - "(?i)secret\\b"
      - "(?i)password\\b"
      - "-----BEGIN (RSA|EC|OPENSSH|PRIVATE) KEY-----"
      - "(?i)authorization:\\s*bearer\\s+"
    on_block:
      message: "检测到疑似明文凭证/密钥信号：停止并输出位置与安全修复建议（仅报告，不提交）。"

steps:
  - id: collect_signal
    action: collect_context
    args:
      from_inputs:
        - context
        - changed_files
      include:
        - "recent_errors"
        - "stack_traces"
        - "failing_ci_summary"

  - id: diff_focus
    action: git_diff_summary
    args:
      base: "{{inputs.baseline_ref}}"
      files_hint: "{{inputs.changed_files}}"
      output:
        - "risk_areas"
        - "hotspots"
        - "suspect_commits"

  - id: static_scan
    action: static_analysis
    args:
      scope:
        files_hint: "{{inputs.changed_files}}"
        max_files: 30
      checks:
        - "type_errors"
        - "nullability"
        - "unhandled_exceptions"
        - "unsafe_parsing"
        - "sql_injection_smells"
        - "redis_dangerous_cmds"
      output:
        - "findings"

  - id: run_checks
    action: run_commands
    args:
      commands: "{{inputs.test_commands}}"
      time_budget_minutes: "{{inputs.time_budget_minutes}}"
      capture:
        - "stdout"
        - "stderr"
        - "exit_codes"
        - "failed_tests"
        - "first_error"
      allow_fail: true

  - id: minimize_repro
    action: reproduce_and_minimize
    args:
      from:
        - "{{steps.collect_signal}}"
        - "{{steps.run_checks}}"
      output:
        - "minimal_repro_steps"
        - "expected_vs_actual"
        - "environment_assumptions"

  - id: isolate_root_cause
    action: root_cause_analysis
    args:
      evidence:
        - "{{steps.diff_focus}}"
        - "{{steps.static_scan}}"
        - "{{steps.run_checks}}"
        - "{{steps.minimize_repro}}"
      output:
        - "top_hypotheses"
        - "most_likely_root_cause"
        - "blast_radius"
        - "regression_or_existing"

  - id: propose_fix_plan
    action: propose_fix
    args:
      constraints:
        - "no_code_changes_in_this_workflow"
        - "prefer_minimal_fix"
        - "tests_required_for_fix"
      output:
        - "suggested_patch_outline"
        - "tests_to_add"
        - "rollback_strategy"
        - "observability_notes"

  - id: generate_bug_report
    action: generate_issue
    args:
      template: |
        ## 现象 / 症状
        {{symptom}}

        ## 最小复现步骤
        {{minimal_repro_steps}}

        ## 期望 vs 实际
        {{expected_vs_actual}}

        ## 影响范围（Blast Radius）
        {{blast_radius}}

        ## 回归判断
        {{regression_or_existing}}

        ## 证据
        - CI/命令结果：{{checks_summary}}
        - 首个错误：{{first_error}}
        - 可疑变更点：{{hotspots}}

        ## 最可能根因（含备选假设）
        {{top_hypotheses}}

        ## 修复建议（不改代码，只给方案）
        {{suggested_patch_outline}}

        ## 需要补的测试
        {{tests_to_add}}

        ## 回滚/降级方案
        {{rollback_strategy}}

        ## 观测建议
        {{observability_notes}}
      fill:
        symptom: "{{inputs.context}}"
        minimal_repro_steps: "{{steps.minimize_repro.minimal_repro_steps}}"
        expected_vs_actual: "{{steps.minimize_repro.expected_vs_actual}}"
        blast_radius: "{{steps.isolate_root_cause.blast_radius}}"
        regression_or_existing: "{{steps.isolate_root_cause.regression_or_existing}}"
        checks_summary: "{{steps.run_checks.failed_tests}}"
        first_error: "{{steps.run_checks.first_error}}"
        hotspots: "{{steps.diff_focus.hotspots}}"
        top_hypotheses: "{{steps.isolate_root_cause.top_hypotheses}}"
        suggested_patch_outline: "{{steps.propose_fix_plan.suggested_patch_outline}}"
        tests_to_add: "{{steps.propose_fix_plan.tests_to_add}}"
        rollback_strategy: "{{steps.propose_fix_plan.rollback_strategy}}"
        observability_notes: "{{steps.propose_fix_plan.observability_notes}}"

outputs:
  - id: bug_report_markdown
    from: "{{steps.generate_bug_report}}"
  - id: triage_summary
    format: text
    value: |
      likely_root_cause: {{steps.isolate_root_cause.most_likely_root_cause}}
      regression: {{steps.isolate_root_cause.regression_or_existing}}
      blast_radius: {{steps.isolate_root_cause.blast_radius}}
      next_actions: {{steps.propose_fix_plan.suggested_patch_outline}}