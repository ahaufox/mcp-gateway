package process

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// EventType 进程事件类型
type EventType int

const (
	// EventStart 进程启动事件
	EventStart EventType = iota
	// EventStop 进程停止事件
	EventStop
	// EventRestart 进程重启事件
	EventRestart
	// EventError 进程错误事件
	EventError
)

// String 返回事件类型的字符串表示
func (e EventType) String() string {
	switch e {
	case EventStart:
		return "EventStart"
	case EventStop:
		return "EventStop"
	case EventRestart:
		return "EventRestart"
	case EventError:
		return "EventError"
	default:
		return "Unknown"
	}
}

// Event 进程事件
type Event struct {
	Type      EventType
	Timestamp time.Time
	Message   string
	PID       int
}

// Manager 进程管理器
type Manager struct {
	name    string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	
	events    chan Event
	stop      chan struct{}
	restart   chan struct{}
	
	mu        sync.Mutex
	stopped   bool
	
	// 配置
	restartDelay    time.Duration
	maxRestartCount int
	restartCount    int
	
	// 回调
	onEvent func(Event)
}

// Config 进程管理器配置
type Config struct {
	// RestartDelay 重启延迟
	RestartDelay time.Duration
	// MaxRestartCount 最大重启次数
	MaxRestartCount int
	// OnEvent 事件回调
	OnEvent func(Event)
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		RestartDelay:     1 * time.Second,
		MaxRestartCount:  5,
		OnEvent:          nil,
	}
}

// NewManager 创建新的进程管理器
func NewManager(name string, cmd *exec.Cmd, config Config) *Manager {
	return &Manager{
		name:            name,
		cmd:             cmd,
		events:          make(chan Event, 100),
		stop:            make(chan struct{}),
		restart:         make(chan struct{}, 1),
		restartDelay:    config.RestartDelay,
		maxRestartCount: config.MaxRestartCount,
		onEvent:         config.OnEvent,
	}
}

// Start 启动进程
func (p *Manager) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.stopped {
		return fmt.Errorf("process manager is stopped")
	}
	
	if err := p.startProcess(); err != nil {
		return err
	}
	
	// 启动监控协程
	go p.monitor()
	
	return nil
}

// startProcess 启动实际进程
func (p *Manager) startProcess() error {
	log.Printf("[process:%s] starting process: %s", p.name, p.cmd.Path)
	
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	
	// 设置 stdin/stdout/stderr
	if p.cmd.Stdin != nil {
		if w, ok := p.cmd.Stdin.(io.WriteCloser); ok {
			p.stdin = w
		}
	}
	if p.cmd.Stdout != nil {
		if r, ok := p.cmd.Stdout.(io.ReadCloser); ok {
			p.stdout = r
		}
	}
	if p.cmd.Stderr != nil {
		if r, ok := p.cmd.Stderr.(io.ReadCloser); ok {
			p.stderr = r
		}
	}
	
	p.emit(Event{
		Type:      EventStart,
		Timestamp: time.Now(),
		PID:       p.cmd.Process.Pid,
		Message:   fmt.Sprintf("Process started with PID %d", p.cmd.Process.Pid),
	})
	
	// 启动 stderr 读取协程
	go p.readStderr()
	
	// 启动进程等待协程
	go p.waitForProcess()
	
	return nil
}

// monitor 监控进程状态
func (p *Manager) monitor() {
	for {
		select {
		case <-p.stop:
			return
		case <-p.restart:
			p.handleRestart()
		}
	}
}

// handleRestart 处理进程重启
func (p *Manager) handleRestart() {
	p.mu.Lock()
	
	if p.stopped {
		p.mu.Unlock()
		return
	}
	
	p.restartCount++
	if p.restartCount > p.maxRestartCount {
		log.Printf("[process:%s] max restart count exceeded (%d), giving up", p.name, p.maxRestartCount)
		p.mu.Unlock()
		p.Close()
		return
	}
	
	log.Printf("[process:%s] restarting process (attempt %d/%d), waiting %v...",
		p.name, p.restartCount, p.maxRestartCount, p.restartDelay)
	
	p.mu.Unlock()
	
	// 等待延迟
	select {
	case <-p.stop:
		return
	case <-time.After(p.restartDelay):
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.stopped {
		return
	}
	
	// 创建新的命令
	newCmd := exec.Command(p.cmd.Path, p.cmd.Args[1:]...)
	newCmd.Env = p.cmd.Env
	newCmd.Dir = p.cmd.Dir
	newCmd.Stdin = nil
	newCmd.Stdout = nil
	newCmd.Stderr = nil
	
	p.cmd = newCmd
	
	if err := p.startProcess(); err != nil {
		log.Printf("[process:%s] failed to restart process: %v", p.name, err)
		p.emit(Event{
			Type:      EventError,
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Failed to restart: %v", err),
		})
		return
	}
	
	p.emit(Event{
		Type:      EventRestart,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Process restarted (attempt %d)", p.restartCount),
	})
}

// readStderr 读取 stderr
func (p *Manager) readStderr() {
	if p.stderr == nil {
		return
	}
	
	scanner := bufio.NewScanner(p.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[process:%s] STDERR: %s", p.name, line)
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("[process:%s] stderr scanner error: %v", p.name, err)
	}
}

// waitForProcess 等待进程结束
func (p *Manager) waitForProcess() {
	err := p.cmd.Wait()
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.stopped {
		return
	}
	
	p.emit(Event{
		Type:      EventStop,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Process exited: %v", err),
	})
	
	// 触发重启
	select {
	case p.restart <- struct{}{}:
	default:
	}
}

// RequestRestart 请求重启进程
func (p *Manager) RequestRestart() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.stopped {
		return
	}
	
	select {
	case p.restart <- struct{}{}:
	default:
	}
}

// Close 关闭进程管理器
func (p *Manager) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.stopped {
		return nil
	}
	
	p.stopped = true
	close(p.stop)
	
	p.emit(Event{
		Type:      EventStop,
		Timestamp: time.Now(),
		Message:   "Process manager closed",
	})
	
	// 关闭 stdin
	if p.stdin != nil {
		p.stdin.Close()
	}
	
	// 杀死进程
	if p.cmd.Process != nil {
		p.cmd.Process.Kill()
		p.cmd.Process = nil
	}
	
	return nil
}

// Stdout 返回 stdout
func (p *Manager) Stdout() io.ReadCloser {
	return p.stdout
}

// Stdin 返回 stdin
func (p *Manager) Stdin() io.WriteCloser {
	return p.stdin
}

// IsRunning 检查进程是否运行中
func (p *Manager) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.stopped || p.cmd.Process == nil {
		return false
	}
	
	// 检查进程是否还在运行
	process, err := os.FindProcess(p.cmd.Process.Pid)
	if err != nil {
		return false
	}
	
	// FindProcess 总是成功，需要检查进程是否真的在运行
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// PID 返回进程 PID
func (p *Manager) PID() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

// emit 发送事件
func (p *Manager) emit(event Event) {
	if p.onEvent != nil {
		p.onEvent(event)
	}
	
	select {
	case p.events <- event:
	default:
		// 事件通道已满，丢弃事件
	}
}

// Events 返回事件通道
func (p *Manager) Events() <-chan Event {
	return p.events
}

// RestartCount 返回重启次数
func (p *Manager) RestartCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.restartCount
}
