package errors

import "fmt"

// ErrorCode 定义了 MCP 网关的错误码类型
type ErrorCode string

const (
    // ErrCodeTimeout 超时错误
    ErrCodeTimeout ErrorCode = "timeout"
    // ErrCodeConnection 连接错误
    ErrCodeConnection ErrorCode = "connection"
    // ErrCodeProtocol 协议错误
    ErrCodeProtocol ErrorCode = "protocol"
    // ErrCodeServer 服务端错误
    ErrCodeServer ErrorCode = "server"
    // ErrCodeToolNotFound 工具不存在错误
    ErrCodeToolNotFound ErrorCode = "tool_not_found"
    // ErrCodeInvalidRequest 请求参数无效错误
    ErrCodeInvalidRequest ErrorCode = "invalid_request"
    // ErrCodeCircuitOpen 熔断器打开错误
    ErrCodeCircuitOpen ErrorCode = "circuit_open"
    // ErrCodeRateLimit 限流错误
    ErrCodeRateLimit ErrorCode = "rate_limit"
    // ErrCodeResourceExhausted 资源耗尽错误
    ErrCodeResourceExhausted ErrorCode = "resource_exhausted"
    // ErrCodeInternal 内部错误
    ErrCodeInternal ErrorCode = "internal"
)

// MCPError 是 MCP 网关的标准错误类型
type MCPError struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Service string    `json:"service,omitempty"`
    Tool    string    `json:"tool,omitempty"`
    Cause   error     `json:"-"`
}

// Error 实现 error 接口
func (e *MCPError) Error() string {
    if e.Service != "" && e.Tool != "" {
        return fmt.Sprintf("[%s] %s (service=%s, tool=%s)", e.Code, e.Message, e.Service, e.Tool)
    }
    if e.Service != "" {
        return fmt.Sprintf("[%s] %s (service=%s)", e.Code, e.Message, e.Service)
    }
    if e.Tool != "" {
        return fmt.Sprintf("[%s] %s (tool=%s)", e.Code, e.Message, e.Tool)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *MCPError) Unwrap() error {
    return e.Cause
}

// New 创建一个新的 MCPError
func New(code ErrorCode, message string) *MCPError {
    return &MCPError{
        Code:    code,
        Message: message,
    }
}

// Newf 创建一个新的带格式化消息的 MCPError
func Newf(code ErrorCode, format string, args ...any) *MCPError {
    return &MCPError{
        Code:    code,
        Message: fmt.Sprintf(format, args...),
    }
}
