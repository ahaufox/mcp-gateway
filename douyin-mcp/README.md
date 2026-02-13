# 抖音 MCP 服务 (Douyin MCP Service)

基于 MCP (Model Context Protocol) 的抖音视频文案提取与无水印下载工具。


## ✨ 功能特性

- 🎬 **无水印已下载** - 获取高质量无水印视频下载链接
- 🎙️ **AI 语音识别** - 使用硅基流动 SenseVoice 自动提取文案
- 📑 **大文件支持** - 自动分段处理超过 1 小时或 50MB 的音频
- 🔌 **MCP 集成** - 支持 Claude Desktop, Cherry Studio 等 AI 应用
- 🛠️ **命令行工具** - 提供便捷的 CLI 脚本用于批量处理

## 🚀 快速开始

### 1. 安装依赖

本项目使用 `uv` 进行包管理：

```bash
cd douyin-mcp
uv sync
```

### 2. 配置 API Key

文案提取功能需要**硅基流动 (SiliconFlow)** 的 API Key。

```bash
export API_KEY="sk-xxxxxxxxxxxxxxxx"
```

> 💡 获取免费 API Key：[硅基流动](https://cloud.siliconflow.cn/)（新用户有免费额度）

---

## 🐳 Docker 部署

如果您不想在本地安装 FFmpeg 或 Python 环境，可以使用 Docker。

### 1. 构建镜像

```bash
docker build -t douyin-mcp .
```

### 2. MCP Server 配置 (Docker)

在 Claude Desktop 中使用 Docker 运行：

```json
{
  "mcpServers": {
    "douyin-mcp": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "API_KEY=sk-xxxxxxxxxxxxxxxx",
        "douyin-mcp"
      ]
    }
  }
}
```

### 3. 使用命令行工具 (Docker)

```bash
# 获取信息
docker run --rm douyin-mcp python scripts/douyin_downloader.py -l "分享链接" -a info

# 下载视频 (挂载当前目录 output 到容器)
docker run --rm -v $(pwd)/output:/app/output douyin-mcp python scripts/douyin_downloader.py -l "分享链接" -a download -o /app/output
```

---

## 🔌 MCP Server 配置 (本地)

### Claude Desktop / Cherry Studio

编辑 MCP 配置文件 (如 `~/Library/Application Support/Claude/claude_desktop_config.json`)，添加以下内容：

```json
{
  "mcpServers": {
    "douyin-mcp": {
      "command": "uv",
      "args": [
        "run",
        "python",
        "/absolute/path/to/douyin-mcp/server.py" 
      ],
      "env": {
        "API_KEY": "sk-xxxxxxxxxxxxxxxx"
      }
    }
  }
}
```
*请注意将 `/absolute/path/to/douyin-mcp/server.py` 替换为实际的绝对路径。*

### 可用工具

| 工具名 | 功能 | 需要 API |
|--------|------|:--------:|
| `parse_douyin_video_info` | 解析视频信息 (标题、ID、封面等) | ❌ |
| `get_douyin_download_link` | 获取无水印下载链接 | ❌ |
| `extract_douyin_text` | 提取视频文案 (下载->音频->识别) | ✅ |

---

## 🛠️ 命令行工具

`scripts/douyin_downloader.py` 提供了独立的命令行功能。

### 基本用法

```bash
# 查看帮助
uv run python scripts/douyin_downloader.py --help

# 获取视频信息（无需 API）
uv run python scripts/douyin_downloader.py -l "分享链接" -a info

# 下载无水印视频
uv run python scripts/douyin_downloader.py -l "分享链接" -a download -o ./videos

# 提取文案（需要 API_KEY）
export API_KEY="sk-xxx"
uv run python scripts/douyin_downloader.py -l "分享链接" -a extract -o ./output

# 提取文案并保存视频
uv run python scripts/douyin_downloader.py -l "分享链接" -a extract -o ./output --save-video
```

### 输出示例

```
output/
└── 7600361826030865707/
    ├── transcript.md    # 文案文件 (Markdown)
    └── *.mp4            # 视频文件（可选）
```

---

## 📋 系统要求

- **Python**: 3.10+
- **UV**: Python 包管理器 (`curl -LsSf https://astral.sh/uv/install.sh | sh`)
- **FFmpeg**: 音视频处理工具 (必须安装)
  - macOS: `brew install ffmpeg`
  - Ubuntu: `apt install ffmpeg`

## 🔧 技术说明

- **语音识别**: 使用 [硅基流动 SenseVoice API](https://cloud.siliconflow.cn/) (模型: `FunAudioLLM/SenseVoiceSmall`)
- **大文件处理**: 超过 API 限制（1小时/50MB）的音频会自动分割处理
