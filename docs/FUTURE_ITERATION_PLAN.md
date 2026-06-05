# MCP Gateway 功能优化迭代方案

## 1. 现状分析

### 1.1 当前支持的协议

| 协议类型 | 客户端支持 | 服务端暴露 | 状态 |
|----------|------------|------------|------|
| **stdio** | ✅ 完整 | ❌ 不支持 | 子进程通信 |
| **SSE** | ✅ 完整 | ✅ 完整 | 双向事件流 |
| **Streamable HTTP** | ✅ 完整 | ✅ 完整 | 无状态 HTTP |
| **WebSocket** | ✅ 独立实现 | ❌ 待开发 | 客户端已实现，待集成 |

### 1.2 核心模块实现状态

| 模块 | 文件 | 实现状态 | 集成状态 |
|------|------|----------|----------|
| 配置加载 | `internal/config/config.go` | ✅ 完成 | ✅ 已集成 |
| 客户端管理 | `internal/core/client.go` | ✅ 完成 | ✅ 已集成 |
| 服务端管理 | `internal/server/mcp_server.go` | ✅ 完成 | ✅ 已集成 |
| HTTP 服务 | `internal/server/server.go` | ✅ 完成 | ✅ 已集成 |
| 熔断器 | `internal/circuitbreaker/breaker.go` | ✅ 完成 | ✅ 已集成 |
| 错误体系 | `internal/errors/` | ✅ 完成 | ✅ 已集成 |
| 超时配置 | `internal/config/config.go` | ✅ 完成 | ✅ 已集成 |
| WebSocket 客户端 | `internal/transport/websocket_client.go` | ✅ 完成 | ⏳ 待集成 |
| 传输层抽象 | `internal/transport/interface.go` | ✅ 完成 | ⏳ 待集成 |
| 工具重命名 | `internal/tools/rewrite.go` | ✅ 完成 | ⏳ 待集成 |
| 工具缓存 | `internal/cache/toolcache.go` | ✅ 完成 | ⏳ 待集成 |
| 批量调用 | `internal/batch/batch.go` | ✅ 完成 | ⏳ 待集成 |
| 钩子机制 | `internal/hook/hook.go` | ✅ 完成 | ⏳ 待集成 |
| 重试中间件 | `internal/retry/middleware.go` | ✅ 完成 | ⏳ 待集成 |
| 进程管理 | `internal/process/manager.go` | ✅ 完成 | ⏳ 待集成 |
| 限流器 | `internal/ratelimit/` | ❌ 未实现 | - |
| 请求队列 | `internal/queue/` | ❌ 未实现 | - |
| 连接池 | `internal/pool/` | ❌ 未实现 | - |
| 协议桥接 | `internal/bridge/` | ❌ 未实现 | - |

### 1.3 当前问题与痛点

#### 协议相关
1. **WebSocket 客户端未接入** — 独立实现但核心流程仍回退 SSE
2. **WebSocket 服务端缺失** — 不支持以 WebSocket 方式暴露网关
3. **协议桥接能力为空** — 无法在不同协议间灵活转换
4. **SSE 连接稳定性** — 长连接断开后重连机制需要优化

#### 集成相关
1. **模块孤岛** — 7 个功能模块代码完整但未接入核心流程
2. **工具重命名未生效** — `tools/rewrite.go` 未在工具注册/聚合流程中调用
3. **缓存未生效** — `cache/toolcache.go` 未集成到 CallTool 调用链
4. **批量调用无路由** — `batch/batch.go` 未挂载到 HTTP API
5. **重试未接入** — `retry/middleware.go` 未在客户端调用链中执行
6. **钩子未串联** — `hook/hook.go` 未注入到请求/响应处理管道
7. **进程管理未替换** — `process/manager.go` 未替代 mcp-go 原生进程管理

#### 功能体验相关
1. **高并发保障缺失** — 缺乏限流、请求队列机制
2. **连接管理粗放** — 缺乏连接池复用
3. **资源监控不足** — 缺乏完善的 metrics 和 tracing

---

## 2. 已完成基础能力 (无需重复开发)

以下能力已实现完整代码，核心工作转为**集成打通**而非重新开发。

### 2.1 传输层抽象

