// Package config handles loading and locating ldctl configuration.
package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"

	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
)

// Config holds the loaded configuration.
type Config struct {
	URL   string `toml:"url"`
	Token string `toml:"token"`
}

// Source records where each config value came from.
type Source struct {
	URL   string // "config file", "env: LINKDING_URL", or "not set"
	Token string // "config file", "env: LINKDING_TOKEN", or "not set"
}

// LoadResult bundles Config and Source together.
type LoadResult struct {
	Config Config
	Source Source
}

// tomlConfig is the on-disk representation.
type tomlConfig struct {
	URL   string `toml:"url"`
	Token string `toml:"token"`
}

// ConfigPath returns the platform-appropriate config file path.
//
//   - Linux/macOS: $XDG_CONFIG_HOME/ldctl/config.toml or ~/.config/ldctl/config.toml
//   - Windows:     %APPDATA%\ldctl\config.toml
func ConfigPath() (string, error) {
	var base string

	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", ldcerr.Newf(ldcerr.ConfigError, "cannot determine config directory: %v", err)
			}
			appdata = home
		}
		base = appdata
	} else {
		xdg := os.Getenv("XDG_CONFIG_HOME")
		if xdg != "" {
			base = xdg
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", ldcerr.Newf(ldcerr.ConfigError, "cannot determine home directory: %v", err)
			}
			base = filepath.Join(home, ".config")
		}
	}

	return filepath.Join(base, "ldctl", "config.toml"), nil
}

// Load reads config from env vars and/or config file.
// Precedence: env var > config file field.
// Returns error if no config found at all.
func Load() (*LoadResult, error) {
	result := &LoadResult{
		Source: Source{
			URL:   "not set",
			Token: "not set",
		},
	}

	// Try reading the config file first (lowest precedence, overridden by env).
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	var fileCfg tomlConfig
	fileExists := false
	if _, statErr := os.Stat(path); statErr == nil {
		fileExists = true
		if _, decodeErr := toml.DecodeFile(path, &fileCfg); decodeErr != nil {
			return nil, ldcerr.Newf(ldcerr.ConfigError, "failed to parse config file %s: %v", path, decodeErr)
		}
		if fileCfg.URL != "" {
			result.Config.URL = fileCfg.URL
			result.Source.URL = "config file"
		}
		if fileCfg.Token != "" {
			result.Config.Token = fileCfg.Token
			result.Source.Token = "config file"
		}
	}

	// Env vars take precedence over config file.
	if envURL := os.Getenv("LINKDING_URL"); envURL != "" {
		result.Config.URL = envURL
		result.Source.URL = "env: LINKDING_URL"
	}
	if envToken := os.Getenv("LINKDING_TOKEN"); envToken != "" {
		result.Config.Token = envToken
		result.Source.Token = "env: LINKDING_TOKEN"
	}

	// If nothing was loaded from any source, return an appropriate error.
	if result.Config.URL == "" && result.Config.Token == "" {
		if !fileExists {
			return nil, ldcerr.NewConfigNotFound(path)
		}
		return nil, ldcerr.Newf(ldcerr.ConfigError, "config file %s exists but contains no url or token", path)
	}

	return result, nil
}
