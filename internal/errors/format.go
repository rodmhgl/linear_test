package errors

import (
	"fmt"
	"io"
)

// hintFor returns an optional actionable hint line for well-known error types.
// Returns empty string when no hint is applicable.
func hintFor(e *Error) string {
	switch e.Type {
	case ConfigError:
		if sug, ok := e.Details["suggestion"].(string); ok {
			return sug
		}
		return "Run 'ldctl config init' to get started."
	case AuthError:
		return "Your API token may be invalid or expired.\nRun 'ldctl config init' to reconfigure."
	case NetworkError:
		if u, ok := e.Details["url"].(string); ok {
			return fmt.Sprintf("Failed to reach: %s\nCheck your network connection and instance URL.", u)
		}
		return "Check your network connection and instance URL."
	default:
		return ""
	}
}

// PrintHuman writes a human-readable error to w in the standard format:
//
//	Error: <message>
//	<optional hint>
func PrintHuman(w io.Writer, e *Error) {
	fmt.Fprintf(w, "Error: %s\n", e.Message)
	if hint := hintFor(e); hint != "" {
		fmt.Fprintf(w, "%s\n", hint)
	}
}
