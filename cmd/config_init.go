package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	ldcconfig "github.com/rodmhgl/ldctl/internal/config"
	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
)

// configInitFlags holds the flags for the config init subcommand.
type configInitFlags struct {
	noVerify bool
	force    bool
}

// configInitOptions bundles all dependencies for config init so they can be
// swapped in tests without touching global state.
type configInitOptions struct {
	flags    *configInitFlags
	rootFlags *rootFlags
	// promptURL reads a URL from the user; defaults to readLineFromStdin.
	promptURL func(prompt string, r io.Reader) (string, error)
	// promptToken reads a masked token from the user; defaults to readPasswordFromTerminal.
	promptToken func(prompt string) (string, error)
	// validateFn calls the LinkDing API to confirm credentials.
	validateFn func(rawURL, token string) error
	// writeFn writes the config file.
	writeFn func(path, rawURL, token string) error
	// stdin is used for non-terminal prompt fallback.
	stdin io.Reader
	// stdout is used for output.
	stdout io.Writer
}

func newConfigInitCmd(rflags *rootFlags) *cobra.Command {
	iflags := &configInitFlags{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize ldctl configuration",
		Long: `Initialize the ldctl configuration file with your LinkDing URL and API token.

The command prompts for your LinkDing instance URL and API token, validates the
credentials against the API, then writes ~/.config/ldctl/config.toml (or
$XDG_CONFIG_HOME/ldctl/config.toml on Linux/macOS, %APPDATA%\ldctl\config.toml
on Windows).

Non-interactive mode: when both LINKDING_URL and LINKDING_TOKEN environment
variables are set the prompts are skipped (validation still runs unless
--no-verify is given).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			opts := &configInitOptions{
				flags:       iflags,
				rootFlags:   rflags,
				promptURL:   readLineFromReader,
				promptToken: readPasswordFromTerminal,
				validateFn:  validateCredentials,
				writeFn:     writeConfigFile,
				stdin:       cmd.InOrStdin(),
				stdout:      cmd.OutOrStdout(),
			}
			return runConfigInit(cmd, opts)
		},
	}

	cmd.Flags().BoolVar(&iflags.noVerify, "no-verify", false, "Skip API validation and write config immediately")
	cmd.Flags().BoolVar(&iflags.force, "force", false, "Overwrite existing config file without prompting")

	return cmd
}

// runConfigInit is the testable core of the config init command.
func runConfigInit(cmd *cobra.Command, opts *configInitOptions) error {
	instanceURL, token, err := collectCredentials(cmd, opts)
	if err != nil {
		return err
	}

	// Normalize the URL.
	instanceURL, err = normalizeURL(instanceURL)
	if err != nil {
		return err
	}

	// Validate credentials unless --no-verify.
	if !opts.flags.noVerify {
		if err := opts.validateFn(instanceURL, token); err != nil {
			return err
		}
	}

	// Determine config file path.
	cfgPath, err := ldcconfig.ConfigPath()
	if err != nil {
		return err
	}

	// Refuse to overwrite unless --force.
	if _, statErr := os.Stat(cfgPath); statErr == nil && !opts.flags.force {
		return ldcerr.Newf(ldcerr.ConfigError,
			"config file already exists at %s; use --force to overwrite", cfgPath)
	}

	// Write the config file.
	if err := opts.writeFn(cfgPath, instanceURL, token); err != nil {
		return err
	}

	if !opts.rootFlags.quiet {
		fmt.Fprintf(opts.stdout, "Configuration written to %s\n", cfgPath)
	}

	return nil
}

// collectCredentials gets URL and token either from env vars or by prompting.
func collectCredentials(cmd *cobra.Command, opts *configInitOptions) (string, string, error) {
	envURL := os.Getenv("LINKDING_URL")
	envToken := os.Getenv("LINKDING_TOKEN")

	// Non-interactive mode: both env vars set.
	if envURL != "" && envToken != "" {
		if !opts.rootFlags.quiet {
			fmt.Fprintln(opts.stdout, "Using credentials from environment variables LINKDING_URL and LINKDING_TOKEN.")
		}
		return envURL, envToken, nil
	}

	// Interactive mode (or partial env — fall through to full prompts).
	_ = cmd // cmd available for future use

	fmt.Fprint(opts.stdout, "LinkDing instance URL: ")
	instanceURL, err := opts.promptURL("", opts.stdin)
	if err != nil {
		return "", "", ldcerr.Newf(ldcerr.IOError, "failed to read URL: %v", err)
	}
	instanceURL = strings.TrimSpace(instanceURL)
	if instanceURL == "" {
		return "", "", ldcerr.New(ldcerr.ValidationError, "URL cannot be empty")
	}

	fmt.Fprint(opts.stdout, "API token: ")
	token, err := opts.promptToken("")
	if err != nil {
		return "", "", ldcerr.Newf(ldcerr.IOError, "failed to read token: %v", err)
	}
	// Print newline after the masked password input.
	fmt.Fprintln(opts.stdout)

	token = strings.TrimSpace(token)
	if token == "" {
		return "", "", ldcerr.New(ldcerr.ValidationError, "token cannot be empty")
	}

	return instanceURL, token, nil
}

// normalizeURL strips trailing slashes and prepends https:// if no scheme is
// present. Returns a validation error for malformed URLs.
func normalizeURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)

	// Prepend https:// if no scheme provided.
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", ldcerr.Newf(ldcerr.ValidationError, "invalid URL %q: %v", rawURL, err)
	}

	// Must have a valid host.
	if parsed.Host == "" {
		return "", ldcerr.Newf(ldcerr.ValidationError, "invalid URL %q: no host", rawURL)
	}

	// Only allow http and https schemes.
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ldcerr.Newf(ldcerr.ValidationError,
			"invalid URL %q: scheme must be http or https", rawURL)
	}

	// Strip trailing slash from path.
	parsed.Path = strings.TrimRight(parsed.Path, "/")

	return parsed.String(), nil
}

// validateCredentials calls GET /api/user/profile/ with the given token.
func validateCredentials(rawURL, token string) error {
	profileURL := strings.TrimRight(rawURL, "/") + "/api/user/profile/"

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, profileURL, nil)
	if err != nil {
		return ldcerr.Newf(ldcerr.NetworkError, "failed to build validation request: %v", err)
	}
	req.Header.Set("Authorization", "Token "+token)

	resp, err := client.Do(req)
	if err != nil {
		return &ldcerr.Error{
			Type:    ldcerr.NetworkError,
			Message: fmt.Sprintf("failed to reach %s: %v", rawURL, err),
			Details: map[string]interface{}{"url": rawURL},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ldcerr.MapHTTPError(resp)
	}

	return nil
}

// writeConfigFile creates the directory if needed and writes url+token to
// the TOML config file at path with 0600 permissions.
func writeConfigFile(path, rawURL, token string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ldcerr.Newf(ldcerr.IOError, "failed to create config directory %s: %v", dir, err)
	}

	// Open/create file with 0600 permissions (no-op for Windows enforcement).
	var perm os.FileMode = 0o600
	if runtime.GOOS == "windows" {
		perm = 0o666 // Windows does not enforce Unix permissions.
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return ldcerr.Newf(ldcerr.IOError, "failed to open config file %s: %v", path, err)
	}
	defer f.Close()

	cfg := struct {
		URL   string `toml:"url"`
		Token string `toml:"token"`
	}{
		URL:   rawURL,
		Token: token,
	}

	enc := toml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		return ldcerr.Newf(ldcerr.IOError, "failed to write config file: %v", err)
	}

	return nil
}

// readLineFromReader reads a single line from r (used for URL prompts).
func readLineFromReader(prompt string, r io.Reader) (string, error) {
	_ = prompt // prompt already printed by caller
	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("unexpected EOF")
}

// readPasswordFromTerminal reads a masked password from the real terminal.
// Falls back to reading from stdin when not a terminal (e.g. in tests).
func readPasswordFromTerminal(_ string) (string, error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	// Fallback: read a plain line from stdin (useful in pipes/tests).
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("unexpected EOF reading token")
}
