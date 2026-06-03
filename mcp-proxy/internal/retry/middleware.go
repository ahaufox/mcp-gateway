package retry

import (
	"context"
	"log"
	"math"
	"strings"
	"time"

	mcperrors "github.com/ahaufox/mcp-gateway/mcp-proxy/internal/errors"
)

// Config 重试配置
type Config struct {
	// MaxRetries 最大重试次数
	MaxRetries int
	// InitialDelay 初始延迟时间
	InitialDelay time.Duration
	// MaxDelay 最大延迟时间
	MaxDelay time.Duration
	// BackoffMultiplier 退避乘数
	BackoffMultiplier float64
	// RetryableErrors 可重试的错误码列表
	RetryableErrors []mcperrors.ErrorCode
}

// DefaultConfig 返回默认重试配置
func DefaultConfig() Config {
	return Config{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableErrors:   []mcperrors.ErrorCode{mcperrors.ErrCodeConnection, mcperrors.ErrCodeTimeout},
	}
}

// calculateDelay 计算延迟时间（指数退避）
func calculateDelay(delay time.Duration, multiplier float64, attempt int) time.Duration {
	// 指数退避: delay * multiplier^attempt
	multiplied := float64(delay) * math.Pow(multiplier, float64(attempt))
	// 添加一些随机抖动（0-25%）
	jitter := multiplied * 0.25 * (1 - 2*float64(attempt%2))
	result := multiplied + jitter
	
	return time.Duration(math.Min(result, math.MaxFloat64))
}

// isRetryable 判断错误是否可重试
func isRetryable(err error, retryableErrors []mcperrors.ErrorCode) bool {
	if err == nil {
		return false
	}
	
	// 检查是否是 MCP 错误
	if mcpErr, ok := err.(*mcperrors.MCPError); ok {
		for _, code := range retryableErrors {
			if mcpErr.Code == code {
				return true
			}
		}
		return false
	}
	
	// 对于标准错误，假设网络错误可重试
	errStr := err.Error()
	retryableKeywords := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"i/o timeout",
		"network",
	}
	
	for _, keyword := range retryableKeywords {
		if strings.Contains(strings.ToLower(errStr), keyword) {
			return true
		}
	}
	
	return false
}

// WithRetry 使用重试策略执行函数
func WithRetry(ctx context.Context, config Config, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// 检查上下文是否取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// 执行函数
		err := fn()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// 判断是否应该重试
		if attempt == config.MaxRetries {
			log.Printf("[retry] max retries (%d) reached, giving up", config.MaxRetries)
			break
		}
		
		if !isRetryable(err, config.RetryableErrors) {
			log.Printf("[retry] error is not retryable: %v", err)
			break
		}
		
		// 计算延迟
		delay := calculateDelay(config.InitialDelay, config.BackoffMultiplier, attempt)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
		
		log.Printf("[retry] attempt %d failed: %v, retrying in %v", attempt+1, err, delay)
		
		// 等待延迟
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	
	return lastErr
}

// RetryableFunc 可重试的函数类型
type RetryableFunc func() error

// Retry 便捷函数，使用默认配置
func Retry(ctx context.Context, fn RetryableFunc) error {
	return WithRetry(ctx, DefaultConfig(), fn)
}

// RetryWithConfig 便捷函数，使用自定义配置
func RetryWithConfig(ctx context.Context, config Config, fn RetryableFunc) error {
	return WithRetry(ctx, config, fn)
}
