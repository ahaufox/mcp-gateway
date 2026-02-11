# 🗺️ 项目开发路线图 (Roadmap)

本文档规划了 MCP Gateway 项目的后续开发里程碑与关键任务。

## 📍 阶段 1: 基础服务建设 (Ready)
> **目标**: 确保核心 MCP 服务功能完备且可独立运行。

- [ ] **完善 Jules MCP Server**: 确保功能完整，通过单元测试。
- [ ] **完善 PyMCPAutoGUI**: 确保 GUI 自动化功能稳定。
- [ ] **完善 MCP Server Chart**: 确保数据可视化服务可用。

## 📍 阶段 2: 容器化交付 (Dockerize)
> **目标**: 实现服务的标准化交付，支持从源码构建 Docker 镜像。

- [ ] **Dockerize Jules**: 编写 Dockerfile，验证构建与运行。
- [ ] **Dockerize PyMCPAutoGUI**: 解决 GUI 环境依赖（Xvfb 等），实现容器化。
- [ ] **Dockerize Chart Server**: 编写 Dockerfile，优化镜像体积。

## 📍 阶段 3: 网关互联 (Integration)
> **目标**: 打通 MCP Gateway 与三个核心子服务的连接。

- [ ] **配置 Gateway**: 在 `config.json` 中添加三个子服务的配置。
- [ ] **网络互通**: 确保 Gateway 容器能访问子服务容器（Docker Compose 或 K8s）。
- [ ] **集成测试**: 验证 Gateway 能正确转发请求并调用子服务的 Tool/Resource。

## 📍 阶段 4: 扩展公共服务 (Public HTTP)
> **目标**: 接入外部公开的 HTTP MCP 服务，验证网关的通用性。

- [ ] **调研公共 MCP 服务**: 寻找稳定的开源/公开 MCP 服务端点。
- [ ] **接入测试**: 配置 Gateway 接入 HTTP 协议的 MCP 服务。
- [ ] **协议兼容性验证**: 验证 SSE 与 HTTP 传输协议的混合支持。

## 📍 阶段 5: 生态扩展 (Ecosystem)
> **目标**: 持续丰富 MCP 服务生态。

- [ ] **接入更多工具**: 集成数据库、搜索、文件系统等更多类型的 MCP 服务。
- [ ] **社区贡献**: 将通用服务回馈给开源社区。
