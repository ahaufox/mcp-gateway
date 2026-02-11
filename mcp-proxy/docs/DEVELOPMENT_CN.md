# mcp-proxy 二次开发指南

本指南面向希望修改 `mcp-proxy` 核心逻辑或增加新协议支持的开发者。

## 💻 环境准备

- **语言**: Go 1.24+
- **工具**: Makefile, Docker (可选)

## 📁 核心结构

- `main.go`: 程序入口，处理命令行参数和服务器启动。
- `config.go`: 配置加载与 V1/V2 迁移逻辑。
- `client.go`: 下游 MCP 客户端解析与初始化逻辑。
- `http.go`: HTTP Server 路由与 SSE/Stream 处理。

## 🛠️ 常用命令 (Makefile)

- `make build`: 在 `./build/` 目录下生成可执行文件。
- `make format`: 执行代码格式化 (gofmt, tidy)。
- `make buildImage`: 构建 Docker 镜像。

## 🧪 核心逻辑扩展

### 添加新客户端类型
如果需要支持新的下游传输协议，请在 `client.go` 中扩展 `McpClientConfig` 解析逻辑，并实现相应的 MCP Client 接口。

### 修改 HTTP 聚合逻辑
如需自定义 Tools/Prompts 的聚合或过滤规则，请关注 `http.go` 中的 Handler 实现。

---
> [!NOTE]
> 提交代码前，请确保运行 `make format` 以保持代码风格一致。
