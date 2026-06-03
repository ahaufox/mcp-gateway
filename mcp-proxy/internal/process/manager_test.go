package process

import (
	"os/exec"
	"testing"
	"time"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		eventType EventType
		name      string
	}{
		{EventStart, "EventStart"},
		{EventStop, "EventStop"},
		{EventRestart, "EventRestart"},
		{EventError, "EventError"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.eventType.String(); got == "" {
				t.Error("event type should have a string representation")
			}
		})
	}
}

func TestEvent_Fields(t *testing.T) {
	event := Event{
		Type:      EventStart,
		Timestamp: time.Now(),
		Message:   "Test message",
		PID:       12345,
	}

	if event.Type != EventStart {
		t.Errorf("expected type EventStart, got %v", event.Type)
	}

	if event.Message != "Test message" {
		t.Errorf("expected message 'Test message', got %s", event.Message)
	}

	if event.PID != 12345 {
		t.Errorf("expected PID 12345, got %d", event.PID)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.RestartDelay != 1*time.Second {
		t.Errorf("expected default RestartDelay of 1s, got %v", config.RestartDelay)
	}

	if config.MaxRestartCount != 5 {
		t.Errorf("expected default MaxRestartCount of 5, got %d", config.MaxRestartCount)
	}

	if config.OnEvent != nil {
		t.Error("expected default OnEvent to be nil")
	}
}

func TestManager_Lifecycle(t *testing.T) {
	// 创建一个简单的测试命令
	cmd := exec.Command("sleep", "10")
	
	config := DefaultConfig()
	config.MaxRestartCount = 3
	config.RestartDelay = 100 * time.Millisecond
	
	manager := NewManager("test-process", cmd, config)

	// 测试 PID 在未启动时为 0
	if pid := manager.PID(); pid != 0 {
		t.Errorf("expected PID 0 before start, got %d", pid)
	}

	// 测试 IsRunning 在未启动时返回 false
	if running := manager.IsRunning(); running {
		t.Error("expected IsRunning to return false before start")
	}

	// 测试 RestartCount 在未启动时为 0
	if count := manager.RestartCount(); count != 0 {
		t.Errorf("expected RestartCount 0 before start, got %d", count)
	}
}

func TestManager_StartFailure(t *testing.T) {
	// 创建一个不存在的命令
	cmd := exec.Command("/nonexistent/command")
	
	config := DefaultConfig()
	manager := NewManager("test-fail", cmd, config)

	err := manager.Start()
	if err == nil {
		t.Error("expected error when starting nonexistent command")
	}

	// 确保状态正确
	if running := manager.IsRunning(); running {
		t.Error("expected IsRunning to return false after failed start")
	}
}

func TestManager_Close(t *testing.T) {
	cmd := exec.Command("sleep", "60")
	
	config := DefaultConfig()
	manager := NewManager("test-close", cmd, config)

	// 未启动时关闭应该成功
	err := manager.Close()
	if err != nil {
		t.Errorf("Close should not error when not started: %v", err)
	}

	// 创建新实例用于启动测试
	cmd2 := exec.Command("sleep", "60")
	manager2 := NewManager("test-close-2", cmd2, config)
	
	err = manager2.Start()
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// 确保进程在运行
	if !manager2.IsRunning() {
		t.Error("expected IsRunning to return true after start")
	}

	err = manager2.Close()
	if err != nil {
		t.Errorf("Close should not error: %v", err)
	}

	// 确保关闭后不在运行
	if manager2.IsRunning() {
		t.Error("expected IsRunning to return false after close")
	}
	
	// 确保 PID 重置
	if pid := manager2.PID(); pid != 0 {
		t.Errorf("expected PID 0 after close, got %d", pid)
	}
}

func TestManager_EventCallback(t *testing.T) {
	eventsReceived := make([]Event, 0)
	
	cmd := exec.Command("sleep", "60")
	
	config := DefaultConfig()
	config.OnEvent = func(event Event) {
		eventsReceived = append(eventsReceived, event)
	}
	
	manager := NewManager("test-callback", cmd, config)

	err := manager.Start()
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	// 等待事件被触发
	time.Sleep(200 * time.Millisecond)

	// 应该收到至少一个事件
	if len(eventsReceived) == 0 {
		t.Error("expected at least one event to be received")
	}

	// 第一个事件应该是启动事件
	if len(eventsReceived) > 0 && eventsReceived[0].Type != EventStart {
		t.Errorf("expected first event to be EventStart, got %v", eventsReceived[0].Type)
	}

	manager.Close()
}
