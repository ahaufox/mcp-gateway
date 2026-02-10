---
description: 对 Python 热点进行剖析与优化：先构造可复现基准，再做最小改动优化，并提供前后对比数据（内测期/生产前通用）
---

## 适用范围
- 技术栈：FastAPI（Python 3.10+）/ 通用 Python 服务与脚本
- 目标：在**不引入新依赖**、**最小 diff**前提下，定位热点并拿到可复现的**前后对比数据**
- 风险级别：默认 **P1**（涉及核心路径或高 QPS）/ 纯内部脚本可按 **P2**

---

## inputs

- id: scenario  
  type: text  
  required: true  
  description: 性能场景描述（入口、请求类型、数据规模、期望现象、是否高并发）

- id: target_metric  
  type: text  
  required: true  
  description: 目标指标（latency / cpu / memory），给出优化目标（如 P95 < 200ms）

- id: baseline  
  type: text  
  required: false  
  description: 基准数据（如当前 P50/P95、CPU%、内存峰值、QPS），可附采样窗口与环境

- id: files_hint  
  type: list  
  required: false  
  description: 相关文件提示（可能的热点路径、路由入口、服务层/CRUD 层、工具函数、测试、bench 脚本）

---

## steps

### 1) read_context
action: read_files

args:
- hints: "{{inputs.files_hint}}"
- max_files: 12
- include:
  - 热点代码路径（函数/模块/类）
  - 调用链入口（FastAPI 路由、Celery/任务入口、CLI）
  - 数据结构与 IO 边界（DB/Redis/HTTP/文件）
  - 现有 profiling/benchmark 脚本（如有）
  - 相关测试（单测/集成/e2e）

输出要求（阅读结论必须记录）：
- 热点候选列表（Top 3）
- 入口与最小复现路径
- IO 与 CPU 的边界判断（主要耗时在 IO 还是计算）
- 可能的锁/阻塞点（同步 IO、线程池、连接池）

---

### 2) propose_plan
action: write_plan

args.must_include:
- reproduction_and_measurement
- profiling_method
- bottleneck_hypotheses
- chosen_optimization
- tradeoffs
- risk_assessment
- verification_commands

计划写作规范（必须）：
- **用中文**
- 明确“先测后改”的顺序
- 只选择**一个**主要优化方向（避免同时改多处导致归因不清）
- 标注风险分级（P0/P1/P2）与回滚方式（代码回滚 / 配置开关）

#### reproduction_and_measurement（最小可复现基准）
必须说明：
- 固定输入规模（数据量、payload 大小、循环次数）
- 固定环境（Python 版本、机器规格、是否容器、是否开启 DEBUG）
- 输出指标口径（P50/P95、平均耗时、吞吐、内存峰值）
- 去噪策略（预热、重复次数、取中位数、关掉多余日志）

#### profiling_method（剖析方法）
按优先级选择（尽量用标准库）：
1. CPU：`cProfile` + `pstats`
2. 微基准：`timeit`（固定输入、固定迭代次数）
3. 内存：`tracemalloc`（标准库）
4. 线上采样（如有）：用现有 APM/指标（不新增依赖）

#### bottleneck_hypotheses（瓶颈假设）
至少 2 条可验证假设，例如：
- 重复序列化/反序列化（JSON/MsgPack）
- N+1 查询 / 未批量化
- 热路径中创建大量临时对象（list/dict）
- 频繁正则/日期解析/字符串拼接
- 不必要的深拷贝、排序、重复计算（可缓存）

#### chosen_optimization（选择的最小优化）
- 只做**最小**、**可回滚**、**可度量**的改动
- 禁止过早复杂化（不引入新架构、不拆大模块）

#### tradeoffs（权衡）
必须写清：
- 可读性 vs 性能
- 内存 vs CPU
- 精度/一致性影响（如缓存/近似计算）
- 对观测与调试的影响

#### risk_assessment（风险评估）
至少包含：
- 是否触及鉴权/安全逻辑（触及则 P0）
- 是否影响核心主流程（影响则 P1）
- 是否改变业务语义（必须写清并加测试锁定）

#### verification_commands（验证命令）
必须列出：
- `pytest`（至少相关测试）
- `lint` / `typecheck`
- 基准命令（timeit/bench 脚本/cProfile 命令）
- 若涉及 DB/Redis：提供本地/CI 可跑的最小化验证方式（可用测试容器/fixtures）

---

### 3) apply_patch
action: edit_code

args.constraints:
- no_new_dependencies（不新增依赖）
- minimal_diff（最小 diff）
- avoid_premature_complexity（避免过早复杂化）

实现规范（必须）：
- 只改动与热点直接相关的代码路径
- 为关键逻辑添加中文注释（解释“为何更快/为何安全”）
- 如引入缓存/预计算：必须有失效策略或明确生命周期
- 任何行为变更必须增加测试覆盖或说明“不变”的证明

---

### 4) add_tests（性能回归哨兵）
action: edit_code

args.goal: add_performance_regression_sentinel  
args.constraints:
- benchmark_stability（稳定性优先）
- fixed_input_scale（固定输入规模）
- time_or_call_count_assertions（时间或调用次数断言）

建议策略（优先级从高到低）：
1. **调用次数/查询次数断言**（最稳定）
   - 例如：断言 DB 查询次数不超过 N（避免 N+1 回归）
2. **固定输入下的相对耗时断言**（谨慎使用）
   - 例如：基于“优化后耗时 <= 基线 * 1.2”
   - 需预热、重复、多次取中位数，并给出足够宽松阈值
3. **复杂度哨兵**
   - 例如：输入翻倍，耗时不应呈平方级增长（用小规模近似验证）

禁止：
- 直接用绝对毫秒阈值卡死 CI（除非环境高度稳定且阈值留足余量）

---

### 5) run_checks
action: run_commands

args.commands:
- "pytest"
- "lint"
- "typecheck"
- "基准命令（timeit/bench 脚本 / cProfile + pstats）"

args.allow_fail: false

执行要求：
- 记录测试环境信息（CPU/内存、Python 版本、是否容器）
- 基准至少跑 3 次，记录中位数（或 P50）与波动范围

---

### 6) generate_commit
action: generate_commit_message

args:
- convention: conventional_commits
- rules:
  - type: perf  # 或 fix(perf)
  - scope_from_paths: true
  - max_subject_len: 72
- include_body:
  - baseline_metrics
  - optimized_metrics
  - test_environment
  - tradeoffs

提交信息模板（参考）：
```
perf(<scope>): 优化 <热点函数/路径> 的 <指标>

- baseline: P95=..., avg=..., cpu=...
- optimized: P95=..., avg=..., cpu=...
- env: Python..., CPU..., OS..., dataset...
- tradeoffs: ...
```

---

## outputs

### patch_summary
必须包含：
- 改动文件列表（含关键 diff 摘要）
- 选择的瓶颈假设与验证结果（为何确定它是热点）
- 优化点说明（为何更快、是否影响语义）

### perf_report
必须包含（用表格呈现）：
- 场景描述（inputs.scenario）
- 指标口径（inputs.target_metric）
- 基线 vs 优化后（至少 3 次统计的中位数/波动）
- 环境信息（硬件/容器/Python 版本）
- 风险与回滚方案

示例表格：
| 项目 | 基线 | 优化后 | 变化 |
|---|---:|---:|---:|
| P95 latency (ms) | 320 | 190 | -40.6% |
| CPU (%) | 85 | 62 | -27.1% |
| 内存峰值 (MB) | 410 | 430 | +4.9% |

---
