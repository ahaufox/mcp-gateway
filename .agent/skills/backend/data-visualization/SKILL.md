---
name: data-visualization
description: 专注于使用 mcp-server-chart 进行高效、专业的数据可视化图表生成。
---

# 📊 数据可视化技能 (Data Visualization)

本技能由“**数据分析师 (Data Analyst)**”与“**信息设计专家 (Information Designer)**”主导，旨在通过 MCP 协议快速生成高质量的图表。

## 🎯 触发场景
- 需要将数据转换为图表（如 `line`, `bar`, `pie`, `scatter` 等）。
- 用户请求分析数据趋势或对比。
- 需要生成可交互或静态的 AntV 图表。

## 🛠️ 核心能力
### 1. 图表类型丰富 (Rich Chart Types)
- **趋势分析**: `generate_line_chart`, `generate_area_chart`。
- **比较分析**: `generate_bar_chart`, `generate_column_chart`, `generate_radar_chart`。
- **构成分析**: `generate_pie_chart`, `generate_liquid_chart`。
- **关系分析**: `generate_scatter_chart`, `generate_network_graph`, `generate_sankey_chart`。
- **地理空间**: `generate_district_map`, `generate_pin_map`。

### 2. 数据与布局 (Data & Layout)
- **多轴支持**: `generate_dual_axes_chart` 处理双 Y 轴场景。
- **层级展示**: `generate_treemap_chart`, `generate_sunburst_chart`。
- **自定义配置**: 支持 AntV 的高级配置项（虽然 MCP 接口可能简化了部分参数）。

## 🚫 负向约束 (Negative Constraints)
- **严禁误导**: 选择的图表类型必须准确反映数据关系（如：不用饼图展示超过 7 个类别）。
- **严禁数据泄露**: 在生成图表时，确保不包含未经脱敏的敏感数据。
- **严禁无效调用**: 确保传入的数据结构符合 AntV G2 的标准 JSON 格式。

## 💡 最佳实践 (CoT Checklist)
1. **类型选择**: 这种数据最适合用什么图表展示？（例如：时间序列用折线图，分类对比用柱状图）。
2. **数据清洗**: 数据是否包含空值或异常值？是否需要在生成前进行预处理？
3. **交互与导出**: 生成后的图表是否满足用户的交互需求？是否需要保存为图片或 HTML？

---
> [!TIP]
> **Proactive Growth**: 探索 `mcp-server-chart` 的更多高级图表类型（如桑基图、词云图），并在合适场景主动推荐。
