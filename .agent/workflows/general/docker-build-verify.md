---
description: 本地 Docker 构建与启动验证，确保镜像生产可用。
---

# Docker 构建验证工作流 (Docker Build Verify)

## 执行步骤

### 1. 多阶段构建
- **命令**: `docker build --target production -t project:latest .`
- **检查**: 观察构建日志，确保没有不必要的开发依赖被打入生产镜像。

### 2. 运行冒烟测试
- **命令**: `docker-compose -f docker-compose.prod.yml up -d`
- **验证**: 访问 `localhost:8080/health` (或对应健康检查接口)，确保容器正常启动且数据库连接成功。

### 3. 性能与体积优化
- **动作**: 检查镜像大小。如果超过 500MB，考虑使用 `alpine` 或 `distroless` 基础镜像，并优化层缓存。

### 4. 清理
- **命令**: `docker-compose down -v`。
