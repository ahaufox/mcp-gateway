# 🗺️ 项目开发路线图 (Roadmap)

本文档规划了 MCP Gateway 项目的后续开发里程碑与关键任务。

## 📍 阶段 1: 基础服务建设 (Ready)
> **目标**: 确保核心 MCP 服务功能完备且可独立运行。

- [x] **完善 Jules MCP Server**: 核心功能已实现，集成至网关。
- [ ] **完善 PyMCPAutoGUI**: 待子模块初始化及功能补全。
- [ ] **完善 MCP Server Chart**: 待子模块初始化及数据可视化验证。

## 📍 阶段 2: 容器化交付 (Dockerize)
> **目标**: 实现服务的标准化交付，支持从源码构建 Docker 镜像。

- [x] **Dockerize Jules**: 已编写 Dockerfile，支持容器化部署。
- [ ] **Dockerize PyMCPAutoGUI**: 待解决 GUI 环境依赖（Xvfb 等）。
- [ ] **Dockerize Chart Server**: 待编写 Dockerfile。

## 📍 阶段 3: 网关互联 (Integration)
> **目标**: 打通 MCP Gateway 与三个核心子服务的连接。

- [x] **配置 Gateway**: 已在 `config.json` 中添加 Jules 与 GitHub 配置。
- [ ] **GitHub 验证**: 待通过真实 Token 验证 GitHub MCP 功能。
- [ ] **网络互通**: 验证 Docker Compose 容器间网络。
- [ ] **集成测试**: 建立全链路集成测试套件。

## 📍 阶段 4: 多租户与安全 (Advanced Features)
> **目标**: 支持多用户隔离，并提供灵活的凭据（Token/API Key）管理方案。

- [ ] **多租户架构**: 实现基础的用户空间隔离。
- [ ] **本地配置注入**: 允许通过客户端本地配置文件动态注入 Token。
- [ ] **服务端凭据存储**: 实现数据库或加密存储，用于服务端持久化 Token。

## 📍 阶段 5: 扩展公共服务 (Public HTTP)
> **目标**: 接入外部公开的 HTTP MCP 服务，验证网关的通用性。

- [ ] **调研公共 MCP 服务**: 寻找稳定的开源/公开 MCP 服务端点。
- [ ] **接入测试**: 配置 Gateway 接入 HTTP 协议的 MCP 服务。
- [ ] **协议兼容性验证**: 验证 SSE 与 HTTP 传输协议的混合支持。

## 📍 阶段 6: 生态扩展 (Ecosystem)
> **目标**: 持续丰富 MCP 服务生态。

- [ ] **接入更多工具**: 集成数据库、搜索、文件系统等更多类型的 MCP 服务。
- [ ] **社区贡献**: 将通用服务回馈给开源社区。
