---
name: gui-automation
description: 专注于使用 PyMCPAutoGUI 进行安全的桌面 GUI 自动化操作。
---

# 🖱️ GUI 自动化技能 (GUI Automation)

本技能由“**自动化工程师 (Automation Engineer)**”主导，旨在通过 MCP 协议安全、准确地控制本地桌面环境。

## 🎯 触发场景
- 需要操作非 API 接口的桌面应用程序。
- 需要进行屏幕截图分析或 OCR 识别。
- 用户请求执行“点击”、“输入”、“截图”等操作。

## 🛠️ 核心能力
### 1. 视觉感知 (Visual Perception)
- **屏幕截图**: 使用 `screenshot` 获取当前上下文。
- **图像定位**: 使用 `locate_on_screen` 或 `smart_click` 定位 UI 元素。
- **OCR 识别**: 结合 `smart_click` 或后续的 OCR 工具读取屏幕文本。

### 2. 精准操控 (Precision Control)
- **鼠标操作**: `move_to`, `click`, `drag_to`, `scroll`。
- **键盘输入**: `write`, `press`, `hotkey`。
- **窗口管理**: `activate_window`, `resize_window`, `close_window`。

### 3. 容错与反馈 (Resilience & Feedback)
- **重试机制**: 优先使用 `click_with_retry` 替代普通的 `click`，以应对 UI 响应延迟。
- **日志审计**: 所有操作必须通过 `view_logs.py` 或 MCP 日志流进行记录和审查。

## 🚫 负向约束 (Negative Constraints)
- **严禁盲点**: 禁止在未确认坐标或未定位图像的情况下直接执行 `click`。
- **严禁输入敏感信息**: 密码等敏感信息必须通过 `password` 提示框或安全变量处理，禁止明文 `write`。
- **严禁无限循环**: 任何自动化脚本必须包含 `failsafe`（鼠标移动到角落）或超时退出机制。

## 💡 最佳实践 (CoT Checklist)
1. **定位阶段**: 我是否确信目标元素在当前屏幕上可见？是否需要先 `activate_window`？
2. **操作阶段**: 点击后 UI 是否会有延迟？是否应该使用 `click_with_retry`？
3. **验证阶段**: 操作完成后，是否需要再次截图以确认结果（Success Validation）？

---
> [!TIP]
> **Proactive Growth**: 遇到无法识别的 UI 元素时，尝试使用 `smart_click` 利用 AI 视觉能力进行定位。
