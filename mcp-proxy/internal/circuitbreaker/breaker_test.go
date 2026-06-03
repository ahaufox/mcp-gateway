package circuitbreaker

import (
    "testing"
    "time"

    "github.com/ahaufox/mcp-gateway/mcp-proxy/internal/errors"
)

func TestCircuitBreaker_Closed(t *testing.T) {
    cb := New("test", DefaultConfig())
    
    if cb.State() != StateClosed {
        t.Error("circuit breaker should start closed")
    }
    
    if err := cb.Allow(); err != nil {
        t.Error("should allow requests when closed")
    }
    
    // 记录成功操作
    cb.RecordResult(true)
    if cb.Failures() != 0 {
        t.Error("success should reset failure count")
    }
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
    config := DefaultConfig()
    config.MaxFailures = 3
    config.ResetTimeout = 100 * time.Millisecond
    cb := New("test", config)
    
    // 多次失败操作
    for i := 0; i < config.MaxFailures; i++ {
        cb.RecordResult(false)
    }
    
    if cb.State() != StateOpen {
        t.Error("circuit breaker should open after max failures")
    }
    
    if !errors.Is(cb.Allow(), errors.ErrCodeCircuitOpen) {
        t.Error("should not allow requests when open")
    }
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
    config := DefaultConfig()
    config.MaxFailures = 2
    config.ResetTimeout = 50 * time.Millisecond
    cb := New("test", config)
    
    // 打开熔断器
    cb.RecordResult(false)
    cb.RecordResult(false)
    
    if cb.State() != StateOpen {
        t.Fatal("circuit breaker should be open")
    }
    
    // 等待重置超时
    time.Sleep(config.ResetTimeout + 10*time.Millisecond)
    
    // 尝试请求，应该允许（半打开）
    err := cb.Allow()
    if err != nil {
        t.Error("should allow test request in half-open state, got", err)
    }
    
    if cb.State() != StateHalfOpen {
        t.Error("should enter half-open state after timeout")
    }
}

func TestCircuitBreaker_HalfOpenSuccess(t *testing.T) {
    config := DefaultConfig()
    config.MaxFailures = 2
    config.ResetTimeout = 50 * time.Millisecond
    cb := New("test", config)
    
    // 打开熔断器
    cb.RecordResult(false)
    cb.RecordResult(false)
    
    // 等待重置超时
    time.Sleep(config.ResetTimeout + 10*time.Millisecond)
    cb.Allow() // 进入半打开
    
    // 记录成功，应该关闭熔断器
    cb.RecordResult(true)
    if cb.State() != StateClosed {
        t.Error("success in half-open should close circuit")
    }
}

func TestCircuitBreaker_HalfOpenFail(t *testing.T) {
    config := DefaultConfig()
    config.MaxFailures = 2
    config.ResetTimeout = 50 * time.Millisecond
    cb := New("test", config)
    
    // 打开熔断器
    cb.RecordResult(false)
    cb.RecordResult(false)
    
    // 等待重置超时
    time.Sleep(config.ResetTimeout + 10*time.Millisecond)
    cb.Allow() // 进入半打开
    
    // 记录失败，应该重新打开
    cb.RecordResult(false)
    if cb.State() != StateOpen {
        t.Error("failure in half-open should reopen circuit")
    }
}

func TestCircuitBreaker_Execute(t *testing.T) {
    config := DefaultConfig()
    config.MaxFailures = 1
    cb := New("test", config)
    
    err := cb.Execute(func() error {
        return nil
    })
    if err != nil {
        t.Error("success should not return error")
    }
    
    // 执行失败
    testErr := errors.Newf(errors.ErrCodeServer, "test error")
    err = cb.Execute(func() error {
        return testErr
    })
    if err != testErr {
        t.Error("expected error to propagate")
    }
    
    // 现在熔断器应该打开
    if cb.State() != StateOpen {
        t.Error("should open after one failure with max failures = 1")
    }
}

func TestCircuitBreaker_Reset(t *testing.T) {
    config := DefaultConfig()
    config.MaxFailures = 1
    cb := New("test", config)
    
    cb.RecordResult(false)
    if cb.State() != StateOpen {
        t.Fatal("should be open")
    }
    
    cb.Reset()
    if cb.State() != StateClosed {
        t.Error("reset should close circuit")
    }
    if cb.Failures() != 0 {
        t.Error("reset should clear failures")
    }
}
