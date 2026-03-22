# PRD: ldctl User Feature

## Executive Summary

ldctl needs a user command group to display the authenticated user's LinkDing profile preferences. This is a read-only, single-endpoint feature — `GET /api/user/profile/` returns display settings, sharing toggles, and search preferences. The command exists purely for visibility; it does not influence CLI behavior in v1. The `config test` command already handles connectivity validation, so `user profile` has no health-check role.

## Problem Statement

**User pain point:** Users have no CLI visibility into their LinkDing profile preferences. Verifying settings like theme, date display format, or sharing toggles requires opening the web UI.

**Dependency:** None. This feature depends on `config.Load()` for credentials but has no downstream dependents.

## Goals & Success Metrics

| Goal | Metric | Target |
|------|--------|--------|
| Full API coverage | User endpoint implemented | 1/1 endpoint |
| CLI-native experience | Follows established output conventions | Supports `--json`, exit codes |
| Consistent UX | Matches existing command patterns | Bare `user` aliases to `user profile` |

## User Stories

### US-001: View User Profile

**As a** terminal user, **I want to** view my LinkDing profile preferences **so that I can** check my settings without opening the web UI.

**Acceptance criteria:**

- `ldctl user profile` fetches `GET /api/user/profile/` and displays all fields
- Default output is flat key-value, one field per line
- Field names rendered in snake_case matching the API (e.g., `bookmark_date_display: relative`)
- Nested `search_preferences` fields flattened with dot notation (e.g., `search_preferences.sort: title_asc`)
- Boolean fields displayed as `true`/`false`
- `--json` outputs the raw API response as-is (no transformation)
- `ldctl user` with no subcommand aliases to `user profile`
- Exit code 0 on success, 1 on failure (auth error, network error, missing config)
- Errors printed to stderr with actionable messages

## Functional Requirements

| ID | Priority | Requirement |
|----|----------|-------------|
| REQ-001 | Must | `user profile` fetches and displays all fields from `GET /api/user/profile/` |
| REQ-002 | Must | Default output: flat key-value with snake_case field names |
| REQ-003 | Must | Nested objects flattened with dot notation (e.g., `search_preferences.sort`) |
| REQ-004 | Must | Boolean fields displayed as `true`/`false` |
| REQ-005 | Must | `--json` outputs raw API response |
| REQ-006 | Must | Bare `ldctl user` aliases to `user profile` |
| REQ-007 | Must | Exit code 0 on success, 1 on failure |
| REQ-008 | Must | Auth/network errors printed to stderr with actionable messages |
| REQ-009 | Must | Always fetches fresh from API (no caching) |

## Non-Functional Requirements

| Category | Requirement |
|----------|-------------|
| Performance | Single API call, no caching or background fetches |
| Error handling | Meaningful error messages on stderr. Non-zero exit codes on failure |
| Exit codes | 0 = success, 1 = failure |
| Consistency | Output conventions match `config show` (key-value alignment, `--json` support) |

## Technical Considerations

### API Response Shape

```json
{
  "theme": "auto",
  "bookmark_date_display": "relative",
  "bookmark_link_target": "_blank",
  "web_archive_integration": "enabled",
  "tag_search": "lax",
  "enable_sharing": true,
  "enable_public_sharing": true,
  "enable_favicons": false,
  "display_url": false,
  "permanent_notes": false,
  "search_preferences": {
    "sort": "title_asc",
    "shared": "off",
    "unread": "off"
  }
}
```

### Architecture

```
internal/
  user/
    user.go           # API client: GetProfile(), response types
    user_test.go      # unit tests
cmd/
  user.go             # user subcommand group (aliases to profile)
  user_profile.go     # user profile command
```

`internal/user/user.go` owns the HTTP call and response deserialization. It accepts the base URL and token (from `config.Load()`), returns a typed struct. The command layer in `cmd/` handles formatting and output.

### Display Formats

**`user profile` default output:**

```
theme:                        auto
bookmark_date_display:        relative
bookmark_link_target:         _blank
web_archive_integration:      enabled
tag_search:                   lax
enable_sharing:               true
enable_public_sharing:        true
enable_favicons:              false
display_url:                  false
permanent_notes:              false
search_preferences.sort:      title_asc
search_preferences.shared:    off
search_preferences.unread:    off
```

**`user profile --json` output:**

Raw API response, no transformation:

```json
{
  "theme": "auto",
  "bookmark_date_display": "relative",
  "bookmark_link_target": "_blank",
  "web_archive_integration": "enabled",
  "tag_search": "lax",
  "enable_sharing": true,
  "enable_public_sharing": true,
  "enable_favicons": false,
  "display_url": false,
  "permanent_notes": false,
  "search_preferences": {
    "sort": "title_asc",
    "shared": "off",
    "unread": "off"
  }
}
```

### Error Messages

| Scenario | stderr message |
|----------|---------------|
| No config | `Error: no configuration found. Run 'ldctl config init' to get started.` |
| Auth failure (401) | `Error: authentication failed. Check your API token.` |
| Network error | `Error: could not connect to <url>: <reason>` |
| Non-200 response | `Error: unexpected response from server (HTTP <code>)` |

## Implementation Roadmap

### Phase 1: Core

1. `internal/user/user.go` — `GetProfile()` function, response struct, HTTP client
2. `cmd/user.go` — user subcommand group with alias to `profile`
3. `cmd/user_profile.go` — flat key-value output, `--json` flag

### Phase 2: Integration

1. Wire `config.Load()` into the user command (via root command's `PersistentPreRunE`, shared with other commands)
2. Verify error paths (no config, bad token, unreachable server)

## Out of Scope

- Profile modification (API is read-only for this endpoint)
- Caching profile data locally
- Using profile preferences to influence other command behavior
- Health check / connection test role (handled by `config test`)
- Field filtering (`--field` flag — use `--json` + `jq` instead)
- `user whoami` or other aliases beyond the bare `user` alias

## Open Questions & Risks

| # | Question/Risk | Status | Notes |
|---|---------------|--------|-------|
| 1 | API may add new profile fields | Low risk | Use a permissive deserialization strategy — unknown fields pass through in JSON mode, appear in key-value output via reflection or map iteration |
| 2 | Field ordering in flat output | Open | API response field order isn't guaranteed. Define a fixed display order in the command to keep output stable |
| 3 | Future write endpoints | Out of scope | If LinkDing adds profile update APIs, a `user profile update` command can be added without restructuring |

## Validation Checkpoints

| Checkpoint | Criteria |
|------------|----------|
| Phase 1 complete | `user profile` displays all profile fields in flat key-value format. `--json` outputs raw API response. Bare `ldctl user` aliases correctly. Unit tests cover response parsing and output formatting |
| Phase 2 complete | `user profile` loads config via shared `config.Load()`. Error messages are correct for missing config, bad auth, and network failures. Exit codes are 0/1 as specified |
