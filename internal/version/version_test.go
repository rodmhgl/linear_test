package version_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rodmhgl/ldctl/internal/version"
)

// setVersion overrides the build-time package variables for the duration of a
// test and restores them via t.Cleanup so parallel tests are not affected.
func setVersion(t *testing.T, ver, commit, buildDate string) {
	t.Helper()

	origVersion := version.Version
	origCommit := version.Commit
	origBuildDate := version.BuildDate

	version.Version = ver
	version.Commit = commit
	version.BuildDate = buildDate

	t.Cleanup(func() {
		version.Version = origVersion
		version.Commit = origCommit
		version.BuildDate = origBuildDate
	})
}

func TestVersionString(t *testing.T) {
	setVersion(t, "1.2.3", "a1b2c3d", "2025-01-27T10:30:00Z")

	expected := fmt.Sprintf(
		"ldctl version 1.2.3 (commit a1b2c3d, built 2025-01-27T10:30:00Z, %s)",
		runtime.Version(),
	)
	assert.Equal(t, expected, version.String())
}

func TestVersionString_DevDefaults(t *testing.T) {
	setVersion(t, "dev", "unknown", "unknown")

	expected := fmt.Sprintf(
		"ldctl version dev (commit unknown, built unknown, %s)",
		runtime.Version(),
	)
	assert.Equal(t, expected, version.String())
}

func TestVersionGet(t *testing.T) {
	setVersion(t, "2.0.0", "deadbeef", "2026-01-01T00:00:00Z")

	info := version.Get()

	require.Equal(t, "2.0.0", info.Version)
	require.Equal(t, "deadbeef", info.Commit)
	require.Equal(t, "2026-01-01T00:00:00Z", info.BuildDate)
	assert.Equal(t, runtime.Version(), info.GoVersion)
	assert.Equal(t, runtime.GOOS, info.OS)
	assert.Equal(t, runtime.GOARCH, info.Arch)
}

func TestVersionGet_DefaultValues(t *testing.T) {
	setVersion(t, "dev", "unknown", "unknown")

	info := version.Get()

	assert.Equal(t, "dev", info.Version)
	assert.Equal(t, "unknown", info.Commit)
	assert.Equal(t, "unknown", info.BuildDate)
	// GoVersion, OS, and Arch are always populated from the runtime.
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.OS)
	assert.NotEmpty(t, info.Arch)
}

func TestVersionGet_AllFieldsPopulated(t *testing.T) {
	info := version.Get()

	// Regardless of build values, runtime fields must never be empty.
	assert.NotEmpty(t, info.GoVersion, "GoVersion must always be set")
	assert.NotEmpty(t, info.OS, "OS must always be set")
	assert.NotEmpty(t, info.Arch, "Arch must always be set")
}
