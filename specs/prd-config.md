# PRD: ldctl Config Feature

## Executive Summary

ldctl needs a configuration subsystem to persist LinkDing instance URL and API token. This is a foundational prerequisite — every other command group (bookmarks, tags, bundles, assets, user) depends on `config.Load()` to obtain connection credentials. The config feature provides three subcommands: `init` (interactive setup with validation), `show` (display effective config with source indication), and `test` (stepwise connectivity diagnostics).

## Problem Statement

**User pain point:** Without persistent configuration, users would need to pass `--url` and `--token` flags on every command, or always set environment variables — neither is ergonomic for daily use.

**Dependency:** The bookmarks feature (and all future command groups) assumes a `config.Load()` function exists. Config must ship first.

## Goals & Success Metrics

| Goal | Metric | Target |
|------|--------|--------|
| Persistent credentials | Config file created with correct permissions | `0600` on file, auto-created directory |
| Secure by default | Token never displayed in cleartext | Masked in `show`, masked during input |
| Config precedence | Env vars override config file | `LINKDING_URL` and `LINKDING_TOKEN` respected everywhere |
| Validation on setup | Init verifies credentials work | API call on init by default |
| Scriptability | All commands support `--json` | `show` and `test` output valid JSON |

## User Stories

### US-001: Initialize Configuration

**As a** new user, **I want to** run an interactive setup command **so that I can** configure my LinkDing connection once and use all other commands without re-entering credentials.

**Acceptance criteria:**

- `ldctl config init` prompts for URL, then token
- URL prompt accepts input, normalizes it (strips trailing slash, prepends `https://` if no scheme)
- Token prompt masks input (password-style, no echo)
- After collecting both values, validates by hitting `GET /api/user/profile/`
- On validation success: writes `~/.config/ldctl/config.toml` with `0600` permissions, creates directory if missing
- On validation failure: displays error, does not write file
- `--no-verify` skips API validation, writes file immediately
- If config file already exists: refuses with message, `--force` overwrites
- Non-interactive mode: when both `LINKDING_URL` and `LINKDING_TOKEN` env vars are set, skips prompts and writes directly (still validates unless `--no-verify`)
- If only one env var is set, falls through to full interactive prompts
- Exit code 0 on success, 1 on failure

### US-002: Show Configuration

**As a** user, **I want to** view my effective configuration and where each value comes from **so that I can** debug connection issues and verify precedence.

**Acceptance criteria:**

- `ldctl config show` displays URL and token with source labels
- Token is partially masked: first 3 and last 3 characters visible (e.g., `abc...xyz`)
- Source indicated per field: `(config file)`, `(env: LINKDING_URL)`, `(env: LINKDING_TOKEN)`
- `--json` outputs structured JSON with values and sources
- If no config exists and no env vars set: displays "No configuration found. Run `ldctl config init` to get started." and exits 1
- If config file has overly permissive permissions (not `0600`): displays warning on stderr
- `ldctl config` with no subcommand aliases to `config show`
- Exit code 0 when config is found, 1 when missing

### US-003: Test Connection

**As a** user, **I want to** verify my LinkDing connection works **so that I can** diagnose connectivity or auth problems.

**Acceptance criteria:**

- `ldctl config test` runs stepwise diagnostics:
  1. Checks config is loadable (file or env vars)
  2. Tests network connectivity to the URL
  3. Tests authentication via `GET /api/user/profile/`
- Output shows each step with pass/fail:
  ```
  Configuration... ok
  Connecting to https://linkding.example.com... ok
  Authenticating... ok
  ✓ Successfully connected to https://linkding.example.com
  ```
- On failure, stops at the failed step with actionable error message
- `--json` outputs structured result with each step's status
- Exit code 0 on full success, 1 on any failure

## Functional Requirements

| ID | Priority | Requirement |
|----|----------|-------------|
| REQ-001 | Must | `config init` interactive prompts for URL and token |
| REQ-002 | Must | Token input masked (no echo) during interactive prompt |
| REQ-003 | Must | URL normalization: strip trailing slash, auto-prepend `https://` |
| REQ-004 | Must | URL validation: reject malformed URLs |
| REQ-005 | Must | Validate credentials via API on init (default behavior) |
| REQ-006 | Must | `--no-verify` flag to skip validation on init |
| REQ-007 | Must | `--force` flag to overwrite existing config |
| REQ-008 | Must | Non-interactive mode when both `LINKDING_URL` and `LINKDING_TOKEN` env vars are set |
| REQ-009 | Must | Config file written at `$XDG_CONFIG_HOME/ldctl/config.toml` (fallback `~/.config/ldctl/config.toml`) |
| REQ-010 | Must | File created with `0600` permissions, directory auto-created |
| REQ-011 | Must | `config show` displays URL and partially-masked token with source labels |
| REQ-012 | Must | `config test` with stepwise diagnostics (config, connectivity, auth) |
| REQ-013 | Must | `config.Load()` implements precedence: env var > config file |
| REQ-014 | Must | All config commands support `--json` output |
| REQ-015 | Must | Bare `ldctl config` aliases to `config show` |
| REQ-016 | Must | Warn on stderr if config file permissions are not `0600` |
| REQ-017 | Must | Malformed TOML: error with "Config file is corrupt. Run `ldctl config init --force` to recreate." |
| REQ-018 | Must | Missing field in config: error naming the specific field (e.g., "Config missing required field: token") |
| REQ-019 | Must | Partial config + env var: env var fills the gap (file has URL, `LINKDING_TOKEN` supplies token) |

