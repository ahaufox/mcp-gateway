package tools

import (
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// RewriteConfig 工具重写配置
type RewriteConfig struct {
	// Namespace 添加命名空间前缀
	Namespace string `json:"namespace,omitempty"`
	// Prefix 添加工具名前缀
	Prefix string `json:"prefix,omitempty"`
	// Rename 精确重命名映射表
	Rename map[string]string `json:"rename,omitempty"`
}

// DefaultRewriteConfig 返回默认配置（不做任何修改）
func DefaultRewriteConfig() *RewriteConfig {
	return &RewriteConfig{}
}

// rewriteToolName 根据配置重写工具名
func rewriteToolName(original string, config *RewriteConfig) string {
	if config == nil {
		return original
	}

	// 1. 先检查精确重命名映射
	if config.Rename != nil {
		if newName, ok := config.Rename[original]; ok {
			return newName
		}
	}

	name := original

	// 2. 添加前缀
	if config.Prefix != "" {
		name = config.Prefix + name
	}

	// 3. 添加命名空间
	if config.Namespace != "" {
		name = config.Namespace + "." + name
	}

	return name
}

// RewriteTools 根据配置重写工具列表
func RewriteTools(tools []mcp.Tool, config *RewriteConfig) []mcp.Tool {
	if config == nil {
		return tools
	}

	rewritten := make([]mcp.Tool, len(tools))
	for i, tool := range tools {
		rewritten[i] = tool
		rewritten[i].Name = rewriteToolName(tool.Name, config)
	}

	return rewritten
}

// ToolNameWithContext 带有完整上下文信息的工具名
type ToolNameWithContext struct {
	OriginalName string
	RewrittenName string
	ServiceName   string
	Namespace     string
	Prefix        string
	IsRenamed     bool
}

// ParseToolName 解析工具名，返回原始名和命名空间
func ParseToolName(name string) (original, namespace string) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 2 {
		return parts[1], parts[0]
	}
	return name, ""
}

// BuildToolName 构建完整的工具名
func BuildToolName(service, tool string, config *RewriteConfig) string {
	if config == nil {
		return tool
	}

	parts := []string{}
	
	if config.Namespace != "" {
		parts = append(parts, config.Namespace)
	}
	
	// 添加服务名作为第二级命名空间
	parts = append(parts, service)
	
	// 添加工具名
	name := tool
	if config.Prefix != "" {
		name = config.Prefix + name
	}
	parts = append(parts, name)
	
	return strings.Join(parts, ".")
}

// ValidateRewriteConfig 验证重写配置
func ValidateRewriteConfig(config *RewriteConfig) error {
	if config == nil {
		return nil
	}

	// 检查重命名映射是否有循环引用
	if config.Rename != nil {
		for original, renamed := range config.Rename {
			if original == renamed {
				return fmt.Errorf("self-rename detected: %s -> %s", original, renamed)
			}
			
			// 检查是否存在 A -> B 和 B -> A 的情况
			if config.Rename[renamed] == original {
				return fmt.Errorf("circular rename detected: %s <-> %s", original, renamed)
			}
		}
	}

	return nil
}

// MergeRewriteConfig 合并多个重写配置
func MergeRewriteConfig(base, override *RewriteConfig) *RewriteConfig {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}

	result := &RewriteConfig{
		Namespace: override.Namespace,
		Prefix:    override.Prefix,
		Rename:    make(map[string]string),
	}

	// 复制基础配置的映射
	for k, v := range base.Rename {
		result.Rename[k] = v
	}

	// 覆盖
	for k, v := range override.Rename {
		result.Rename[k] = v
	}

	return result
}
