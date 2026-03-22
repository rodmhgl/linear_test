package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rodmhgl/ldctl/cmd"
	"github.com/rodmhgl/ldctl/internal/version"
)

// executeVersionCmd runs the version subcommand with the supplied args and
// returns the captured stdout output and the command error (if any).
func executeVersionCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer
	root := cmd.NewRootCmd(version.Version)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(append([]string{"version"}, args...))

	err := root.Execute()
	return buf.String(), err
}

// TestVersion_Short_OutputIsSemverOnly verifies that --short prints only the
// version string followed by a single newline, with no other text.
func TestVersion_Short_OutputIsSemverOnly(t *testing.T) {
	t.Parallel()

	out, err := executeVersionCmd(t, "--short")

	require.NoError(t, err)

	// The output must be exactly "<version>\n" — one line, nothing else.
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	require.Len(t, lines, 1, "expected exactly one line of output, got: %q", out)
	assert.Equal(t, version.Version, lines[0])
}

// TestVersion_Short_NoBuildMetadata verifies that --short does not include
// commit hash, build date, or Go version in its output.
func TestVersion_Short_NoBuildMetadata(t *testing.T) {
	t.Parallel()

	out, err := executeVersionCmd(t, "--short")

	require.NoError(t, err)
	assert.NotContains(t, out, "commit", "--short output must not contain commit info")
	assert.NotContains(t, out, "built", "--short output must not contain build date")
	assert.NotContains(t, out, "go", "--short output must not contain Go version prefix text")
}

// TestVersion_Short_TrailingNewlineOnly ensures the output ends with exactly
// one newline — critical for shell variable capture via VERSION=$(ldctl version --short).
func TestVersion_Short_TrailingNewlineOnly(t *testing.T) {
	t.Parallel()

	out, err := executeVersionCmd(t, "--short")

	require.NoError(t, err)
	require.NotEmpty(t, out, "output must not be empty")

	// Must end with exactly one '\n'.
	assert.True(t, strings.HasSuffix(out, "\n"), "output must end with a newline")
	// Must NOT end with two consecutive newlines (no blank trailing line).
	assert.False(t, strings.HasSuffix(out, "\n\n"), "output must not end with two newlines")
}

// TestVersion_Short_CapturesCorrectly simulates VERSION=$(ldctl version --short)
// by stripping the trailing newline (as a shell would) and confirms the result
// equals the semver string with no whitespace.
func TestVersion_Short_CapturesCorrectly(t *testing.T) {
	t.Parallel()

	out, err := executeVersionCmd(t, "--short")

	require.NoError(t, err)

	// Shell command substitution strips trailing newlines.
	captured := strings.TrimRight(out, "\n")
	assert.Equal(t, version.Version, captured,
		"shell-captured value must equal the version string exactly")
	assert.Equal(t, strings.TrimSpace(captured), captured,
		"captured value must have no surrounding whitespace")
}

// TestVersion_Short_ExitCodeZero confirms the command exits with code 0.
// (No error return from RunE means exit 0 in cobra.)
func TestVersion_Short_ExitCodeZero(t *testing.T) {
	t.Parallel()

	_, err := executeVersionCmd(t, "--short")

	assert.NoError(t, err, "ldctl version --short must exit 0")
}

// TestVersion_NoFlags_HumanOutput verifies that the default (no flags) output
// contains the human-readable version string with all metadata fields.
func TestVersion_NoFlags_HumanOutput(t *testing.T) {
	t.Parallel()

	out, err := executeVersionCmd(t)

	require.NoError(t, err)
	assert.Contains(t, out, "ldctl version")
	assert.Contains(t, out, "commit")
	assert.Contains(t, out, "built")
}

// TestVersion_JSON_ContainsExpectedFields verifies that --json output is valid
// JSON containing the expected top-level keys.
func TestVersion_JSON_ContainsExpectedFields(t *testing.T) {
	t.Parallel()

	out, err := executeVersionCmd(t, "--json")

	require.NoError(t, err)
	assert.Contains(t, out, `"version"`)
	assert.Contains(t, out, `"commit"`)
	assert.Contains(t, out, `"buildDate"`)
	assert.Contains(t, out, `"goVersion"`)
}

// TestVersion_Short_Flag_Defined confirms the --short flag is registered on
// the version subcommand so CLI users can discover it via --help.
func TestVersion_Short_Flag_Defined(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCmd("0.0.0")
	versionCmd, _, err := root.Find([]string{"version"})

	require.NoError(t, err)
	require.NotNil(t, versionCmd)

	flag := versionCmd.Flags().Lookup("short")
	require.NotNil(t, flag, "expected --short flag to be defined on version command")
	assert.Equal(t, "false", flag.DefValue, "--short must default to false")
}

// TestVersion_Short_WithInjectedVersion verifies behavior when the version
// package variable is set to a realistic semver value (simulating an ldflag build).
func TestVersion_Short_WithInjectedVersion(t *testing.T) {
	// Not parallel: mutates package-level variable.
	orig := version.Version
	version.Version = "1.5.2"
	t.Cleanup(func() { version.Version = orig })

	var buf bytes.Buffer
	root := cmd.NewRootCmd(version.Version)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"version", "--short"})

	err := root.Execute()
	require.NoError(t, err)

	out := strings.TrimRight(buf.String(), "\n")
	assert.Equal(t, "1.5.2", out)
}
