package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setConfigEnv overrides XDG_CONFIG_HOME (or USERPROFILE on Windows) so that
// config.ConfigPath() resolves to a temp dir we control, then restores it on
// cleanup.
func setConfigDir(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		old := os.Getenv("APPDATA")
		t.Setenv("APPDATA", dir)
		t.Cleanup(func() { os.Setenv("APPDATA", old) })
	} else {
		old := os.Getenv("XDG_CONFIG_HOME")
		t.Setenv("XDG_CONFIG_HOME", dir)
		t.Cleanup(func() { os.Setenv("XDG_CONFIG_HOME", old) })
	}
}

// prepareConfigDir creates the ldctl subdirectory inside a temp root and
// returns both the root and the full config file path.
func prepareConfigDir(t *testing.T) (root string, cfgFile string) {
	t.Helper()
	root = t.TempDir()
	ldctlDir := filepath.Join(root, "ldctl")
	require.NoError(t, os.MkdirAll(ldctlDir, 0o755))
	cfgFile = filepath.Join(ldctlDir, "config.toml")
	return root, cfgFile
}

// runConfigShow executes "ldctl config show [args...]" against a fresh root
// command and returns stdout, stderr, and the returned error.
func runConfigShowCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root := NewRootCmd("test")
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	root.SetArgs(append([]string{"config", "show"}, args...))
	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}

// runConfigBare executes "ldctl config" (bare, no subcommand).
func runConfigBare(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	root := NewRootCmd("test")
	root.SetOut(outBuf)
	root.SetErr(errBuf)
	root.SetArgs(append([]string{"config"}, args...))
	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}

// ---- maskToken unit tests ----

func TestMaskToken_LongToken(t *testing.T) {
	// Normal token: show first 3 and last 3.
	result := maskToken("abcdefghijk")
	assert.Equal(t, "abc...ijk", result)
}

func TestMaskToken_ExactlySixChars(t *testing.T) {
	// Exactly 6 chars → fully masked.
	result := maskToken("abcdef")
	assert.Equal(t, "***", result)
}

func TestMaskToken_FewChars(t *testing.T) {
	result := maskToken("abc")
	assert.Equal(t, "***", result)
}

func TestMaskToken_SevenChars(t *testing.T) {
	// 7 chars: first 3, last 3.
	result := maskToken("abcdefg")
	assert.Equal(t, "abc...efg", result)
}

func TestMaskToken_Empty(t *testing.T) {
	result := maskToken("")
	assert.Equal(t, "***", result)
}

// ---- config show integration tests ----

func TestConfigShow_FromFile(t *testing.T) {
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("url = \"https://ld.example.com\"\ntoken = \"myverylongtoken\"\n"),
		0o600,
	))

	stdout, stderr, err := runConfigShowCmd(t)
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "https://ld.example.com")
	assert.Contains(t, stdout, "myv...ken")
	assert.Contains(t, stdout, "(config file)")
}

func TestConfigShow_FromEnvVars(t *testing.T) {
	// No config file; both values from env.
	root, _ := prepareConfigDir(t)
	setConfigDir(t, root)
	t.Setenv("LINKDING_URL", "https://env.example.com")
	t.Setenv("LINKDING_TOKEN", "envtoken1234567")

	stdout, _, err := runConfigShowCmd(t)
	require.NoError(t, err)
	assert.Contains(t, stdout, "https://env.example.com")
	assert.Contains(t, stdout, "(env: LINKDING_URL)")
	assert.Contains(t, stdout, "(env: LINKDING_TOKEN)")
	assert.Contains(t, stdout, "env...567")
}

func TestConfigShow_EnvOverridesFile(t *testing.T) {
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("url = \"https://file.example.com\"\ntoken = \"filetoken123456\"\n"),
		0o600,
	))
	t.Setenv("LINKDING_URL", "https://env-override.example.com")

	stdout, _, err := runConfigShowCmd(t)
	require.NoError(t, err)
	// URL should come from env, token from file.
	assert.Contains(t, stdout, "https://env-override.example.com")
	assert.Contains(t, stdout, "(env: LINKDING_URL)")
	assert.Contains(t, stdout, "(config file)")
}

