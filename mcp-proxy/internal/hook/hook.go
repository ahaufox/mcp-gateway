package hook

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// RequestHook 工具调用请求钩子
type RequestHook func(ctx context.Context, req *mcp.CallToolRequest) error

// ResponseHook 工具调用响应钩子
type ResponseHook func(ctx context.Context, req *mcp.CallToolRequest, res *mcp.CallToolResult, err error) (*mcp.CallToolResult, error)

// Chain 钩子链
type Chain struct {
	requestHooks  []RequestHook
	responseHooks []ResponseHook
}

// NewChain 创建新的钩子链
func NewChain() *Chain {
	return &Chain{
		requestHooks:  make([]RequestHook, 0),
		responseHooks: make([]ResponseHook, 0),
	}
}

// AddRequestHook 添加请求钩子
func (c *Chain) AddRequestHook(hook RequestHook) *Chain {
	c.requestHooks = append(c.requestHooks, hook)
	return c
}

// AddResponseHook 添加响应钩子
func (c *Chain) AddResponseHook(hook ResponseHook) *Chain {
	c.responseHooks = append(c.responseHooks, hook)
	return c
}

// BeforeCall 执行请求钩子
func (c *Chain) BeforeCall(ctx context.Context, req *mcp.CallToolRequest) error {
	for _, hook := range c.requestHooks {
		if err := hook(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// AfterCall 执行响应钩子
func (c *Chain) AfterCall(ctx context.Context, req *mcp.CallToolRequest, res *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
	var finalRes *mcp.CallToolResult = res
	var finalErr error = err

	for _, hook := range c.responseHooks {
		finalRes, finalErr = hook(ctx, req, finalRes, finalErr)
	}

	return finalRes, finalErr
}

// Clear 清空所有钩子
func (c *Chain) Clear() {
	c.requestHooks = c.requestHooks[:0]
	c.responseHooks = c.responseHooks[:0]
}

// RequestHooksCount 返回请求钩子数量
func (c *Chain) RequestHooksCount() int {
	return len(c.requestHooks)
}

// ResponseHooksCount 返回响应钩子数量
func (c *Chain) ResponseHooksCount() int {
	return len(c.responseHooks)
}
