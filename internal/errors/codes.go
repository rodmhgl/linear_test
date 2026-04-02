// Package errors defines typed errors with exit code mapping for the ldctl CLI.
package errors

// Exit codes for the ldctl CLI.
const (
	// ExitSuccess indicates the command completed successfully.
	ExitSuccess = 0
	// ExitError indicates a general operational error (API, validation, I/O, not found).
	ExitError = 1
	// ExitConfigError indicates a configuration, authentication, or connectivity error.
	ExitConfigError = 2
)
