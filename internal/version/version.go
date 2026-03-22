// Package version provides build-time version information for ldctl.
package version

import (
	"fmt"
	"runtime"
)

// Version, Commit, and BuildDate are set via -ldflags at build time.
//
//nolint:gochecknoglobals // set via -ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Info holds structured version information.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns a populated Info struct.
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a human-readable version string.
func String() string {
	return fmt.Sprintf(
		"ldctl version %s (commit %s, built %s, %s)",
		Version, Commit, BuildDate, runtime.Version(),
	)
}
