package circuitbreaker

import (
    "sync"
    "time"

    "github.com/ahaufox/mcp-gateway/mcp-proxy/internal/errors"
)

// State 表示熔断器的状态
type State int

const (
    // StateClosed 表示熔断器处于关闭状态，允许请求通过
    StateClosed State = iota
    // StateOpen 表示熔断器处于打开状态，阻止请求通过
    StateOpen
    // StateHalfOpen 表示熔断器处于半打开状态，允许少量请求通过以测试服务是否恢复
    StateHalfOpen
)

// Config 定义熔断器配置
type Config struct {
    // MaxFailures 最大失败次数，达到此值后打开熔断器
    MaxFailures int
    // ResetTimeout 重置超时时间，熔断器打开后等待此时间后进入半打开状态
    ResetTimeout time.Duration
    // HalfOpenMax 半打开状态允许通过的最大请求数
    HalfOpenMax int
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
    return Config{
        MaxFailures: 5,
        ResetTimeout: 30 * time.Second,
        HalfOpenMax: 1,
    }
}

// CircuitBreaker 是熔断器的核心实现
type CircuitBreaker struct {
    name       string
    state      State
    failures   int
    lastError  error
    lastChange time.Time
    config     Config
    mu         sync.RWMutex
}

// New 创建新的熔断器
func New(name string, config Config) *CircuitBreaker {
    return &CircuitBreaker{
        name:       name,
        state:      StateClosed,
        failures:   0,
        config:     config,
        lastChange: time.Now(),
    }
}

// Allow 检查是否允许请求通过
func (cb *CircuitBreaker) Allow() error {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.state {
    case StateClosed:
        return nil
    case StateOpen:
        if time.Since(cb.lastChange) >= cb.config.ResetTimeout {
            // 切换到半打开状态，允许测试请求
            cb.state = StateHalfOpen
            cb.failures = 0
            cb.lastChange = time.Now()
            return nil
        }
        return errors.Newf(
            errors.ErrCodeCircuitOpen,
            "circuit breaker is open, will reset after %v",
            cb.config.ResetTimeout,
        )
    case StateHalfOpen:
        // 半打开状态：允许部分请求通过
        if cb.failures >= cb.config.HalfOpenMax {
            return errors.Newf(
                errors.ErrCodeCircuitOpen,
                "circuit breaker is half-open, request limit exceeded",
            )
        }
        return nil
    default:
        return errors.Newf(errors.ErrCodeInternal, "unknown state: %d", cb.state)
    }
}

// RecordResult 记录执行结果
func (cb *CircuitBreaker) RecordResult(success bool) {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    if success {
        // 成功操作
        switch cb.state {
        case StateClosed:
            // 保持关闭状态
            cb.failures = 0
        case StateHalfOpen:
            // 半打开状态下成功，关闭熔断器
            cb.state = StateClosed
            cb.failures = 0
        case StateOpen:
            // 打开状态下不应该有成功操作
            return
        }
    } else {
        // 失败操作
        switch cb.state {
        case StateClosed:
            cb.failures++
            cb.lastError = errors.New(errors.ErrCodeServer, "operation failed")
            if cb.failures >= cb.config.MaxFailures {
                cb.state = StateOpen
                cb.lastChange = time.Now()
            }
        case StateOpen:
            // 已经是打开状态，不做处理
            return
        case StateHalfOpen:
            // 半打开状态下失败，重新打开熔断器
            cb.state = StateOpen
            cb.lastChange = time.Now()
        }
    }
    cb.lastChange = time.Now()
}

// State 返回当前状态
func (cb *CircuitBreaker) State() State {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.state
}

// Failures 返回当前失败次数
func (cb *CircuitBreaker) Failures() int {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.failures
}

// Execute 执行给定的函数，并记录执行结果
func (cb *CircuitBreaker) Execute(fn func() error) error {
    if err := cb.Allow(); err != nil {
        return err
    }

    err := fn()
    cb.RecordResult(err == nil)
    return err
}

// Reset 重置熔断器到关闭状态
func (cb *CircuitBreaker) Reset() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    cb.state = StateClosed
    cb.failures = 0
    cb.lastChange = time.Now()
}

// String 返回状态的字符串表示
func (s State) String() string {
    switch s {
    case StateClosed:
        return "closed"
    case StateOpen:
        return "open"
    case StateHalfOpen:
        return "half-open"
    default:
        return "unknown"
    }
}
