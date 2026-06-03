package transport

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mark3labs/mcp-go/mcp"
)

// WebSocketConfig WebSocket 配置
type WebSocketConfig struct {
	// URL WebSocket 地址
	URL string
	// Headers 自定义请求头
	Headers map[string]string
	// HandshakeTimeout 握手超时时间
	HandshakeTimeout time.Duration
	// ReadBufferSize 读取缓冲区大小
	ReadBufferSize int
	// WriteBufferSize 写入缓冲区大小
	WriteBufferSize int
	// EnableCompression 启用压缩
	EnableCompression bool
	// PingInterval 心跳间隔
	PingInterval time.Duration
	// ReconnectDelay 重连延迟
	ReconnectDelay time.Duration
	// MaxReconnectAttempts 最大重连次数
	MaxReconnectAttempts int
}

// DefaultWebSocketConfig 返回默认配置
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		HandshakeTimeout:    10 * time.Second,
		ReadBufferSize:      4096,
		WriteBufferSize:     4096,
		EnableCompression:   false,
		PingInterval:        30 * time.Second,
		ReconnectDelay:       5 * time.Second,
		MaxReconnectAttempts: 5,
	}
}

// WebSocketClient WebSocket MCP 客户端
type WebSocketClient struct {
	BaseTransport
	config    WebSocketConfig
	conn      *websocket.Conn
	mu        sync.RWMutex
	closeCh   chan struct{}
	readCh    chan *mcp.JSONRPCMessage
	writeCh   chan *mcp.JSONRPCMessage
	errorCh   chan error
	doneCh    chan struct{}
	
	// 重连状态
	reconnectAttempts int
	shouldReconnect   bool
}

// NewWebSocketClient 创建新的 WebSocket 客户端
func NewWebSocketClient(config WebSocketConfig) *WebSocketClient {
	return &WebSocketClient{
		BaseTransport: BaseTransport{status: StatusDisconnected},
		config:        config,
		closeCh:       make(chan struct{}),
		readCh:        make(chan *mcp.JSONRPCMessage, 100),
		writeCh:       make(chan *mcp.JSONRPCMessage, 100),
		errorCh:       make(chan error, 10),
		doneCh:        make(chan struct{}),
		shouldReconnect: true,
	}
}

// Connect 建立 WebSocket 连接
func (c *WebSocketClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	if c.status == StatusConnected {
		c.mu.Unlock()
		return nil
	}
	c.SetStatus(StatusConnecting)
	c.mu.Unlock()

	// 构建 HTTP 请求头
	header := make(http.Header)
	for k, v := range c.config.Headers {
		header.Set(k, v)
	}

	// 建立连接
	dialer := websocket.Dialer{
		HandshakeTimeout: c.config.HandshakeTimeout,
		ReadBufferSize:   c.config.ReadBufferSize,
		WriteBufferSize:  c.config.WriteBufferSize,
		EnableCompression: c.config.EnableCompression,
	}

	conn, _, err := dialer.DialContext(ctx, c.config.URL, header)
	if err != nil {
		c.SetStatus(StatusError)
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.reconnectAttempts = 0
	c.shouldReconnect = true
	c.SetStatus(StatusConnected)
	c.mu.Unlock()

	// 启动读写协程
	go c.readLoop()
	go c.writeLoop()
	go c.pingLoop()

	log.Printf("[websocket] connected to %s", c.config.URL)
	return nil
}

// readLoop 读取消息循环
func (c *WebSocketClient) readLoop() {
	for {
		select {
		case <-c.closeCh:
			return
		case <-c.doneCh:
			return
		default:
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				if !c.shouldReconnect {
					return
				}
				c.handleReadError(err)
				return
			}

			var rpcMsg mcp.JSONRPCMessage
			if err := json.Unmarshal(msg, &rpcMsg); err != nil {
				log.Printf("[websocket] failed to unmarshal message: %v", err)
				continue
			}

			select {
			case c.readCh <- &rpcMsg:
			case <-c.closeCh:
				return
			}
		}
	}
}

