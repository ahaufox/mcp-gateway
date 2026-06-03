# MCP Gateway 功能优化迭代方案

## 1. 现状分析

### 1.1 当前支持的协议

| 协议类型 | 客户端支持 | 服务端暴露 | 状态 |
|----------|------------|------------|------|
| **stdio** | ✅ 完整 | ❌ 不支持 | 子进程通信 |
| **SSE** | ✅ 完整 | ✅ 完整 | 双向事件流 |
| **Streamable HTTP** | ✅ 完整 | ✅ 完整 | 无状态 HTTP |

### 1.2 核心模块梳理

| 模块 | 文件 | 功能 |
|------|------|------|
| 配置加载 | [config.go](file:///workspace/mcp-proxy/internal/config/config.go) | 配置解析、环境变量展开 |
| 客户端管理 | [client.go](file:///workspace/mcp-proxy/internal/core/client.go) | MCP 客户端连接、重连、工具聚合 |
| 服务端管理 | [mcp_server.go](file:///workspace/mcp-proxy/internal/server/mcp_server.go) | SSE/HTTP 服务暴露 |
| HTTP 服务 | [server.go](file:///workspace/mcp-proxy/internal/server/server.go) | 路由、Dashboard |

### 1.3 当前问题与痛点

#### 协议相关
1. **WebSocket 协议缺失** - 不支持 WebSocket 作为传输协议
2. **协议转换能力有限** - 无法在不同协议间灵活转换（如 stdio → WebSocket）
3. **缺乏协议适配层** - 新增协议需要修改核心代码
4. **SSE 连接稳定性** - 长连接断开后重连机制需要优化

#### 稳定性相关
1. **错误隔离不足** - 单个服务异常可能影响其他服务
2. **超时配置单一** - 不同服务/工具的超时策略无法定制
3. **请求排队处理** - 高并发下缺乏请求队列和限流
4. **资源泄漏风险** - stdio 进程管理存在资源泄漏隐患
5. **大型消息处理** - 大数据量的请求/响应缺乏流式处理优化

#### 功能体验相关
1. **工具聚合冲突** - 多个服务同名工具无重命名机制
2. **批量工具调用** - 缺乏批量调用能力
3. **工具调用缓存** - 重复调用相同工具无法缓存结果
4. **请求/响应转换** - 工具参数和结果无法灵活转换

---

## 2. 协议支持扩展

### 2.1 新增 WebSocket 协议支持

#### 设计目标
- 支持 WebSocket 作为客户端传输协议
- 支持通过 WebSocket 暴露网关服务
- 保持与现有 SSE/HTTP 一致的使用体验

#### 技术实现

```go
package protocol

// WebSocket 客户端配置
type WebsocketMCPClientConfig struct {
    URL               string            `json:"url"`
    Headers           map[string]string `json:"headers"`
    HandshakeTimeout  time.Duration     `json:"handshakeTimeout"`
    ReadBufferSize    int               `json:"readBufferSize"`
    WriteBufferSize   int               `json:"writeBufferSize"`
    EnableCompression bool              `json:"enableCompression"`
}

// WebSocket 服务端配置
type WebsocketMCPServerConfig struct {
    Path              string        `json:"path"`
    ReadBufferSize    int           `json:"readBufferSize"`
    WriteBufferSize   int           `json:"writeBufferSize"`
    HandshakeTimeout  time.Duration `json:"handshakeTimeout"`
    EnableCompression bool          `json:"enableCompression"`
}
```

#### 配置示例
```json
{
  "mcpServers": {
    "websocket-service": {
      "transportType": "websocket",
      "url": "ws://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer token"
      },
      "options": {
        "handshakeTimeout": "10s",
        "enableCompression": true
      }
    }
  }
}
```

### 2.2 协议转换桥接

#### 桥接模式
支持任意协议间的双向转换：
- stdio ↔ SSE
- stdio ↔ WebSocket
- SSE ↔ WebSocket
- Streamable HTTP ↔ 任意协议

#### 桥接实现
```go
package bridge

// ProtocolBridge 协议桥接器
type ProtocolBridge struct {
    source ProtocolAdapter
    target ProtocolAdapter
}

type ProtocolAdapter interface {
    Start(ctx context.Context) error
    Read(ctx context.Context) (*mcp.JSONRPCMessage, error)
    Write(ctx context.Context, msg *mcp.JSONRPCMessage) error
    Close() error
}

func (b *ProtocolBridge) Run(ctx context.Context) error {
    // 双向转发消息
    go b.forward(ctx, b.source, b.target)
    go b.forward(ctx, b.target, b.source)
    <-ctx.Done()
    return nil
}
```

### 2.3 统一协议适配层

#### 抽象接口设计
```go
package transport

// Transport 统一传输接口
type Transport interface {
    // 初始化连接
    Connect(ctx context.Context) error
    
    // 发送消息
    Send(ctx context.Context, msg *mcp.JSONRPCMessage) error
    
    // 接收消息
    Receive(ctx context.Context) (*mcp.JSONRPCMessage, error)
    
    // 健康检查
    Ping(ctx context.Context) error
    
    // 关闭连接
    Close() error
    
    // 获取状态
    Status() TransportStatus
}

// TransportFactory 传输工厂
type TransportFactory interface {
    Create(config any) (Transport, error)
    Supports(transportType string) bool
}
```

### 2.4 gRPC 协议支持（可选扩展）

```protobuf
syntax = "proto3";

package mcp.gateway;

service MCPGateway {
  rpc Initialize(InitializeRequest) returns (InitializeResponse);
  rpc ListTools(ListToolsRequest) returns (ListToolsResponse);
  rpc CallTool(stream CallToolRequest) returns (stream CallToolResponse);
  rpc ListResources(ListResourcesRequest) returns (ListResourcesResponse);
  rpc ReadResource(ReadResourceRequest) returns (ReadResourceResponse);
  rpc ListPrompts(ListPromptsRequest) returns (ListPromptsResponse);
  rpc GetPrompt(GetPromptRequest) returns (GetPromptResponse);
}
```

---

## 3. 稳定性优化

### 3.1 服务隔离与熔断

#### 熔断器设计
```go
package circuitbreaker

type State int

const (
    StateClosed State = iota   // 正常
    StateOpen                  // 熔断
    StateHalfOpen              // 半开
)

type CircuitBreaker struct {
    name           string
    state          State
    failureCount   int
    successCount   int
    lastFailure    time.Time
    config         Config
    
    mu             sync.RWMutex
}

type Config struct {
    MaxFailures    int           // 最大失败次数
    ResetTimeout   time.Duration // 熔断重置时间
    HalfOpenMax    int           // 半开状态最大请求数
    Timeout        time.Duration // 请求超时
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    if !cb.Allow() {
        return ErrCircuitOpen
    }
    
    err := fn()
    cb.RecordResult(err == nil)
    return err
}
```

#### 集成到客户端
```go
// 在 core/client.go 中集成熔断器
type Client struct {
    // ... 现有字段
    circuitBreaker *circuitbreaker.CircuitBreaker
}

func (c *Client) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    return c.circuitBreaker.Execute(func() error {
        return c.client.CallTool(ctx, req)
    })
}
```

### 3.2 超时与重试策略

#### 分级超时配置
```go
// config.go 扩展
type OptionsV2 struct {
    // ... 现有字段
    
    // 超时配置
    CallTimeout         time.Duration `json:"callTimeout,omitempty"`
    InitializeTimeout   time.Duration `json:"initializeTimeout,omitempty"`
    ListToolsTimeout    time.Duration `json:"listToolsTimeout,omitempty"`
    
    // 重试配置
    MaxRetries          int           `json:"maxRetries,omitempty"`
    RetryDelay          time.Duration `json:"retryDelay,omitempty"`
    RetryBackoff        float64       `json:"retryBackoff,omitempty"` // 退避因子
    RetryableErrors     []string      `json:"retryableErrors,omitempty"`
    
    // 限流配置
    RateLimit           float64       `json:"rateLimit,omitempty"` // 每秒请求数
    RateLimitBurst      int           `json:"rateLimitBurst,omitempty"`
}
```

#### 重试中间件
```go
package retry

func WithRetry(maxRetries int, delay time.Duration, backoff float64, fn func() error) error {
    var lastErr error
    for i := 0; i <= maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        lastErr = err
        
        if i == maxRetries {
            break
        }
        
        select {
        case <-time.After(calculateDelay(delay, backoff, i)):
        }
    }
    return lastErr
}
```

### 3.3 请求队列与限流

#### 令牌桶限流
```go
package ratelimit

import "golang.org/x/time/rate"

type Limiter struct {
    limiter *rate.Limiter
}

func NewLimiter(r rate.Limit, b int) *Limiter {
    return &Limiter{
        limiter: rate.NewLimiter(r, b),
    }
}

func (l *Limiter) Wait(ctx context.Context) error {
    return l.limiter.Wait(ctx)
}
```

#### 请求队列
```go
package queue

type Request struct {
    ctx      context.Context
    fn       func() error
    resultCh chan error
}

type RequestQueue struct {
    queue     chan *Request
    workers   int
    wg        sync.WaitGroup
}

func NewRequestQueue(queueSize, workers int) *RequestQueue {
    q := &RequestQueue{
        queue:   make(chan *Request, queueSize),
        workers: workers,
    }
    q.start()
    return q
}

func (q *RequestQueue) Submit(ctx context.Context, fn func() error) error {
    req := &Request{
        ctx:      ctx,
        fn:       fn,
        resultCh: make(chan error, 1),
    }
    
    select {
    case q.queue <- req:
    case <-ctx.Done():
        return ctx.Err()
    }
    
    select {
    case err := <-req.resultCh:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### 3.4 资源管理优化

#### stdio 进程守护
```go
package process

type ManagedProcess struct {
    cmd       *exec.Cmd
    stdout    io.ReadCloser
    stderr    io.ReadCloser
    stdin     io.WriteCloser
    
    restart   chan struct{}
    stop      chan struct{}
    mu        sync.Mutex
}

func (p *ManagedProcess) Start() error {
    // 启动进程
    // 监控进程状态
    // 异常时自动重启
}

func (p *ManagedProcess) monitor() {
    for {
        select {
        case <-p.stop:
            return
        case <-p.restart:
            p.restartProcess()
        }
    }
}
```

#### 连接池管理
```go
package pool

type ConnectionPool struct {
    pool      chan Transport
    factory   func() (Transport, error)
    maxIdle   int
    maxActive int
    
    mu        sync.Mutex
    active    int
}

func (p *ConnectionPool) Get(ctx context.Context) (Transport, error) {
    // 获取连接或创建新连接
}

func (p *ConnectionPool) Put(conn Transport) {
    // 归还连接或关闭
}
```

### 3.5 错误处理增强

#### 错误分类与包装
```go
package errors

type ErrorCode string

const (
    ErrCodeTimeout        ErrorCode = "timeout"
    ErrCodeConnection     ErrorCode = "connection"
    ErrCodeProtocol       ErrorCode = "protocol"
    ErrCodeServer         ErrorCode = "server"
    ErrCodeToolNotFound   ErrorCode = "tool_not_found"
    ErrCodeInvalidRequest ErrorCode = "invalid_request"
)

type MCPError struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Service string    `json:"service,omitempty"`
    Tool    string    `json:"tool,omitempty"`
    Cause   error     `json:"-"`
}

func (e *MCPError) Error() string {
    return fmt.Sprintf("[%s] %s (service=%s, tool=%s)", e.Code, e.Message, e.Service, e.Tool)
}
```

---

## 4. 功能增强

### 4.1 工具重命名与命名空间

#### 配置示例
```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "options": {
        "toolNamespace": "github",
        "toolPrefix": "gh_",
        "toolRename": {
          "create_issue": "create_github_issue",
          "search_repos": "search_github_repos"
        }
      }
    }
  }
}
```

#### 实现代码
```go
type ToolRewriteConfig struct {
    Namespace string            `json:"namespace,omitempty"`
    Prefix    string            `json:"prefix,omitempty"`
    Rename    map[string]string `json:"rename,omitempty"`
}

