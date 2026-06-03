package errors

import "fmt"

// Wrap 将普通错误包装为 MCPError
func Wrap(err error, code ErrorCode, message string) *MCPError {
    if err == nil {
        return nil
    }
    if mcpErr, ok := err.(*MCPError); ok {
        return mcpErr
    }
    return &MCPError{
        Code:    code,
        Message: message,
        Cause:   err,
    }
}

// Wrapf 将普通错误包装为带格式化消息的 MCPError
func Wrapf(err error, code ErrorCode, format string, args ...any) *MCPError {
    if err == nil {
        return nil
    }
    if mcpErr, ok := err.(*MCPError); ok {
        return mcpErr
    }
    return &MCPError{
        Code:    code,
        Message: fmt.Sprintf(format, args...),
        Cause:   err,
    }
}

// WithService 添加服务名称到错误
func (e *MCPError) WithService(service string) *MCPError {
    e.Service = service
    return e
}

// WithTool 添加工具名称到错误
func (e *MCPError) WithTool(tool string) *MCPError {
    e.Tool = tool
    return e
}

// WithCause 添加原因到错误
func (e *MCPError) WithCause(cause error) *MCPError {
    e.Cause = cause
    return e
}

// Is 检查错误是否是指定的错误码
func Is(err error, code ErrorCode) bool {
    if mcpErr, ok := err.(*MCPError); ok {
        return mcpErr.Code == code
    }
    return false
}

// CodeOf 获取错误的错误码
func CodeOf(err error) ErrorCode {
    if mcpErr, ok := err.(*MCPError); ok {
        return mcpErr.Code
    }
    return ErrCodeInternal
}

// CauseOf 获取错误的原始原因
func CauseOf(err error) error {
    if mcpErr, ok := err.(*MCPError); ok {
        if mcpErr.Cause != nil {
            return CauseOf(mcpErr.Cause)
        }
        return nil
    }
    return err
}
