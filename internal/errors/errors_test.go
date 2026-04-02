package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// --- ExitCode mapping tests ---

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		errType  Type
		wantCode int
	}{
		{"config_error → 2", ConfigError, ExitConfigError},
		{"auth_error → 2", AuthError, ExitConfigError},
		{"network_error → 2", NetworkError, ExitConfigError},
		{"api_error → 1", APIError, ExitError},
		{"validation_error → 1", ValidationError, ExitError},
		{"io_error → 1", IOError, ExitError},
		{"not_found → 1", NotFound, ExitError},
		{"user_cancelled → 1", UserCancelled, ExitError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Error{Type: tt.errType, Message: "test"}
			if got := e.ExitCode(); got != tt.wantCode {
				t.Errorf("ExitCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

// --- Error interface tests ---

func TestErrorInterface(t *testing.T) {
	e := &Error{Type: APIError, Message: "something broke"}
	var err error = e
	if err.Error() != "something broke" {
		t.Errorf("Error() = %q, want %q", err.Error(), "something broke")
	}
}

func TestUnwrap(t *testing.T) {
	cause := fmt.Errorf("underlying")
	e := &Error{Type: IOError, Message: "wrapper", Cause: cause}
	if !errors.Is(e, cause) {
		t.Error("Unwrap chain should contain the cause")
	}
}

func TestAs(t *testing.T) {
	e := &Error{Type: NotFound, Message: "gone"}
	wrapped := fmt.Errorf("context: %w", e)

	var target *Error
	if !As(wrapped, &target) {
		t.Fatal("As should find *Error in wrapped chain")
	}
	if target.Type != NotFound {
		t.Errorf("Type = %q, want %q", target.Type, NotFound)
	}
}

// --- Constructor tests ---

func TestNewConfigNotFound(t *testing.T) {
	e := NewConfigNotFound("/home/user/.config/ldctl/config.toml")
	if e.Type != ConfigError {
		t.Errorf("Type = %q, want %q", e.Type, ConfigError)
	}
	if e.ExitCode() != ExitConfigError {
		t.Errorf("ExitCode() = %d, want %d", e.ExitCode(), ExitConfigError)
	}
	if e.Details["config_path"] != "/home/user/.config/ldctl/config.toml" {
		t.Error("missing config_path detail")
	}
	if e.Details["suggestion"] == nil {
		t.Error("missing suggestion detail")
	}
}

func TestNewAuthFailed(t *testing.T) {
	e := NewAuthFailed("https://ld.example.com", http.StatusUnauthorized)
	if e.Type != AuthError {
		t.Errorf("Type = %q, want %q", e.Type, AuthError)
	}
	if !strings.Contains(e.Message, "401") {
		t.Errorf("Message should contain status code, got %q", e.Message)
	}
	if e.Details["http_status"] != http.StatusUnauthorized {
		t.Error("missing http_status detail")
	}
	if e.Details["instance_url"] != "https://ld.example.com" {
		t.Error("missing instance_url detail")
	}
}

func TestNewNotFound(t *testing.T) {
	e := NewNotFound("bookmark", 42)
	if e.Type != NotFound {
		t.Errorf("Type = %q, want %q", e.Type, NotFound)
	}
	if e.Message != "bookmark not found" {
		t.Errorf("Message = %q, want %q", e.Message, "bookmark not found")
	}
	if e.Details["resource"] != "bookmark" {
		t.Error("missing resource detail")
	}
	if e.Details["id"] != 42 {
		t.Error("missing id detail")
	}
}

func TestNewValidation(t *testing.T) {
	e := NewValidation("invalid URL format", map[string]any{
		"field": "url",
		"value": "not-a-url",
	})
	if e.Type != ValidationError {
		t.Errorf("Type = %q, want %q", e.Type, ValidationError)
	}
	if e.ExitCode() != ExitError {
		t.Errorf("ExitCode() = %d, want %d", e.ExitCode(), ExitError)
	}
}

func TestNewNetworkError(t *testing.T) {
	cause := fmt.Errorf("dial tcp: connection refused")
	e := NewNetworkError("https://ld.example.com", cause)
	if e.Type != NetworkError {
		t.Errorf("Type = %q, want %q", e.Type, NetworkError)
	}
	if e.ExitCode() != ExitConfigError {
		t.Errorf("ExitCode() = %d, want %d", e.ExitCode(), ExitConfigError)
	}
	if e.Details["url"] != "https://ld.example.com" {
		t.Error("missing url detail")
	}
	if !errors.Is(e, cause) {
		t.Error("Unwrap chain should contain the cause")
	}
}

func TestNewNetworkErrorNilCause(t *testing.T) {
	e := NewNetworkError("https://ld.example.com", nil)
	if _, ok := e.Details["error"]; ok {
		t.Error("should not include 'error' detail when cause is nil")
	}
}

func TestNewIOError(t *testing.T) {
	cause := fmt.Errorf("permission denied")
	e := NewIOError("cannot write to file", cause)
	if e.Type != IOError {
		t.Errorf("Type = %q, want %q", e.Type, IOError)
	}
	if e.ExitCode() != ExitError {
		t.Errorf("ExitCode() = %d, want %d", e.ExitCode(), ExitError)
	}
	if !errors.Is(e, cause) {
		t.Error("Unwrap chain should contain the cause")
	}
}

// --- HTTP status mapping tests ---

func TestFromHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		wantType Type
		wantCode int
		wantMsg  string
	}{
		{"401 → auth_error", 401, AuthError, ExitConfigError, "authentication failed (401 Unauthorized)"},
		{"403 → auth_error", 403, AuthError, ExitConfigError, "authentication failed (403 Forbidden)"},
		{"404 → not_found", 404, NotFound, ExitError, "resource not found"},
		{"400 → validation_error", 400, ValidationError, ExitError, "validation failed"},
		{"429 → api_error", 429, APIError, ExitError, "rate limit exceeded"},
		{"500 → api_error", 500, APIError, ExitError, "server error (500 Internal Server Error)"},
		{"502 → api_error", 502, APIError, ExitError, "server error (502 Bad Gateway)"},
		{"503 → api_error", 503, APIError, ExitError, "server error (503 Service Unavailable)"},
		{"504 → api_error", 504, APIError, ExitError, "server error (504 Gateway Timeout)"},
		{"418 → api_error (unexpected)", 418, APIError, ExitError, "unexpected response (418 I'm a teapot)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := FromHTTPStatus(tt.status, "https://ld.example.com")
			if e.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", e.Type, tt.wantType)
			}
			if e.ExitCode() != tt.wantCode {
				t.Errorf("ExitCode() = %d, want %d", e.ExitCode(), tt.wantCode)
			}
			if e.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", e.Message, tt.wantMsg)
			}
			if e.Details["http_status"] != tt.status {
				t.Error("missing http_status detail")
			}
		})
	}
}

