package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// jsonErrorEnvelope is the wire format for --json error output.
// Details is omitted from JSON entirely when nil (omitempty).
type jsonErrorEnvelope struct {
	Error jsonErrorBody `json:"error"`
}

type jsonErrorBody struct {
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// PrintJSON writes the JSON error envelope to w and returns any encoding error.
// The envelope schema is: {"error": {"type": "...", "message": "...", "details": {...}}}
func PrintJSON(w io.Writer, e *Error) error {
	env := jsonErrorEnvelope{
		Error: jsonErrorBody{
			Type:    string(e.Type),
			Message: e.Message,
			Details: e.Details,
		},
	}
	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", b)
	return err
}

// MapHTTPError converts an HTTP response status to a typed *Error.
// The response body is NOT read; callers should read/close it separately.
func MapHTTPError(resp *http.Response) *Error {
	status := resp.StatusCode
	statusText := http.StatusText(status)

	switch {
	case status == 401 || status == 403:
		return &Error{
			Type:    AuthError,
			Message: fmt.Sprintf("authentication failed (%d %s)", status, statusText),
			Details: map[string]interface{}{"http_status": status},
		}
	case status == 404:
		return &Error{
			Type:    NotFound,
			Message: "resource not found",
			Details: map[string]interface{}{"http_status": status},
		}
	case status == 400:
		return &Error{
			Type:    ValidationError,
			Message: "validation failed",
			Details: map[string]interface{}{"http_status": status},
		}
	case status == 429:
		return &Error{
			Type:    APIError,
			Message: "rate limit exceeded",
			Details: map[string]interface{}{
				"http_status": status,
				"retry_after": resp.Header.Get("Retry-After"),
			},
		}
	case status >= 500 && status <= 599:
		return &Error{
			Type:    APIError,
			Message: fmt.Sprintf("server error (%d %s)", status, statusText),
			Details: map[string]interface{}{"http_status": status},
		}
	default:
		return &Error{
			Type:    APIError,
			Message: fmt.Sprintf("unexpected response (%d %s)", status, statusText),
			Details: map[string]interface{}{"http_status": status},
		}
	}
}