func rewriteToolName(original string, config *ToolRewriteConfig) string {
    if config == nil {
        return original
    }
    
    // 先检查精确重命名
    if newName, ok := config.Rename[original]; ok {
        return newName
    }
    
    // 添加前缀
    name := original
    if config.Prefix != "" {
        name = config.Prefix + name
    }
    
    // 添加命名空间
    if config.Namespace != "" {
        name = config.Namespace + "." + name
    }
    
    return name
}
```

### 4.2 工具调用缓存

#### 缓存配置
```go
type CacheConfig struct {
    Enabled      bool          `json:"enabled"`
    TTL          time.Duration `json:"ttl"`
    MaxSize      int           `json:"maxSize"`
    CacheableTools []string    `json:"cacheableTools,omitempty"`
}
```

#### 缓存实现
```go
package cache

import "github.com/hashicorp/golang-lru/v2"

type ToolCallCache struct {
    cache *lru.Cache[string, cacheEntry]
    ttl   time.Duration
}

type cacheEntry struct {
    result    *mcp.CallToolResult
    timestamp time.Time
}

func (c *ToolCallCache) Get(service, tool string, args map[string]any) (*mcp.CallToolResult, bool) {
    key := c.buildKey(service, tool, args)
    entry, ok := c.cache.Get(key)
    if !ok {
        return nil, false
    }
    
    if time.Since(entry.timestamp) > c.ttl {
        c.cache.Remove(key)
        return nil, false
    }
    
    return entry.result, true
}

