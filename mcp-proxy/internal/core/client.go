package core

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ahaufox/mcp-gateway/mcp-proxy/internal/circuitbreaker"
	"github.com/ahaufox/mcp-gateway/mcp-proxy/internal/config"
	mcperrors "github.com/ahaufox/mcp-gateway/mcp-proxy/internal/errors"
)

var DefaultPingInterval = 30 * time.Second

var blockedEnvKeys = map[string]struct{}{
	"PATH":                  {},
	"LD_PRELOAD":            {},
	"SHELL":                 {},
	"HOME":                  {},
	"USER":                  {},
	"LD_LIBRARY_PATH":       {},
	"DYLD_INSERT_LIBRARIES": {},
	"DYLD_LIBRARY_PATH":     {},
	"BASH_ENV":              {},
	"IFS":                   {},
}

func sanitizeEnvKey(key string) bool {
	_, blocked := blockedEnvKeys[key]
	return !blocked
}

type gzipDecompressor struct {
	transport http.RoundTripper
}

func (d *gzipDecompressor) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := d.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body = &readCloserWrapper{Reader: gz, Closer: resp.Body}
		resp.Header.Del("Content-Encoding")
		resp.Header.Del("Content-Length")
	}
	return resp, nil
}

type readCloserWrapper struct {
	io.Reader
	io.Closer
}

// WebSocketClient WebSocket MCP 客户端包装器
type WebSocketClient struct {
	mu       sync.RWMutex
	conn     *websocket.Conn
	url      string
	headers  map[string]string
	closed   bool
	readCh   chan *mcp.JSONRPCMessage
	writeCh  chan *mcp.JSONRPCMessage
	errorCh  chan error
}

func NewWebSocketClient(url string, headers map[string]string) (*WebSocketClient, error) {
	return &WebSocketClient{
		url:     url,
		headers: headers,
		readCh:  make(chan *mcp.JSONRPCMessage, 100),
		writeCh: make(chan *mcp.JSONRPCMessage, 100),
		errorCh: make(chan error, 10),
	}, nil
}

func (w *WebSocketClient) Connect(ctx context.Context) error {
	header := http.Header{}
	for k, v := range w.headers {
		header.Set(k, v)
	}

	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(ctx, w.url, header)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.conn = conn
	w.closed = false
	w.mu.Unlock()

	go w.readLoop()
	go w.writeLoop()

	return nil
}

func (w *WebSocketClient) readLoop() {
	for {
		_, msg, err := w.conn.ReadMessage()
		if err != nil {
			if !w.isClosed() {
				w.errorCh <- err
			}
			return
		}

		var rpcMsg mcp.JSONRPCMessage
		if err := json.Unmarshal(msg, &rpcMsg); err != nil {
			log.Printf("[websocket] failed to unmarshal message: %v", err)
			continue
		}

		select {
		case w.readCh <- &rpcMsg:
		default:
		}
	}
}

func (w *WebSocketClient) writeLoop() {
	for msg := range w.writeCh {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("[websocket] failed to marshal message: %v", err)
			continue
		}

		if err := w.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			w.errorCh <- err
		}
	}
}

func (w *WebSocketClient) isClosed() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.closed
}

func (w *WebSocketClient) Send(ctx context.Context, msg *mcp.JSONRPCMessage) error {
	select {
	case w.writeCh <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *WebSocketClient) Receive(ctx context.Context) (*mcp.JSONRPCMessage, error) {
	select {
	case msg := <-w.readCh:
		return msg, nil
	case err := <-w.errorCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (w *WebSocketClient) Ping(ctx context.Context) error {
	w.mu.RLock()
	conn := w.conn
	w.mu.RUnlock()

	if conn == nil {
		return errors.New("not connected")
	}

	pingCh := make(chan error, 1)
	go func() {
		pingCh <- conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
	}()

	select {
	case err := <-pingCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return errors.New("ping timeout")
	}
}

func (w *WebSocketClient) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	close(w.writeCh)

	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}

// Client MCP 客户端
type Client struct {
	mu              sync.RWMutex
	Name            string
	NeedPing        bool
	NeedManualStart bool
	Client          *client.Client
	Options         *config.OptionsV2
	Status          string
	LastError       string

	// WebSocket 客户端（如果使用 WebSocket）
	wsClient *WebSocketClient

	// Metadata for dashboard
	Description string
	Tools       []mcp.Tool
	Prompts     []mcp.Prompt
	Resources   []mcp.Resource

	// Internal state for reconnection
	clientInfo mcp.Implementation
	mcpServer  *server.MCPServer
	
	// 熔断器
	circuitBreaker *circuitbreaker.CircuitBreaker
}

func (c *Client) GetStatus() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Status
}

func (c *Client) SetStatus(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Status = s
}

func (c *Client) GetLastError() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.LastError
}

