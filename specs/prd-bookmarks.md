# PRD: ldctl Bookmarks Feature

## Executive Summary

ldctl needs a complete bookmarks command group to manage LinkDing bookmarks from the terminal. This PRD covers all CRUD operations mirroring the LinkDing API, plus value-add features (bulk ops, import/export, browser open, smart tag management) that make the CLI more powerful than the web UI for power users.

## Problem Statement

**User pain point:** LinkDing users who live in the terminal have no CLI interface for managing bookmarks. Every interaction requires switching to the browser, breaking workflow.

**Business impact:** Without a CLI, LinkDing can't be integrated into shell scripts, automation pipelines, or terminal-first workflows — limiting its utility for technical users.

## Goals & Success Metrics

| Goal | Metric | Target |
|------|--------|--------|
| Full API coverage | Bookmark endpoints implemented | 10/10 endpoints |
| CLI-native experience | Commands follow Unix conventions | All commands support --json, stdin piping, exit codes |
| Value beyond API | Features not possible in web UI | 5 value-add features shipped |
| Scriptability | Machine-readable output | All commands support --json flag |

## User Stories

### US-001: List and Search Bookmarks

**As a** terminal user, **I want to** list and search my bookmarks with filters **so that I can** quickly find what I'm looking for without opening a browser.

**Acceptance criteria:**

- Default output shows first page (100 items) in key-value list format
- `--query` supports LinkDing search syntax (`#tag`, `!#exclude`, `title:`, `url:`, etc.)
- `--archived` switches to archived bookmarks endpoint
- `--limit` and `--offset` provide manual pagination
- `--all` auto-fetches all pages sequentially
- `--json` outputs raw JSON array
- Non-zero exit code on API errors

### US-002: View Single Bookmark

**As a** terminal user, **I want to** retrieve a bookmark by ID **so that I can** see its full details.

**Acceptance criteria:**

- `ldctl bookmarks get <id>` displays full bookmark in key-value format
- `--json` outputs raw JSON object
- Exit code 1 on 404

### US-003: Add Bookmarks

**As a** terminal user, **I want to** add bookmarks from the command line **so that I can** save URLs without leaving my workflow.

**Acceptance criteria:**

- `ldctl bookmarks add <url>` creates a bookmark with the given URL
- Optional flags: `--title`, `--description`, `--notes`, `--tags "tag1 tag2"`, `--archived`, `--unread`, `--shared`
- Tags are space-separated: `--tags "go cli tools"`
- Duplicate URLs silently upsert (matches API behavior)
- Displays created/updated bookmark on success
- Exit code 0 on success, 1 on error

### US-004: Check URL

**As a** terminal user, **I want to** check if a URL is already bookmarked **so that I can** see existing bookmark data, scraped metadata, and auto-tags before deciding to add it.

**Acceptance criteria:**

- `ldctl bookmarks check <url>` calls the `/check` endpoint
- Shows existing bookmark if found, or "Not bookmarked" message
- Displays scraped metadata (title, description) and auto-tags
- `--json` outputs raw JSON response

### US-005: Update Bookmarks

**As a** terminal user, **I want to** update bookmark fields selectively **so that I can** modify specific attributes without resending everything.

**Acceptance criteria:**

- `ldctl bookmarks update <id>` with any combination of flags
- Uses PATCH semantics: only sends provided flags to API
- Supports: `--title`, `--description`, `--notes`, `--tags`, `--unread`, `--shared`
- `--add-tags "newtag"` fetches current tags, merges, and patches
- `--remove-tags "oldtag"` fetches current tags, removes, and patches
- `--add-tags` and `--remove-tags` can be used together
- `--tags` replaces entire tag list (cannot combine with `--add-tags`/`--remove-tags`)
- Displays updated bookmark on success

### US-006: Archive and Unarchive

**As a** terminal user, **I want to** archive and unarchive bookmarks **so that I can** manage my bookmark lifecycle.

**Acceptance criteria:**

- `ldctl bookmarks archive <id>` archives a bookmark
- `ldctl bookmarks unarchive <id>` unarchives a bookmark
- Both support multiple IDs: `ldctl bookmarks archive 1 2 3`
- Both support `--stdin` for piped input
- Displays confirmation message per bookmark
- Exit code 0 on success, 1 if any operation fails

### US-007: Delete Bookmarks

**As a** terminal user, **I want to** delete bookmarks with a safety net **so that I can** clean up without accidental data loss.

**Acceptance criteria:**

