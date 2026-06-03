package batch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockToolCaller 模拟工具调用器
type mockToolCaller struct {
	delay     time.Duration
	shouldErr bool
}

func (m *mockToolCaller) CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if m.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(m.delay):
		}
	}

	if m.shouldErr {
		return nil, errors.New("tool call failed")
	}

	return &mcp.CallToolResult{}, nil
}

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     BatchCallRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: BatchCallRequest{
				Calls: []ToolCallRequest{
					{ID: "1", ToolName: "tool1"},
					{ID: "2", ToolName: "tool2"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty calls",
			req: BatchCallRequest{
				Calls: []ToolCallRequest{},
			},
			wantErr: true,
		},
		{
			name: "missing call ID",
			req: BatchCallRequest{
				Calls: []ToolCallRequest{
					{ID: "", ToolName: "tool1"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing tool name",
			req: BatchCallRequest{
				Calls: []ToolCallRequest{
					{ID: "1", ToolName: ""},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate IDs",
			req: BatchCallRequest{
				Calls: []ToolCallRequest{
					{ID: "1", ToolName: "tool1"},
					{ID: "1", ToolName: "tool2"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteSequential(t *testing.T) {
	caller := &mockToolCaller{}
	executor := NewExecutor(caller)

	req := BatchCallRequest{
		Calls: []ToolCallRequest{
			{ID: "1", ToolName: "tool1"},
			{ID: "2", ToolName: "tool2"},
		},
		Parallel: false,
	}

	ctx := context.Background()
	result, err := executor.Execute(ctx, req)

	require.NoError(t, err)
	require.Len(t, result.Results, 2)
	assert.Equal(t, "1", result.Results[0].ID)
	assert.True(t, result.Results[0].Success)
	assert.Equal(t, "2", result.Results[1].ID)
	assert.True(t, result.Results[1].Success)
}

func TestExecuteParallel(t *testing.T) {
	caller := &mockToolCaller{delay: 10 * time.Millisecond}
	executor := NewExecutor(caller)

	req := BatchCallRequest{
		Calls: []ToolCallRequest{
			{ID: "1", ToolName: "tool1"},
			{ID: "2", ToolName: "tool2"},
			{ID: "3", ToolName: "tool3"},
		},
		Parallel: true,
	}

	start := time.Now()
	ctx := context.Background()
	result, err := executor.Execute(ctx, req)
	duration := time.Since(start)

	require.NoError(t, err)
	require.Len(t, result.Results, 3)
	assert.True(t, duration < 50*time.Millisecond, "parallel execution should be faster than sequential")

	for _, res := range result.Results {
		assert.True(t, res.Success)
	}
}

func TestExecuteWithErrors(t *testing.T) {
	caller := &mockToolCaller{shouldErr: true}
	executor := NewExecutor(caller)

	req := BatchCallRequest{
		Calls: []ToolCallRequest{
			{ID: "1", ToolName: "tool1"},
			{ID: "2", ToolName: "tool2"},
		},
		Parallel: false,
	}

	ctx := context.Background()
	result, err := executor.Execute(ctx, req)

	require.NoError(t, err)
	require.Len(t, result.Results, 2)
	assert.False(t, result.Results[0].Success)
	assert.NotEmpty(t, result.Results[0].Error)
	assert.False(t, result.Results[1].Success)
	assert.NotEmpty(t, result.Results[1].Error)
}

func TestExecuteWithContextCancel(t *testing.T) {
	caller := &mockToolCaller{delay: 100 * time.Millisecond}
	executor := NewExecutor(caller)

	req := BatchCallRequest{
		Calls: []ToolCallRequest{
			{ID: "1", ToolName: "tool1"},
			{ID: "2", ToolName: "tool2"},
		},
		Parallel: true,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := executor.Execute(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestExecuteEmpty(t *testing.T) {
	caller := &mockToolCaller{}
	executor := NewExecutor(caller)

	req := BatchCallRequest{
		Calls:    []ToolCallRequest{},
		Parallel: false,
	}

	ctx := context.Background()
	result, err := executor.Execute(ctx, req)

	require.NoError(t, err)
	require.Len(t, result.Results, 0)
}
