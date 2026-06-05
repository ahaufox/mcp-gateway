# 🐳 Docker Best Practices

本规则旨在标准化 Docker 镜像构建流程，确保安全性、可维护性与最小化体积。

## 1. 基础镜像 (Base Image)
- **推荐**: 使用官方 slim 版本或 alpine 版本（如 `python:3.11-slim`, `node:18-alpine`）。
- **理由**: 减小攻击面，加快构建与拉取速度。

## 2. 多阶段构建 (Multi-stage Builds)
- **必选**: 对于编译型语言或含构建产物的项目（如 Node.js, Go），必须使用多阶段构建。
- **示例**:
  ```dockerfile
  # Build stage
  FROM node:18 AS builder
  WORKDIR /app
  COPY package*.json ./
  RUN npm ci
  COPY . .
  RUN npm run build

  # Run stage
  FROM node:18-alpine
  WORKDIR /app
  COPY --from=builder /app/dist ./dist
  CMD ["node", "dist/index.js"]
  ```

## 3. 安全性 (Security)
- **非 Root 用户**: 生产环境容器不应以 root 身份运行。
  ```dockerfile
  RUN addgroup -S appgroup && adduser -S appuser -G appgroup
  USER appuser
  ```
- **敏感信息**: 严禁将 secrets (API keys, passwords) 硬编码进 Dockerfile。应通过环境变量或 secret mount 注入。

## 4. 优化 (Optimization)
- **.dockerignore**: 必须包含 `.git`, `node_modules`, `__pycache__`, `venv` 等非必要文件。
- **Layer Caching**: 将变更频率低的指令（如 `npm install`）放在前面，变更频率高的（如 `COPY . .`）放在后面。

## 5. 验证 (Verification)
- 构建后应运行简单的健康检查或版本检查指令确保镜像可用。