- `ldctl bookmarks delete <id>` prompts for confirmation: `Delete bookmark #42 "Title"? [y/N]`
- `--force` skips confirmation
- Supports multiple IDs: `ldctl bookmarks delete 1 2 3`
- Supports `--stdin` for piped input
- With multiple IDs, confirms each unless `--force`
- Displays confirmation message per deleted bookmark
- Exit code 0 on success, 1 if any deletion fails

### US-008: Open in Browser

**As a** terminal user, **I want to** open a bookmark's URL in my browser **so that I can** quickly visit a saved link.

**Acceptance criteria:**

- `ldctl bookmarks open <id>` fetches bookmark and opens URL in default browser
- Uses `xdg-open` (Linux), `open` (macOS), or `start` (Windows)
- Displays "Opening: <url>" message
- Exit code 1 if bookmark not found

### US-009: Export Bookmarks

**As a** terminal user, **I want to** export all my bookmarks **so that I can** back them up or import them into a browser.

**Acceptance criteria:**

- `ldctl bookmarks export` auto-paginates and fetches all bookmarks
- `--format json` (default): JSON array of bookmark objects
- `--format csv`: CSV with header row
- `--format html`: Netscape bookmark format (browser-importable)
- `--output <file>`: write to file instead of stdout
- `--archived`: include archived bookmarks (fetches from both endpoints)
- Progress indicator on stderr during multi-page fetch

### US-010: Import Bookmarks

**As a** terminal user, **I want to** import bookmarks from a file **so that I can** migrate from browsers or restore from backup.

**Acceptance criteria:**

- `ldctl bookmarks import <file>` reads file and bulk-creates bookmarks
- `--format html`: parse Netscape HTML bookmark format (Chrome/Firefox export)
- `--format json`: parse JSON array matching ldctl export format
- Auto-detects format if `--format` not specified
- Displays progress: `Imported 42/100 bookmarks...`
- Reports failures without stopping: `Failed: <url> - <error>`
- Summary on completion: `Imported 98/100 (2 failures)`

## Functional Requirements

| ID | Priority | Requirement |
|----|----------|-------------|
| REQ-001 | Must | `bookmarks list` with pagination (`--limit`, `--offset`, `--all`), search (`--query`), and archive filter (`--archived`) |
| REQ-002 | Must | `bookmarks get <id>` retrieves single bookmark |
| REQ-003 | Must | `bookmarks add <url>` creates bookmark with optional metadata flags |
| REQ-004 | Must | `bookmarks check <url>` queries check endpoint |
| REQ-005 | Must | `bookmarks update <id>` with PATCH semantics and selective field flags |
| REQ-006 | Must | `bookmarks archive <id>` and `bookmarks unarchive <id>` |
| REQ-007 | Must | `bookmarks delete <id>` with confirmation prompt and `--force` |
| REQ-008 | Must | All commands support `--json` flag for machine-readable output |
| REQ-009 | Must | `bm` alias for `bookmarks` subcommand |
| REQ-010 | Must | Human-readable key-value list output format (see Display Format section) |
| REQ-011 | Should | `--add-tags` and `--remove-tags` on update with fetch-merge-patch logic |
| REQ-012 | Should | Bulk operations: multi-ID args for delete, archive, unarchive |
| REQ-013 | Should | `--stdin` flag for piped input on bulk operations |
| REQ-014 | Should | `bookmarks open <id>` opens URL in default browser |
| REQ-015 | Should | `bookmarks export` with `--format json|csv|html` and `--output` |
| REQ-016 | Should | `bookmarks import <file>` with JSON and Netscape HTML support |
| REQ-017 | Could | Auto-detect import format when `--format` not specified |
| REQ-018 | Could | Progress indicator on stderr for export/import operations |

## Non-Functional Requirements

