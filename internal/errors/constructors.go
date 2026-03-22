package errors

import "fmt"

// NewNetworkError returns a network_error for a connection failure reaching url.
// The original transport-level cause is wrapped for errors.Is/As inspection.
func NewNetworkError(url string, cause error) *Error {
	return &Error{
		Type:    NetworkError,
		Message: fmt.Sprintf("failed to connect to %s", url),
		Details: map[string]interface{}{
			"url": url,
		},
		Cause: cause,
	}
}

// NewAPIError returns an api_error for a 5xx HTTP response. httpStatus is
// included in Details so callers can surface it to the user.
func NewAPIError(httpStatus int, url string) *Error {
	return &Error{
		Type:    APIError,
		Message: fmt.Sprintf("server error (%d)", httpStatus),
		Details: map[string]interface{}{
			"http_status": httpStatus,
			"url":         url,
		},
	}
}

// NewValidationError returns a validation_error for an invalid field value.
// field is the name of the offending field and message describes the violation.
func NewValidationError(field, message string) *Error {
	return &Error{
		Type:    ValidationError,
		Message: fmt.Sprintf("validation failed: %s", message),
		Details: map[string]interface{}{
			"field":   field,
			"message": message,
		},
	}
}

// NewIOError returns an io_error for a file-system operation failure at path.
// The original OS-level cause is wrapped for errors.Is/As inspection.
func NewIOError(path string, cause error) *Error {
	return &Error{
		Type:    IOError,
		Message: fmt.Sprintf("I/O error on %s", path),
		Details: map[string]interface{}{
			"path": path,
		},
		Cause: cause,
	}
}

// NewUserCancelled returns a user_cancelled error indicating the user aborted
// an interactive operation before it completed.
func NewUserCancelled() *Error {
	return &Error{
		Type:    UserCancelled,
		Message: "operation cancelled by user",
	}
}
