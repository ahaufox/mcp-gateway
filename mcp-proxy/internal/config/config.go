package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	nethttp "net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sphere/confstore"
	"github.com/go-sphere/confstore/codec"
	"github.com/go-sphere/confstore/provider"
	"github.com/go-sphere/confstore/provider/file"
	"github.com/go-sphere/confstore/provider/http"
	"github.com/tbxark/optional-go"
)

// Trae 兼容: 标准 MCP 配置识别字段
// 参考: https://docs.trae.cn/ide_model-context-protocol 与 https://docs.trae.cn/ide_add-mcp-servers
const (
	// TraeStartTimeoutKey Trae 标准的启动超时字段名。
	// stdio 放在 env 中, HTTP/SSE 放在 headers 中。
	TraeStartTimeoutKey = "START_MCP_TIMEOUT_MS"
	// TraeRunTimeoutKey Trae 标准的 tools 调用超时字段名。
	// stdio 放在 env 中, HTTP/SSE 放在 headers 中。
	TraeRunTimeoutKey = "RUN_MCP_TIMEOUT_MS"
	// TraeWorkspaceVar Trae 配置中允许的变量占位符。
	TraeWorkspaceVar = "${workspaceFolder}"

	// EnvWorkspaceFolder 工作区路径的兜底环境变量。
	EnvWorkspaceFolder = "MCP_WORKSPACE_FOLDER"
)

type StdioMCPClientConfig struct {
	Command string            `json:"command"`
	Env     map[string]string `json:"env"`
	Args    []string          `json:"args"`
}

type SSEMCPClientConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type StreamableMCPClientConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Timeout time.Duration     `json:"timeout"`
}

type WebSocketMCPClientConfig struct {
	URL                  string            `json:"url"`
	Headers              map[string]string `json:"headers"`
	HandshakeTimeout     time.Duration     `json:"handshakeTimeout,omitempty"`
	PingInterval         time.Duration     `json:"pingInterval,omitempty"`
	ReconnectDelay       time.Duration     `json:"reconnectDelay,omitempty"`
	MaxReconnectAttempts int               `json:"maxReconnectAttempts,omitempty"`
}

type MCPClientType string

const (
	MCPClientTypeStdio      MCPClientType = "stdio"
	MCPClientTypeSSE        MCPClientType = "sse"
	MCPClientTypeStreamable MCPClientType = "streamable-http"
	MCPClientTypeWebSocket  MCPClientType = "websocket"
)

type MCPServerType string

const (
	MCPServerTypeSSE        MCPServerType = "sse"
	MCPServerTypeStreamable MCPServerType = "streamable-http"
)

// ---- V2 ----

type ToolFilterMode string

const (
	ToolFilterModeAllow ToolFilterMode = "allow"
	ToolFilterModeBlock ToolFilterMode = "block"
)

type ToolFilterConfig struct {
	Mode ToolFilterMode `json:"mode,omitempty"`
	List []string       `json:"list,omitempty"`
}

type CircuitBreakerConfig struct {
	Enabled      bool          `json:"enabled,omitempty"`
	MaxFailures  int           `json:"maxFailures,omitempty"`
	ResetTimeout time.Duration `json:"resetTimeout,omitempty"`
	HalfOpenMax  int           `json:"halfOpenMax,omitempty"`
}

type OptionsV2 struct {
	PanicIfInvalid      optional.Field[bool] `json:"panicIfInvalid,omitempty"`
	LogEnabled          optional.Field[bool] `json:"logEnabled,omitempty"`
	DisablePing         optional.Field[bool] `json:"disablePing,omitempty"`
	MaintenanceInterval time.Duration        `json:"maintenanceInterval,omitempty"`
	AuthTokens          []string             `json:"authTokens,omitempty"`
	ToolFilter          *ToolFilterConfig    `json:"toolFilter,omitempty"`
	Disabled            bool                 `json:"disabled,omitempty"`

	// 超时配置
	CallTimeout       time.Duration `json:"callTimeout,omitempty"`
	InitializeTimeout time.Duration `json:"initializeTimeout,omitempty"`
	ListToolsTimeout  time.Duration `json:"listToolsTimeout,omitempty"`

	// 重试配置
	MaxRetries   int           `json:"maxRetries,omitempty"`
	RetryDelay   time.Duration `json:"retryDelay,omitempty"`
	RetryBackoff float64       `json:"retryBackoff,omitempty"`

	// 熔断器配置
	CircuitBreaker *CircuitBreakerConfig `json:"circuitBreaker,omitempty"`
}