// --- FormatHuman tests ---

func TestFormatHuman(t *testing.T) {
	t.Run("with suggestion", func(t *testing.T) {
		e := NewConfigNotFound("/path")
		got := FormatHuman(e)
		if !strings.HasPrefix(got, "Error: ") {
			t.Errorf("should start with 'Error: ', got %q", got)
		}
		if !strings.Contains(got, "ldctl config init") {
			t.Error("should include suggestion text")
		}
	})

	t.Run("without suggestion", func(t *testing.T) {
		e := &Error{Type: APIError, Message: "server error"}
		got := FormatHuman(e)
		if got != "Error: server error" {
			t.Errorf("got %q, want %q", got, "Error: server error")
		}
	})

	t.Run("nil details", func(t *testing.T) {
		e := &Error{Type: APIError, Message: "oops", Details: nil}
		got := FormatHuman(e)
		if got != "Error: oops" {
			t.Errorf("got %q, want %q", got, "Error: oops")
		}
	})
}

// --- FormatJSON tests ---

func TestFormatJSON(t *testing.T) {
	t.Run("valid JSON structure", func(t *testing.T) {
		e := NewConfigNotFound("/home/user/.config/ldctl/config.toml")
		jsonStr := FormatJSON(e)

		var envelope struct {
			Error struct {
				Type    string         `json:"type"`
				Message string         `json:"message"`
				Details map[string]any `json:"details"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &envelope); err != nil {
			t.Fatalf("invalid JSON: %v\nraw: %s", err, jsonStr)
		}
		if envelope.Error.Type != string(ConfigError) {
			t.Errorf("type = %q, want %q", envelope.Error.Type, ConfigError)
		}
		if envelope.Error.Message != "no configuration found" {
			t.Errorf("message = %q, want %q", envelope.Error.Message, "no configuration found")
		}
		if envelope.Error.Details["config_path"] != "/home/user/.config/ldctl/config.toml" {
			t.Error("missing config_path in details")
		}
	})

	t.Run("nil details becomes empty object", func(t *testing.T) {
		e := &Error{Type: APIError, Message: "test", Details: nil}
		jsonStr := FormatJSON(e)

		var envelope struct {
			Error struct {
				Details map[string]any `json:"details"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &envelope); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if envelope.Error.Details == nil {
			t.Error("details should be empty object, not null")
		}
	})

	t.Run("all error types produce valid JSON", func(t *testing.T) {
		types := []Type{ConfigError, AuthError, NetworkError, APIError, ValidationError, IOError, NotFound, UserCancelled}
		for _, typ := range types {
			e := &Error{Type: typ, Message: "msg", Details: map[string]any{"key": "value"}}
			jsonStr := FormatJSON(e)
			var raw map[string]any
			if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
				t.Errorf("type %q produced invalid JSON: %v", typ, err)
			}
		}
	})
}

// --- Exit code constants ---

func TestExitCodeConstants(t *testing.T) {
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
	if ExitError != 1 {
		t.Errorf("ExitError = %d, want 1", ExitError)
	}
	if ExitConfigError != 2 {
		t.Errorf("ExitConfigError = %d, want 2", ExitConfigError)
	}
}
