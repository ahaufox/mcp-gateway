package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	mcperrors "github.com/ahaufox/mcp-gateway/mcp-proxy/internal/errors"
)

func TestWithRetry_Success(t *testing.T) {
	config := DefaultConfig()
	config.MaxRetries = 3
	
	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}
	
	err := WithRetry(context.Background(), config, fn)
	if err != nil {
		t.Error("should not return error on success")
	}
	if callCount != 1 {
		t.Errorf("should call function once, got %d", callCount)
	}
}

func TestWithRetry_RetryOnFailure(t *testing.T) {
	config := DefaultConfig()
	config.MaxRetries = 2
	config.InitialDelay = 1 * time.Millisecond
	
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return mcperrors.Newf(mcperrors.ErrCodeConnection, "connection failed")
		}
		return nil
	}
	
	err := WithRetry(context.Background(), config, fn)
	if err != nil {
		t.Error("should succeed after retries")
	}
	if callCount != 3 {
		t.Errorf("should call function 3 times, got %d", callCount)
	}
}

func TestWithRetry_MaxRetriesExceeded(t *testing.T) {
	config := DefaultConfig()
	config.MaxRetries = 2
	config.InitialDelay = 1 * time.Millisecond
	
	callCount := 0
	fn := func() error {
		callCount++
		return mcperrors.Newf(mcperrors.ErrCodeConnection, "connection failed")
	}
	
	err := WithRetry(context.Background(), config, fn)
	if err == nil {
		t.Error("should return error after max retries")
	}
	if callCount != 3 {
		t.Errorf("should call function 3 times (initial + 2 retries), got %d", callCount)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	config := DefaultConfig()
	config.MaxRetries = 3
	config.InitialDelay = 1 * time.Millisecond
	
	callCount := 0
	fn := func() error {
		callCount++
		return mcperrors.Newf(mcperrors.ErrCodeServer, "server error")
	}
	
	err := WithRetry(context.Background(), config, fn)
	if err == nil {
		t.Error("should return error")
	}
	if callCount != 1 {
		t.Errorf("should not retry non-retryable error, got %d calls", callCount)
	}
}

func TestWithRetry_ContextCancelled(t *testing.T) {
	config := DefaultConfig()
	config.MaxRetries = 10
	config.InitialDelay = 100 * time.Millisecond
	
	ctx, cancel := context.WithCancel(context.Background())
	
	callCount := 0
	fn := func() error {
		callCount++
		return mcperrors.Newf(mcperrors.ErrCodeConnection, "connection failed")
	}
	
	// Cancel context after 1st retry
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	
	err := WithRetry(ctx, config, fn)
	if err != context.Canceled {
		t.Errorf("should return context.Canceled, got %v", err)
	}
	// Should stop retrying after context is cancelled
	if callCount > 2 {
		t.Errorf("should stop after context cancelled, got %d calls", callCount)
	}
}

func TestCalculateDelay(t *testing.T) {
	tests := []struct {
		name       string
		initial    time.Duration
		multiplier float64
		attempt    int
		wantMin    time.Duration
		wantMax    time.Duration
	}{
		{
			name:       "first attempt",
			initial:    100 * time.Millisecond,
			multiplier: 2.0,
			attempt:    0,
			wantMin:    75 * time.Millisecond,  // 100 * 1.0 + jitter
			wantMax:    125 * time.Millisecond,
		},
		{
			name:       "second attempt",
			initial:    100 * time.Millisecond,
			multiplier: 2.0,
			attempt:    1,
			wantMin:    150 * time.Millisecond, // 100 * 2.0 + jitter
			wantMax:    250 * time.Millisecond,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateDelay(tt.initial, tt.multiplier, tt.attempt)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateDelay() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	config := DefaultConfig()
	
	tests := []struct {
		name    string
		err     error
		retry   bool
	}{
		{
			name:  "nil error",
			err:   nil,
			retry: false,
		},
		{
			name:  "connection error",
			err:   mcperrors.Newf(mcperrors.ErrCodeConnection, "connection refused"),
			retry: true,
		},
		{
			name:  "timeout error",
			err:   mcperrors.Newf(mcperrors.ErrCodeTimeout, "timeout"),
			retry: true,
		},
		{
			name:  "server error",
			err:   mcperrors.Newf(mcperrors.ErrCodeServer, "server error"),
			retry: false,
		},
		{
			name:  "network keyword",
			err:   errors.New("network error"),
			retry: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryable(tt.err, config.RetryableErrors)
			if got != tt.retry {
				t.Errorf("isRetryable() = %v, want %v", got, tt.retry)
			}
		})
	}
}
