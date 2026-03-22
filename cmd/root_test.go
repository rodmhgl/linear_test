package cmd_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rodmhgl/ldctl/cmd"
	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
)

func TestRoot_NoArgs_ShowsHelp(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	root := cmd.NewRootCmd("1.0.0")
	root.SetOut(&buf)
	root.SetErr(&buf)

	err := root.Execute()
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "ldctl")
	assert.Contains(t, out, "config")
	assert.Contains(t, out, "bookmarks")
	assert.Contains(t, out, "Global Flags")
	assert.Contains(t, out, "Examples")
}

func TestRoot_HelpFits80Cols(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	root := cmd.NewRootCmd("1.0.0")
	root.SetOut(&buf)
	root.SetErr(&buf)

	err := root.Execute()
	require.NoError(t, err)

	for i, line := range strings.Split(buf.String(), "\n") {
		if len(line) > 80 {
			t.Errorf("line %d exceeds 80 cols (%d chars): %q", i+1, len(line), line)
		}
	}
}

func TestRoot_GlobalFlags_Defined(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCmd("1.0.0")
	pf := root.PersistentFlags()

	for _, name := range []string{"json", "quiet", "verbose", "version"} {
		assert.NotNil(t, pf.Lookup(name), "expected persistent flag --%s to be defined", name)
	}
}

func TestRoot_MutualExclusivity(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCmd("1.0.0")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--quiet", "--verbose"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use")
}

func TestRoot_BookmarksAlias(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCmd("1.0.0")

	bm, _, err := root.Find([]string{"bookmarks"})
	require.NoError(t, err)
	require.NotNil(t, bm)
	assert.Contains(t, bm.Aliases, "bm")
}

func TestRoot_CompletionDisabled(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCmd("1.0.0")

	for _, sub := range root.Commands() {
		assert.NotEqual(t, "completion", sub.Name(),
			"completion command should be disabled")
	}
}

func TestRoot_VersionFlag(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	root := cmd.NewRootCmd("1.2.3")
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--version"})

	// Execute returns nil because errVersionRequested is swallowed by Execute().
	// But NewRootCmd returns the raw command — so we call Execute directly.
	err := root.Execute()
	// PersistentPreRunE returns errVersionRequested, which cobra propagates.
	// NewRootCmd doesn't swallow it — only cmd.Execute() does.
	// The test just checks the output contains the version.
	_ = err

	assert.Contains(t, buf.String(), "1.2.3")
}

func TestRoot_SilenceFlags(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCmd("1.0.0")
	assert.True(t, root.SilenceErrors)
	assert.True(t, root.SilenceUsage)
}

func TestRoot_UserBare_OK(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCmd("1.0.0")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"user"})

	err := root.Execute()
	require.NoError(t, err)
}

func TestRoot_ConfigBare_OK(t *testing.T) {
	t.Parallel()

	root := cmd.NewRootCmd("1.0.0")
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"config"})

	err := root.Execute()
	require.NoError(t, err)
}

func TestHandleError_LdctlConfigError_Returns2(t *testing.T) {
	t.Parallel()

	var errBuf bytes.Buffer
	e := ldcerr.NewConfigNotFound("/cfg/config.toml")

	code := cmd.HandleError(e, false, &errBuf)

	assert.Equal(t, 2, code)
	assert.Contains(t, errBuf.String(), "Error: no configuration found")
}

func TestHandleError_LdctlAPIError_Returns1(t *testing.T) {
	t.Parallel()

	var errBuf bytes.Buffer
	e := ldcerr.NewNotFound("bookmark", 99)

	code := cmd.HandleError(e, false, &errBuf)

	assert.Equal(t, 1, code)
	assert.Contains(t, errBuf.String(), "Error: bookmark not found")
}

func TestHandleError_LdctlError_JSONMode(t *testing.T) {
	t.Parallel()

	var errBuf bytes.Buffer
	e := ldcerr.NewAuthFailed("https://ld.example.com", 401)

	code := cmd.HandleError(e, true, &errBuf)

	assert.Equal(t, 2, code)
	assert.Contains(t, errBuf.String(), `"type"`)
	assert.Contains(t, errBuf.String(), `"auth_error"`)
}

func TestHandleError_PlainError_Returns1(t *testing.T) {
	t.Parallel()

	var errBuf bytes.Buffer
	code := cmd.HandleError(fmt.Errorf("unexpected"), false, &errBuf)

	assert.Equal(t, 1, code)
	assert.Contains(t, errBuf.String(), "Error: unexpected")
}
