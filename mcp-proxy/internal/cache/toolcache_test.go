package cache

import (
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewToolCallCache(t *testing.T) {
	config := DefaultCacheConfig()
	cache := NewToolCallCache(config)

	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if !cache.config.Enabled == config.Enabled {
		t.Error("enabled should match")
	}
	if cache.config.TTL != config.TTL {
		t.Errorf("TTL should be %v, got %v", config.TTL, cache.config.TTL)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = true
	config.TTL = 1 * time.Hour
	config.MaxSize = 100

	cache := NewToolCallCache(config)

	result := &mcp.CallToolResult{}

	// Set cache
	cache.Set("test-service", "test-tool", map[string]any{"param": "value"}, result)

	// Get cache
	cachedResult, ok := cache.Get("test-service", "test-tool", map[string]any{"param": "value"})

	if !ok {
		t.Fatal("expected cache hit")
	}
	if cachedResult == nil {
		t.Error("expected non-nil result")
	}
}

func TestCacheMiss(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = true
	cache := NewToolCallCache(config)

	_, ok := cache.Get("unknown", "tool", map[string]any{})
	if ok {
		t.Error("expected cache miss")
	}

	if cache.Stats().Misses != 1 {
		t.Errorf("expected 1 miss, got %d", cache.Stats().Misses)
	}
}

func TestCacheDisabled(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = false
	cache := NewToolCallCache(config)

	result := &mcp.CallToolResult{}
	cache.Set("service", "tool", map[string]any{}, result)

	_, ok := cache.Get("service", "tool", map[string]any{})
	if ok {
		t.Error("should not cache when disabled")
	}
}

func TestCacheableToolsList(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = true
	config.CacheableTools = []string{"allowed1", "allowed2"}

	cache := NewToolCallCache(config)

	result := &mcp.CallToolResult{}

	// Should cache
	cache.Set("service", "allowed1", map[string]any{}, result)
	cached, ok := cache.Get("service", "allowed1", map[string]any{})
	if !ok || cached == nil {
		t.Error("should cache allowed tool")
	}

	// Should not cache
	cache.Set("service", "disallowed", map[string]any{}, result)
	cached, ok = cache.Get("service", "disallowed", map[string]any{})
	if ok || cached != nil {
		t.Error("should not cache disallowed tool")
	}
}

func TestCacheExpiration(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = true
	config.TTL = 10 * time.Millisecond

	cache := NewToolCallCache(config)

	result := &mcp.CallToolResult{}
	cache.Set("service", "tool", map[string]any{}, result)

	// Should hit immediately
	cached, ok := cache.Get("service", "tool", map[string]any{})
	if !ok {
		t.Error("should hit before expiration")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Should miss after expiration
	cached, ok = cache.Get("service", "tool", map[string]any{})
	if ok || cached != nil {
		t.Error("should miss after expiration")
	}
}

func TestCacheEviction(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = true
	config.MaxSize = 3

	cache := NewToolCallCache(config)

	result := &mcp.CallToolResult{}

	for i := 1; i <= 5; i++ {
		args := map[string]any{"id": i}
		cache.Set("service", "tool", args, result)
	}

	if len(cache.cache) > config.MaxSize {
		t.Errorf("cache size should not exceed max size, got %d", len(cache.cache))
	}
}

func TestCacheStats(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = true
	cache := NewToolCallCache(config)

	result := &mcp.CallToolResult{}

	cache.Set("service", "tool", map[string]any{"id": 1}, result)
	cache.Set("service", "tool", map[string]any{"id": 2}, result)
	cache.Set("service", "tool", map[string]any{"id": 3}, result)

	// Hits
	cache.Get("service", "tool", map[string]any{"id": 1}) // hit
	cache.Get("service", "tool", map[string]any{"id": 2}) // hit
	cache.Get("service", "tool", map[string]any{"id": 4}) // miss

	stats := cache.Stats()

	if stats.Hits != 2 {
		t.Errorf("expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
	if stats.HitRate != 0.6666666666666666 {
		t.Errorf("expected hit rate ~0.666, got %f", stats.HitRate)
	}
	if stats.Size != 3 {
		t.Errorf("expected size 3, got %d", stats.Size)
	}
}

func TestCacheClear(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = true
	cache := NewToolCallCache(config)

	result := &mcp.CallToolResult{}
	cache.Set("service", "tool", map[string]any{}, result)

	if len(cache.cache) != 1 {
		t.Error("cache should have one entry")
	}

	cache.Clear()

	if len(cache.cache) != 0 {
		t.Error("cache should be empty after clear")
	}
}

func TestCacheUpdateConfig(t *testing.T) {
	config := DefaultCacheConfig()
	config.Enabled = true
	cache := NewToolCallCache(config)

	newConfig := &CacheConfig{
		Enabled: false,
		TTL:     10 * time.Minute,
		MaxSize: 2000,
	}

	cache.UpdateConfig(newConfig)

	if cache.config.Enabled != false {
		t.Error("enabled should be false after update")
	}
	if cache.config.TTL != 10*time.Minute {
		t.Errorf("TTL should be updated")
	}
}
