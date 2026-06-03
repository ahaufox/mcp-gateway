package transport

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// Status 表示传输层状态
type Status int

const (
	// StatusDisconnected 未连接
	StatusDisconnected Status = iota
	// StatusConnecting 连接中
	StatusConnecting
	// StatusConnected 已连接
	StatusConnected
	// StatusReconnecting 重连中
	StatusReconnecting
	// StatusError 错误状态
	StatusError
)

// String 返回状态的字符串表示
func (s Status) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	case StatusReconnecting:
		return "reconnecting"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// Transport 统一的传输层接口
type Transport interface {
	// Connect 建立连接
	Connect(ctx context.Context) error
	
	// Send 发送消息
	Send(ctx context.Context, msg *mcp.JSONRPCMessage) error
	
	// Receive 接收消息（阻塞）
	Receive(ctx context.Context) (*mcp.JSONRPCMessage, error)
	
	// Ping 健康检查
	Ping(ctx context.Context) error
	
	// Close 关闭连接
	Close() error
	
	// Status 获取连接状态
	Status() Status
}

// TransportFactory 传输工厂接口
type TransportFactory interface {
	// Create 创建传输实例
	Create(config any) (Transport, error)
	
	// Supports 检查是否支持指定类型
	Supports(transportType string) bool
	
	// Name 返回工厂名称
	Name() string
}

// BaseTransport 提供 Transport 接口的基础实现
type BaseTransport struct {
	status Status
}

// Status 返回当前状态
func (b *BaseTransport) Status() Status {
	return b.status
}

// SetStatus 设置状态
func (b *BaseTransport) SetStatus(s Status) {
	b.status = s
}
