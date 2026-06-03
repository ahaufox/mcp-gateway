package transport

import (
	"testing"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusDisconnected, "disconnected"},
		{StatusConnecting, "connecting"},
		{StatusConnected, "connected"},
		{StatusReconnecting, "reconnecting"},
		{StatusError, "error"},
		{Status(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("Status.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBaseTransport(t *testing.T) {
	bt := &BaseTransport{status: StatusDisconnected}

	if bt.Status() != StatusDisconnected {
		t.Error("expected initial status to be disconnected")
	}

	bt.SetStatus(StatusConnected)
	if bt.Status() != StatusConnected {
		t.Error("expected status to be connected after SetStatus")
	}

	bt.SetStatus(StatusError)
	if bt.Status() != StatusError {
		t.Error("expected status to be error after SetStatus")
	}
}