func (c *ToolCallCache) Set(service, tool string, args map[string]any, result *mcp.CallToolResult) {
    key := c.buildKey(service, tool, args)
    c.cache.Add(key, cacheEntry{
        result:    result,
        timestamp: time.Now(),
    })
}
```

### 4.3 批量工具调用

#### 批量调用 API
```go
type BatchCallRequest struct {
    Calls []struct {
        Service string         `json:"service"`
        Tool    string         `json:"tool"`
        Args    map[string]any `json:"arguments"`
    } `json:"calls"`
    Parallel bool `json:"parallel,omitempty"`
}

type BatchCallResponse struct {
    Results []struct {
        Success bool                   `json:"success"`
        Result  *mcp.CallToolResult    `json:"result,omitempty"`
        Error   string                 `json:"error,omitempty"`
    } `json:"results"`
}
```

#### 批量调用实现
```go
func (s *Server) BatchCallTools(ctx context.Context, req BatchCallRequest) (*BatchCallResponse, error) {
    resp := &BatchCallResponse{
        Results: make([]struct {
            Success bool
            Result  *mcp.CallToolResult
            Error   string
        }, len(req.Calls)),
    }
    
    if req.Parallel {
        // 并行执行
        var wg sync.WaitGroup
        for i, call := range req.Calls {
            wg.Add(1)
            go func(idx int, c struct{ Service, Tool string; Args map[string]any }) {
                defer wg.Done()
                result, err := s.callSingleTool(ctx, c.Service, c.Tool, c.Args)
                if err != nil {
                    resp.Results[idx].Error = err.Error()
                } else {
                    resp.Results[idx].Success = true
                    resp.Results[idx].Result = result
                }
            }(i, call)
        }
        wg.Wait()
    } else {
        // 串行执行
        for i, call := range req.Calls {
            result, err := s.callSingleTool(ctx, call.Service, call.Tool, call.Args)
            if err != nil {
                resp.Results[i].Error = err.Error()
            } else {
                resp.Results[i].Success = true
                resp.Results[i].Result = result
            }
        }
    }
    
    return resp, nil
}
```

### 4.4 请求/响应转换钩子

#### 转换钩子接口
```go
package hook