已定义统一 `Transport` 接口及工厂模式：

```go
// transport/interface.go
type Transport interface {
    Connect(ctx context.Context) error
    Send(ctx context.Context, msg *mcp.JSONRPCMessage) error
    Receive(ctx context.Context) (*mcp.JSONRPCMessage, error)
    Ping(ctx context.Context) error
    Close() error
    Status() TransportStatus
}

type TransportFactory interface {
    Create(config any) (Transport, error)
    Supports(transportType string) bool
    Name() string
}
```

**待完成**: 将现有 Stdio/SSE/HTTP 客户端迁移至 Transport 接口。

### 2.2 WebSocket 客户端

`internal/transport/websocket_client.go` 已实现完整 WebSocket 客户端：
- 连接建立（可配置 Dialer、Headers、超时）
- 读写循环（goroutine + channel）
- 心跳检测（PingMessage 定期发送）
- 自动重连（指数退避）
- WebSocketFactory 工厂

**待完成**: 在 `core/client.go` 中注册 WebSocket 路由，替代现有的 SSE 回退逻辑。

### 2.3 熔断器

`internal/circuitbreaker/breaker.go` 已实现完整三态熔断器（Closed/Open/HalfOpen），已在 `core/client.go` 中集成：
- `Allow()` / `RecordResult()` / `Execute()`
- 并发安全（RWMutex）
- 可配置（MaxFailures / ResetTimeout / HalfOpenMax）

### 2.4 错误体系

`internal/errors/` 已实现完整错误体系，已在 `client.go` 中广泛使用：
- 10 种标准错误码
- `MCPError` 结构体 + `Wrap`/`New` 工厂函数
- `WithService`/`WithTool`/`WithCause` 链式构建

### 2.5 超时与重试

- 超时配置: `OptionsV2` 中 `CallTimeout`/`InitializeTimeout`/`ListToolsTimeout` 已定义，`client.go` 已应用
- 重试中间件: `internal/retry/middleware.go` 已实现（指数退避 + 随机抖动 + 可重试错误码匹配）

**待完成**: 将重试中间件接入 `client.go` 的工具调用链路。

### 2.6 工具重命名

`internal/tools/rewrite.go` 已实现完整工具名重写机制：
- 精确重命名（map 映射）
- 前缀添加
- 命名空间添加
- 配置校验（自引用/循环引用检测）
- 配置合并

**待完成**: 在 Client 初始化工具列表时调用 `RewriteTools()`。

### 2.7 工具调用缓存

`internal/cache/toolcache.go` 已实现完整缓存机制：
- LRU 淘汰
- TTL 过期
- 缓存白名单（CacheableTools）
- 命中率统计
- 并发安全

**待完成**: 在 Client.CallTool 流程中插入缓存查询/写入逻辑。

### 2.8 批量调用

`internal/batch/batch.go` 已实现：
- 顺序执行
- 并行执行（goroutine + WaitGroup）
- 请求校验
- Context 取消传播

**待完成**: 挂载到 HTTP API 路由。

### 2.9 钩子机制

`internal/hook/hook.go` 已实现基础钩子链：
- RequestHook / ResponseHook 函数类型
- Chain 链式执行
- 错误传递

**待完成**: 注册到服务端请求处理管道；实现内置具体钩子（参数注入、结果转换等）。

### 2.10 进程管理

`internal/process/manager.go` 已实现：
- 进程启动/监控/自动重启
- 优雅关闭
- 事件系统（Start/Stop/Restart/Error）

**待完成**: 替换 `client.go` 中 direct stdio 进程创建逻辑。

---

## 3. 待开发能力

### 3.1 限流器

#### 设计目标
- 基于令牌桶算法
- 支持服务级和方法级限流
- 与熔断器联动

#### 实现方案

```go
// internal/ratelimit/limiter.go (新增)
package ratelimit

import "golang.org/x/time/rate"

type Config struct {
    Enabled    bool    `json:"enabled"`
    Rate       float64 `json:"rate"`       // 每秒令牌数
    Burst      int     `json:"burst"`      // 突发大小
    PerMethod  bool    `json:"perMethod"`  // 是否按方法隔离
}

type Limiter struct {
    limiter *rate.Limiter
}

func NewLimiter(cfg Config) *Limiter {
    return &Limiter{
        limiter: rate.NewLimiter(rate.Limit(cfg.Rate), cfg.Burst),
    }
}

func (l *Limiter) Wait(ctx context.Context) error {
    return l.limiter.Wait(ctx)
}
```

