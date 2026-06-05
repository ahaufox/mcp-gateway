---
trigger: always_on
description: 规定了本项目的视觉风格、核心调色板、UI 组件的设计标准及响应式/暗黑模式要求（唯一 UI 规范入口）。
---

# 视觉设计标准与品牌规范

本规范为全栈（客户端/管理端）唯一的 UI 视觉与样式基准。所有 UI 开发必须严格遵循以下硬性约束。

## 1. 核心布局理念
- **编辑式不对称（Editorial Asymmetry）**：通过充足留白与高对比字号层级组织信息。
- **色调分层（Tonal Layering）**：优先使用背景色差产生纵深，禁止使用 1px 灰线围栏。

## 2. 核心调色板与 Design Tokens

所有颜色引用必须使用 Design Token，**严禁硬编码十六进制值 (Hex)、RGB 或 Tailwind 原始色阶（如 `bg-[#F8F9FA]`、`text-slate-500`）**。

| 角色 | Token名称 | 十六进制 (亮色) | 描述 |
| :--- | :--- | :--- | :--- |
| **主品牌色** | `primary` | `#FACC15` | 荧光金 - 核心品牌色，用于高亮和主要行动点 |
| **主背景** | `background` | `#F8F9FA` | 冷灰色 - 页面基础底色 (桌面) |
| **卡片/表面** | `surface` | `#FFFFFF` | 纯白色 - 主卡片、工作区"纸面" |
| **表面容器** | `surface-container` | `#F1F5F9` | 灰底色 - 用于次级内容区或输入框默认背景 |
| **主要文本** | `on-surface` | `#111827` | 深藏青 - 用于标题、主要文本 |
| **次要文本** | `on-surface-variant` | `#6B7280` | 暖灰褐 - 辅助文字、微文案、失效状态 |
| **三级强调** | `tertiary` | `#2563EB` | 专业蓝 - 链接、时间戳、进度指示器 |
| **边框/分割** | `border` | `#E2E8F0` | 极浅灰 - 仅在必须时使用的物理分割线 |

> **开发落地规则 (Tailwind v4)**：
> 使用 `bg-background`、`text-on-surface`、`border-border` 等。

### 2.1 Design Token 优先原则

严禁在代码中硬编码任何与主题相关的颜色值。
- **禁止项**：
  - 内联样式中使用颜色：`style={{ backgroundColor: '#fff' }}`
  - Tailwind 类名中使用具体颜色：`className="bg-white dark:bg-[#141414]"`
  - CSS 文件中使用颜色：`color: #333;`
- **推荐项**：
  - 使用 `theme.useToken()` 获取系统变量：`const { token } = theme.useToken();`
  - 将样式绑定到 Token：`style={{ backgroundColor: token.colorBgContainer }}`

## 3. "无边线"规则 (No-Hard-Border)

**明确指令：禁止使用 1px 实线灰边框来包区块或做分割！**
- **边界通过背景区分**：在 `background` 上放置 `surface` 卡片，或在 `surface-container` 上放置 `surface`，利用色差完成分区。
- **阴影兜底**：静态卡片默认不要阴影（依靠色调分层）；浮层型组件（模态框、下拉）使用扩散型环境阴影（不可用纯黑）。
- **幽灵边框兜底**：若因高密度表格等必须要有边框，使用极低不透明度的边框（如 `border-white/20` 或 `border-slate-200/50`），做到"可感知而非可见"。

## 4. 排版与字体系统 (Typography)

- **字体体系**：优先使用 `Inter` / `Manrope`，或系统无衬线字体。
- **层次对比**：
  - **标题 (Display/Headline)**：大字号配合 `font-black` 或 `font-bold`，`tracking-tight`，营造权威感与速度感。
  - **正文 (Body)**：不要使用 100% 纯黑（#000），必须使用 `on-surface`。行高设定为 `leading-relaxed`。
  - **标签 (Label)**：用于元数据，使用小字号（10px-12px）、`font-bold`、`uppercase` 配合超宽字距 `tracking-widest`。

## 5. 关键组件与交互规范

### 5.1 按钮 (Buttons)
- **主按钮**：背景使用 `primary`，文本使用深色。可加入极轻微的阴影或交互放大效果。
- **次按钮 (幽灵)**：无背景，使用极浅边框或无边框，文本色匹配 `primary` 或 `on-surface-variant`。

### 5.2 沉浸式弹窗 (Modals & Dialogs)
- **极简一体化**：弹窗主体、表单控制件、底部动作行必须无缝平铺在单层连续的盒子内。**禁止在弹窗内使用灰色底页脚或横线割裂**。
- **极致圆角**：弹窗容器使用 `rounded-[32px]`，深弥散阴影 `shadow-2xl`，以及全屏高斯模糊遮罩 `backdrop-blur-xl`。
- **悬浮感**：利用半透墨色遮罩 (`bg-gray-900/40`) 营造脱离桌面的空间纵深。

### 5.3 玻璃拟态 (Glassmorphism)
- 用于悬浮控制栏、吸顶 Header 或命令面板。
- 背景设定为 `bg-white/70` 或 `bg-gray-900/80`，配合高斯模糊 `backdrop-blur-xl`。

## 6. 暗黑模式适配 (Dark Mode)

系统支持一键切换暗黑模式，所有自定义组件必须自动适配。
- **主题自适应**：在任何业务场景中，禁止在组件类名中固定写入亮色限定字样。所有 UI 元素必须通过 Design Token 自动响应主题切换（依赖 CSS 变量）。
- **颜色层级**：
  - 容器背景：使用 `surface` Token。
  - 页面背景：使用 `background` Token。
  - 标题文字：使用 `on-surface` Token。
  - 次要文字：使用 `on-surface-variant` Token。
  - 分割线：使用 `border` Token。
- **验证要求**：在提交 UI 变更前，必须手动切换暗黑模式进行视觉验证，确保没有黑底黑字或白底白字的情况。

## 7. 响应式布局 (Responsive Design)

- **断点布局**：采用 Mobile-First 原则，通过 Tailwind 前缀（`sm:`, `md:`, `lg:`）控制网格和分栏。禁止给容器设定写死的 `width`，使用 `max-w` 和弹性布局。
- **弹性尺寸**：避免给容器设置固定的 `width`，应优先使用 `maxWidth` 或百分比布局。
- **间距一致性**：使用 Token 中的间距变量，如 `token.paddingLG`, `token.marginMD`，确保页面节奏感统一。
- **独立滚动 (Pane Scrolling)**：在双栏/多栏布局中，主体内容区域应支持独立 `overflow-y-auto` 滚动，严禁全局 Body 滚动导致整体导航栏位移。

## 8. 应做与禁做

### ✅ 应做 (Do)
- 把留白当作功能性工具，用于归组相关法律概念。
- 使用不对称布局（左宽栏 + 右窄元数据栏）营造高端杂志感。
- 在提交 UI 前，切换"暗黑模式"肉眼验证没有黑底黑字或白底白字的情况。

### ❌ 禁做 (Don't)
- 不要直接使用 `#000` 或 `#FFF` 字符。
- 不要使用标准 1px 灰色物理边框。若看起来像 Excel 表格，即违规。
- 不要在弹窗/按钮上使用硬质黑色阴影。阴影应表现为环境光扩散。

## 9. 红线检查 (Checklist)
- [ ] 是否完全移除了硬编码的十六进制/RGB 颜色值？
- [ ] 在暗黑模式下，对比度是否符合阅读标准？
- [ ] 在 1280px 和 1920px 宽度下，布局是否依然美观且未溢出？
- [ ] 关键交互元素是否使用了系统主题色（`colorPrimary`）？