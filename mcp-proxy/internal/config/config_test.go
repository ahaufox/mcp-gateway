package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFile drops body into path, creating any missing parent directories.
func writeFile(t *testing.T, path, body string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
}

// sampleBase is a reusable base.json fixture.
const sampleBase = `{
  "mcpProxy": {
    "baseURL": "http://localhost:9090",
    "addr": ":9090",
    "name": "MCP Proxy",
    "version": "1.0.0",
    "type": "streamable-http",
    "options": {
      "panicIfInvalid": false,
      "logEnabled": true,
      "authTokens": ["default-token"]
    }
  }
}`

func TestPartialConfigMergeInto_McpProxyLastWriteWins(t *testing.T) {
	a := &PartialConfig{McpProxy: &MCPProxyConfigV2{Name: "first"}}
	b := &PartialConfig{McpProxy: &MCPProxyConfigV2{Name: "second"}}
	a.mergeInto(b)
	require.NotNil(t, a.McpProxy)
	assert.Equal(t, "second", a.McpProxy.Name)
}

func TestPartialConfigMergeInto_McpServersByName(t *testing.T) {
	a := &PartialConfig{McpServers: map[string]*MCPClientConfigV2{
		"alpha": {Command: "alpha-a"},
		"beta":  {Command: "beta-a"},
	}}
	b := &PartialConfig{McpServers: map[string]*MCPClientConfigV2{
		"beta":  {Command: "beta-b"},
		"gamma": {Command: "gamma-b"},
	}}
	a.mergeInto(b)

	assert.Equal(t, "alpha-a", a.McpServers["alpha"].Command, "untouched entry kept")
	assert.Equal(t, "beta-b", a.McpServers["beta"].Command, "overlapping entry replaced")
	assert.Equal(t, "gamma-b", a.McpServers["gamma"].Command, "new entry added")
}

func TestPartialConfigMergeInto_NilOtherIsNoop(t *testing.T) {
	a := &PartialConfig{McpProxy: &MCPProxyConfigV2{Name: "kept"}}
	a.mergeInto(nil)
	require.NotNil(t, a.McpProxy)
	assert.Equal(t, "kept", a.McpProxy.Name)
}

func TestCollectConfigFiles_LexicalOrder(t *testing.T) {
	root := t.TempDir()
	// Create files in intentionally non-sorted order; the walker must
	// return them sorted by relative path.
	writeFile(t, filepath.Join(root, "overrides", "z.json"), `{}`)
	writeFile(t, filepath.Join(root, "base.json"), `{}`)
	writeFile(t, filepath.Join(root, "categories", "stdio.json"), `{}`)
	writeFile(t, filepath.Join(root, "categories", "sse.json"), `{}`)
	// Hidden basenames and non-JSON files are skipped.
	writeFile(t, filepath.Join(root, "categories", ".hidden.json"), `{}`)
	writeFile(t, filepath.Join(root, "categories", "notes.md"), `# nope`)

	files, err := collectConfigFiles(root)
	require.NoError(t, err)
	expected := []string{
		"base.json",
		"categories/sse.json",
		"categories/stdio.json",
		"overrides/z.json",
	}
	assert.Equal(t, expected, files)
}

func TestLoadDir_MergesAcrossCategories(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "base.json"), sampleBase)
	writeFile(t, filepath.Join(root, "categories", "stdio.json"), `{
      "mcpServers": {
        "github": {
          "command": "npx",
          "args": ["-y", "@modelcontextprotocol/server-github"]
        }
      }
    }`)
	writeFile(t, filepath.Join(root, "categories", "sse.json"), `{
      "mcpServers": {
        "jules": { "transportType": "sse", "url": "http://jules/sse" }
      }
    }`)
	writeFile(t, filepath.Join(root, "categories", "streamable-http.json"), `{
      "mcpServers": {
        "stitch": {
          "transportType": "streamable-http",
          "url": "https://stitch.googleapis.com/mcp"
        }
      }
    }`)

	cfg, err := LoadDir(root, false, false, "", 0)
	require.NoError(t, err)
	require.NotNil(t, cfg.McpProxy)
	assert.Equal(t, "MCP Proxy", cfg.McpProxy.Name)
	assert.Equal(t, []string{"default-token"}, cfg.McpProxy.Options.AuthTokens,
		"base auth tokens should be split, not joined")

	require.Len(t, cfg.McpServers, 3, "all three categories contribute servers")
	assert.Equal(t, "npx", cfg.McpServers["github"].Command)
	assert.Equal(t, "http://jules/sse", cfg.McpServers["jules"].URL)
	assert.Equal(t, "https://stitch.googleapis.com/mcp", cfg.McpServers["stitch"].URL)
}