### 3.2 请求队列

#### 设计目标
- 缓冲高峰请求
- 支持优先级
- 背压保护

#### 实现方案

```go
// internal/queue/queue.go (新增)
package queue

type Request struct {
    ctx      context.Context
    fn       func() error
    priority int
    resultCh chan error
}

type Queue struct {
    queues   []chan *Request  // 多级优先级队列
    workers  int
    wg       sync.WaitGroup
}

func NewQueue(depth, workers, levels int) *Queue {
    q := &Queue{
        queues:  make([]chan *Request, levels),
        workers: workers,
    }
    for i := range q.queues {
        q.queues[i] = make(chan *Request, depth)
    }
    q.start()
    return q
}
```

### 3.3 连接池

#### 设计目标
- 复用长连接（SSE/WebSocket）
- 减少握手开销
- 自动健康检查

#### 实现方案

```go
// internal/pool/pool.go (新增)
package pool

type Pool struct {
    mu       sync.Mutex
    idle     []*Conn
    active   int
    maxIdle  int
    maxOpen  int
    factory  func() (*Conn, error)
    closeFn  func(*Conn)
}

func (p *Pool) Get(ctx context.Context) (*Conn, error) {
    p.mu.Lock()
    if n := len(p.idle); n > 0 {
        c := p.idle[n-1]
        p.idle = p.idle[:n-1]
        p.active++
        p.mu.Unlock()
        return c, nil
    }
    p.mu.Unlock()
    return p.create(ctx)
}
```

### 3.4 协议桥接

#### 设计目标
- 任意协议间双向转换
- 零拷贝消息转发
- 健壮的错误传递

#### 实现方案

```go
// internal/bridge/bridge.go (新增)
package bridge

type Adapter interface {
    Start(ctx context.Context) error
    Read(ctx context.Context) (*mcp.JSONRPCMessage, error)
    Write(ctx context.Context, msg *mcp.JSONRPCMessage) error
    Close() error
}

type Bridge struct {
    source Adapter
    target Adapter
}

func (b *Bridge) Run(ctx context.Context) error {
    eg, ctx := errgroup.WithContext(ctx)
    eg.Go(func() error { return b.forward(ctx, b.source, b.target) })
    eg.Go(func() error { return b.forward(ctx, b.target, b.source) })
    return eg.Wait()
}
```

### 3.5 WebSocket 服务端

复用已有 WebSocket 客户端的基础设施，在服务端增加 WebSocket 升级端点：

```go
// internal/server/websocket_server.go (新增)
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    upgrader := websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
    }
    conn, err := upgrader.Upgrade(w, r, nil)
    // 创建 WebSocketTransport 并接入 MCP 消息循环
}
```

---

## 4. 模块集成计划

核心工作：将 7 个独立模块接入核心流程。

### 4.1 集成依赖关系

```
集成顺序 (按依赖):
1. transport/interface.go    ← 无依赖，基础设施
2. process/manager.go        ← 无依赖，基础设施
3. cache/toolcache.go        ← 无依赖，基础设施
4. tools/rewrite.go          ← 依赖 transport（工具列表来源）
5. hook/hook.go              ← 依赖 transport/tools 等
6. retry/middleware.go       ← 依赖 errors
7. batch/batch.go            ← 依赖 transport + hook
8. transport/websocket_client.go ← 依赖 transport/interface
```

### 4.2 集成点详述

#### 集成点 1: transport 适配器迁移

| 项目 | 内容 |
|------|------|
| **任务** | 将 core/client.go 中的 Stdio/SSE/HTTP 客户端迁移到 Transport 接口 |
| **改动文件** | `internal/core/client.go`, `internal/transport/stdio.go`(新增), `internal/transport/sse.go`(新增), `internal/transport/http.go`(新增) |
| **工作量** | 2-3 天 |
| **说明** | 现有 client.go 直接使用 mcp-go SDK 的原生客户端，需包装适配 |