func (c *Client) SetLastError(e string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastError = e
}

func (c *Client) GetTools() []mcp.Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Tools
}

func (c *Client) SetTools(t []mcp.Tool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Tools = t
}

func (c *Client) GetPrompts() []mcp.Prompt {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Prompts
}

func (c *Client) SetPrompts(p []mcp.Prompt) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Prompts = p
}

func (c *Client) GetResources() []mcp.Resource {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Resources
}

func (c *Client) SetResources(r []mcp.Resource) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Resources = r
}

func NewMCPClient(name string, conf *config.MCPClientConfigV2) (*Client, error) {
	clientInfo, pErr := config.ParseMCPClientConfigV2(conf)
	if pErr != nil {
		return nil, pErr
	}
	
	// 创建熔断器
	var cb *circuitbreaker.CircuitBreaker
	if conf.Options != nil && conf.Options.CircuitBreaker != nil && conf.Options.CircuitBreaker.Enabled {
		cfg := circuitbreaker.DefaultConfig()
		if conf.Options.CircuitBreaker.MaxFailures > 0 {
			cfg.MaxFailures = conf.Options.CircuitBreaker.MaxFailures
		}
		if conf.Options.CircuitBreaker.ResetTimeout > 0 {
			cfg.ResetTimeout = conf.Options.CircuitBreaker.ResetTimeout
		}
		if conf.Options.CircuitBreaker.HalfOpenMax > 0 {
			cfg.HalfOpenMax = conf.Options.CircuitBreaker.HalfOpenMax
		}
		cb = circuitbreaker.New(name, cfg)
	}
	
	switch v := clientInfo.(type) {
	case *config.StdioMCPClientConfig:
		envs := make([]string, 0, len(v.Env))
		for kk, vv := range v.Env {
			if !sanitizeEnvKey(kk) {
				log.Printf("<%s> Blocked dangerous environment variable: %s", name, kk)
				continue
			}
			envs = append(envs, fmt.Sprintf("%s=%s", kk, vv))
		}
		mcpClient, err := client.NewStdioMCPClient(v.Command, envs, v.Args...)
		if err != nil {
			return nil, err
		}

		return &Client{
			Name:           name,
			Description:    conf.Description,
			Client:         mcpClient,
			Options:        conf.Options,
			circuitBreaker: cb,
		}, nil
	case *config.SSEMCPClientConfig:
		var options []transport.ClientOption
		if len(v.Headers) > 0 {
			options = append(options, client.WithHeaders(v.Headers))
		}
		mcpClient, err := client.NewSSEMCPClient(v.URL, options...)
		if err != nil {
			return nil, err
		}
		return &Client{
			Name:            name,
			NeedPing:        !conf.Options.DisablePing.OrElse(false),
			NeedManualStart: true,
			Client:          mcpClient,
			Options:         conf.Options,
			circuitBreaker:  cb,
		}, nil
	case *config.StreamableMCPClientConfig:
		var options []transport.StreamableHTTPCOption
		if len(v.Headers) > 0 {
			options = append(options, transport.WithHTTPHeaders(v.Headers))
		}
		httpClient := &http.Client{
			Transport: &gzipDecompressor{transport: http.DefaultTransport},
		}
		if v.Timeout > 0 {
			httpClient.Timeout = v.Timeout
		}
		options = append(options, transport.WithHTTPBasicClient(httpClient))
		mcpClient, err := client.NewStreamableHttpClient(v.URL, options...)
		if err != nil {
			return nil, err
		}
		return &Client{
			Name:            name,
			NeedPing:        !conf.Options.DisablePing.OrElse(false),
			NeedManualStart: true,
			Client:          mcpClient,
			Options:         conf.Options,
			circuitBreaker:  cb,
		}, nil
	case *config.WebSocketMCPClientConfig:
		wsClient, err := NewWebSocketClient(v.URL, v.Headers)
		if err != nil {
			return nil, err
		}
		return &Client{
			Name:            name,
			NeedPing:        !conf.Options.DisablePing.OrElse(false),
			NeedManualStart: true,
			wsClient:        wsClient,
			Options:         conf.Options,
			circuitBreaker:  cb,
		}, nil
	}
	return nil, errors.New("invalid client type")
}