func TestLoadDir_OverridesLastWriteWins(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "base.json"), sampleBase)
	// categories/stdio.json declares github with one set of args.
	writeFile(t, filepath.Join(root, "categories", "stdio.json"), `{
      "mcpServers": {
        "github": {
          "command": "npx",
          "args": ["-y", "old-version"]
        }
      }
    }`)
	// overrides/local.json lives under the overrides/ subdir which sorts
	// after categories/, so it must win.
	writeFile(t, filepath.Join(root, "overrides", "local.json"), `{
      "mcpServers": {
        "github": {
          "command": "npx",
          "args": ["-y", "new-version"]
        }
      }
    }`)

	cfg, err := LoadDir(root, false, false, "", 0)
	require.NoError(t, err)
	require.Contains(t, cfg.McpServers, "github")
	assert.Equal(t, []string{"-y", "new-version"}, cfg.McpServers["github"].Args,
		"override file should win over category file")
}

func TestLoadDir_OverrideCanAddNewServer(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "base.json"), sampleBase)
	writeFile(t, filepath.Join(root, "overrides", "extra.json"), `{
      "mcpServers": {
        "personal": { "command": "node", "args": ["./personal-mcp.js"] }
      }
    }`)

	cfg, err := LoadDir(root, false, false, "", 0)
	require.NoError(t, err)
	require.Contains(t, cfg.McpServers, "personal")
	assert.Equal(t, "node", cfg.McpServers["personal"].Command)
}

func TestLoadDir_EmptyDirIsError(t *testing.T) {
	root := t.TempDir()
	_, err := LoadDir(root, false, false, "", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no .json config files found")
}

func TestLoadDir_MissingDirIsError(t *testing.T) {
	_, err := LoadDir(filepath.Join(t.TempDir(), "does-not-exist"), false, false, "", 0)
	require.Error(t, err)
}

func TestLoadDir_FilePathIsError(t *testing.T) {
	// Passing a file (not a dir) must be rejected clearly, not silently
	// fall through to an empty load.
	root := t.TempDir()
	file := filepath.Join(root, "single.json")
	writeFile(t, file, sampleBase)
	_, err := LoadDir(file, false, false, "", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestLoadDir_MissingBaseMcpProxyIsError(t *testing.T) {
	root := t.TempDir()
	// Only categories, no base.json -> mcpProxy is nil after merge.
	writeFile(t, filepath.Join(root, "categories", "stdio.json"), `{
      "mcpServers": { "github": { "command": "npx" } }
    }`)
	_, err := LoadDir(root, false, false, "", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no mcpProxy block")
}

func TestLoadDir_BadJSONIsError(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "base.json"), `{ not valid json`)
	_, err := LoadDir(root, false, false, "", 0)
	require.Error(t, err)
}

func TestLoadDir_ServerOptionsInheritFromProxy(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "base.json"), `{
      "mcpProxy": {
        "baseURL": "http://localhost:9090",
        "addr": ":9090",
        "name": "MCP Proxy",
        "version": "1.0.0",
        "options": {
          "panicIfInvalid": false,
          "logEnabled": true,
          "authTokens": ["global-token"]
        }
      }
    }`)
	writeFile(t, filepath.Join(root, "categories", "stdio.json"), `{
      "mcpServers": {
        "github": { "command": "npx" }
      }
    }`)

	cfg, err := LoadDir(root, false, false, "", 0)
	require.NoError(t, err)
	gopts := cfg.McpServers["github"].Options
	require.NotNil(t, gopts)
	assert.Equal(t, []string{"global-token"}, gopts.AuthTokens,
		"per-server auth tokens should inherit from mcpProxy when unset")
}

func TestLoadDir_SplitCommaAuthTokens(t *testing.T) {
	root := t.TempDir()
	// Comma-separated token values are split into multiple entries,
	// matching the behavior of single-file Load.
	writeFile(t, filepath.Join(root, "base.json"), `{
      "mcpProxy": {
        "baseURL": "http://localhost:9090",
        "addr": ":9090",
        "name": "MCP Proxy",
        "version": "1.0.0",
        "options": { "authTokens": ["a,b , c"] }
      }
    }`)
	cfg, err := LoadDir(root, false, false, "", 0)
	require.NoError(t, err)
	sort.Strings(cfg.McpProxy.Options.AuthTokens)
	assert.Equal(t, []string{"a", "b", "c"}, cfg.McpProxy.Options.AuthTokens)
}

func TestLoadDir_MergedConfigIsJSONSerializable(t *testing.T) {
	// /api/config dumps the in-memory merged view via json.Marshal; this test
	// guards against accidentally adding unexported fields to Config that
	// would break that endpoint.
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "base.json"), sampleBase)
	writeFile(t, filepath.Join(root, "categories", "stdio.json"), `{
      "mcpServers": { "github": { "command": "npx" } }
    }`)

	cfg, err := LoadDir(root, false, false, "", 0)
	require.NoError(t, err)
	b, err := json.Marshal(cfg)
	require.NoError(t, err)
	s := string(b)
	assert.True(t, strings.Contains(s, `"mcpProxy"`))
	assert.True(t, strings.Contains(s, `"mcpServers"`))
	assert.True(t, strings.Contains(s, `"github"`))
}
