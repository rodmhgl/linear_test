# ldctl Implementation Plan

> Gap analysis: specs (`specs/prd-*.md`) + `AGENTS.md` vs current repository state.
>
> **Current state**: Only `main.go` exists (calls `cmd.Execute()` which doesn't exist yet). `cmd/`, `internal/`, `docs/` directories are empty (`.gitkeep` only). `go.mod` and `Makefile` are configured. **Zero implementation code exists.**
>
> **Every task below traces to an unimplemented spec requirement or a foundational dependency required by specs.**

---

## P0 — Foundational / Blocking

Nothing else can start without these. Ordered by internal dependency chain.

- [ ] **P0** | Error types and exit code infrastructure | ~medium
  - Acceptance: `internal/errors/` package exists with `Error` struct implementing `error` interface. All 8 error types defined as constants (`config_error`, `auth_error`, `network_error`, `api_error`, `validation_error`, `io_error`, `not_found`, `user_cancelled`). `ExitCode()` returns 2 for config/auth/network errors, 1 for all others (per prd-error-handling.md). JSON serialization outputs `{"error":{"type":"...","message":"...","details":{...}}}` to stderr. Human-readable format: `"Error: <description>"` with optional context and resolution lines. Constructor helpers exist: `NewConfigNotFound`, `NewAuthFailed`, `NewNotFound`, `NewValidation`, `NewNetworkError`, `NewIOError`. HTTP status mapping function covers 400, 401, 403, 404, 429, 5xx. Unit tests pass covering: all type→exit code mappings, JSON serialization validity, human format, each constructor.
  - Files: `internal/errors/errors.go`, `internal/errors/codes.go`, `internal/errors/format.go`, `internal/errors/json.go`, `internal/errors/errors_test.go`
  - Depends on: none
  - Spec trace: prd-error-handling.md REQ-001 through REQ-011

- [ ] **P0** | Version package with build-time injection | ~small
  - Acceptance: `internal/version/` package exports `Version`, `Commit`, `BuildDate` string vars (defaults: `"dev"`, `"unknown"`, `"unknown"`) and `GoVersion` via `runtime.Version()`. `Info` struct with JSON tags: `version`, `commit`, `buildDate`, `goVersion`, `os`, `arch`. `Get()` returns populated `Info`. `String()` returns `"ldctl version X (commit Y, built Z, W)"`. `make build` injects real values via existing Makefile ldflags. Unit tests verify `String()` format and `Get()` field population.
  - Files: `internal/version/version.go`, `internal/version/version_test.go`
  - Depends on: none
  - Spec trace: prd-version.md REQ-004, REQ-005

- [ ] **P0** | Data models for all API resources | ~small
  - Acceptance: Go structs defined with JSON tags matching LinkDing API field names (snake_case): `Bookmark` (id, url, title, description, notes, web_archive_snapshot_url, favicon_url, preview_image_url, is_archived, unread, shared, tag_names []string, date_added, date_modified), `BookmarkCheck` (bookmark *Bookmark, metadata with title/description/auto_tags), `Tag` (id, name, date_added), `Bundle` (id, name, search, any_tags, all_tags, excluded_tags, order, date_created, date_modified), `Asset` (id, bookmark_id, asset_type, status, content_type, display_name, date_created), `UserProfile` (all fields from prd-user.md including nested SearchPreferences struct), `PaginatedResponse[T]` generic (count int, next *string, previous *string, results []T). Unit tests verify JSON round-trip marshaling for each type.
  - Files: `internal/models/bookmark.go`, `internal/models/tag.go`, `internal/models/bundle.go`, `internal/models/asset.go`, `internal/models/user.go`, `internal/models/pagination.go`, `internal/models/models_test.go`
  - Depends on: none
  - Spec trace: prd-bookmarks.md (display format), prd-tags.md (tag object), prd-bundles.md (bundle object), prd-assets.md (asset fields), prd-user.md (profile response shape)

- [ ] **P0** | Output helpers (JSON, key-value, progress, verbose) | ~medium
  - Acceptance: `internal/output/` package provides: `PrintJSON(data interface{})` — pretty-prints to stdout with 2-space indent, snake_case keys, nulls as `null`, empty arrays as `[]`, ISO 8601 timestamps; `PrintData(format, args)` — to stdout; `PrintError(format, args)` — to stderr with "Error: " prefix; `PrintProgress(msg)` — to stderr, suppressed when quiet; `PrintVerbose(format, args)` — to stderr with "[DEBUG]" prefix, only when verbose. Quiet/verbose/json state configurable via `SetFlags(json, quiet, verbose bool)`. Unit tests verify: stdout vs stderr routing, quiet suppression, verbose gating, JSON format validity.
  - Files: `internal/output/output.go`, `internal/output/output_test.go`
  - Depends on: none
  - Spec trace: prd-root-command.md REQ-002 through REQ-004, REQ-009, REQ-010

- [ ] **P0** | Configuration subsystem (Load, file I/O, env vars) | ~large
  - Acceptance: `internal/config/` package implements `Load() (*Config, error)` where `Config` contains `URL`, `Token`, and per-field `Source` ("config_file" or "env"). File path: `$XDG_CONFIG_HOME/ldctl/config.toml` (fallback `~/.config/ldctl/config.toml`; Windows: `%APPDATA%\ldctl\config.toml`). Reads TOML with `BurntSushi/toml` (already in go.mod). Env vars `LINKDING_URL`/`LINKDING_TOKEN` override config file per-field independently (REQ-013, REQ-019). Permission check: warns on stderr if file not `0600` on Unix (REQ-016, skip on Windows). Malformed TOML: returns error with "Config file is corrupt. Run `ldctl config init --force` to recreate." (REQ-017). Missing field: error naming specific field (REQ-018). `Save(url, token string)` writes TOML with `0600` perms, auto-creates directory (REQ-010). URL normalization: strip trailing `/`, prepend `https://` if no scheme, reject malformed (REQ-003, REQ-004). `ConfigPath()` returns resolved path. Unit tests cover: env-only, file-only, mixed precedence, partial config+env gap-fill, missing field, malformed TOML, permission warning, URL normalization edge cases, XDG resolution.
  - Files: `internal/config/config.go`, `internal/config/config_test.go`
  - Depends on: Error types and exit code infrastructure
  - Spec trace: prd-config.md REQ-001 through REQ-019

- [ ] **P0** | Root command with global flags | ~medium
  - Acceptance: `cmd/root.go` defines root Cobra command. `ldctl` with no args shows help matching prd-root-command.md template (description, usage, available commands, global flags, examples). `Execute()` returns `int` exit code (matching `main.go` contract `os.Exit(cmd.Execute())`). Global persistent flags: `--json` (bool), `--quiet`/`-q` (bool), `--verbose`/`-v` (bool), `--version` (bool). `--quiet` + `--verbose` returns error "cannot use --quiet and --verbose together" (REQ-007). `--version` short-circuits to version display, exit 0 (REQ-005). `PersistentPreRunE`: validates flag exclusivity, loads config (skips for `config`, `version`, `help` commands), sets output flags. `SilenceErrors = true`, `SilenceUsage = true`. Custom help template. Default completion command disabled. Error handling: typed errors → appropriate exit code, `--json` → JSON error to stderr. Unit tests: help output, flag mutual exclusivity error, version short-circuit, config skip for exempt commands.
  - Files: `cmd/root.go`, `cmd/root_test.go`
  - Depends on: Error types and exit code infrastructure, Version package, Configuration subsystem, Output helpers
  - Spec trace: prd-root-command.md REQ-001 through REQ-012

- [ ] **P0** | API client core (HTTP, auth, pagination) | ~medium
  - Acceptance: `internal/api/client.go` defines `Client` struct with base URL, token, `*http.Client`. `NewClient(url, token string) *Client`. All requests include `Authorization: Token <token>`, `Accept: application/json`, `User-Agent: ldctl/<version>` headers. Internal `do(method, path, body)` method builds request, reads response, maps HTTP errors to `errors.Error` types (401/403→AuthError exit 2, 404→NotFound exit 1, 400→ValidationError, 429→APIError with Retry-After, 5xx→APIError). `Paginate[T](path, params) ([]T, error)` generic helper fetches all pages from LinkDing paginated responses (reads `count`, `next`, `results`). Verbose mode logs request/response to stderr via output helpers. Unit tests with `httptest.NewServer`: auth header present, pagination across 3+ pages, each HTTP error code mapping, User-Agent header.
  - Files: `internal/api/client.go`, `internal/api/client_test.go`
  - Depends on: Error types and exit code infrastructure, Version package, Data models, Output helpers
  - Spec trace: prd-bookmarks.md (API client design), AGENTS.md (stdlib HTTP, token auth)

---

## P1 — Core Functionality

Primary features that define the CLI. Ordered by dependency chain within groups.

### Version Command

- [ ] **P1** | `ldctl version` command | ~small
  - Acceptance: `ldctl version` outputs full version string to stdout. `ldctl version --short` outputs only semver (e.g., `1.2.3`). `ldctl version --json` outputs JSON object with `version`, `commit`, `buildDate`, `goVersion`, `os`, `arch`. Exit code 0. No network calls. Works without config file. Dev builds show "dev". `ldctl --version` (global flag) also works from any subcommand context. Unit tests: default format, short format, JSON validity, no-config requirement.
  - Files: `cmd/version.go`, `cmd/version_test.go`
  - Depends on: Root command with global flags, Version package
  - Spec trace: prd-version.md REQ-001, REQ-002, REQ-006

### Config Commands

- [ ] **P1** | `ldctl config` command group + `config init` | ~medium
  - Acceptance: `cmd/config.go` registers `config` subcommand group; bare `ldctl config` aliases to `config show` (prd-config.md REQ-015). `config init`: prompts interactively for URL then token. Token input masked via `golang.org/x/term` (REQ-002). URL normalized before write (REQ-003). Validates credentials via `GET /api/user/profile/` by default (REQ-005). `--no-verify` skips validation (REQ-006). Writes config file with `0600` perms and auto-created directory (REQ-010). Refuses if file exists with message; `--force` overwrites (REQ-007). Non-interactive mode: both `LINKDING_URL` and `LINKDING_TOKEN` set → skip prompts, still validate unless `--no-verify` (REQ-008). Only one env var set → full interactive prompts. Exit code 0 success, 1 failure. Unit tests: URL normalization cases, force overwrite, non-interactive detection, partial env var fallthrough.
  - Files: `cmd/config.go`, `cmd/config_init.go`, `cmd/config_init_test.go`
  - Depends on: Root command with global flags, Configuration subsystem, API client core
  - Spec trace: prd-config.md US-001, REQ-001 through REQ-010

- [ ] **P1** | `ldctl config show` command | ~small
  - Acceptance: Displays URL and token with source labels: `(config file)`, `(env: LINKDING_URL)`, `(env: LINKDING_TOKEN)` (REQ-011). Token partially masked: first 3 + last 3 chars visible (e.g., `abc...xyz`). Token masked in JSON output too. `--json` outputs `{"url":{"value":"...","source":"config_file"},"token":{"value":"abc...xyz","source":"env"}}`. If no config and no env vars: "No configuration found. Run `ldctl config init` to get started." exit 1. Warns on stderr if file permissions not `0600` (REQ-016). Exit code 0 found, 1 missing. Unit tests: source labels, masking logic, missing config message, JSON format, permission warning.
  - Files: `cmd/config_show.go`, `cmd/config_show_test.go`
  - Depends on: `ldctl config` command group + `config init`, Configuration subsystem
  - Spec trace: prd-config.md US-002, REQ-011, REQ-014, REQ-015, REQ-016

- [ ] **P1** | `ldctl config test` command | ~small
  - Acceptance: Runs 3 stepwise diagnostics: (1) config loadable, (2) network connectivity to URL, (3) auth via `GET /api/user/profile/`. Output per step: `"Configuration... ok"`, `"Connecting to <url>... ok"`, `"Authenticating... ok"`, `"✓ Successfully connected to <url>"`. On failure, stops at failed step with actionable message. `--json` outputs `{"config":{"status":"ok"},"connectivity":{"status":"ok"},"auth":{"status":"ok","url":"..."},"success":true}`. Failed step shows `"status":"failed","error":"..."`, subsequent steps show `"status":"skipped"`. Exit code 0 full success, 1 any failure. Unit tests with mock HTTP: all-pass, config fail, connectivity fail, auth fail, JSON output.
  - Files: `cmd/config_test_cmd.go`, `cmd/config_test_cmd_test.go`
  - Depends on: `ldctl config` command group + `config init`, Configuration subsystem, API client core
  - Spec trace: prd-config.md US-003, REQ-012, REQ-014

### Bookmark Commands

- [ ] **P1** | Bookmark API client methods | ~medium
  - Acceptance: `internal/api/bookmarks.go` on `Client`: `ListBookmarks(params ListParams)` returns `PaginatedResponse[Bookmark]` with `limit`, `offset`, `query`, `archived` filter (uses `/api/bookmarks/archived/` when archived=true). `GetBookmark(id int)` returns `*Bookmark`. `CreateBookmark(input CreateBookmarkInput)` returns `*Bookmark`. `CheckBookmark(url string)` returns `*BookmarkCheck`. `UpdateBookmark(id int, input map[string]interface{})` (PATCH) returns `*Bookmark`. `ArchiveBookmark(id int)` returns error. `UnarchiveBookmark(id int)` returns error. `DeleteBookmark(id int)` returns error. Unit tests with mock HTTP: each method happy path, error responses, pagination for list, archived endpoint switch.
  - Files: `internal/api/bookmarks.go`, `internal/api/bookmarks_test.go`
  - Depends on: API client core, Data models
  - Spec trace: prd-bookmarks.md (all CRUD user stories)

- [ ] **P1** | `ldctl bookmarks` command group + `bm` alias | ~small
  - Acceptance: `cmd/bookmarks.go` registers `bookmarks` subcommand with `bm` alias (REQ-009). Running `ldctl bookmarks` (no subcommand) shows help listing all bookmark subcommands. Help text includes all subcommands: list, get, add, check, update, archive, unarchive, delete.
  - Files: `cmd/bookmarks.go`
  - Depends on: Root command with global flags
  - Spec trace: prd-bookmarks.md REQ-009, prd-root-command.md (subcommand help example)

- [ ] **P1** | `ldctl bookmarks list` command | ~medium
  - Acceptance: Default output: first page (100 items) in key-value format per prd-bookmarks.md display format (ID, URL, Title, Description, Tags, Added, Modified, Unread, Shared, Archived). Bookmarks separated by blank lines. Tags comma-separated in display. `--query` passes through LinkDing search syntax. `--archived` switches to archived endpoint. `--limit` and `--offset` manual pagination. `--all` auto-fetches all pages. `--json` outputs JSON array. Empty result set: exit 0, appropriate message (not error). Exit code 1 on API error. Unit tests: output format, each flag, empty results, pagination.
  - Files: `cmd/bookmarks_list.go`, `cmd/bookmarks_list_test.go`
  - Depends on: `ldctl bookmarks` command group, Bookmark API client methods, Output helpers
  - Spec trace: prd-bookmarks.md US-001, REQ-001, REQ-008, REQ-010

- [ ] **P1** | `ldctl bookmarks get` command | ~small
  - Acceptance: `ldctl bookmarks get <id>` displays full bookmark in key-value format. `--json` outputs JSON object. Validates ID is numeric (returns "Expected numeric ID, got: <input>" on invalid). Exit code 1 on 404. Unit tests: success, 404, invalid ID, JSON.
  - Files: `cmd/bookmarks_get.go`, `cmd/bookmarks_get_test.go`
  - Depends on: `ldctl bookmarks` command group, Bookmark API client methods
  - Spec trace: prd-bookmarks.md US-002, REQ-002

- [ ] **P1** | `ldctl bookmarks add` command | ~small
  - Acceptance: `ldctl bookmarks add <url>` creates bookmark. Flags: `--title`, `--description`, `--notes`, `--tags "tag1 tag2"` (space-separated), `--archived`, `--unread`, `--shared`. Tags split on spaces into `tag_names` array. Duplicate URLs silently upsert (API behavior, per CLAUDE.md quirk #1). Displays created/updated bookmark. `--json` outputs JSON. Exit code 0 success, 1 error. Unit tests: minimal, all flags, space-separated tags.
  - Files: `cmd/bookmarks_add.go`, `cmd/bookmarks_add_test.go`
  - Depends on: `ldctl bookmarks` command group, Bookmark API client methods
  - Spec trace: prd-bookmarks.md US-003, REQ-003

- [ ] **P1** | `ldctl bookmarks check` command | ~small
  - Acceptance: `ldctl bookmarks check <url>` calls `/api/bookmarks/check/?url=<url>`. Shows existing bookmark if found, or "Not bookmarked" message. Displays scraped metadata (title, description) and auto-tags when available. `--json` outputs raw JSON response. Unit tests: found, not found, JSON.
  - Files: `cmd/bookmarks_check.go`, `cmd/bookmarks_check_test.go`
  - Depends on: `ldctl bookmarks` command group, Bookmark API client methods
  - Spec trace: prd-bookmarks.md US-004, REQ-004

- [ ] **P1** | `ldctl bookmarks update` command (basic) | ~small
  - Acceptance: `ldctl bookmarks update <id>` with PATCH semantics — only provided flags sent. Flags: `--title`, `--description`, `--notes`, `--tags` (replaces entire tag list, space-separated), `--unread`, `--shared`. Displays updated bookmark. `--json` outputs JSON. Exit code 0 success, 1 error. Unit tests: single field, multiple fields, tag replacement.
  - Files: `cmd/bookmarks_update.go`, `cmd/bookmarks_update_test.go`
  - Depends on: `ldctl bookmarks` command group, Bookmark API client methods
  - Spec trace: prd-bookmarks.md US-005, REQ-005 (basic subset)

- [ ] **P1** | `ldctl bookmarks archive` and `unarchive` commands | ~small
  - Acceptance: `ldctl bookmarks archive <id>` archives via `POST /api/bookmarks/<id>/archive/`. `ldctl bookmarks unarchive <id>` unarchives via `POST /api/bookmarks/<id>/unarchive/`. Displays confirmation message. Exit code 0 success, 1 failure. Unit tests: success, 404.
  - Files: `cmd/bookmarks_archive.go`, `cmd/bookmarks_archive_test.go`
  - Depends on: `ldctl bookmarks` command group, Bookmark API client methods
  - Spec trace: prd-bookmarks.md US-006, REQ-006

- [ ] **P1** | `ldctl bookmarks delete` command | ~small
  - Acceptance: `ldctl bookmarks delete <id>` fetches bookmark (for title), prompts `Delete bookmark #<id> "<title>"? [y/N]`. Default "N". `--force` skips confirmation. Displays confirmation on success. Exit code 0 success, 1 failure or cancel. Unit tests: force, user cancel, 404.
  - Files: `cmd/bookmarks_delete.go`, `cmd/bookmarks_delete_test.go`
  - Depends on: `ldctl bookmarks` command group, Bookmark API client methods
  - Spec trace: prd-bookmarks.md US-007, REQ-007

### Tag Commands

- [ ] **P1** | Tag API client methods | ~small
  - Acceptance: `internal/api/tags.go` on `Client`: `ListTags(limit, offset int)` returns `PaginatedResponse[Tag]`, `GetTag(id int)` returns `*Tag`, `CreateTag(name string)` returns `*Tag`. Unit tests with mock HTTP.
  - Files: `internal/api/tags.go`, `internal/api/tags_test.go`
  - Depends on: API client core, Data models
  - Spec trace: prd-tags.md (API endpoints table)

- [ ] **P1** | `ldctl tags` command group + subcommands | ~small
  - Acceptance: `ldctl tags list` shows paginated tag list in key-value format (ID, Name, Added), blank-line separated. `--limit`, `--offset`, `--all` pagination. `--json` outputs JSON array. `ldctl tags get <id>` displays single tag in key-value, exit 1 on 404. `ldctl tags create <name>` creates tag, displays result. `tags create --help` includes note: "Tags are also created implicitly when used in bookmark tag_names arrays" (REQ-006). All support `--json`. Exit codes per spec. Unit tests per subcommand.
  - Files: `cmd/tags.go`, `cmd/tags_list.go`, `cmd/tags_get.go`, `cmd/tags_create.go`, `cmd/tags_test.go`
  - Depends on: Root command with global flags, Tag API client methods
  - Spec trace: prd-tags.md US-001 through US-003, REQ-001 through REQ-006

### User Command

- [ ] **P1** | User API client method | ~small
  - Acceptance: `internal/api/user.go` on `Client`: `GetProfile()` returns `*UserProfile`. **Ambiguity note**: prd-user.md places this in `internal/user/user.go` but for consistency with all other API methods on the shared `Client` struct, using `internal/api/user.go`. Unit tests with mock HTTP.
  - Files: `internal/api/user.go`, `internal/api/user_test.go`
  - Depends on: API client core, Data models
  - Spec trace: prd-user.md (API response shape)

- [ ] **P1** | `ldctl user profile` command | ~small
  - Acceptance: `ldctl user profile` fetches `GET /api/user/profile/`, displays all fields in flat key-value format with snake_case names. Nested `search_preferences` flattened with dot notation (e.g., `search_preferences.sort: title_asc`). Booleans as `true`/`false`. `--json` outputs raw API response (no transformation). Bare `ldctl user` aliases to `user profile` (REQ-006). Exit code 0 success, 1 failure. Error messages per prd-user.md: no config → "Run 'ldctl config init'", auth 401 → "Check your API token", network → "could not connect to <url>: <reason>", other → "unexpected response (HTTP <code>)". Unit tests: output format, dot-notation flattening, JSON passthrough, alias behavior.
  - Files: `cmd/user.go`, `cmd/user_profile.go`, `cmd/user_test.go`
  - Depends on: Root command with global flags, User API client method
  - Spec trace: prd-user.md US-001, REQ-001 through REQ-009

---

## P2 — Enhanced Features

Secondary capabilities that extend the core.

### Bookmark Enhancements

- [ ] **P2** | `--add-tags` and `--remove-tags` on bookmarks update | ~medium
  - Acceptance: `--add-tags "newtag"` on update: fetches current bookmark, merges new tags into `tag_names`, PATCHes merged list. `--remove-tags "oldtag"`: fetches current, removes specified tags (case-insensitive match), PATCHes. Both usable together in one command. `--tags` mutually exclusive with `--add-tags`/`--remove-tags` — validated before any API call. Edge cases: empty current tags, removing non-existent tag (no-op, no error), adding duplicate (idempotent). Unit tests: merge logic, remove logic, combined use, mutual exclusivity error, case-insensitive removal, empty-tags edge case.
  - Files: `cmd/bookmarks_update.go` (modify), `cmd/bookmarks_update_test.go` (modify)
  - Depends on: `ldctl bookmarks update` command (basic)
  - Spec trace: prd-bookmarks.md US-005 (--add-tags/--remove-tags), REQ-011

- [ ] **P2** | Bulk operations (multi-ID + stdin) for archive/unarchive/delete | ~medium
  - Acceptance: `ldctl bookmarks archive 1 2 3`, `unarchive 1 2 3`, `delete 1 2 3` process multiple IDs. `--stdin` reads newline-delimited IDs from stdin; also accepts JSON array extracting `.id` from each object. Delete with multiple IDs confirms each unless `--force`. Per-item result displayed. Exit code 1 if any operation fails (partial success still reports). Unit tests: multi-ID, stdin plain IDs, stdin JSON, partial failure, delete multi-confirm.
  - Files: `cmd/bookmarks_archive.go` (modify), `cmd/bookmarks_delete.go` (modify), `cmd/bookmarks_archive_test.go` (modify), `cmd/bookmarks_delete_test.go` (modify)
  - Depends on: `ldctl bookmarks archive` and `unarchive` commands, `ldctl bookmarks delete` command
  - Spec trace: prd-bookmarks.md US-006 (multi-ID + stdin), US-007 (multi-ID + stdin), REQ-012, REQ-013

- [ ] **P2** | `ldctl bookmarks open` command | ~small
  - Acceptance: `ldctl bookmarks open <id>` fetches bookmark, opens URL in default browser using `xdg-open` (Linux), `open` (macOS), `start` (Windows). Displays "Opening: <url>" to stderr. Exit code 1 if bookmark not found. Unit tests: success (mock exec), 404.
  - Files: `cmd/bookmarks_open.go`, `cmd/bookmarks_open_test.go`
  - Depends on: `ldctl bookmarks` command group, Bookmark API client methods
  - Spec trace: prd-bookmarks.md US-008, REQ-014

### Bundle Commands

- [ ] **P2** | Bundle API client methods | ~small
  - Acceptance: `internal/api/bundles.go` on `Client`: `ListBundles(limit, offset)` returns `PaginatedResponse[Bundle]`, `GetBundle(id)` returns `*Bundle`, `CreateBundle(input)` returns `*Bundle`, `UpdateBundle(id, input map[string]interface{})` (PATCH) returns `*Bundle`, `DeleteBundle(id)` returns error. Unit tests with mock HTTP per method.
  - Files: `internal/api/bundles.go`, `internal/api/bundles_test.go`
  - Depends on: API client core, Data models
  - Spec trace: prd-bundles.md (API endpoints table)

- [ ] **P2** | `ldctl bundles` CRUD commands | ~medium
  - Acceptance: `bundles list`: paginated, key-value format (ID, Name, Search, Any Tags, All Tags, Excluded Tags, Order, Created, Modified), empty fields as `(none)`, blank-line separated, `--limit`/`--offset`/`--all`, `--json`. `bundles get <id>`: all fields shown, `(none)` for empties, exit 1 on 404. `bundles create <name>`: `--search`, `--any-tag`, `--all-tag`, `--exclude-tag` (repeated flags joined to space-separated strings per API), `--order` (omitted if not provided). `bundles update <id>`: PATCH semantics, same flags as create + `--name`, empty string clears fields (`--search ''`). `bundles delete <id>`: confirmation prompt, `--force` skips. All support `--json`. Unit tests per subcommand including: repeated tag flag joining, field clearing, delete confirmation.
  - Files: `cmd/bundles.go`, `cmd/bundles_list.go`, `cmd/bundles_get.go`, `cmd/bundles_create.go`, `cmd/bundles_update.go`, `cmd/bundles_delete.go`, `cmd/bundles_test.go`
  - Depends on: Root command with global flags, Bundle API client methods
  - Spec trace: prd-bundles.md US-001 through US-005, REQ-001 through REQ-010

- [ ] **P2** | `ldctl bundles view` command | ~small
  - Acceptance: `ldctl bundles view <id>` lists bookmarks matching bundle via `GET /api/bookmarks/?bundle=<id>`. Output matches `bookmarks list` format exactly. `--limit`, `--offset`, `--all` pagination. `--json` outputs JSON array of bookmark objects. Exit code 1 if bundle not found. Reuses bookmark list output formatting logic. Unit tests: results, empty, 404.
  - Files: `cmd/bundles_view.go`, `cmd/bundles_view_test.go`
  - Depends on: `ldctl bundles` CRUD commands, Bookmark API client methods, `ldctl bookmarks list` command
  - Spec trace: prd-bundles.md US-006, REQ-009

### Asset Commands

- [ ] **P2** | Asset API client methods | ~medium
  - Acceptance: `internal/api/assets.go` on `Client`: `ListAssets(bookmarkID int)` returns `[]Asset`, `GetAsset(bookmarkID, assetID int)` returns `*Asset`, `DownloadAsset(bookmarkID, assetID int)` returns `(io.ReadCloser, contentType string, error)`, `UploadAsset(bookmarkID int, filename string, body io.Reader, contentType string)` returns `*Asset` (multipart/form-data), `DeleteAsset(bookmarkID, assetID int)` returns error. Unit tests with mock HTTP including multipart upload verification.
  - Files: `internal/api/assets.go`, `internal/api/assets_test.go`
  - Depends on: API client core, Data models
  - Spec trace: prd-assets.md (API client methods section)

- [ ] **P2** | `ldctl assets list`, `get`, `delete` commands + table formatter | ~medium
  - Acceptance: `assets list <bookmark-id>`: compact table (ID, Type, Status, Content-Type, Name, Created). `assets get <bookmark-id> <asset-id>`: key-value (ID, Bookmark, Type, Status, Content-Type, Name, Created). `assets delete <bookmark-id> <asset-id>`: confirmation `Delete asset #<aid> from bookmark #<bid>? [y/N]`, `--force` skips. All support `--json`. Table formatter reusable for other tabular output. Unit tests per command.
  - Files: `cmd/assets.go`, `cmd/assets_list.go`, `cmd/assets_get.go`, `cmd/assets_delete.go`, `cmd/assets_test.go`, `internal/output/table.go`, `internal/output/table_test.go`
  - Depends on: Root command with global flags, Asset API client methods, Output helpers
  - Spec trace: prd-assets.md US-001, US-003, US-006, REQ-001, REQ-002, REQ-005, REQ-006, REQ-012

- [ ] **P2** | `ldctl assets download` command | ~medium
  - Acceptance: `assets download <bookmark-id> <asset-id>` saves file. Default filename: `asset-{bid}-{aid}.{ext}` (ext from `mime.ExtensionsByType`, fallback `.bin`). `--output <path>`: specific file path. `--output-dir <dir>`: auto-named file in directory. `--output` and `--output-dir` mutually exclusive (REQ-009). Pre-checks asset status via GET; if not "complete": `"Asset #X is not available (status: <status>)"`, exit 1 (REQ-010). `--force` skips status check AND allows overwrite (REQ-011). Refuses overwrite without `--force`: `"Error: file already exists: <path>"`. Displays "Saved: <filepath>" on success. Unit tests: auto-naming, status check fail, force skip, overwrite protection, output flag exclusivity.
  - Files: `cmd/assets_download.go`, `cmd/assets_download_test.go`
  - Depends on: `ldctl assets list`, `get`, `delete` commands, Asset API client methods
  - Spec trace: prd-assets.md US-004, REQ-003, REQ-007 through REQ-011

- [ ] **P2** | `ldctl assets upload` command | ~small
  - Acceptance: `assets upload <bookmark-id> <file>` uploads via multipart/form-data. Content-Type inferred from extension using `mime.TypeByExtension`, fallback `application/octet-stream`. Displays created asset metadata. `--json` outputs JSON. Exit code 1: file missing, bookmark not found, upload fail. Unit tests: success, missing file, MIME inference.
  - Files: `cmd/assets_upload.go`, `cmd/assets_upload_test.go`
  - Depends on: `ldctl assets list`, `get`, `delete` commands, Asset API client methods
  - Spec trace: prd-assets.md US-005, REQ-004

---

## P3 — Polish / Optional

Nice-to-haves, advanced features, and cross-cutting concerns.

- [ ] **P3** | `ldctl bookmarks export` command | ~large
  - Acceptance: Auto-paginates all bookmarks. `--format json` (default): JSON array. `--format csv`: CSV with header row. `--format html`: Netscape bookmark format (browser-importable). `--output <file>`: write to file instead of stdout. `--archived`: fetches from both main and archived endpoints. Progress indicator on stderr during multi-page fetch. Streams pages (doesn't buffer 10k+ bookmarks in memory). Unit tests: each format output, file output, archived flag, streaming verification.
  - Files: `cmd/bookmarks_export.go`, `cmd/bookmarks_export_test.go`, `internal/export/json.go`, `internal/export/csv.go`, `internal/export/netscape.go`
  - Depends on: `ldctl bookmarks list` command, Bookmark API client methods
  - Spec trace: prd-bookmarks.md US-009, REQ-015

- [ ] **P3** | `ldctl bookmarks import` command | ~large
  - Acceptance: `bookmarks import <file>` bulk-creates bookmarks. `--format html`: Netscape HTML format. `--format json`: JSON array matching export format. Auto-detects format if `--format` omitted (REQ-017). Progress: `"Imported 42/100 bookmarks..."`. Reports failures without stopping: `"Failed: <url> - <error>"`. Summary: `"Imported 98/100 (2 failures)"`. Exit code 0 all succeed, 1 any fail. Unit tests: JSON, HTML, auto-detection, partial failure, progress.
  - Files: `cmd/bookmarks_import.go`, `cmd/bookmarks_import_test.go`, `internal/export/import.go`, `internal/export/netscape_parse.go`
  - Depends on: `ldctl bookmarks` command group, Bookmark API client methods
  - Spec trace: prd-bookmarks.md US-010, REQ-016, REQ-017, REQ-018

- [ ] **P3** | `ldctl assets list --all` (cross-bookmark listing) | ~medium
  - Acceptance: `assets list --all` paginates all bookmarks, fetches assets per bookmark, merges into table with added Bookmark ID and Bookmark Title columns. Progress on stderr: `"Scanning bookmark 42/150..."`. `--json` outputs flat JSON array with bookmark context. `--limit`/`--offset` control bookmark pagination. Handles zero-asset bookmarks (skip). Unit tests: aggregation, progress, empty bookmarks.
  - Files: `cmd/assets_list.go` (modify), `cmd/assets_list_test.go` (modify)
  - Depends on: `ldctl assets list`, `get`, `delete` commands, Bookmark API client methods
  - Spec trace: prd-assets.md US-002, REQ-013, REQ-014, REQ-015

- [ ] **P3** | Verbose mode HTTP request/response logging | ~small
  - Acceptance: `--verbose` logs all HTTP traffic to stderr: `[DEBUG] Loading config from <path>`, `[DEBUG] GET <url>`, `[DEBUG] Response: <status> (<ms>ms)`, `[DEBUG] Request headers` (token masked), response headers. Format matches prd-root-command.md verbose output examples. Suppressed when `--verbose` not set. Unit tests: verbose on → logs present, verbose off → no logs.
  - Files: `internal/api/client.go` (modify), `internal/api/client_test.go` (modify)
  - Depends on: API client core, Output helpers
  - Spec trace: prd-root-command.md REQ-004

- [ ] **P3** | Quiet mode behavior across all commands | ~small
  - Acceptance: `--quiet` suppresses: progress indicators, confirmation messages (acts like `--force` for delete prompts per prd-root-command.md), informational messages, non-critical warnings. Still shows: requested data output, error messages, critical warnings. Verified across export/import progress, delete confirmations, asset download messages. Unit tests verify suppression.
  - Files: `internal/output/output.go` (modify), various `cmd/` files (modify)
  - Depends on: Output helpers, all command implementations
  - Spec trace: prd-root-command.md REQ-003

- [ ] **P3** | JSON error output in `--json` mode | ~small
  - Acceptance: When `--json` active, all errors output as structured JSON to stderr: `{"error":{"type":"...","message":"...","details":{...}}}`. Error types map to correct JSON `type` field. Works across all commands. Standard errors (non-typed) wrapped with `api_error` type. Unit tests verify JSON structure per error type.
  - Files: `cmd/root.go` (modify), `internal/errors/json.go` (modify)
  - Depends on: Error types and exit code infrastructure, Root command with global flags
  - Spec trace: prd-error-handling.md REQ-004

---

## Summary

| Priority | Count | Scope |
|----------|-------|-------|
| P0       | 7     | Scaffolding: errors, version, models, output, config, root cmd, API client |
| P1       | 16    | Core: version cmd, config cmds (3), bookmark API + cmds (8), tags API + cmds, user API + cmd |
| P2       | 10    | Enhanced: tag merge, bulk ops, open, bundles API + CRUD + view, assets API + CRUD + download + upload |
| P3       | 6     | Polish: export, import, cross-bookmark assets, verbose logging, quiet mode, JSON errors |
| **Total** | **39** | |

### Dependency Graph (Critical Path)

```
Error types ──┬──> Config subsystem ──┐
              │                       │
Version pkg ──┤                       ├──> Root command ──> All P1/P2/P3 commands
              │                       │
Data models ──┤                       │
              │                       │
Output helpers┴──> API client core ───┘
```

### Ambiguities Noted

1. **User API location**: prd-user.md specifies `internal/user/user.go` but all other API methods live on `internal/api/client.Client`. Recommend `internal/api/user.go` for consistency. Noted in User API task.
2. **`--quiet` and delete prompts**: prd-root-command.md says quiet suppresses confirmations. This implies `--quiet` on delete skips prompts (like `--force`). This interaction should be verified during implementation.
3. **Stdin JSON format**: prd-bookmarks.md says `--stdin` accepts both plain IDs and JSON arrays with auto-detection. Detection heuristic (try JSON parse first, fall back to line-delimited) should be documented in code.