func TestConfigShow_NoConfig(t *testing.T) {
	root, _ := prepareConfigDir(t)
	setConfigDir(t, root)
	// Ensure no env vars are set.
	t.Setenv("LINKDING_URL", "")
	t.Setenv("LINKDING_TOKEN", "")

	_, stderr, err := runConfigShowCmd(t)
	require.Error(t, err)
	assert.Contains(t, stderr, "No configuration found")
	assert.Contains(t, stderr, "ldctl config init")
}

func TestConfigShow_JSONOutput(t *testing.T) {
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("url = \"https://json.example.com\"\ntoken = \"jsontoken12345\"\n"),
		0o600,
	))

	stdout, _, err := runConfigShowCmd(t, "--json")
	require.NoError(t, err)

	var out configShowJSON
	require.NoError(t, json.Unmarshal([]byte(stdout), &out))

	assert.Equal(t, "https://json.example.com", out.URL.Value)
	assert.Equal(t, "config file", out.URL.Source)
	// Token should be masked in JSON too.
	assert.Equal(t, "jso...345", out.Token.Value)
	assert.Equal(t, "config file", out.Token.Source)
	// Must NOT contain the raw token.
	assert.NotContains(t, stdout, "jsontoken12345")
}

func TestConfigShow_CorruptConfig(t *testing.T) {
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("this is [ not valid toml !!!"),
		0o600,
	))

	_, _, err := runConfigShowCmd(t)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "corrupt")
}

func TestConfigShow_MissingURLField(t *testing.T) {
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	// Token present but no URL.
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("token = \"tokenonly12345\"\n"),
		0o600,
	))
	t.Setenv("LINKDING_URL", "")

	_, _, err := runConfigShowCmd(t)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "url")
}

func TestConfigShow_MissingTokenField(t *testing.T) {
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	// URL present but no token.
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("url = \"https://notoken.example.com\"\n"),
		0o600,
	))
	t.Setenv("LINKDING_TOKEN", "")

	_, _, err := runConfigShowCmd(t)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}

func TestConfigBare_AliasesToShow(t *testing.T) {
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("url = \"https://bare.example.com\"\ntoken = \"baretokenvalue\"\n"),
		0o600,
	))

	stdout, _, err := runConfigBare(t)
	require.NoError(t, err)
	assert.Contains(t, stdout, "https://bare.example.com")
	assert.Contains(t, stdout, "bar...lue")
}

func TestConfigShow_PermissionWarning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission check not applicable on Windows")
	}
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("url = \"https://perm.example.com\"\ntoken = \"permtoken12345\"\n"),
		0o644, // too permissive
	))

	_, stderr, err := runConfigShowCmd(t)
	require.NoError(t, err)
	assert.Contains(t, stderr, "overly permissive")
	assert.Contains(t, stderr, "chmod 600")
	// Should also still show the config.
	// (stdout not captured here but err == nil confirms it ran successfully)
}

func TestConfigShow_TokenMaskedInOutput(t *testing.T) {
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	secretToken := "supersecrettoken999"
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("url = \"https://secret.example.com\"\ntoken = \""+secretToken+"\"\n"),
		0o600,
	))

	stdout, _, err := runConfigShowCmd(t)
	require.NoError(t, err)
	// Raw token must NOT appear in output.
	assert.NotContains(t, stdout, secretToken)
	// Masked form should appear.
	assert.Contains(t, stdout, "sup...999")
}

func TestConfigShow_JSONTokenMaskedAndNotExposed(t *testing.T) {
	root, cfgFile := prepareConfigDir(t)
	setConfigDir(t, root)
	secretToken := "topsecrettoken777"
	require.NoError(t, os.WriteFile(cfgFile,
		[]byte("url = \"https://topsecret.example.com\"\ntoken = \""+secretToken+"\"\n"),
		0o600,
	))

	stdout, _, err := runConfigShowCmd(t, "--json")
	require.NoError(t, err)
	assert.NotContains(t, stdout, secretToken)
	assert.True(t, strings.Contains(stdout, "top...777"))
}