func (c *Client) AddToMCPServer(ctx context.Context, clientInfo mcp.Implementation, mcpServer *server.MCPServer) error {
	c.clientInfo = clientInfo
	c.mcpServer = mcpServer

	err := c.connectAndRegister(ctx)
	if err != nil {
		log.Printf("<%s> Initial connection failed: %v", c.Name, err)
		if c.Options != nil && c.Options.PanicIfInvalid.OrElse(false) {
			return err
		}
		// For network clients, we always start a background task even if first connection fails
		if c.NeedManualStart {
			c.SetStatus("Failed")
			c.SetLastError(err.Error())
			go c.startMaintenanceTask(ctx)
			return nil
		}
		return err
	}

	if c.NeedPing {
		go c.startMaintenanceTask(ctx)
	}
	return nil
}

func (c *Client) connectAndRegister(ctx context.Context) error {
	if c.NeedManualStart {
		log.Printf("<%s> Starting client transport", c.Name)
		
		if c.wsClient != nil {
			// WebSocket 连接
			err := c.wsClient.Connect(ctx)
			if err != nil {
				return fmt.Errorf("failed to connect websocket: %w", err)
			}
		} else if c.Client != nil {
			err := c.Client.Start(ctx)
			if err != nil {
				return fmt.Errorf("failed to start transport: %w", err)
			}
		}
	}

	// 初始化
	if c.wsClient != nil {
		// WebSocket MCP 初始化
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = c.clientInfo
		initRequest.Params.Capabilities = mcp.ClientCapabilities{
			Experimental: make(map[string]interface{}),
		}
		
		err := c.wsClient.Send(ctx, initRequest.JSONRPCRequest())
		if err != nil {
			return fmt.Errorf("failed to send initialize: %w", err)
		}
		
		// 等待响应
		_, err = c.wsClient.Receive(ctx)
		if err != nil {
			return fmt.Errorf("failed to receive initialize response: %w", err)
		}
	} else {
		initRequest := mcp.InitializeRequest{}
		initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
		initRequest.Params.ClientInfo = c.clientInfo
		initRequest.Params.Capabilities = mcp.ClientCapabilities{
			Experimental: make(map[string]interface{}),
		}

		_, err := c.Client.Initialize(ctx, initRequest)
		if err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}
	}

	err := c.addToolsToServer(ctx, c.mcpServer)
	if err != nil {
		return fmt.Errorf("failed to add tools: %w", err)
	}
	if err := c.addPromptsToServer(ctx, c.mcpServer); err != nil {
		log.Printf("<%s> Failed to load prompts: %v", c.Name, err)
	}
	if err := c.addResourcesToServer(ctx, c.mcpServer); err != nil {
		log.Printf("<%s> Failed to load resources: %v", c.Name, err)
	}
	if err := c.addResourceTemplatesToServer(ctx, c.mcpServer); err != nil {
		log.Printf("<%s> Failed to load resource templates: %v", c.Name, err)
	}

	c.SetStatus("Connected")
	c.SetLastError("")
	log.Printf("<%s> Successfully (re)connected and initialized", c.Name)
	return nil
}

func (c *Client) startMaintenanceTask(ctx context.Context) {
	interval := DefaultPingInterval
	if c.Options != nil && c.Options.MaintenanceInterval > 0 {
		interval = c.Options.MaintenanceInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	failCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Printf("<%s> Context done, stopping maintenance", c.Name)
			return
		case <-ticker.C:
			if c.GetStatus() == "Connected" {
				var err error
				if c.wsClient != nil {
					err = c.wsClient.Ping(ctx)
				} else if c.Client != nil {
					err = c.Client.Ping(ctx)
				}
				
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						return
					}
					failCount++
					c.SetStatus("Unhealthy")
					c.SetLastError(fmt.Sprintf("Ping failed: %v", err))
					log.Printf("<%s> Ping failed: %v (count=%d)", c.Name, err, failCount)
				} else {
					if failCount > 0 {
						log.Printf("<%s> Recovered", c.Name)
						failCount = 0
					}
				}
			}

			if c.GetStatus() != "Connected" {
				log.Printf("<%s> Attempting to reconnect", c.Name)
				if err := c.connectAndRegister(ctx); err != nil {
					log.Printf("<%s> Reconnection failed: %v", c.Name, err)
					c.SetLastError(err.Error())
				}
			}
		}
	}
}

