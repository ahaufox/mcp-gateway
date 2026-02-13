package server

import (
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/server"
	"github.com/tbxark/mcp-proxy/internal/config"
)

type Server struct {
	Tokens    []string
	McpServer *server.MCPServer
	Handler   http.Handler
}

func NewMCPServer(name string, serverConfig *config.MCPProxyConfigV2, clientConfig *config.MCPClientConfigV2) (*Server, error) {
	serverOpts := []server.ServerOption{
		server.WithResourceCapabilities(true, true),
		server.WithRecovery(),
	}

	if clientConfig.Options.LogEnabled.OrElse(false) {
		serverOpts = append(serverOpts, server.WithLogging())
	}
	mcpServer := server.NewMCPServer(
		name,
		serverConfig.Version,
		serverOpts...,
	)

	var handler http.Handler

	switch serverConfig.Type {
	case config.MCPServerTypeSSE:
		handler = server.NewSSEServer(
			mcpServer,
			server.WithStaticBasePath(name),
			server.WithBaseURL(serverConfig.BaseURL),
		)
	case config.MCPServerTypeStreamable:
		handler = server.NewStreamableHTTPServer(
			mcpServer,
			server.WithStateLess(true),
		)
	default:
		return nil, fmt.Errorf("unknown server type: %s", serverConfig.Type)
	}
	srv := &Server{
		McpServer: mcpServer,
		Handler:   handler,
	}

	if clientConfig.Options != nil && len(clientConfig.Options.AuthTokens) > 0 {
		srv.Tokens = clientConfig.Options.AuthTokens
	}

	return srv, nil
}