#### 集成点 2: WebSocket 路由注册

| 项目 | 内容 |
|------|------|
| **任务** | 在 client.go 中注册 WebSocket 传输类型，替代 SSE 回退 |
| **改动文件** | `internal/core/client.go`, `internal/config/config.go` |
| **工作量** | 1 天 |
| **说明** | 已有完整 WebSocket 客户端，只需在 ParseMCPClientConfigV2 中正确路由 |

#### 集成点 3: 工具重命名串联

| 项目 | 内容 |
|------|------|
| **任务** | 在 addToolsToServer / ListTools 流程中调用 RewriteTools |
| **改动文件** | `internal/core/client.go` |
| **工作量** | 0.5 天 |
| **说明** | 调用点明确，需确保反向映射（调用时还原原始名） |

#### 集成点 4: 缓存集成

| 项目 | 内容 |
|------|------|
| **任务** | 在 CallTool 中插入缓存查询/写入 |
| **改动文件** | `internal/core/client.go` |
| **工作量** | 0.5 天 |
| **说明** | Get 命中直接返回，Set 在成功响应后写入 |

#### 集成点 5: 重试串联

| 项目 | 内容 |
|------|------|
| **任务** | 在 createWrappedCallTool 中应用重试中间件 |
| **改动文件** | `internal/core/client.go` |
| **工作量** | 0.5 天 |
| **说明** | 熔断器已在调用链中，在熔断之后包装重试 |

#### 集成点 6: 钩子注入

| 项目 | 内容 |
|------|------|
| **任务** | 在服务端请求处理管道中注册钩子链 |
| **改动文件** | `internal/server/server.go`, `internal/server/mcp_server.go` |
| **工作量** | 1 天 |
| **说明** | 先实现空链，后续通过配置注入具体钩子 |

#### 集成点 7: 批量调用 API

| 项目 | 内容 |
|------|------|
| **任务** | 将 batch.Executor 挂载到 HTTP 路由 |
| **改动文件** | `internal/server/server.go` |
| **工作量** | 0.5 天 |

#### 集成点 8: 进程管理器替换

| 项目 | 内容 |
|------|------|
| **任务** | 用 process.Manager 替代 mcp-go 原生进程启动 |
| **改动文件** | `internal/core/client.go` |
| **工作量** | 1-2 天 |
| **说明** | 需保持与现有 StdioMCPClient 接口兼容 |

---

## 5. 分阶段实施计划

### 阶段一：模块集成打通（2-3 周）

**目标**：将已实现的独立模块接入核心流程，释放既有开发成果。

| 任务 | 优先级 | 工作量 | 交付物 |
|------|--------|--------|--------|
| transport 适配器迁移 | P0 | 2-3d | stdio/sse/http 适配器 + 客户端改造 |
| WebSocket 路由注册 | P0 | 1d | WebSocket 客户端正式可用 |
| 工具重命名串联 | P0 | 0.5d | 工具名重写在注册中生效 |
| 缓存集成 | P0 | 0.5d | CallTool 走缓存查询 |
| 重试串联 | P1 | 0.5d | 工具调用带重试 |
| 钩子注入 | P1 | 1d | 请求处理管道可执行钩子 |
| 批量调用 API | P1 | 0.5d | HTTP 端点 `/api/batch` |
| 进程管理器替换 | P1 | 1-2d | stdio 进程带守护管理 |

**总计**: P0 4-5d + P1 3-4d ≈ **7-9 个工作日**

### 阶段二：新功能开发（2-3 周）

**目标**：实现限流、队列、连接池等高级特性。

| 任务 | 优先级 | 工作量 | 交付物 |
|------|--------|--------|--------|
| 限流器实现 | P1 | 1d | 令牌桶 + 配置 |
| 限流器集成 | P1 | 1d | HTTP 中间件 |
| 请求队列实现 | P1 | 1.5d | 优先级队列 + worker pool |
| 请求队列集成 | P1 | 1d | 服务端请求管道接入 |
| WebSocket 服务端 | P1 | 2d | ws:// 端点暴露 |
| 协议桥接 | P2 | 2d | 任意协议双向转换 |
| 连接池 | P2 | 2d | 长连接复用 + 健康检查 |

