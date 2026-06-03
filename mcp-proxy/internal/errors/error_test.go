package errors

import (
    "errors"
    "testing"
)

func TestNew(t *testing.T) {
    tests := []struct {
        name    string
        code    ErrorCode
        message string
    }{
        {
            name:    "timeout error",
            code:    ErrCodeTimeout,
            message: "request timed out",
        },
        {
            name:    "connection error",
            code:    ErrCodeConnection,
            message: "connection failed",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := New(tt.code, tt.message)
            if err.Code != tt.code {
                t.Errorf("expected code %q, got %q", tt.code, err.Code)
            }
            if err.Message != tt.message {
                t.Errorf("expected message %q, got %q", tt.message, err.Message)
            }
        })
    }
}

func TestNewf(t *testing.T) {
    err := Newf(ErrCodeToolNotFound, "tool %q not found", "test-tool")
    if err.Code != ErrCodeToolNotFound {
        t.Errorf("expected code %q, got %q", ErrCodeToolNotFound, err.Code)
    }
    if err.Message != "tool \"test-tool\" not found" {
        t.Errorf("unexpected message: %q", err.Message)
    }
}

func TestWrap(t *testing.T) {
    originalErr := errors.New("original error")
    wrapped := Wrap(originalErr, ErrCodeConnection, "wrapped error")
    
    if wrapped.Code != ErrCodeConnection {
        t.Errorf("expected code %q, got %q", ErrCodeConnection, wrapped.Code)
    }
    if wrapped.Message != "wrapped error" {
        t.Errorf("unexpected message: %q", wrapped.Message)
    }
    if wrapped.Cause != originalErr {
        t.Errorf("expected cause to be original error")
    }
}

func TestWithService(t *testing.T) {
    err := New(ErrCodeServer, "test error").WithService("test-service")
    if err.Service != "test-service" {
        t.Errorf("expected service %q, got %q", "test-service", err.Service)
    }
}

func TestWithTool(t *testing.T) {
    err := New(ErrCodeServer, "test error").WithTool("test-tool")
    if err.Tool != "test-tool" {
        t.Errorf("expected tool %q, got %q", "test-tool", err.Tool)
    }
}

func TestIs(t *testing.T) {
    err := New(ErrCodeTimeout, "timeout")
    if !Is(err, ErrCodeTimeout) {
        t.Error("expected Is to return true")
    }
    if Is(err, ErrCodeConnection) {
        t.Error("expected Is to return false")
    }
}

func TestCodeOf(t *testing.T) {
    err := New(ErrCodeConnection, "connection failed")
    if CodeOf(err) != ErrCodeConnection {
        t.Errorf("expected code %q, got %q", ErrCodeConnection, CodeOf(err))
    }
    
    standardErr := errors.New("standard error")
    if CodeOf(standardErr) != ErrCodeInternal {
        t.Errorf("expected code %q for standard error, got %q", ErrCodeInternal, CodeOf(standardErr))
    }
}

func TestErrorFormat(t *testing.T) {
    tests := []struct {
        name string
        err  *MCPError
        want string
    }{
        {
            name: "basic error",
            err:  New(ErrCodeServer, "simple error"),
            want: "[server] simple error",
        },
        {
            name: "with service",
            err:  New(ErrCodeServer, "service error").WithService("my-service"),
            want: "[server] service error (service=my-service)",
        },
        {
            name: "with tool",
            err:  New(ErrCodeServer, "tool error").WithTool("my-tool"),
            want: "[server] tool error (tool=my-tool)",
        },
        {
            name: "with service and tool",
            err:  New(ErrCodeServer, "both error").WithService("my-service").WithTool("my-tool"),
            want: "[server] both error (service=my-service, tool=my-tool)",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.err.Error(); got != tt.want {
                t.Errorf("Error() = %q, want %q", got, tt.want)
            }
        })
    }
}

func TestUnwrap(t *testing.T) {
    original := errors.New("root cause")
    wrapped := Wrap(original, ErrCodeConnection, "wrapped")
    
    if !errors.Is(wrapped, original) {
        t.Error("errors.Is should match original")
    }
}
