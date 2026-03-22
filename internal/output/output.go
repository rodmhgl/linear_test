// Package output provides helpers for writing consistent CLI output.
// It supports human-readable text, JSON, progress messages, and verbose/debug
// logging. Callers pass an [io.Writer] explicitly so output can be redirected
// in tests without patching global state.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// PrintJSON writes v as pretty-printed JSON (2-space indent) to w.
// Keys use snake_case when the supplied value uses json struct tags.
// Null values are rendered as null, empty slices as [].
// It returns any marshalling or write error.
func PrintJSON(w io.Writer, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("output: json marshal: %w", err)
	}
	b = append(b, '\n')
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("output: write: %w", err)
	}
	return nil
}

// PrintData writes human-readable formatted data to w.
// It behaves like [fmt.Fprintf] and discards write errors because CLI output
// errors are non-actionable at the call site; use a capturing writer in tests.
func PrintData(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, format, args...)
}

// PrintError writes a formatted error message to w (typically [os.Stderr]).
// The message is prefixed with "Error: " to be consistent with cobra's own
// error output style.
func PrintError(w io.Writer, format string, args ...any) {
	fmt.Fprintf(w, "Error: "+format, args...)
}

// PrintProgress writes a formatted progress/status message to w.
// When quiet is true the call is a no-op so callers never need to branch.
func PrintProgress(w io.Writer, quiet bool, format string, args ...any) {
	if quiet {
		return
	}
	fmt.Fprintf(w, format, args...)
}

// PrintVerbose writes a debug message prefixed with "[DEBUG] " to w.
// When verbose is false the call is a no-op.
// The caller is responsible for passing [os.Stderr] when appropriate; the
// function itself does not hard-code a destination so it remains testable.
func PrintVerbose(w io.Writer, verbose bool, format string, args ...any) {
	if !verbose {
		return
	}
	fmt.Fprintf(w, "[DEBUG] "+format, args...)
}