func (c *Client) addToolsToServer(ctx context.Context, mcpServer *server.MCPServer) error {
	toolsRequest := mcp.ListToolsRequest{}
	filterFunc := func(toolName string) bool {
		return true
	}

	if c.Options != nil && c.Options.ToolFilter != nil && len(c.Options.ToolFilter.List) > 0 {
		filterSet := make(map[string]struct{})
		mode := config.ToolFilterMode(strings.ToLower(string(c.Options.ToolFilter.Mode)))
		for _, toolName := range c.Options.ToolFilter.List {
			filterSet[toolName] = struct{}{}
		}
		switch mode {
		case config.ToolFilterModeAllow:
			filterFunc = func(toolName string) bool {
				_, inList := filterSet[toolName]
				if !inList {
					log.Printf("<%s> Ignoring tool %s as it is not in allow list", c.Name, toolName)
				}
				return inList
			}
		case config.ToolFilterModeBlock:
			filterFunc = func(toolName string) bool {
				_, inList := filterSet[toolName]
				if inList {
					log.Printf("<%s> Ignoring tool %s as it is in block list", c.Name, toolName)
				}
				return !inList
			}
		default:
			log.Printf("<%s> Unknown tool filter mode: %s, skipping tool filter", c.Name, mode)
		}
	}

	for {
		var tools *mcp.ListToolsResult
		var err error
		
		if c.wsClient != nil {
			// WebSocket 客户端
			listToolsReq := mcp.ListToolsRequest{}
			err = c.wsClient.Send(ctx, listToolsReq.JSONRPCRequest())
			if err != nil {
				return err
			}
			
			resp, recvErr := c.wsClient.Receive(ctx)
			if recvErr != nil {
				return recvErr
			}
			
			tools = &mcp.ListToolsResult{}
			if resp.Result != nil {
				if data, marshalErr := json.Marshal(resp.Result); marshalErr == nil {
					json.Unmarshal(data, tools)
				}
			}
		} else {
			tools, err = c.Client.ListTools(ctx, toolsRequest)
		}
		
		if err != nil {
			return err
		}
		if len(tools.Tools) == 0 {
			break
		}
		log.Printf("<%s> Successfully listed %d tools", c.Name, len(tools.Tools))
		c.SetTools(append(c.GetTools(), tools.Tools...))
		for _, tool := range tools.Tools {
			if filterFunc(tool.Name) {
				log.Printf("<%s> Adding tool %s", c.Name, tool.Name)
				wrappedCall := c.createWrappedCallTool(tool.Name)
				mcpServer.AddTool(tool, wrappedCall)
			}
		}
		if tools.NextCursor == "" {
			break
		}
		toolsRequest.Params.Cursor = tools.NextCursor
	}

	return nil
}

func (c *Client) createWrappedCallTool(toolName string) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 如果配置了熔断器，先检查
		if c.circuitBreaker != nil {
			if err := c.circuitBreaker.Allow(); err != nil {
				return nil, mcperrors.Wrap(err, mcperrors.ErrCodeCircuitOpen, "circuit breaker open").WithService(c.Name).WithTool(toolName)
			}
		}
		
		// 应用超时
		callCtx := ctx
		if c.Options != nil && c.Options.CallTimeout > 0 {
			var cancel context.CancelFunc
			callCtx, cancel = context.WithTimeout(ctx, c.Options.CallTimeout)
			defer cancel()
		}
		
		var result *mcp.CallToolResult
		var err error
		
		if c.wsClient != nil {
			// WebSocket 调用
			err = c.wsClient.Send(callCtx, req.JSONRPCRequest())
			if err != nil {
				result = nil
			} else {
				resp, recvErr := c.wsClient.Receive(callCtx)
				if recvErr != nil {
					err = recvErr
				} else if resp.Error != nil {
					err = errors.New(resp.Error.Message)
				} else if resp.Result != nil {
					result = &mcp.CallToolResult{}
					if data, marshalErr := json.Marshal(resp.Result); marshalErr == nil {
						json.Unmarshal(data, result)
					}
				}
			}
		} else {
			result, err = c.Client.CallTool(callCtx, req)
		}
		
		// 记录熔断器结果
		if c.circuitBreaker != nil {
			c.circuitBreaker.RecordResult(err == nil)
		}
		
		// 包装错误
		if err != nil {
			wrappedErr := mcperrors.Wrap(err, mcperrors.ErrCodeServer, "tool call failed").WithService(c.Name).WithTool(toolName)
			return nil, wrappedErr
		}
		
		return result, nil
	}
}