type MCPProxyConfigV2 struct {
	BaseURL         string        `json:"baseURL"`
	Addr            string        `json:"addr"`
	Name            string        `json:"name"`
	Version         string        `json:"version"`
	Type            MCPServerType `json:"type,omitempty"`
	WorkspaceFolder string        `json:"workspaceFolder,omitempty"`
	Options         *OptionsV2    `json:"options,omitempty"`
}

type MCPClientConfigV2 struct {
	Description   string        `json:"description,omitempty"`
	TransportType MCPClientType `json:"transportType,omitempty"`

	// Stdio
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	// SSE or Streamable HTTP
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`

	Options *OptionsV2 `json:"options,omitempty"`
}

// popTraeTimeout 从 env (stdio) 或 headers (HTTP/SSE) 中读取并删除 Trae 标准超时键。
// 优先读取 env，env 不存在时再尝试 headers。返回的数值单位为毫秒。
func popTraeTimeout(env *map[string]string, headers *map[string]string, key string) (time.Duration, bool) {
	if env != nil {
		if v, ok := (*env)[key]; ok {
			delete(*env, key)
			if n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil && n > 0 {
				return time.Duration(n) * time.Millisecond, true
			}
		}
	}
	if headers != nil {
		if v, ok := (*headers)[key]; ok {
			delete(*headers, key)
			if n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil && n > 0 {
				return time.Duration(n) * time.Millisecond, true
			}
		}
	}
	return 0, false
}

// applyTraeCompat 将 Trae 风格的 MCP server 配置归一化为 mcp-proxy 内部表示:
//  1. 从 env (stdio) / headers (HTTP/SSE) 中提取 START_MCP_TIMEOUT_MS /
//     RUN_MCP_TIMEOUT_MS，映射到 Options.InitializeTimeout / CallTimeout，
//     并在原 map 中删除，避免泄漏给子进程或上游服务。
//  2. 将 args / env 值 / header 值中的 ${workspaceFolder} 占位符替换为
//     实际的工作区路径。其它占位符 (例如 ${VAR}) 由 go-sphere confstore 在
//     Load 阶段处理，不在此处展开。
func applyTraeCompat(conf *MCPClientConfigV2, workspaceFolder string) {
	if conf == nil {
		return
	}
	if conf.Options == nil {
		conf.Options = &OptionsV2{}
	}

	// 1) 提取并移除 Trae 超时字段
	if d, ok := popTraeTimeout(&conf.Env, &conf.Headers, TraeStartTimeoutKey); ok && conf.Options.InitializeTimeout == 0 {
		conf.Options.InitializeTimeout = d
	}
	if d, ok := popTraeTimeout(&conf.Env, &conf.Headers, TraeRunTimeoutKey); ok && conf.Options.CallTimeout == 0 {
		conf.Options.CallTimeout = d
	}

	// 2) 替换 ${workspaceFolder}
	if workspaceFolder == "" {
		return
	}
	for i, arg := range conf.Args {
		conf.Args[i] = strings.ReplaceAll(arg, TraeWorkspaceVar, workspaceFolder)
	}
	for k, v := range conf.Env {
		conf.Env[k] = strings.ReplaceAll(v, TraeWorkspaceVar, workspaceFolder)
	}
	for k, v := range conf.Headers {
		conf.Headers[k] = strings.ReplaceAll(v, TraeWorkspaceVar, workspaceFolder)
	}
}

// resolveWorkspaceFolder 解析 ${workspaceFolder} 的替换值。
// 优先级: mcpProxy.workspaceFolder > MCP_WORKSPACE_FOLDER 环境变量 > 当前工作目录。
func resolveWorkspaceFolder(configured string) string {
	if v := strings.TrimSpace(configured); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv(EnvWorkspaceFolder)); v != "" {
		return v
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return ""
}

func ParseMCPClientConfigV2(conf *MCPClientConfigV2) (any, error) {
	if conf.Command != "" || conf.TransportType == MCPClientTypeStdio {
		if conf.Command == "" {
			return nil, errors.New("command is required for stdio transport")
		}
		return &StdioMCPClientConfig{
			Command: conf.Command,
			Env:     conf.Env,
			Args:    conf.Args,
		}, nil
	}
	if conf.URL != "" {
		switch conf.TransportType {
		case MCPClientTypeStreamable:
			return &StreamableMCPClientConfig{
				URL:     conf.URL,
				Headers: conf.Headers,
				Timeout: conf.Timeout,
			}, nil
		case MCPClientTypeWebSocket:
			return &WebSocketMCPClientConfig{
				URL:     conf.URL,
				Headers: conf.Headers,
			}, nil
		default:
			// 默认 SSE
			return &SSEMCPClientConfig{
				URL:     conf.URL,
				Headers: conf.Headers,
			}, nil
		}
	}
	return nil, errors.New("invalid server type")
}

// ---- Config ----

type Config struct {
	McpProxy   *MCPProxyConfigV2             `json:"mcpProxy"`
	McpServers map[string]*MCPClientConfigV2 `json:"mcpServers"`
}

type FullConfig struct {
	DeprecatedServerV1  *MCPProxyConfigV1             `json:"server"`
	DeprecatedClientsV1 map[string]*MCPClientConfigV1 `json:"clients"`

	McpProxy   *MCPProxyConfigV2             `json:"mcpProxy"`
	McpServers map[string]*MCPClientConfigV2 `json:"mcpServers"`
}

// PartialConfig is the per-file shape used by LoadDir. Any field is optional;
// the loader merges multiple PartialConfigs (last write wins) before the same
// post-processing applied to a single-file FullConfig.
type PartialConfig struct {
	DeprecatedServerV1  *MCPProxyConfigV1             `json:"server,omitempty"`
	DeprecatedClientsV1 map[string]*MCPClientConfigV1 `json:"clients,omitempty"`

	McpProxy   *MCPProxyConfigV2             `json:"mcpProxy,omitempty"`
	McpServers map[string]*MCPClientConfigV2 `json:"mcpServers,omitempty"`
}

// mergeInto copies non-empty fields from other into p. McpServers and
// DeprecatedClientsV1 are merged by name (other wins). McpProxy and
// DeprecatedServerV1 are replaced if other sets them. A nil other is a no-op.
func (p *PartialConfig) mergeInto(other *PartialConfig) {
	if other == nil {
		return
	}
	if other.McpProxy != nil {
		p.McpProxy = other.McpProxy
	}
	if len(other.McpServers) > 0 {
		if p.McpServers == nil {
			p.McpServers = make(map[string]*MCPClientConfigV2, len(other.McpServers))
		}
		for k, v := range other.McpServers {
			p.McpServers[k] = v
		}
	}
	if other.DeprecatedServerV1 != nil {
		p.DeprecatedServerV1 = other.DeprecatedServerV1
	}
	if len(other.DeprecatedClientsV1) > 0 {
		if p.DeprecatedClientsV1 == nil {
			p.DeprecatedClientsV1 = make(map[string]*MCPClientConfigV1, len(other.DeprecatedClientsV1))
		}
		for k, v := range other.DeprecatedClientsV1 {
			p.DeprecatedClientsV1[k] = v
		}
	}
}

// toFullConfig projects a PartialConfig into the FullConfig used by finalize.
func (p *PartialConfig) toFullConfig() *FullConfig {
	return &FullConfig{
		DeprecatedServerV1:  p.DeprecatedServerV1,
		DeprecatedClientsV1: p.DeprecatedClientsV1,
		McpProxy:            p.McpProxy,
		McpServers:          p.McpServers,
	}
}

func IsRemoteURL(path string) bool {
	return http.IsRemoteURL(path)
}

func newConfProvider(path string, insecure, expandEnv bool, httpHeaders string, httpTimeout int) (provider.Provider, error) {
	if http.IsRemoteURL(path) {
		var opts []http.Option
		httpClient := nethttp.DefaultClient
		if insecure {
			transport := nethttp.DefaultTransport.(*nethttp.Transport).Clone()
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			log.Printf("WARNING: TLS certificate verification disabled (insecure mode)")
			httpClient = &nethttp.Client{Transport: transport}
		}
		if httpTimeout > 0 {
			httpClient.Timeout = time.Duration(httpTimeout) * time.Second
		}
		opts = append(opts, http.WithClient(httpClient))
		if httpHeaders != "" {
			// format: 'Key1:Value1;Key2:Value2'
			headers := make(nethttp.Header)
			for _, kv := range strings.Split(httpHeaders, ";") {
				parts := strings.SplitN(kv, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					if key != "" && value != "" {
						headers.Add(key, value)
					}
				}
			}
			if len(headers) > 0 {
				opts = append(opts, http.WithHeaders(headers))
			}
		}
		pro := http.New(path, opts...)
		if expandEnv {
			return provider.NewExpandEnv(pro), nil
		} else {
			return pro, nil
		}
	}
	if file.IsLocalPath(path) {
		if expandEnv {
			return provider.NewExpandEnv(file.New(path, file.WithExpandEnv())), nil
		} else {
			return file.New(path), nil
		}
	}
	return nil, errors.New("unsupported config path")
}

func Load(path string, insecure, expandEnv bool, httpHeaders string, httpTimeout int) (*Config, error) {
	pro, err := newConfProvider(path, insecure, expandEnv, httpHeaders, httpTimeout)
	if err != nil {
		return nil, err
	}
	conf, err := confstore.Load[FullConfig](pro, codec.JsonCodec())
	if err != nil {
		return nil, err
	}
	adaptMCPClientConfigV1ToV2(conf)
	return conf.finalize()
}

// splitAuthTokens 将 ${VAR} 形式的占位符在逗号分隔的字符串里展开成数组。
func splitAuthTokens(tokens []string) []string {
	var result []string
	for _, token := range tokens {
		for _, t := range strings.Split(token, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				result = append(result, t)
			}
		}
	}
	return result
}

// finalize applies shared post-processing to a FullConfig and projects it
// into the immutable Config consumed by the rest of the application:
//   - V1 -> V2 migration has already happened at this point.
//   - Auth tokens are split on commas.
//   - Trae-specific env/header keys are extracted into Options.* timeouts and
//     ${workspaceFolder} is expanded.
//   - Per-server Options inherit defaults from mcpProxy.Options.
func (c *FullConfig) finalize() (*Config, error) {
	if c.McpProxy == nil {
		return nil, errors.New("mcpProxy is required")
	}
	if c.McpProxy.Options == nil {
		c.McpProxy.Options = &OptionsV2{}
	}
	c.McpProxy.Options.AuthTokens = splitAuthTokens(c.McpProxy.Options.AuthTokens)

	// 解析 ${workspaceFolder} 替换值，作用于所有下游 server。
	workspaceFolder := resolveWorkspaceFolder(c.McpProxy.WorkspaceFolder)

	for _, clientConfig := range c.McpServers {
		if clientConfig == nil {
			continue
		}
		if clientConfig.Options == nil {
			clientConfig.Options = &OptionsV2{}
		}

		// Trae 标准格式 -> mcp-proxy 内部格式。
		// 必须在 AuthTokens 继承等逻辑之前执行, 因为它会修改 Options。
		applyTraeCompat(clientConfig, workspaceFolder)

		clientConfig.Options.AuthTokens = splitAuthTokens(clientConfig.Options.AuthTokens)

		if clientConfig.Options.AuthTokens == nil {
			clientConfig.Options.AuthTokens = c.McpProxy.Options.AuthTokens
		}
		if !clientConfig.Options.PanicIfInvalid.Present() {
			clientConfig.Options.PanicIfInvalid = c.McpProxy.Options.PanicIfInvalid
		}
		if !clientConfig.Options.LogEnabled.Present() {
			clientConfig.Options.LogEnabled = c.McpProxy.Options.LogEnabled
		}
		if !clientConfig.Options.DisablePing.Present() {
			clientConfig.Options.DisablePing = c.McpProxy.Options.DisablePing
		}
	}

	if c.McpProxy.Type == "" {
		c.McpProxy.Type = MCPServerTypeSSE // default to SSE
	}

	return &Config{
		McpProxy:   c.McpProxy,
		McpServers: c.McpServers,
	}, nil
}

// LoadDir walks dir for *.json files (alphabetical, deterministic), merges
// each file as a PartialConfig, and returns a finalized *Config.
//
// Merge order (later wins):
//
//	<dir>/<file>.json                    (e.g. base.json)
//	<dir>/<sub>/<file>.json              (subdirs processed in lexical order)
//
// Within a subdir, alphabetical file order determines precedence. A typical
// layout — `base.json` + `categories/*.json` + `overrides/*.json` — gets
// `base` < `categories` < `overrides` for free, because the lexical ordering
// of the relative paths is `base.json` < `categories/...` < `overrides/...`.
//
// Hidden files (basenames starting with "."), non-JSON files, and
// directories that cannot be entered are skipped. A non-existent dir is an
// error; a directory that simply contains no JSON files returns an error so
// operators notice the misconfiguration.
func LoadDir(dir string, insecure, expandEnv bool, httpHeaders string, httpTimeout int) (*Config, error) {
	if dir == "" {
		return nil, errors.New("config dir is required")
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat config dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("config path %q is not a directory", dir)
	}

	files, err := collectConfigFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no .json config files found under %q", dir)
	}

	merged := &PartialConfig{}
	for _, rel := range files {
		partial, err := loadPartialConfig(dir, rel, insecure, expandEnv, httpHeaders, httpTimeout)
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", rel, err)
		}
		merged.mergeInto(partial)
	}

	full := merged.toFullConfig()
	adaptMCPClientConfigV1ToV2(full)
	if full.McpProxy == nil {
		return nil, fmt.Errorf("no mcpProxy block found in any file under %q (expected at least base.json)", dir)
	}
	return full.finalize()
}

// collectConfigFiles returns every *.json file under root in deterministic,
// lexical order of its path relative to root. Hidden basenames (starting with
// ".") are excluded so dotfiles like .DS_Store are not picked up.
func collectConfigFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// loadPartialConfig reads a single file under root and parses it as a
// PartialConfig. The same provider pipeline is reused so ${VAR} expansion,
// remote URLs and HTTP fetch options work uniformly between Load and LoadDir.
// Each file is resolved to an absolute path before being handed to the
// provider, so the loader behaves the same regardless of the caller's CWD.
func loadPartialConfig(root, relPath string, insecure, expandEnv bool, httpHeaders string, httpTimeout int) (*PartialConfig, error) {
	absPath := relPath
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(root, relPath)
	}
	pro, err := newConfProvider(absPath, insecure, expandEnv, httpHeaders, httpTimeout)
	if err != nil {
		return nil, err
	}
	partial, err := confstore.Load[PartialConfig](pro, codec.JsonCodec())
	if err != nil {
		return nil, err
	}
	return partial, nil
}
