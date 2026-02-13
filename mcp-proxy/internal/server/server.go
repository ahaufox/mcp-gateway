package server

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tbxark/mcp-proxy/internal/config"
	"github.com/tbxark/mcp-proxy/internal/core"
	"golang.org/x/sync/errgroup"
)

//go:embed templates/*
var res embed.FS

var (
	templates = template.Must(template.ParseFS(res, "templates/*.html"))
)

type ServerInfo struct {
	Name   string
	Route  string
	Status string
	Error  string
}

type MiddlewareFunc func(http.Handler) http.Handler

func chainMiddleware(h http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for _, mw := range middlewares {
		h = mw(h)
	}
	return h
}

func newAuthMiddleware(tokens []string) MiddlewareFunc {
	tokenSet := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		tokenSet[token] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(tokens) != 0 {
				token := r.Header.Get("Authorization")
				token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer "))
				if token == "" {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				if _, ok := tokenSet[token]; !ok {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func loggerMiddleware(prefix string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("<%s> Request [%s] %s", prefix, r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

func recoverMiddleware(prefix string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("<%s> Recovered from panic: %v", prefix, err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func StartHTTPServer(cfg *config.Config) error {
	baseURL, uErr := url.Parse(cfg.McpProxy.BaseURL)
	if uErr != nil {
		return uErr
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var errorGroup errgroup.Group
	httpMux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:    cfg.McpProxy.Addr,
		Handler: httpMux,
	}
	info := mcp.Implementation{
		Name: cfg.McpProxy.Name,
	}

	mcpClients := make(map[string]*core.Client)
	for name, clientConfig := range cfg.McpServers {
		if clientConfig.Options.Disabled {
			log.Printf("<%s> Disabled", name)
			continue
		}
		mcpClient, err := core.NewMCPClient(name, clientConfig)
		if err != nil {
			return err
		}
		mcpClients[name] = mcpClient
		srv, err := NewMCPServer(name, cfg.McpProxy, clientConfig)
		if err != nil {
			return err
		}
		errorGroup.Go(func() error {
			log.Printf("<%s> Connecting", name)
			addErr := mcpClient.AddToMCPServer(ctx, info, srv.McpServer)
			if addErr != nil {
				log.Printf("<%s> Failed to add client to server: %v", name, addErr)
				if clientConfig.Options.PanicIfInvalid.OrElse(false) {
					return addErr
				}
				return nil
			}
			log.Printf("<%s> Connected", name)

			middlewares := make([]MiddlewareFunc, 0)
			middlewares = append(middlewares, recoverMiddleware(name))
			if clientConfig.Options.LogEnabled.OrElse(false) {
				middlewares = append(middlewares, loggerMiddleware(name))
			}
			if len(clientConfig.Options.AuthTokens) > 0 {
				middlewares = append(middlewares, newAuthMiddleware(clientConfig.Options.AuthTokens))
			}
			mcpRoute := path.Join(baseURL.Path, name)
			if !strings.HasPrefix(mcpRoute, "/") {
				mcpRoute = "/" + mcpRoute
			}
			if !strings.HasSuffix(mcpRoute, "/") {
				mcpRoute += "/"
			}
			log.Printf("<%s> Handling requests at %s", name, mcpRoute)
			httpMux.Handle(mcpRoute, chainMiddleware(srv.Handler, middlewares...))
			httpServer.RegisterOnShutdown(func() {
				log.Printf("<%s> Shutting down", name)
				_ = mcpClient.Close()
			})
			return nil
		})
	}

	go func() {
		err := errorGroup.Wait()
		if err != nil {
			log.Fatalf("Failed to add clients: %v", err)
		}
		log.Printf("All clients initialized")
	}()

	httpMux.HandleFunc("/docs/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.ExecuteTemplate(w, "converter.html", map[string]interface{}{
			"ActivePage": "converter",
		}); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	httpMux.HandleFunc("/changelog/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := templates.ExecuteTemplate(w, "changelog.html", map[string]interface{}{
			"ActivePage": "changelog",
		}); err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: Hit root handler: %s", r.URL.Path)
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		activeServers := getActiveServers(cfg, baseURL, mcpClients)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := templates.ExecuteTemplate(w, "dashboard.html", map[string]interface{}{
			"Servers":    activeServers,
			"ActivePage": "dashboard",
		})
		if err != nil {
			log.Printf("Template execution error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	httpMux.HandleFunc("/api/servers", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: Hit /api/servers handler")
		activeServers := getActiveServers(cfg, baseURL, mcpClients)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(activeServers); err != nil {
			log.Printf("JSON encoding error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	go func() {
		log.Printf("Starting %s server", cfg.McpProxy.Type)
		log.Printf("%s server listening on %s", cfg.McpProxy.Type, cfg.McpProxy.Addr)
		hErr := httpServer.ListenAndServe()
		if hErr != nil && !errors.Is(hErr, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", hErr)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	err := httpServer.Shutdown(shutdownCtx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func getActiveServers(cfg *config.Config, baseURL *url.URL, mcpClients map[string]*core.Client) []ServerInfo {
	var activeServers []ServerInfo
	for name := range cfg.McpServers {
		if cfg.McpServers[name].Options.Disabled {
			continue
		}
		mcpRoute := path.Join(baseURL.Path, name)
		if !strings.HasPrefix(mcpRoute, "/") {
			mcpRoute = "/" + mcpRoute
		}
		status := "Unknown"
		lastError := ""
		if client, ok := mcpClients[name]; ok {
			status = client.Status
			lastError = client.LastError
		}
		activeServers = append(activeServers, ServerInfo{
			Name:   name,
			Route:  mcpRoute,
			Status: status,
			Error:  lastError,
		})
	}

	sort.Slice(activeServers, func(i, j int) bool {
		return activeServers[i].Name < activeServers[j].Name
	})
	return activeServers
}
