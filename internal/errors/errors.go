package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Type classifies an error for exit code mapping and JSON serialization.
type Type string

// All recognized error types. Config, Auth, and Network errors map to exit code 2;
// all others map to exit code 1.
const (
	ConfigError     Type = "config_error"
	AuthError       Type = "auth_error"
	NetworkError    Type = "network_error"
	APIError        Type = "api_error"
	ValidationError Type = "validation_error"
	IOError         Type = "io_error"
	NotFound        Type = "not_found"
	UserCancelled   Type = "user_cancelled"
)

// Error is a structured error carrying a type, human-readable message,
// optional detail map, and an optional wrapped cause.
type Error struct {
	Type    Type
	Message string
	Details map[string]any
	Cause   error
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause, enabling errors.Is / errors.As chains.
func (e *Error) Unwrap() error {
	return e.Cause
}

// ExitCode returns the process exit code for this error type.
// Config, auth, and network errors return 2; everything else returns 1.
func (e *Error) ExitCode() int {
	switch e.Type {
	case ConfigError, AuthError, NetworkError:
		return ExitConfigError
	default:
		return ExitError
	}
}

// --- Constructor helpers ---

// NewConfigNotFound creates a config_error for a missing configuration file.
func NewConfigNotFound(path string) *Error {
	return &Error{
		Type:    ConfigError,
		Message: "no configuration found",
		Details: map[string]any{
			"config_path": path,
			"suggestion":  "Run 'ldctl config init' to get started",
		},
	}
}

// NewAuthFailed creates an auth_error for authentication failures.
func NewAuthFailed(url string, status int) *Error {
	return &Error{
		Type:    AuthError,
		Message: fmt.Sprintf("authentication failed (%d %s)", status, http.StatusText(status)),
		Details: map[string]any{
			"http_status":  status,
			"instance_url": url,
		},
	}
}

// NewNotFound creates a not_found error for missing resources.
func NewNotFound(resource string, id interface{}) *Error {
	return &Error{
		Type:    NotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: map[string]any{
			"resource": resource,
			"id":       id,
		},
	}
}

// NewValidation creates a validation_error for invalid user input.
func NewValidation(message string, details map[string]any) *Error {
	return &Error{
		Type:    ValidationError,
		Message: message,
		Details: details,
	}
}

// NewNetworkError creates a network_error for connectivity failures.
func NewNetworkError(url string, cause error) *Error {
	details := map[string]any{
		"url": url,
	}
	if cause != nil {
		details["error"] = cause.Error()
	}
	return &Error{
		Type:    NetworkError,
		Message: "could not connect to LinkDing instance",
		Details: details,
		Cause:   cause,
	}
}

// NewIOError creates an io_error for file system operation failures.
func NewIOError(message string, cause error) *Error {
	details := map[string]any{}
	if cause != nil {
		details["error"] = cause.Error()
	}
	return &Error{
		Type:    IOError,
		Message: message,
		Details: details,
		Cause:   cause,
	}
}

// FromHTTPStatus maps an HTTP response status code to a typed Error.
// It covers 400, 401, 403, 404, 429, and 5xx status codes.
func FromHTTPStatus(statusCode int, url string) *Error {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &Error{
			Type:    AuthError,
			Message: fmt.Sprintf("authentication failed (%d %s)", statusCode, http.StatusText(statusCode)),
			Details: map[string]any{
				"http_status":  statusCode,
				"instance_url": url,
			},
		}
	case http.StatusNotFound:
		return &Error{
			Type:    NotFound,
			Message: "resource not found",
			Details: map[string]any{
				"http_status": statusCode,
			},
		}
	case http.StatusBadRequest:
		return &Error{
			Type:    ValidationError,
			Message: "validation failed",
			Details: map[string]any{
				"http_status": statusCode,
			},
		}
	case http.StatusTooManyRequests:
		return &Error{
			Type:    APIError,
			Message: "rate limit exceeded",
			Details: map[string]any{
				"http_status": statusCode,
			},
		}
	default:
		if statusCode >= 500 && statusCode < 600 {
			return &Error{
				Type:    APIError,
				Message: fmt.Sprintf("server error (%d %s)", statusCode, http.StatusText(statusCode)),
				Details: map[string]any{
					"http_status": statusCode,
				},
			}
		}
		return &Error{
			Type:    APIError,
			Message: fmt.Sprintf("unexpected response (%d %s)", statusCode, http.StatusText(statusCode)),
			Details: map[string]any{
				"http_status": statusCode,
			},
		}
	}
}

// As is a convenience re-export of the standard library errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}
