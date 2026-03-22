package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- normalizeURL tests ---

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{
			name:  "adds https scheme",
			input: "example.com",
			want:  "https://example.com",
		},
		{
			name:  "strips trailing slash",
			input: "https://example.com/",
			want:  "https://example.com",
		},
		{
			name:  "strips multiple trailing slashes",
			input: "https://example.com///",
			want:  "https://example.com",
		},
		{
			name:  "preserves http scheme",
			input: "http://example.com",
			want:  "http://example.com",
		},
		{
			name:  "preserves path without trailing slash",
			input: "https://example.com/linkding",
			want:  "https://example.com/linkding",
		},
		{
			name:  "strips trailing slash with path",
			input: "https://example.com/linkding/",
			want:  "https://example.com/linkding",
		},
		{
			name:  "strips trailing slash, adds scheme",
			input: "example.com/",
			want:  "https://example.com",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			errMsg:  "no host",
		},
		{
			name:    "invalid scheme",
			input:   "ftp://example.com",
			wantErr: true,
			errMsg:  "scheme must be http or https",
		},
		{
			name:  "url with port",
			input: "https://example.com:9090",
			want:  "https://example.com:9090",
		},
		{
			name:  "url with port, no scheme",
			input: "example.com:9090",
			want:  "https://example.com:9090",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeURL(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// --- validateCredentials tests ---

func TestValidateCredentials_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/user/profile/", r.URL.Path)
		assert.Equal(t, "Token mytoken", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"username":"alice"}`)
	}))
	defer srv.Close()

	err := validateCredentials(srv.URL, "mytoken")
	require.NoError(t, err)
}

func TestValidateCredentials_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	err := validateCredentials(srv.URL, "bad-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestValidateCredentials_NetworkError(t *testing.T) {
	err := validateCredentials("http://127.0.0.1:19999", "token")
	require.Error(t, err)
	// Should be a network error.
	assert.Contains(t, err.Error(), "127.0.0.1:19999")
}

// --- writeConfigFile tests ---

func TestWriteConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ldctl", "config.toml")

	err := writeConfigFile(path, "https://example.com", "mytoken123")
	require.NoError(t, err)

	// File must exist.
	info, err := os.Stat(path)
	require.NoError(t, err)

	// Check permissions on non-Windows.
	if runtime.GOOS != "windows" {
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}

	// Check contents.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, `url = "https://example.com"`)
	assert.Contains(t, content, `token = "mytoken123"`)
}

func TestWriteConfigFile_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	// deep nested path — directory must be auto-created.
	path := filepath.Join(dir, "a", "b", "c", "config.toml")

	err := writeConfigFile(path, "https://example.com", "tok")
	require.NoError(t, err)

	_, err = os.Stat(path)
	require.NoError(t, err)
}

// --- runConfigInit integration tests ---

// buildTestOpts creates a configInitOptions suitable for unit testing.
// validateFn succeeds by default; override fields as needed.
// It also redirects XDG_CONFIG_HOME (Linux/macOS) or APPDATA (Windows) to cfgDir
// so that ConfigPath() resolves under the temp directory.
func buildTestOpts(t *testing.T, cfgDir string, iflags *configInitFlags, rflags *rootFlags) (*configInitOptions, *bytes.Buffer) {
	t.Helper()

	// Redirect the config home so ConfigPath() points inside cfgDir.
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", cfgDir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", cfgDir)
	}

	out := &bytes.Buffer{}

	opts := &configInitOptions{
		flags:     iflags,
		rootFlags: rflags,
		stdout:    out,
		// Successful no-op validator.
		validateFn: func(_, _ string) error { return nil },
		// writeFn defaults to the real implementation; callers may override.
		writeFn: writeConfigFile,
		// promptURL returns a canned URL.
		promptURL: func(_ string, _ io.Reader) (string, error) {
			return "https://ld.example.com", nil
		},
		// promptToken returns a canned token.
		promptToken: func(_ string) (string, error) {
			return "secrettoken", nil
		},
		stdin: strings.NewReader(""),
	}
	return opts, out
}

func TestRunConfigInit_Interactive(t *testing.T) {
	dir := t.TempDir()
	iflags := &configInitFlags{noVerify: false, force: false}
	rflags := &rootFlags{}

	opts, out := buildTestOpts(t, dir, iflags, rflags)

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Configuration written to")

	// Confirm the file was written at the expected path.
	cfgPath := filepath.Join(dir, "ldctl", "config.toml")
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "https://ld.example.com")
	assert.Contains(t, string(data), "secrettoken")
}

func TestRunConfigInit_NoVerify(t *testing.T) {
	dir := t.TempDir()
	iflags := &configInitFlags{noVerify: true, force: false}
	rflags := &rootFlags{}

	opts, _ := buildTestOpts(t, dir, iflags, rflags)
	validateCalled := false
	realValidate := opts.validateFn
	opts.validateFn = func(u, tok string) error {
		validateCalled = true
		return realValidate(u, tok)
	}

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.NoError(t, err)
	assert.False(t, validateCalled, "validateFn should not be called with --no-verify")
}

func TestRunConfigInit_ForceOverwrite(t *testing.T) {
	dir := t.TempDir()

	// buildTestOpts sets XDG_CONFIG_HOME=dir, so ConfigPath() → dir/ldctl/config.toml
	iflags := &configInitFlags{noVerify: true, force: true}
	rflags := &rootFlags{}
	opts, _ := buildTestOpts(t, dir, iflags, rflags)

	// Pre-create the config file at the resolved path.
	cfgPath := filepath.Join(dir, "ldctl", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(cfgPath), 0o755))
	require.NoError(t, os.WriteFile(cfgPath, []byte("old content"), 0o600))

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.NoError(t, err)

	// File should have been overwritten.
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	assert.NotEqual(t, "old content", string(data))
}

func TestRunConfigInit_RefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	iflags := &configInitFlags{noVerify: true, force: false}
	rflags := &rootFlags{}

	// buildTestOpts sets XDG_CONFIG_HOME=dir → ConfigPath() = dir/ldctl/config.toml.
	opts, _ := buildTestOpts(t, dir, iflags, rflags)

	// Pre-create the config file at the resolved path.
	cfgPath := filepath.Join(dir, "ldctl", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(cfgPath), 0o755))
	require.NoError(t, os.WriteFile(cfgPath, []byte("existing"), 0o600))

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRunConfigInit_EnvVarNonInteractive(t *testing.T) {
	t.Setenv("LINKDING_URL", "https://env.example.com/")
	t.Setenv("LINKDING_TOKEN", "envtoken")

	dir := t.TempDir()
	iflags := &configInitFlags{noVerify: true, force: false}
	rflags := &rootFlags{}
	opts, out := buildTestOpts(t, dir, iflags, rflags)

	promptCalled := false
	opts.promptURL = func(_ string, _ io.Reader) (string, error) {
		promptCalled = true
		return "", nil
	}
	opts.promptToken = func(_ string) (string, error) {
		promptCalled = true
		return "", nil
	}

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.NoError(t, err)

	assert.False(t, promptCalled, "should not prompt when both env vars set")
	assert.Contains(t, out.String(), "environment variables")
}

func TestRunConfigInit_EnvVarPartialFallsThrough(t *testing.T) {
	// Only one env var set → should still prompt.
	t.Setenv("LINKDING_URL", "https://env.example.com")
	t.Setenv("LINKDING_TOKEN", "") // explicitly empty

	dir := t.TempDir()
	iflags := &configInitFlags{noVerify: true, force: false}
	rflags := &rootFlags{}
	opts, _ := buildTestOpts(t, dir, iflags, rflags)

	promptCalled := false
	opts.promptURL = func(_ string, _ io.Reader) (string, error) {
		promptCalled = true
		return "https://prompt.example.com", nil
	}

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.NoError(t, err)

	assert.True(t, promptCalled, "should prompt when only one env var set")
}

func TestRunConfigInit_ValidationFailureDoesNotWriteFile(t *testing.T) {
	dir := t.TempDir()
	iflags := &configInitFlags{noVerify: false, force: false}
	rflags := &rootFlags{}
	opts, _ := buildTestOpts(t, dir, iflags, rflags)

	writeFileCalled := false
	opts.writeFn = func(_, _, _ string) error {
		writeFileCalled = true
		return nil
	}
	opts.validateFn = func(_, _ string) error {
		return fmt.Errorf("authentication failed")
	}

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.Error(t, err)
	assert.False(t, writeFileCalled, "config file must not be written on validation failure")
}

func TestRunConfigInit_EmptyURLRejected(t *testing.T) {
	dir := t.TempDir()
	iflags := &configInitFlags{noVerify: true}
	rflags := &rootFlags{}
	opts, _ := buildTestOpts(t, dir, iflags, rflags)
	opts.promptURL = func(_ string, _ io.Reader) (string, error) {
		return "", nil // empty
	}

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL cannot be empty")
}

func TestRunConfigInit_EmptyTokenRejected(t *testing.T) {
	dir := t.TempDir()
	iflags := &configInitFlags{noVerify: true}
	rflags := &rootFlags{}
	opts, _ := buildTestOpts(t, dir, iflags, rflags)
	opts.promptToken = func(_ string) (string, error) {
		return "", nil // empty
	}

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token cannot be empty")
}

func TestRunConfigInit_QuietMode(t *testing.T) {
	// Use env vars so collectCredentials skips prompts entirely.
	t.Setenv("LINKDING_URL", "https://quiet.example.com")
	t.Setenv("LINKDING_TOKEN", "quiettoken")

	dir := t.TempDir()
	iflags := &configInitFlags{noVerify: true, force: false}
	rflags := &rootFlags{quiet: true}
	opts, out := buildTestOpts(t, dir, iflags, rflags)

	root := NewRootCmd("test")
	err := runConfigInit(root, opts)
	require.NoError(t, err)
	assert.Empty(t, out.String(), "quiet mode should produce no output on success")
}

// ldcConfigPathForTest returns the expected config path relative to the given base.
// This mirrors the logic in ldcconfig.ConfigPath().
func ldcConfigPathForTest(base string) (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Join(base, "ldctl", "config.toml"), nil
	}
	return filepath.Join(base, "ldctl", "config.toml"), nil
}
