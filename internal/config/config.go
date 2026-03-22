// Package config handles loading and validating ldctl configuration.
// Configuration may come from a TOML file (~/.config/ldctl/config.toml)
// or from environment variables (LINKDING_URL, LINKDING_TOKEN).
// Environment variables take precedence over the config file.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"

	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
)

// Config holds the effective configuration values.
type Config struct {
	URL   string
	Token string
}

// Source records where each configuration value came from.
// Possible values: "config file", "env: LINKDING_URL", "env: LINKDING_TOKEN", or "not set".
type Source struct {
	URL   string
	Token string
}

// LoadResult combines the effective Config and its Sources.
type LoadResult struct {
	Config Config
	Source Source
}

// tomlConfig mirrors the TOML file layout for decoding.
type tomlConfig struct {
	URL   string `toml:"url"`
	Token string `toml:"token"`
}

// ConfigPath returns the absolute path to the ldctl config file.
// On all platforms it uses os.UserConfigDir() as the base.
func ConfigPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", ldcerr.Newf(ldcerr.IOError, "cannot determine config directory: %v", err)
	}
	return filepath.Join(base, "ldctl", "config.toml"), nil
}

// Load reads configuration from the TOML config file and/or environment
// variables, with environment variables taking precedence.
//
// Error cases:
//   - TOML parse error      → ConfigError "Config file is corrupt…"
//   - Missing required field → ConfigError "Config missing required field: <field>"
//   - No config anywhere    → ConfigError via NewConfigNotFound
func Load() (*LoadResult, error) {
	result := &LoadResult{
		Source: Source{
			URL:   "not set",
			Token: "not set",
		},
	}

	// Attempt to load config file first.
	cfgPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	fileExists := false
	if _, statErr := os.Stat(cfgPath); statErr == nil {
		fileExists = true
	}

	if fileExists {
		if err := loadFromFile(cfgPath, result); err != nil {
			return nil, err
		}
	}

	// Environment variables override file values.
	applyEnv(result)

	// Validate that we have at least something usable.
	if result.Config.URL == "" && result.Config.Token == "" {
		return nil, ldcerr.NewConfigNotFound(cfgPath)
	}
	if result.Config.URL == "" {
		return nil, ldcerr.New(ldcerr.ConfigError, "Config missing required field: url")
	}
	if result.Config.Token == "" {
		return nil, ldcerr.New(ldcerr.ConfigError, "Config missing required field: token")
	}

	return result, nil
}

// loadFromFile decodes the TOML config file and populates result.
func loadFromFile(path string, result *LoadResult) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return ldcerr.Newf(ldcerr.IOError, "cannot read config file: %v", err)
	}

	var tc tomlConfig
	if _, err := toml.Decode(string(data), &tc); err != nil {
		return ldcerr.New(
			ldcerr.ConfigError,
			"Config file is corrupt. Run 'ldctl config init --force' to recreate.",
		)
	}

	if tc.URL != "" {
		result.Config.URL = tc.URL
		result.Source.URL = "config file"
	}
	if tc.Token != "" {
		result.Config.Token = tc.Token
		result.Source.Token = "config file"
	}

	return nil
}

// applyEnv checks LINKDING_URL and LINKDING_TOKEN environment variables
// and applies them on top of whatever was loaded from the config file.
func applyEnv(result *LoadResult) {
	if v := os.Getenv("LINKDING_URL"); v != "" {
		result.Config.URL = v
		result.Source.URL = "env: LINKDING_URL"
	}
	if v := os.Getenv("LINKDING_TOKEN"); v != "" {
		result.Config.Token = v
		result.Source.Token = "env: LINKDING_TOKEN"
	}
}

// FilePermissionsOK returns true when the config file either does not exist or
// has permissions of exactly 0600. On Windows this always returns true because
// POSIX-style file mode bits are not meaningful there.
func FilePermissionsOK(path string) (bool, error) {
	if runtime.GOOS == "windows" {
		return true, nil
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, fmt.Errorf("cannot stat config file: %w", err)
	}
	return info.Mode().Perm() == 0o600, nil
}
