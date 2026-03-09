package core

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ahaufox/mcp-gateway/mcp-proxy/internal/config"
)

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

type Client struct {
	Name            string
	NeedPing        bool
	NeedManualStart bool
	Client          *client.Client
	Options         *config.OptionsV2
	Status          string
	LastError       string

	// Metadata for dashboard
	Description string
	Tools       []mcp.Tool
	Prompts     []mcp.Prompt
	Resources   []mcp.Resource

	// Internal state for reconnection
	clientInfo mcp.Implementation
	mcpServer  *server.MCPServer
}

func NewMCPClient(name string, conf *config.MCPClientConfigV2) (*Client, error) {
	clientInfo, pErr := config.ParseMCPClientConfigV2(conf)
	if pErr != nil {
		return nil, pErr
	}
	switch v := clientInfo.(type) {
	case *config.StdioMCPClientConfig:
		envs := make([]string, 0, len(v.Env))
		for kk, vv := range v.Env {
			envs = append(envs, fmt.Sprintf("%s=%s", kk, vv))
		}
		mcpClient, err := client.NewStdioMCPClient(v.Command, envs, v.Args...)
		if err != nil {
			return nil, err
		}

		return &Client{
			Name:        name,
			Description: conf.Description,
			Client:      mcpClient,
			Options:     conf.Options,
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
			c.Status = "Failed"
			c.LastError = err.Error()
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
		err := c.Client.Start(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transport: %w", err)
		}
	}

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

	err = c.addToolsToServer(ctx, c.mcpServer)
	if err != nil {
		return fmt.Errorf("failed to add tools: %w", err)
	}
	_ = c.addPromptsToServer(ctx, c.mcpServer)
	_ = c.addResourcesToServer(ctx, c.mcpServer)
	_ = c.addResourceTemplatesToServer(ctx, c.mcpServer)

	c.Status = "Connected"
	c.LastError = ""
	log.Printf("<%s> Successfully (re)connected and initialized", c.Name)
	return nil
}

func (c *Client) startMaintenanceTask(ctx context.Context) {
	interval := 30 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	failCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Printf("<%s> Context done, stopping maintenance", c.Name)
			return
		case <-ticker.C:
			if c.Status == "Connected" {
				if err := c.Client.Ping(ctx); err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						return
					}
					failCount++
					c.Status = "Unhealthy"
					c.LastError = fmt.Sprintf("Ping failed: %v", err)
					log.Printf("<%s> Ping failed: %v (count=%d)", c.Name, err, failCount)
				} else {
					if failCount > 0 {
						log.Printf("<%s> Recovered", c.Name)
						failCount = 0
					}
				}
			}

			if c.Status != "Connected" {
				log.Printf("<%s> Attempting to reconnect", c.Name)
				if err := c.connectAndRegister(ctx); err != nil {
					log.Printf("<%s> Reconnection failed: %v", c.Name, err)
					c.LastError = err.Error()
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
		tools, err := c.Client.ListTools(ctx, toolsRequest)
		if err != nil {
			return err
		}
		if len(tools.Tools) == 0 {
			break
		}
		log.Printf("<%s> Successfully listed %d tools", c.Name, len(tools.Tools))
		c.Tools = append(c.Tools, tools.Tools...)
		for _, tool := range tools.Tools {
			if filterFunc(tool.Name) {
				log.Printf("<%s> Adding tool %s", c.Name, tool.Name)
				mcpServer.AddTool(tool, c.Client.CallTool)
			}
		}
		if tools.NextCursor == "" {
			break
		}
		toolsRequest.Params.Cursor = tools.NextCursor
	}

	return nil
}

func (c *Client) addPromptsToServer(ctx context.Context, mcpServer *server.MCPServer) error {
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
		c.Prompts = append(c.Prompts, prompts.Prompts...)
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
		c.Resources = append(c.Resources, resources.Resources...)
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
	if c.Client != nil {
		return c.Client.Close()
	}
	return nil
}
