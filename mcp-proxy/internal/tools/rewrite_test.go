package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestRewriteConfig_DefaultConfig(t *testing.T) {
	config := DefaultRewriteConfig()
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Namespace != "" {
		t.Errorf("expected empty namespace, got %s", config.Namespace)
	}
	if config.Prefix != "" {
		t.Errorf("expected empty prefix, got %s", config.Prefix)
	}
	if config.Rename != nil {
		t.Error("expected nil rename map")
	}
}

func TestRewriteToolName_NoConfig(t *testing.T) {
	result := rewriteToolName("test-tool", nil)
	if result != "test-tool" {
		t.Errorf("expected 'test-tool', got %s", result)
	}
}

func TestRewriteToolName_Namespace(t *testing.T) {
	config := &RewriteConfig{Namespace: "github"}
	result := rewriteToolName("create_issue", config)
	if result != "github.create_issue" {
		t.Errorf("expected 'github.create_issue', got %s", result)
	}
}

func TestRewriteToolName_Prefix(t *testing.T) {
	config := &RewriteConfig{Prefix: "gh_"}
	result := rewriteToolName("create_issue", config)
	if result != "gh_create_issue" {
		t.Errorf("expected 'gh_create_issue', got %s", result)
	}
}

func TestRewriteToolName_Rename(t *testing.T) {
	config := &RewriteConfig{
		Rename: map[string]string{
			"create_issue": "new_issue",
		},
	}
	result := rewriteToolName("create_issue", config)
	if result != "new_issue" {
		t.Errorf("expected 'new_issue', got %s", result)
	}
}

func TestRewriteToolName_FullConfig(t *testing.T) {
	config := &RewriteConfig{
		Namespace: "github",
		Prefix:   "gh_",
		Rename: map[string]string{
			"create_issue": "create_github_issue",
		},
	}
	
	// 精确重命名优先
	result := rewriteToolName("create_issue", config)
	if result != "create_github_issue" {
		t.Errorf("expected 'create_github_issue', got %s", result)
	}
	
	// 其他工具使用命名空间+前缀
	result = rewriteToolName("search_repos", config)
	if result != "github.gh_search_repos" {
		t.Errorf("expected 'github.gh_search_repos', got %s", result)
	}
}

func TestRewriteTools(t *testing.T) {
	tools := []mcp.Tool{
		{Name: "tool1", Description: "Tool 1"},
		{Name: "tool2", Description: "Tool 2"},
	}
	
	config := &RewriteConfig{Prefix: "test_"}
	rewritten := RewriteTools(tools, config)
	
	if len(rewritten) != len(tools) {
		t.Fatalf("expected %d tools, got %d", len(tools), len(rewritten))
	}
	
	if rewritten[0].Name != "test_tool1" {
		t.Errorf("expected 'test_tool1', got %s", rewritten[0].Name)
	}
	
	// 原始工具不应该被修改
	if tools[0].Name != "tool1" {
		t.Errorf("original tool should not be modified, got %s", tools[0].Name)
	}
}

func TestParseToolName(t *testing.T) {
	tests := []struct {
		input        string
		wantOriginal string
		wantNS       string
	}{
		{"tool", "tool", ""},
		{"github.tool", "tool", "github"},
		{"a.b.c", "b.c", "a"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			original, ns := ParseToolName(tt.input)
			if original != tt.wantOriginal {
				t.Errorf("ParseToolName(%q) original = %q, want %q", tt.input, original, tt.wantOriginal)
			}
			if ns != tt.wantNS {
				t.Errorf("ParseToolName(%q) namespace = %q, want %q", tt.input, ns, tt.wantNS)
			}
		})
	}
}

func TestBuildToolName(t *testing.T) {
	config := &RewriteConfig{
		Namespace: "github",
		Prefix:   "gh_",
	}
	
	result := BuildToolName("my-service", "create_issue", config)
	expected := "github.my-service.gh_create_issue"
	if result != expected {
		t.Errorf("BuildToolName() = %q, want %q", result, expected)
	}
}

func TestValidateRewriteConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *RewriteConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "empty config",
			config:  &RewriteConfig{},
			wantErr: false,
		},
		{
			name: "self-rename",
			config: &RewriteConfig{
				Rename: map[string]string{
					"tool": "tool",
				},
			},
			wantErr: true,
		},
		{
			name: "circular rename",
			config: &RewriteConfig{
				Rename: map[string]string{
					"tool1": "tool2",
					"tool2": "tool1",
				},
			},
			wantErr: true,
		},
		{
			name: "valid rename",
			config: &RewriteConfig{
				Rename: map[string]string{
					"old_name": "new_name",
				},
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRewriteConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRewriteConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMergeRewriteConfig(t *testing.T) {
	base := &RewriteConfig{
		Namespace: "base-ns",
		Prefix:    "base_",
		Rename: map[string]string{
			"base_tool": "base_renamed",
		},
	}
	
	override := &RewriteConfig{
		Namespace: "override-ns",
		Prefix:    "override_",
		Rename: map[string]string{
			"override_tool": "override_renamed",
		},
	}
	
	merged := MergeRewriteConfig(base, override)
	
	if merged.Namespace != "override-ns" {
		t.Errorf("expected Namespace 'override-ns', got %s", merged.Namespace)
	}
	
	if merged.Prefix != "override_" {
		t.Errorf("expected Prefix 'override_', got %s", merged.Prefix)
	}
	
	// base rename 应该被保留（因为 override 中没有覆盖它）
	if _, ok := merged.Rename["base_tool"]; !ok {
		t.Error("base rename should be preserved")
	}
	
	// override rename 应该存在
	if _, ok := merged.Rename["override_tool"]; !ok {
		t.Error("override rename should exist")
	}
}