## Non-Functional Requirements

| Category | Requirement |
|----------|-------------|
| Security | Token never logged, displayed, or written to stdout in cleartext |
| Security | Config file `0600` permissions enforced on creation |
| Security | Permission warning on read if file is group/world-readable |
| Compatibility | XDG Base Directory spec compliant (`$XDG_CONFIG_HOME` respected) |
| Compatibility | Linux, macOS, Windows (Windows uses `%APPDATA%\ldctl\config.toml` as fallback) |
| Error handling | Meaningful error messages on stderr. Non-zero exit codes on failure |
| Exit codes | 0 = success, 1 = failure |

## Technical Considerations

### Config File Format

```toml
url = "https://linkding.example.com"
token = "abcdef1234567890"
```

Two fields only. No sections, no profiles for v1.

### Config Precedence

```
Environment variable > Config file
```

`LINKDING_URL` overrides `url` from file. `LINKDING_TOKEN` overrides `token` from file. Both are checked independently — env var can fill a gap in a partial config file.

### `config.Load()` Interface

```go
// Load returns the effective configuration by checking env vars and config file.
// Returns an error if neither source provides both URL and token.
func Load() (url string, token string, err error)
```

This is the contract that all other command groups depend on.

### URL Normalization

Applied during `config init` before writing:

1. If no scheme, prepend `https://`
2. Parse as URL — reject if invalid
3. Strip trailing `/`

Example: `linkding.example.com/` becomes `https://linkding.example.com`

### Permission Check

On every `config.Load()` call that reads the file, check permissions. If group-readable (`0640`) or world-readable (`0644`), emit warning to stderr:

```
Warning: config file /home/user/.config/ldctl/config.toml has overly permissive permissions.
Run: chmod 600 /home/user/.config/ldctl/config.toml
```

### Architecture

```
internal/
  config/
    config.go          # Load(), file path resolution, env var handling
    config_test.go     # unit tests
cmd/
  config.go            # config subcommand group (aliases to show)
  config_init.go       # config init
  config_show.go       # config show
  config_test_cmd.go   # config test (avoid name collision with _test.go)
```

### Display Formats

**`config show` default output:**

```
URL:   https://linkding.example.com  (config file)
Token: abc...xyz                     (env: LINKDING_TOKEN)
```

**`config show --json` output:**

```json
{
  "url": {
    "value": "https://linkding.example.com",
    "source": "config_file"
  },
  "token": {
    "value": "abc...xyz",
    "source": "env"
  }
}
```

Token remains masked in JSON output.

**`config test --json` output:**

```json
{
  "config": { "status": "ok" },
  "connectivity": { "status": "ok" },
  "auth": { "status": "ok", "url": "https://linkding.example.com" },
  "success": true
}
```

On failure:

```json
{
  "config": { "status": "ok" },
  "connectivity": { "status": "failed", "error": "connection refused" },
  "auth": { "status": "skipped" },
  "success": false
}
```

## Implementation Roadmap

### Phase 1: Core Config

1. `internal/config/config.go` — file path resolution (XDG), TOML read/write, `Load()`, permission checks
2. `cmd/config.go` — config subcommand group with alias to `show`
3. `cmd/config_init.go` — interactive prompts, env var non-interactive mode, URL normalization, `--force`, `--no-verify`
4. `cmd/config_show.go` — display with masking, source labels, `--json`
5. `cmd/config_test_cmd.go` — stepwise diagnostics, `--json`

### Phase 2: Integration

1. Wire `config.Load()` into root command's `PersistentPreRunE` for use by all subcommands
2. Verify bookmarks commands can consume config (integration point)

## Out of Scope

- Multiple profiles / named configurations
- `config delete` / `config reset` commands (users can `rm` the file)
- Config file encryption at rest
- `--unmask` flag on `config show`
- GUI or TUI config editor
- Config migration from other LinkDing tools

## Open Questions & Risks

| # | Question/Risk | Status | Notes |
|---|---------------|--------|-------|
| 1 | Windows XDG equivalent | Open | Use `%APPDATA%\ldctl\config.toml` on Windows. File permission model differs — skip `0600` enforcement on Windows |
| 2 | Token rotation UX | Low risk | User runs `config init --force` to replace. No incremental update command needed for v1 |
| 3 | Concurrent config access | Low risk | CLI is single-process. No locking needed |

## Validation Checkpoints

| Checkpoint | Criteria |
|------------|----------|
| Phase 1 complete | `config init` creates valid TOML file with correct permissions. `config show` displays masked config with sources. `config test` passes against live LinkDing instance. All three support `--json`. Non-interactive mode works with env vars |
| Phase 2 complete | `bookmarks list` successfully loads config via `config.Load()` and hits the API |