func (c *Client) addPromptsToServer(ctx context.Context, mcpServer *server.MCPServer) error {
	if c.wsClient != nil {
		// WebSocket 暂不支持 prompts
		return nil
	}
	
	promptsRequest := mcp.ListPromptsRequest{}
	for {
		prompts, err := c.Client.ListPrompts(ctx, promptsRequest)
		if err != nil {
			return err
		}
		if len(prompts.Prompts) == 0 {
			break
		}
		log.Printf("<%s> Successfully listed %d prompts", c.Name, len(prompts.Prompts))
		c.SetPrompts(append(c.GetPrompts(), prompts.Prompts...))
		for _, prompt := range prompts.Prompts {
			log.Printf("<%s> Adding prompt %s", c.Name, prompt.Name)
			mcpServer.AddPrompt(prompt, c.Client.GetPrompt)
		}
		if prompts.NextCursor == "" {
			break
		}
		promptsRequest.Params.Cursor = prompts.NextCursor
	}
	return nil
}

func (c *Client) addResourcesToServer(ctx context.Context, mcpServer *server.MCPServer) error {
	if c.wsClient != nil {
		// WebSocket 暂不支持 resources
		return nil
	}
	
	resourcesRequest := mcp.ListResourcesRequest{}
	for {
		resources, err := c.Client.ListResources(ctx, resourcesRequest)
		if err != nil {
			return err
		}
		if len(resources.Resources) == 0 {
			break
		}
		log.Printf("<%s> Successfully listed %d resources", c.Name, len(resources.Resources))
		c.SetResources(append(c.GetResources(), resources.Resources...))
		for _, resource := range resources.Resources {
			log.Printf("<%s> Adding resource %s", c.Name, resource.Name)
			mcpServer.AddResource(resource, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				readResource, e := c.Client.ReadResource(ctx, request)
				if e != nil {
					return nil, e
				}
				return readResource.Contents, nil
			})
		}
		if resources.NextCursor == "" {
			break
		}
		resourcesRequest.Params.Cursor = resources.NextCursor
	}

	return nil
}

func (c *Client) addResourceTemplatesToServer(ctx context.Context, mcpServer *server.MCPServer) error {
	if c.wsClient != nil {
		// WebSocket 暂不支持 resource templates
		return nil
	}
	
	resourceTemplatesRequest := mcp.ListResourceTemplatesRequest{}
	for {
		resourceTemplates, err := c.Client.ListResourceTemplates(ctx, resourceTemplatesRequest)
		if err != nil {
			return err
		}
		if len(resourceTemplates.ResourceTemplates) == 0 {
			break
		}
		log.Printf("<%s> Successfully listed %d resource templates", c.Name, len(resourceTemplates.ResourceTemplates))
		for _, resourceTemplate := range resourceTemplates.ResourceTemplates {
			log.Printf("<%s> Adding resource template %s", c.Name, resourceTemplate.Name)
			mcpServer.AddResourceTemplate(resourceTemplate, func(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
				readResource, e := c.Client.ReadResource(ctx, request)
				if e != nil {
					return nil, e
				}
				return readResource.Contents, nil
			})
		}
		if resourceTemplates.NextCursor == "" {
			break
		}
		resourceTemplatesRequest.Params.Cursor = resourceTemplates.NextCursor
	}
	return nil
}

func (c *Client) Close() error {
	if c.wsClient != nil {
		return c.wsClient.Close()
	}
	if c.Client != nil {
		return c.Client.Close()
	}
	return nil
}
