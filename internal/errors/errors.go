package errors

import "fmt"

// Type identifies the category of an ldctl error.
type Type string

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

// Error is ldctl's structured error type. Use a pointer (*Error) everywhere so
// errors.As can locate it in an error chain.
type Error struct {
	Type    Type
	Message string
	Details map[string]interface{}
	Cause   error
}

// Error satisfies the error interface. The plain message is returned; callers
// that need the full context use PrintHuman or PrintJSON.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap exposes the wrapped cause for errors.Is / errors.As chaining.
func (e *Error) Unwrap() error {
	return e.Cause
}

// ExitCode returns 2 for configuration/auth/network errors and 1 for all others.
func (e *Error) ExitCode() int {
	switch e.Type {
	case ConfigError, AuthError, NetworkError:
		return ExitConfigError
	default:
		return ExitError
	}
}

// New is a low-level constructor. Prefer the named constructors below.
func New(typ Type, message string) *Error {
	return &Error{Type: typ, Message: message}
}

// Newf formats a message and returns a new Error.
func Newf(typ Type, format string, args ...any) *Error {
	return &Error{Type: typ, Message: fmt.Sprintf(format, args...)}
}

// NewConfigNotFound returns a config_error for a missing configuration file.
func NewConfigNotFound(path string) *Error {
	return &Error{
		Type:    ConfigError,
		Message: "no configuration found",
		Details: map[string]interface{}{
			"config_path": path,
			"suggestion":  "Run 'ldctl config init' to get started",
		},
	}
}

// NewAuthFailed returns an auth_error for a 401/403 response.
func NewAuthFailed(instanceURL string, httpStatus int) *Error {
	return &Error{
		Type:    AuthError,
		Message: "authentication failed",
		Details: map[string]interface{}{
			"http_status":  httpStatus,
			"instance_url": instanceURL,
		},
	}
}

// NewNotFound returns a not_found error for a missing resource.
func NewNotFound(resource string, id interface{}) *Error {
	return &Error{
		Type:    NotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Details: map[string]interface{}{
			"resource": resource,
			"id":       id,
		},
	}
}