**总计**: P1 5.5d + P2 4d ≈ **9.5 个工作日**

### 阶段三：运维与可观测性（1-2 周）

**目标**：完善监控、指标、调试能力。

| 任务 | 优先级 | 工作量 | 说明 |
|------|--------|--------|------|
| Prometheus 指标 | P1 | 2d | 请求量/延迟/错误率/熔断状态 |
| 健康检查增强 | P1 | 1d | 各服务独立健康状态 |
| 调试 API | P2 | 1d | 实时查看缓存/熔断/队列状态 |
| 配置热加载 | P2 | 2d | 不重启更新服务配置 |

**总计**: P1 3d + P2 3d ≈ **6 个工作日**

---

## 6. 依赖关系与里程碑

### 6.1 任务依赖图

```
集成阶段 (阶段一)                     新功能阶段 (阶段二)
                                    
transport 适配器 ─→ WebSocket 路由   限流器 ─→ 限流集成
                                                ↓
工具重命名 ─→ 缓存集成                请求队列 ─→ 队列集成
                                                ↓
重试串联 ─→ 批量 API                  WebSocket 服务端

钩子注入 ─→ 进程管理器                协议桥接 ─→ 连接池
```

### 6.2 里程碑

| 里程碑 | 版本 | 内容 | 目标 |
|--------|------|------|------|
| **M1** | v1.1.0 | Transport 适配器迁移完成，WebSocket 正式可用 | 第 2 周末 |
| **M2** | v1.2.0 | 工具重命名+缓存+重试集成完成 | 第 3 周末 |
| **M3** | v1.3.0 | 所有模块集成完成，批量 API+钩子+进程管理 | 第 4 周末 |
| **M4** | v2.0.0 | 限流器+请求队列+WebSocket 服务端 | 第 7 周末 |
| **M5** | v2.1.0 | 协议桥接+连接池+可观测性 | 第 9 周末 |

---

## 7. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| Transport 适配器迁移破坏现有逻辑 | 高 | 中 | 逐协议迁移，充分单元测试 + 集成测试 |
| WebSocket 客户端与核心接口不匹配 | 中 | 中 | 提前接口对齐，代码审查 |
| 限流/队列影响正常请求延迟 | 中 | 低 | 压测验证，可配置关闭 |
| 连接池资源泄漏 | 高 | 低 | 严格的新建/归还配对，泄漏检测 |
| 模块集成后发现设计缺陷 | 中 | 中 | 预留缓冲时间，小步提交 |

---

## 8. 配置完整示例

```json
{
  "mcpProxy": {
    "baseURL": "http://localhost:9090",
    "addr": ":9090",
    "name": "MCP Gateway",
    "version": "2.0.0",
    "type": "sse",
    "options": {
      "callTimeout": "30s",
      "maxRetries": 3,
      "retryDelay": "1s",
      "retryBackoff": 2.0,
      "rateLimit": 100.0,
      "rateLimitBurst": 50
    }
  },
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "options": {
        "toolNamespace": "github",
        "callTimeout": "60s",
        "maxRetries": 2,
        "circuitBreaker": {
          "maxFailures": 5,
          "resetTimeout": "30s"
        },
        "cache": {
          "enabled": true,
          "ttl": "5m",
          "maxSize": 1000,
          "cacheableTools": ["search_repos", "get_repository"]
        }
      }
    },
    "my-service": {
      "transportType": "websocket",
      "url": "ws://localhost:8080/mcp",
      "options": {
        "handshakeTimeout": "10s",
        "enableCompression": true
      }
    }
  }
}
```

---

## 9. 预期收益

| 指标 | 当前 | 目标 |
|------|------|------|
| 协议支持 | 3 种 | 4+ 种（含 WebSocket） |
| 服务可用性 | - | 提升 20-30%（熔断+重试+限流） |
| 重复调用开销 | 100% | 降低 60%+（缓存命中时） |
| 工具名冲突 | 可能 | 完全解决（重命名机制） |
| 并发处理 | 无控制 | 限流+队列保障 |
| 资源泄漏 | 偶发 | 降低 90%+（进程管理+连接池） |
| 开发扩展 | 高成本 | 新协议接入成本降低 70%（Transport 接口） |
