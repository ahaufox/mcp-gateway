package batch

import (
	"context"
	"fmt"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
)

// BatchCallRequest 批量工具调用请求
type BatchCallRequest struct {
	Calls    []ToolCallRequest `json:"calls"`
	Parallel bool              `json:"parallel"`
}

// ToolCallRequest 单个工具调用请求
type ToolCallRequest struct {
	ID       string                 `json:"id"`
	ToolName string                 `json:"toolName"`
	Arguments map[string]interface{} `json:"arguments"`
}

// BatchCallResult 批量工具调用结果
type BatchCallResult struct {
	Results []ToolCallResult `json:"results"`
}

// ToolCallResult 单个工具调用结果
type ToolCallResult struct {
	ID      string                 `json:"id"`
	Success bool                   `json:"success"`
	Result  *mcp.CallToolResult    `json:"result,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// ToolCaller 工具调用器接口
type ToolCaller interface {
	CallTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// Executor 批量执行器
type Executor struct {
	caller ToolCaller
}

// NewExecutor 创建批量执行器
func NewExecutor(caller ToolCaller) *Executor {
	return &Executor{
		caller: caller,
	}
}

// Execute 执行批量调用
func (e *Executor) Execute(ctx context.Context, req BatchCallRequest) (*BatchCallResult, error) {
	if len(req.Calls) == 0 {
		return &BatchCallResult{Results: []ToolCallResult{}}, nil
	}

	if req.Parallel {
		return e.executeParallel(ctx, req.Calls)
	}
	return e.executeSequential(ctx, req.Calls)
}

// executeSequential 顺序执行
func (e *Executor) executeSequential(ctx context.Context, calls []ToolCallRequest) (*BatchCallResult, error) {
	results := make([]ToolCallResult, 0, len(calls))

	for _, call := range calls {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result := e.executeSingle(ctx, call)
		results = append(results, result)
	}

	return &BatchCallResult{Results: results}, nil
}

// executeParallel 并行执行
func (e *Executor) executeParallel(ctx context.Context, calls []ToolCallRequest) (*BatchCallResult, error) {
	results := make([]ToolCallResult, len(calls))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, call := range calls {
		wg.Add(1)
		go func(index int, c ToolCallRequest) {
			defer wg.Done()
			
			select {
			case <-ctx.Done():
				mu.Lock()
				results[index] = ToolCallResult{
					ID:      c.ID,
					Success: false,
					Error:   ctx.Err().Error(),
				}
				mu.Unlock()
				return
			default:
			}

			result := e.executeSingle(ctx, c)
			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, call)
	}

	wg.Wait()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return &BatchCallResult{Results: results}, nil
}

// executeSingle 执行单个工具调用
func (e *Executor) executeSingle(ctx context.Context, call ToolCallRequest) ToolCallResult {
	req := mcp.CallToolRequest{}
	req.Params.Name = call.ToolName
	req.Params.Arguments = call.Arguments

	result, err := e.caller.CallTool(ctx, req)
	if err != nil {
		return ToolCallResult{
			ID:      call.ID,
			Success: false,
			Error:   err.Error(),
		}
	}

	return ToolCallResult{
		ID:      call.ID,
		Success: true,
		Result:  result,
	}
}

// ValidateRequest 验证批量请求
func ValidateRequest(req BatchCallRequest) error {
	if len(req.Calls) == 0 {
		return fmt.Errorf("no calls provided")
	}

	seenIDs := make(map[string]struct{})
	for _, call := range req.Calls {
		if call.ID == "" {
			return fmt.Errorf("call ID is required")
		}
		if call.ToolName == "" {
			return fmt.Errorf("tool name is required for call %s", call.ID)
		}
		if _, exists := seenIDs[call.ID]; exists {
			return fmt.Errorf("duplicate call ID: %s", call.ID)
		}
		seenIDs[call.ID] = struct{}{}
	}

	return nil
}
