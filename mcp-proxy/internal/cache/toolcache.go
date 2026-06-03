package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// CacheConfig 缓存配置
type CacheConfig struct {
	// Enabled 是否启用缓存
	Enabled bool `json:"enabled"`
	// TTL 缓存有效期
	TTL time.Duration `json:"ttl,omitempty"`
	// MaxSize 最大缓存条目数
	MaxSize int `json:"maxSize,omitempty"`
	// CacheableTools 可缓存的工具列表（空表示全部可缓存）
	CacheableTools []string `json:"cacheableTools,omitempty"`
}

// DefaultCacheConfig 返回默认缓存配置
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Enabled: false,
		TTL:      5 * time.Minute,
		MaxSize: 1000,
	}
}

// cacheEntry 缓存条目
type cacheEntry struct {
	result    *mcp.CallToolResult
	timestamp time.Time
}

// CacheKey 缓存键组成部分
type CacheKey struct {
	service string
	tool    string
	args    map[string]any
}

// hash 生成哈希键
func (k CacheKey) hash() (string, error) {
	keyData := struct {
		Service string
		Tool    string
		Args    map[string]any
	}{
		Service: k.service,
		Tool:    k.tool,
		Args:    k.args,
	}

	jsonBytes, err := json.Marshal(keyData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cache key: %w", err)
	}

	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:]), nil
}

// ToolCallCache 工具调用缓存
type ToolCallCache struct {
	mu      sync.RWMutex
	cache   map[string]cacheEntry
	config  CacheConfig
	hits    int
	misses  int
}

// NewToolCallCache 创建新的工具调用缓存
func NewToolCallCache(config *CacheConfig) *ToolCallCache {
	if config == nil {
		config = DefaultCacheConfig()
	}
	if config.MaxSize <= 0 {
		config.MaxSize = 1000
	}
	if config.TTL <= 0 {
		config.TTL = 5 * time.Minute
	}

	return &ToolCallCache{
		cache:  make(map[string]cacheEntry, config.MaxSize),
		config: *config,
	}
}

// isCacheable 检查工具是否可缓存
func (c *ToolCallCache) isCacheable(tool string) bool {
	if !c.config.Enabled {
		return false
	}

	if len(c.config.CacheableTools) == 0 {
		return true
	}

	for _, t := range c.config.CacheableTools {
		if t == tool {
			return true
		}
	}

	return false
}

// evictIfNecessary 按需驱逐旧条目
func (c *ToolCallCache) evictIfNecessary() {
	if len(c.cache) < c.config.MaxSize {
		return
	}

	// 简单的 LRU 实现：删除最早的条目
	var oldestKey string
	var oldestTime time.Time

	for k, entry := range c.cache {
		if oldestKey == "" || entry.timestamp.Before(oldestTime) {
			oldestKey = k
			oldestTime = entry.timestamp
		}
	}

	delete(c.cache, oldestKey)
}

// Get 获取缓存结果
func (c *ToolCallCache) Get(service, tool string, args map[string]any) (*mcp.CallToolResult, bool) {
	if !c.isCacheable(tool) {
		return nil, false
	}

	key := CacheKey{service: service, tool: tool, args: args}
	hashKey, err := key.hash()
	if err != nil {
		return nil, false
	}

	c.mu.RLock()
	entry, ok := c.cache[hashKey]
	c.mu.RUnlock()

	if !ok {
		c.misses++
		return nil, false
	}

	// 检查是否过期
	if time.Since(entry.timestamp) > c.config.TTL {
		c.mu.Lock()
		delete(c.cache, hashKey)
		c.mu.Unlock()
		c.misses++
		return nil, false
	}

	c.hits++
	return entry.result, true
}

// Set 缓存结果
func (c *ToolCallCache) Set(service, tool string, args map[string]any, result *mcp.CallToolResult) {
	if !c.isCacheable(tool) {
		return
	}

	key := CacheKey{service: service, tool: tool, args: args}
	hashKey, err := key.hash()
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// 驱逐旧条目
	c.evictIfNecessary()

	c.cache[hashKey] = cacheEntry{
		result:    result,
		timestamp: time.Now(),
	}
}

// Stats 返回缓存统计信息
type CacheStats struct {
	Hits    int
	Misses  int
	HitRate float64
	Size    int
	MaxSize int
}

func (c *ToolCallCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:    c.hits,
		Misses:  c.misses,
		HitRate: hitRate,
		Size:    len(c.cache),
		MaxSize: c.config.MaxSize,
	}
}

// Clear 清空缓存
func (c *ToolCallCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]cacheEntry, c.config.MaxSize)
}

// Remove 移除特定服务的缓存（简化版）
func (c *ToolCallCache) Remove(service string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 清空所有缓存，因为我们无法从哈希键中反推出服务名
	// 简化起见，清空所有
	c.cache = make(map[string]cacheEntry, c.config.MaxSize)
}

// UpdateConfig 更新缓存配置
func (c *ToolCallCache) UpdateConfig(config *CacheConfig) {
	if config == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if config.MaxSize > 0 {
		c.config.MaxSize = config.MaxSize
	}
	if config.TTL > 0 {
		c.config.TTL = config.TTL
	}
	c.config.Enabled = config.Enabled
	c.config.CacheableTools = config.CacheableTools
}
