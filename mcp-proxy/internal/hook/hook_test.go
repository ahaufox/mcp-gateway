package hook

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewChain(t *testing.T) {
	chain := NewChain()
	assert.NotNil(t, chain)
	assert.Equal(t, 0, chain.RequestHooksCount())
	assert.Equal(t, 0, chain.ResponseHooksCount())
}

func TestAddRequestHook(t *testing.T) {
	chain := NewChain()

	hook1 := func(ctx context.Context, req *mcp.CallToolRequest) error { return nil }
	hook2 := func(ctx context.Context, req *mcp.CallToolRequest) error { return nil }

	chain.AddRequestHook(hook1).AddRequestHook(hook2)

	assert.Equal(t, 2, chain.RequestHooksCount())
}

func TestAddResponseHook(t *testing.T) {
	chain := NewChain()

	hook1 := func(ctx context.Context, req *mcp.CallToolRequest, res *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
		return res, err
	}
	hook2 := func(ctx context.Context, req *mcp.CallToolRequest, res *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
		return res, err
	}

	chain.AddResponseHook(hook1).AddResponseHook(hook2)

	assert.Equal(t, 2, chain.ResponseHooksCount())
}

func TestBeforeCall(t *testing.T) {
	chain := NewChain()

	calledCount := 0
	hook := func(ctx context.Context, req *mcp.CallToolRequest) error {
		calledCount++
		return nil
	}

	chain.AddRequestHook(hook)

	req := &mcp.CallToolRequest{}
	ctx := context.Background()

	err := chain.BeforeCall(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 1, calledCount)
}

func TestBeforeCallWithError(t *testing.T) {
	chain := NewChain()

	hook1 := func(ctx context.Context, req *mcp.CallToolRequest) error { return nil }
	hook2 := func(ctx context.Context, req *mcp.CallToolRequest) error { return errors.New("hook error") }
	hook3 := func(ctx context.Context, req *mcp.CallToolRequest) error { return nil }

	chain.AddRequestHook(hook1).AddRequestHook(hook2).AddRequestHook(hook3)

	req := &mcp.CallToolRequest{}
	ctx := context.Background()

	err := chain.BeforeCall(ctx, req)
	require.Error(t, err)
	assert.Equal(t, "hook error", err.Error())
}

func TestAfterCall(t *testing.T) {
	chain := NewChain()

	originalRes := &mcp.CallToolResult{}
	modifiedRes := &mcp.CallToolResult{}

	hook := func(ctx context.Context, req *mcp.CallToolRequest, res *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
		return modifiedRes, nil
	}

	chain.AddResponseHook(hook)

	req := &mcp.CallToolRequest{}
	ctx := context.Background()

	result, err := chain.AfterCall(ctx, req, originalRes, nil)
	require.NoError(t, err)
	assert.Equal(t, modifiedRes, result)
}

func TestAfterCallChain(t *testing.T) {
	chain := NewChain()

	hook1 := func(ctx context.Context, req *mcp.CallToolRequest, res *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
		return res, errors.New("first error")
	}
	hook2 := func(ctx context.Context, req *mcp.CallToolRequest, res *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
		if err != nil {
			return res, errors.New("wrapped: " + err.Error())
		}
		return res, err
	}

	chain.AddResponseHook(hook1).AddResponseHook(hook2)

	req := &mcp.CallToolRequest{}
	ctx := context.Background()
	originalRes := &mcp.CallToolResult{}

	result, err := chain.AfterCall(ctx, req, originalRes, nil)
	require.Error(t, err)
	assert.Equal(t, "wrapped: first error", err.Error())
	assert.Equal(t, originalRes, result)
}

func TestClear(t *testing.T) {
	chain := NewChain()

	chain.AddRequestHook(func(ctx context.Context, req *mcp.CallToolRequest) error { return nil })
	chain.AddResponseHook(func(ctx context.Context, req *mcp.CallToolRequest, res *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
		return res, err
	})

	assert.Equal(t, 1, chain.RequestHooksCount())
	assert.Equal(t, 1, chain.ResponseHooksCount())

	chain.Clear()

	assert.Equal(t, 0, chain.RequestHooksCount())
	assert.Equal(t, 0, chain.ResponseHooksCount())
}

func TestRequestModification(t *testing.T) {
	chain := NewChain()

	var called bool

	hook := func(ctx context.Context, req *mcp.CallToolRequest) error {
		called = true
		return nil
	}

	chain.AddRequestHook(hook)

	req := &mcp.CallToolRequest{}
	ctx := context.Background()

	err := chain.BeforeCall(ctx, req)
	require.NoError(t, err)
	assert.True(t, called)
}
