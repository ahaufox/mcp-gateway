---
description: 快速调研工作流。针对开发项目需求（技术选型/架构方案/疑难杂症），自动适配调研策略与信源，生成结构化简报保存到 00-tmp 目录。
---

### 1. 需求识别与分类 (Identify & Classify)
*   **输入**: 获取用户的技术调研需求（如 "PostgreSQL vs MongoDB" 或 "Next.js ISR 原理"）。
*   **分类**: 确定话题类型，这将决定后续的信源选择与文档模板。
    *   **Type A: 技术选型/选型对比 (Tech Selection)**: 关注优劣势、Benchmark、生态支持、维护成本 (e.g., FastAPI vs Go Gin, Redis vs Dragonfly)。
    *   **Type B: 架构设计/最佳实践 (Architecture & Best Practices)**: 关注模式、扩展性、解耦、安全性 (e.g., 微服务拆分准则, 鉴权中心化设计)。
    *   **Type C: 疑难杂症/底层原理 (Deep Dive & Debug)**: 关注报错定位、性能分析、源码逻辑、系统限制 (e.g., Python 内存泄漏发现, TCP 连接超时深挖)。

### 2. 动态信源选择 (Adaptive Source Selection)
根据分类优先选择权威信源。

*   **通用权威源**: 官方文档 (docs.xxx.com)、GitHub Issues/Discussions/PRs、StackOverflow、Reddit (r/programming, r/rust 等)。
*   **特定领域 Tier 1**:
    *   *[Type A 选型]*: 技术雷达 (Thoughtworks), DB-Engines, Web Framework Benchmarks (TechEmpower), YouTube 开发频道对比。
    *   *[Type B 架构]*: 厂商技术博客 (Netflix/Meta/Uber Eng Blog), InfoQ 架构频道, Medium (Tech 专题), DDD/系统设计指南。
    *   *[Type C 调试]*: 源码仓库, Sentry 社区, Linux Manual, 个人硬核技术博客 (如 Julia Evans, Cloudflare Blog)。
*   **⚠️ Red Flags (通用禁区)**:
    - AI 生成的垃圾代码片段 (未经校验的 CSDN/博客园复制粘贴)
    - 严重过时的技术专栏 (e.g. 5年前的框架版本教程)
    - 带有明显商业倾向的“软文”对比报告。

### 3. 执行调研 (Execution)
采用“广度扫描 -> 深度聚焦 -> 交叉验证”的漏斗模型。

*   **Phase A: 广度扫描 (Broad Scan)**
    *   对核心关键词进行**多角度搜索** (至少 3 组 Query):
        *   Comparison: `"{Keyword1}" vs "{Keyword2}" benchmarks 2024/2025`
        *   Architecture: `"{Keyword}" high level architecture design patterns`
        *   Debug: `"{Keyword}" common pitfalls/known issues/bottlenecks`
    *   **重要**: 快速浏览搜索结果，筛选出最近 2 年内的 3-5 个最有价值的讨论或文档。
*   **Phase B: 深度挖掘 (Deep Dive)**
    *   针对 Phase A 发现的技术痛点或限制进行二次搜索验证。
    *   *Tip*: 技术调研需验证“官方声称”与“社区反馈”的差异 (Expectations vs Reality)。

### 4. 生成文档 (Synthesize)
根据话题类型选择对应的**文档模板**。

#### **模板 A: 技术选型 (Tech Selection)**
1.  **Overview**: 相关技术或方案的极简定义。
2.  **Comparison Matrix**: 核心维度对比表格 (性能, 开发效率, 社区活跃度, 部署成本)。
3.  **Strengths & Weaknesses**: 各方的杀手锏与致命缺陷。
4.  **Selection Advice**: 推荐适用场景建议 (e.g., “如果团队熟悉 Python, 选 FastAPI”)。
5.  **Ecosystem**: 核心周边工具链是否完备。

#### **模板 B: 架构设计 (Architecture)**
1.  **Core Objective**: 本架构要解决的核心痛点 (如 高并发、高可用)。
2.  **Design Pattern**: 采用的设计模式或核心思想 (e.g., Event-driven, Layered Architecture)。
3.  **Pros & Cons**: 带来的好处与引入的复杂性。
4.  **Implementation Steps**: 关键实施步骤简述。
5.  **Security/Compliance**: 鉴权、日志、脱敏等安全性考虑。

#### **模板 C: 疑难杂症/原理 (Deep Dive)**
1.  **Issue Definition**: 现象精准描述或原理核心问题。
2.  **Root Cause Analysis**: 底层原因分析 (含可能的调用链、参数配置)。
3.  **Reproduction/Logic**: 复现步骤或核心逻辑代码段解析。
4.  **Solution/Workaround**: 最终修复方案或临时绕过策略。
5.  **Prevention**: 如何在未来通过 Linter、监控或 CI 流程规避。

---

### 5. 质量自检 & 交付
*   [ ] **时效性**: 调研的信息是否基于最新的稳定版本？
*   [ ] **实战性**: 结论是否包含可执行的建议或代码示例？
*   [ ] **信源合规**: 核心结论是否引用了官方或权威博文？

*   **保存路径**: `docs/自动调研：{Type}_{话题}_{YYMMDD}.md` (e.g. `docs/自动调研：Selection_DB_250211.md`).
*   **操作**: 使用 `write_to_file` 保存，并 `notify_user`。