| Category | Requirement |
|----------|-------------|
| Performance | All single-resource commands complete in < 2s (network-bound) |
| Performance | Export/import handle 10,000+ bookmarks without memory issues (stream, don't buffer) |
| Compatibility | Linux, macOS, Windows. Single static binary via `go build` |
| Error handling | Meaningful error messages on stderr. Non-zero exit codes on failure |
| Exit codes | 0 = success, 1 = error (API error, not found, validation failure) |
| Output | stdout for data, stderr for progress/errors/confirmations. Allows clean piping |
| Auth | Read token from config file, then env var `LINKDING_TOKEN`. Env var overrides config. Same for `LINKDING_URL` |

## Technical Considerations

### Architecture

```
cmd/
  root.go              # root command, global flags (--json)
  bookmarks.go         # bookmarks subcommand group + alias
  bookmarks_list.go    # each subcommand in its own file
  bookmarks_get.go
  bookmarks_add.go
  bookmarks_check.go
  bookmarks_update.go
  bookmarks_archive.go
  bookmarks_delete.go
  bookmarks_open.go
  bookmarks_export.go
  bookmarks_import.go
internal/
  api/
    client.go          # HTTP client, auth, base URL
    bookmarks.go       # bookmark API methods
  config/
    config.go          # config file + env var loading
  format/
    output.go          # key-value list formatter, JSON output
    netscape.go        # Netscape HTML bookmark parser/writer
    csv.go             # CSV formatter
```

### API Client Design

- Single `Client` struct holding base URL, token, http.Client
- Each API resource gets its own file with typed methods
- Methods return Go structs, not raw JSON
- Pagination handled by a generic `Paginate` helper that yields pages

### Display Format

Default human-readable format for single bookmark:

```
ID:          1
URL:         https://example.com/article
Title:       Article Title Here
Description: Article description text here
Tags:        kubernetes, self-hosted
Added:       2025-11-08 01:30:19
Modified:    2025-11-18 14:46:55
Unread:      false
Shared:      false
Archived:    true
```

For `list` output, bookmarks are separated by blank lines. Tags displayed comma-separated (even though input is space-separated) for readability.

### Tag Merge Logic (--add-tags / --remove-tags)

```
1. GET /api/bookmarks/<id>/  →  current tag_names
2. Merge: add new tags, remove specified tags (case-insensitive match)
3. PATCH /api/bookmarks/<id>/  →  { "tag_names": merged_list }
```

`--tags` is mutually exclusive with `--add-tags`/`--remove-tags`. CLI validates this before making any API calls.

### Stdin Piping

`--stdin` reads newline-delimited bookmark IDs from stdin:

```bash
ldctl bookmarks list --json | jq -r '.[].id' | ldctl bookmarks archive --stdin
```

Alternatively, reads JSON array and extracts `.id` from each object if input is valid JSON.

### Config Precedence

```
CLI flag > Environment variable > Config file > Default
```

Config subsystem is a separate PRD. For this feature, assume a `config.Load()` function exists that returns `(baseURL, token, error)`.

## Implementation Roadmap

### Phase 1: Core CRUD

Prerequisite: config subsystem exists (separate PRD).

1. API client (`internal/api/client.go`, `internal/api/bookmarks.go`)
2. Output formatter (`internal/format/output.go`)
3. `bookmarks list` with pagination and filters
4. `bookmarks get`
5. `bookmarks add`
6. `bookmarks check`
7. `bookmarks update` (basic — `--tags` only, no add/remove)
8. `bookmarks archive` / `bookmarks unarchive` (single ID)
9. `bookmarks delete` with confirmation

### Phase 2: Value-Add

1. `--add-tags` and `--remove-tags` on update
2. Bulk operations (multi-ID args + `--stdin`)
3. `bookmarks open`

### Phase 3: Import/Export

1. `bookmarks export` with JSON format
2. CSV export format
3. Netscape HTML export format
4. `bookmarks import` with JSON format
5. Netscape HTML import
6. Auto-detect import format

## Out of Scope

- Config subsystem (`config init`, `config show`, `config test`) — separate PRD
- Tags subcommand (`tags list`, `tags get`, `tags create`)
- Bundles subcommand
- Assets subcommand
- User subcommand
- Shell completions (future enhancement)
- Interactive/TUI mode (future enhancement)
- Bookmark notes editing via `$EDITOR` (future enhancement)

## Open Questions & Risks

| # | Question/Risk | Status | Notes |
|---|---------------|--------|-------|
| 1 | Config subsystem must ship before or alongside bookmarks | Open | Bookmarks depend on `config.Load()` for auth. Could stub with env-var-only support initially |
| 2 | Netscape HTML bookmark format parsing complexity | Low risk | Well-documented format, Go HTML parsing is straightforward |
| 3 | Large export memory usage | Low risk | Stream pages to output instead of buffering all bookmarks in memory |
| 4 | stdin format ambiguity | Open | Should `--stdin` accept plain IDs, JSON objects, or both? PRD specifies both with auto-detection |

## Validation Checkpoints

| Checkpoint | Criteria |
|------------|----------|
| Phase 1 complete | All 8 core commands work against a live LinkDing instance. `--json` works on all. Manual testing pass |
| Phase 2 complete | Tag merge logic handles edge cases (empty tags, duplicates). Bulk ops work with 50+ IDs. Stdin piping works in a pipeline |
| Phase 3 complete | Round-trip test: export JSON → import JSON produces identical bookmarks. Netscape HTML imports into Chrome/Firefox successfully |
