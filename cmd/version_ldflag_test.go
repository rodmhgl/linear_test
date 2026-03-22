//go:build integration

package cmd_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildBinary compiles the ldctl binary with the supplied ldflags into a
// temporary directory and returns the path to the resulting executable.
// The binary is automatically removed when the test completes.
func buildBinary(t *testing.T, ldflags string) string {
	t.Helper()

	dir := t.TempDir()
	binaryName := "ldctl"
	if runtime.GOOS == "windows" {
		binaryName = "ldctl.exe"
	}
	binaryPath := filepath.Join(dir, binaryName)

	// Find the module root (two levels up from cmd/).
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	moduleRoot := filepath.Join(filepath.Dir(thisFile), "..")

	args := []string{"build"}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	args = append(args, "-o", binaryPath, ".")

	build := exec.Command("go", args...)
	build.Dir = moduleRoot
	out, err := build.CombinedOutput()
	require.NoError(t, err, "go build failed:\n%s", string(out))

	return binaryPath
}

// runBinary executes the compiled binary with the given arguments and returns
// combined stdout+stderr output and the command error.
func runBinary(t *testing.T, binaryPath string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// TestIntegration_LdflagsInjected_Version verifies that the version string
// injected via -ldflags at build time appears in `ldctl version` output.
func TestIntegration_LdflagsInjected_Version(t *testing.T) {
	const (
		wantVersion = "9.8.7"
		wantCommit  = "cafebabe"
		wantDate    = "2026-01-02T03:04:05Z"
		module      = "github.com/rodmhgl/ldctl"
	)

	ldflags := strings.Join([]string{
		"-X " + module + "/internal/version.Version=" + wantVersion,
		"-X " + module + "/internal/version.Commit=" + wantCommit,
		"-X " + module + "/internal/version.BuildDate=" + wantDate,
	}, " ")

	binary := buildBinary(t, ldflags)

	t.Run("human output contains injected values", func(t *testing.T) {
		out, err := runBinary(t, binary, "version")
		require.NoError(t, err)
		assert.Contains(t, out, wantVersion)
		assert.Contains(t, out, wantCommit)
		assert.Contains(t, out, wantDate)
	})

	t.Run("short flag outputs only semver", func(t *testing.T) {
		out, err := runBinary(t, binary, "version", "--short")
		require.NoError(t, err)
		trimmed := strings.TrimRight(out, "\n")
		assert.Equal(t, wantVersion, trimmed)
		// No extra lines.
		assert.NotContains(t, trimmed, "\n")
	})

	t.Run("json output contains all injected fields", func(t *testing.T) {
		out, err := runBinary(t, binary, "version", "--json")
		require.NoError(t, err)

		var info struct {
			Version   string `json:"version"`
			Commit    string `json:"commit"`
			BuildDate string `json:"buildDate"`
			GoVersion string `json:"goVersion"`
			OS        string `json:"os"`
			Arch      string `json:"arch"`
		}
		require.NoError(t, json.Unmarshal([]byte(out), &info), "output must be valid JSON: %s", out)

		assert.Equal(t, wantVersion, info.Version)
		assert.Equal(t, wantCommit, info.Commit)
		assert.Equal(t, wantDate, info.BuildDate)
		assert.NotEmpty(t, info.GoVersion)
		assert.NotEmpty(t, info.OS)
		assert.NotEmpty(t, info.Arch)
	})
}

// TestIntegration_LdflagsNotSet_DefaultsApply verifies that when no ldflags
// are injected, the binary falls back to the "dev" / "unknown" defaults.
func TestIntegration_LdflagsNotSet_DefaultsApply(t *testing.T) {
	binary := buildBinary(t, "")

	out, err := runBinary(t, binary, "version")
	require.NoError(t, err)
	assert.Contains(t, out, "dev")

	shortOut, err := runBinary(t, binary, "version", "--short")
	require.NoError(t, err)
	assert.Equal(t, "dev\n", shortOut)
}

// TestIntegration_ExitCode verifies that version sub-commands exit 0.
func TestIntegration_ExitCode(t *testing.T) {
	binary := buildBinary(t, "")

	for _, args := range [][]string{
		{"version"},
		{"version", "--short"},
		{"version", "--json"},
	} {
		args := args
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := exec.Command(binary, args...)
			err := cmd.Run()
			assert.NoError(t, err, "command %v must exit 0", args)
		})
	}
}

// TestIntegration_ShortCapture simulates VERSION=$(ldctl version --short) and
// confirms the shell-captured value is clean.
func TestIntegration_ShortCapture(t *testing.T) {
	const (
		wantVersion = "3.2.1"
		module      = "github.com/rodmhgl/ldctl"
	)

	ldflags := "-X " + module + "/internal/version.Version=" + wantVersion
	binary := buildBinary(t, ldflags)

	out, err := runBinary(t, binary, "version", "--short")
	require.NoError(t, err)

	// Simulate shell command substitution: strip trailing newlines.
	captured := strings.TrimRight(out, "\n")
	assert.Equal(t, wantVersion, captured)
	assert.Equal(t, strings.TrimSpace(captured), captured, "no surrounding whitespace")

	// Ensure the binary is executable and findable (smoke test for PATH usage).
	_, statErr := os.Stat(binary)
	require.NoError(t, statErr)
}
