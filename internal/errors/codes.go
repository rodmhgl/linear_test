// Package errors defines ldctl's typed errors, exit codes, and formatting.
package errors

// Exit code constants. Every command returns one of these three values.
const (
	ExitSuccess     = 0 // Command completed successfully.
	ExitError       = 1 // Operational error (API, validation, I/O, not-found).
	ExitConfigError = 2 // Configuration/auth/network error.
)
