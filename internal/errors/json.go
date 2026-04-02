package errors

import "encoding/json"

// jsonEnvelope is the top-level JSON error structure written to stderr.
type jsonEnvelope struct {
	Error jsonError `json:"error"`
}

// jsonError is the inner error object within the JSON envelope.
type jsonError struct {
	Type    Type                   `json:"type"`
	Message string                 `json:"message"`
	Details map[string]any `json:"details"`
}

// FormatJSON serializes the error as a JSON string matching the schema:
//
//	{"error":{"type":"...","message":"...","details":{...}}}
//
// Returns the JSON bytes or a fallback JSON string if marshaling fails.
func FormatJSON(e *Error) string {
	details := e.Details
	if details == nil {
		details = map[string]any{}
	}

	envelope := jsonEnvelope{
		Error: jsonError{
			Type:    e.Type,
			Message: e.Message,
			Details: details,
		},
	}

	b, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		// Fallback: produce minimal valid JSON even if details can't be serialized.
		return `{"error":{"type":"` + string(e.Type) + `","message":"` + e.Message + `","details":{}}}`
	}
	return string(b)
}