// writeLoop 写入消息循环
func (c *WebSocketClient) writeLoop() {
	for {
		select {
		case <-c.closeCh:
			return
		case <-c.doneCh:
			return
		case msg := <-c.writeCh:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()
			
			if conn == nil {
				continue
			}

			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("[websocket] failed to marshal message: %v", err)
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[websocket] failed to write message: %v", err)
				c.errorCh <- err
			}
		}
	}
}

// pingLoop 心跳循环
func (c *WebSocketClient) pingLoop() {
	ticker := time.NewTicker(c.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.closeCh:
			return
		case <-c.doneCh:
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				continue
			}

			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				log.Printf("[websocket] ping failed: %v", err)
			}
		}
	}
}

// handleReadError 处理读取错误
func (c *WebSocketClient) handleReadError(err error) {
	// 检查是否是正常的关闭
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
		log.Printf("[websocket] connection error: %v", err)
	}

	c.mu.Lock()
	c.SetStatus(StatusError)
	c.mu.Unlock()

	// 尝试重连
	if c.shouldReconnect && c.reconnectAttempts < c.config.MaxReconnectAttempts {
		c.reconnectAttempts++
		log.Printf("[websocket] attempting reconnect (attempt %d/%d) in %v",
			c.reconnectAttempts, c.config.MaxReconnectAttempts, c.config.ReconnectDelay)
		
		c.SetStatus(StatusReconnecting)
		
		select {
		case <-c.closeCh:
			return
		case <-time.After(c.config.ReconnectDelay):
			if err := c.Connect(context.Background()); err != nil {
				log.Printf("[websocket] reconnect failed: %v", err)
				c.errorCh <- err
			}
		}
	} else {
		c.errorCh <- err
	}
}

// Send 发送消息
func (c *WebSocketClient) Send(ctx context.Context, msg *mcp.JSONRPCMessage) error {
	select {
	case c.writeCh <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-c.closeCh:
		return errors.New("connection closed")
	}
}

// Receive 接收消息
func (c *WebSocketClient) Receive(ctx context.Context) (*mcp.JSONRPCMessage, error) {
	select {
	case msg := <-c.readCh:
		return msg, nil
	case err := <-c.errorCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.closeCh:
		return nil, errors.New("connection closed")
	}
}

// Ping 健康检查
func (c *WebSocketClient) Ping(ctx context.Context) error {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return errors.New("not connected")
	}

	// 发送 ping 并等待 pong
	pingCh := make(chan error, 1)
	go func() {
		pingCh <- conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
	}()

	select {
	case err := <-pingCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return errors.New("ping timeout")
	}
}

// Close 关闭连接
func (c *WebSocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status == StatusDisconnected {
		return nil
	}

	c.shouldReconnect = false
	close(c.closeCh)
	close(c.doneCh)

	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.SetStatus(StatusDisconnected)
		return err
	}

	c.SetStatus(StatusDisconnected)
	return nil
}

// ReadChannel 返回读取通道（用于集成）
func (c *WebSocketClient) ReadChannel() <-chan *mcp.JSONRPCMessage {
	return c.readCh
}

// WriteChannel 返回写入通道（用于集成）
func (c *WebSocketClient) WriteChannel() chan<- *mcp.JSONRPCMessage {
	return c.writeCh
}

// ErrorChannel 返回错误通道
func (c *WebSocketClient) ErrorChannel() <-chan error {
	return c.errorCh
}

// WebSocketFactory WebSocket 传输工厂
type WebSocketFactory struct{}

// Create 创建 WebSocket 传输
func (f *WebSocketFactory) Create(config any) (Transport, error) {
	cfg, ok := config.(WebSocketConfig)
	if !ok {
		return nil, errors.New("invalid config type")
	}
	return NewWebSocketClient(cfg), nil
}

// Supports 检查是否支持
func (f *WebSocketFactory) Supports(transportType string) bool {
	return strings.EqualFold(transportType, "websocket") || 
		   strings.EqualFold(transportType, "ws")
}

// Name 返回工厂名称
func (f *WebSocketFactory) Name() string {
	return "websocket"
}
