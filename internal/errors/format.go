package errors

import "fmt"

// FormatHuman returns a human-readable multi-line error string.
// Format:
//
//	Error: <message>
//	<suggestion if present in details>
func FormatHuman(e *Error) string {
	out := fmt.Sprintf("Error: %s", e.Message)

	if e.Details != nil {
		if suggestion, ok := e.Details["suggestion"].(string); ok && suggestion != "" {
			out += "\n" + suggestion
		}
	}

	return out
}