type ToolCallHook interface {
    BeforeCall(ctx context.Context, service, tool string, args map[string]any) (map[string]any, error)
    AfterCall(ctx context.Context, service, tool string, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error)
}

type PromptHook interface {
    BeforeGetPrompt(ctx context.Context, service, prompt string, args map[string]any) (map[string]any, error)
    AfterGetPrompt(ctx context.Context, service, prompt string, result *mcp.GetPromptResult, err error) (*mcp.GetPromptResult, error)
}
```

#### 配置示例
```json
{
  "mcpServers": {
    "my-service": {
      "command": "my-server",
      "options": {
        "hooks": {
          "toolCall": {
            "before": [
              {
                "type": "add_default_args",
                "config": {
                  "defaults": {
                    "api_version": "v2"
                  }
                }
              }
            ],
            "after": [
              {
                "type": "transform_result",
                "config": {
                  "template": "{ \"data\": {{.content}} }"
                }
              }
            ]
          }
        }
      }
    }
  }
}
```

---

## 5. 分阶段实施计划

### 阶段一：稳定性基础（2-3 周）

**目标**：提升核心稳定性，建立错误处理和熔断机制

| 任务 | 优先级 | 交付物 |
|------|--------|--------|
| 分级超时配置 | P0 | 超时配置选项 + 应用 |
| 熔断器实现 | P0 | CircuitBreaker + 集成 |
| 错误分类与包装 | P0 | MCPError + 错误定义 |
| stdio 进程管理优化 | P1 | ManagedProcess |
| 基础重试机制 | P1 | 重试中间件 |

### 阶段二：协议扩展（3-4 周）

**目标**：支持 WebSocket，建立统一协议抽象层

| 任务 | 优先级 | 交付物 |
|------|--------|--------|
| WebSocket 客户端支持 | P0 | WebSocket MCPClient 实现 |
| 统一协议适配层 | P0 | Transport 抽象接口 |
| WebSocket 服务端暴露 | P1 | WebSocket 服务端点 |
| 协议转换桥接 | P1 | ProtocolBridge 实现 |

### 阶段三：功能增强（3-4 周）

**目标**：工具管理、缓存、批量调用等功能

| 任务 | 优先级 | 交付物 |
|------|--------|--------|
| 工具重命名与命名空间 | P0 | ToolRewriteConfig + 实现 |
| 工具调用缓存 | P0 | ToolCallCache + 集成 |
| 批量工具调用 | P1 | BatchCall API |
| 转换钩子机制 | P1 | Hook 接口 + 内置钩子 |

### 阶段四：高级特性（2-3 周）

**目标**：限流、队列、更多优化

| 任务 | 优先级 | 交付物 |
|------|--------|--------|
| 限流器实现 | P1 | 令牌桶限流 + 集成 |
| 请求队列 | P1 | RequestQueue |
| 连接池管理 | P2 | ConnectionPool |
| gRPC 协议支持（可选） | P2 | gRPC 服务 |

---

## 6. 配置文件完整示例

```json
{
  "mcpProxy": {
    "baseURL": "http://localhost:9090",
    "addr": ":9090",
    "name": "MCP Gateway",
    "version": "2.0.0",
    "type": "sse",
    "options": {
      "panicIfInvalid": false,
      "logEnabled": true,
      "callTimeout": "30s",
      "maxRetries": 3,
      "retryDelay": "1s",
      "retryBackoff": 2.0
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
    "websocket-service": {
      "transportType": "websocket",
      "url": "ws://localhost:8080/mcp",
      "headers": {
        "Authorization": "Bearer token"
      },
      "options": {
        "disablePing": false,
        "maintenanceInterval": "30s",
        "handshakeTimeout": "10s"
      }
    }
  }
}
```

---

## 7. 总结

本方案聚焦于**功能层面优化**，重点包括：

### 核心优化方向

1. **协议支持扩展**
   - 新增 WebSocket 协议（客户端 + 服务端）
   - 统一协议抽象层
   - 协议转换桥接

2. **稳定性提升**
   - 熔断器与服务隔离
   - 分级超时与智能重试
   - 请求队列与限流
   - 资源管理优化
   - 完善的错误处理

3. **功能增强**
   - 工具重命名与命名空间
   - 工具调用缓存
   - 批量工具调用
   - 请求/响应转换钩子

### 预期收益

| 指标 | 改善 |
|------|------|
| 协议支持 | 3 种 → 4+ 种（含 WebSocket） |
| 服务可用性 | 提升 20-30%（熔断+重试） |
| 资源泄漏 | 降低 90%+ |
| 开发效率 | 新协议接入成本降低 70% |
| 工具冲突 | 完全解决（重命名机制） |

通过分阶段实施，逐步将 MCP Gateway 打造成更稳定、更灵活、功能更丰富的 MCP 聚合平台。
