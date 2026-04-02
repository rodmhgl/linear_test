# PRD: ldctl Bundles Feature

## Executive Summary

ldctl needs a bundles command group to manage LinkDing saved search bundles from the terminal. Bundles are saved filter configurations that combine free-text search with tag-based inclusion/exclusion logic. The CLI provides full CRUD plus a `view` convenience command that lists bookmarks matching a bundle's filters.

## Problem Statement

**User pain point:** Users who organize bookmarks with complex tag-based filters need to create, update, and preview bundles without switching to the browser. Bundles are the primary way to build reusable bookmark views in LinkDing.

**Dependency:** Requires `config.Load()` from the config subsystem and the bookmark list API client (for `bundles view`).

## Goals & Success Metrics

| Goal | Metric | Target |
|------|--------|--------|
| Full API coverage | Bundle endpoints implemented | 6/6 endpoints |
| Consistency | Commands follow established PRD patterns | Same output format, flags, exit codes |
| Scriptability | Machine-readable output | All commands support `--json` |
| Cross-resource UX | `bundles view` works seamlessly | Output matches `bookmarks list` format |

## User Stories

### US-001: List Bundles

**As a** terminal user, **I want to** list all my bundles **so that I can** see available saved searches.

**Acceptance criteria:**

- `ldctl bundles list` returns paginated bundle list in key-value format
- `--limit` and `--offset` provide manual pagination
- `--all` auto-fetches all pages sequentially
- `--json` outputs raw JSON array
- Bundles separated by blank lines in default output
- Exit code 0 on success, 1 on API error

### US-002: View Single Bundle

**As a** terminal user, **I want to** retrieve a bundle by ID **so that I can** see its filter configuration.

**Acceptance criteria:**

- `ldctl bundles get <id>` displays bundle in key-value format
- All fields displayed including empty ones (shown as `(none)`)
- `--json` outputs raw JSON object
- Exit code 1 on 404

### US-003: Create Bundle

**As a** terminal user, **I want to** create a bundle **so that I can** save a reusable bookmark filter.

**Acceptance criteria:**

- `ldctl bundles create <name>` creates a bundle with the given name
- Optional flags: `--search`, `--any-tag`, `--all-tag`, `--exclude-tag`, `--order`
- Tag flags accept repeated values: `--any-tag reading --any-tag dev` → API receives `"any_tags": "reading dev"`
- `--order` omitted from request if not provided (API auto-assigns)
- Displays full created bundle on success
- `--json` outputs raw JSON object
- Exit code 0 on success, 1 on error

### US-004: Update Bundle

**As a** terminal user, **I want to** update bundle fields selectively **so that I can** modify specific attributes without re-specifying everything.

**Acceptance criteria:**

- `ldctl bundles update <id>` with any combination of flags
- Uses PATCH semantics: only sends provided flags to API
- Supports: `--name`, `--search`, `--any-tag`, `--all-tag`, `--exclude-tag`, `--order`
- Tag flags accept repeated values (same as create)
- To clear a field, pass empty string: `--search ''` or `--any-tags ''`
- Displays updated bundle on success
- `--json` outputs raw JSON object
- Exit code 0 on success, 1 on error

### US-005: Delete Bundle

**As a** terminal user, **I want to** delete a bundle with a safety net **so that I can** clean up without accidental removal.

**Acceptance criteria:**

- `ldctl bundles delete <id>` prompts for confirmation: `Delete bundle #5 "Work Resources"? [y/N]`
- `--force` skips confirmation
- Displays confirmation message on success
- Exit code 0 on success, 1 on error

### US-006: View Bundle Bookmarks

**As a** terminal user, **I want to** see the bookmarks matching a bundle **so that I can** preview a bundle's filter results without opening the browser.

**Acceptance criteria:**

- `ldctl bundles view <id>` lists bookmarks matching the bundle's filters
- Uses `/api/bookmarks/?bundle=<id>` endpoint under the hood
- Output format matches `bookmarks list` exactly (bookmark key-value format)
- `--limit`, `--offset`, `--all` pagination flags supported
- `--json` outputs raw JSON array of bookmark objects
- Exit code 1 if bundle not found

## Functional Requirements

| ID | Priority | Requirement |
|----|----------|-------------|
| REQ-001 | Must | `bundles list` with pagination (`--limit`, `--offset`, `--all`) |
| REQ-002 | Must | `bundles get <id>` retrieves single bundle |
| REQ-003 | Must | `bundles create <name>` with tag flags using repeated-flag pattern |
| REQ-004 | Must | `bundles update <id>` with PATCH semantics |
| REQ-005 | Must | `bundles delete <id>` with confirmation prompt and `--force` |
| REQ-006 | Must | All commands support `--json` flag |
| REQ-007 | Must | Key-value output format consistent with other PRDs |
| REQ-008 | Must | Empty fields displayed as `(none)` |
| REQ-009 | Should | `bundles view <id>` lists matching bookmarks with full pagination |
| REQ-010 | Should | Clear fields via empty string on update |

## Non-Functional Requirements

| Category | Requirement |
|----------|-------------|
| Consistency | Output format, exit codes, and flag conventions match bookmarks and tags PRDs |
| Error handling | Meaningful error messages on stderr. Non-zero exit codes on failure |
| Exit codes | 0 = success, 1 = error (API error, not found) |
| Output | stdout for data, stderr for errors/confirmations. Allows clean piping |

## Technical Considerations

### Bundle Object

```json
{
  "id": 1,
  "name": "Work Resources",
  "search": "productivity tools",
  "any_tags": "work productivity",
  "all_tags": "",
  "excluded_tags": "personal",
  "order": 0,
  "date_created": "2020-09-26T09:46:23.006313Z",
  "date_modified": "2020-09-26T16:01:14.275335Z"
}
```

### Architecture

```
cmd/
  bundles.go             # bundles subcommand group
  bundles_list.go        # bundles list
  bundles_get.go         # bundles get
  bundles_create.go      # bundles create
  bundles_update.go      # bundles update
  bundles_delete.go      # bundles delete
  bundles_view.go        # bundles view (cross-resource)
internal/
  api/
    bundles.go           # bundle API methods (List, Get, Create, Update, Delete)
```

### Display Format

Single bundle:

```
ID:             1
Name:           Work Resources
Search:         productivity tools
Any Tags:       work productivity
All Tags:       (none)
Excluded Tags:  personal
Order:          0
Created:        2020-09-26 09:46:23
Modified:       2020-09-26 16:01:14
```

All fields always shown. Empty values display `(none)`.

`bundles list` separates entries with blank lines. `--json` outputs raw API response.

`bundles view` output matches `bookmarks list` format exactly.

### Tag Flag Design

CLI uses repeated flags that get joined into space-separated strings for the API:

```bash
# Create with multiple any_tags
ldctl bundles create "Dev Resources" --any-tag golang --any-tag cli --exclude-tag archived

# API receives: { "name": "Dev Resources", "any_tags": "golang cli", "excluded_tags": "archived" }
```

To clear a tag field on update:

```bash
ldctl bundles update 5 --any-tags ''
# API receives: { "any_tags": "" }
```

### Cross-Resource: bundles view

`bundles view <id>` calls the bookmarks list endpoint with the bundle filter:

```
GET /api/bookmarks/?bundle=<id>&limit=<limit>&offset=<offset>
```

This reuses the existing bookmark list API client and output formatter. The `bundles_view.go` command delegates to bookmark list logic with an injected `bundle` query parameter.

### API Endpoints

| Command | Method | Endpoint | Parameters |
|---------|--------|----------|------------|
| `bundles list` | GET | `/api/bundles/` | `limit`, `offset` |
| `bundles get` | GET | `/api/bundles/<id>/` | — |
| `bundles create` | POST | `/api/bundles/` | `name` (required), `search`, `any_tags`, `all_tags`, `excluded_tags`, `order` |
| `bundles update` | PATCH | `/api/bundles/<id>/` | Any subset of writable fields |
| `bundles delete` | DELETE | `/api/bundles/<id>/` | — |
| `bundles view` | GET | `/api/bookmarks/?bundle=<id>` | `limit`, `offset` |

### API Quirks

- **Tag fields are space-separated strings**, not arrays. `"any_tags": "work productivity"` means tags `work` and `productivity`.
- **Bundle filter logic**: `any_tags` = OR, `all_tags` = AND, `excluded_tags` = NOT ANY.
- **No duplicate detection**: API may allow bundles with identical names.

## Implementation Roadmap

### Phase 1: Core CRUD

Prerequisite: config subsystem and API client exist (from config and bookmarks PRDs).

1. `internal/api/bundles.go` — List, Get, Create, Update, Delete methods on existing API client
2. `cmd/bundles.go` — bundles subcommand group
3. `cmd/bundles_list.go` — list with pagination
4. `cmd/bundles_get.go` — single bundle retrieval
5. `cmd/bundles_create.go` — bundle creation with repeated-flag tag input
6. `cmd/bundles_update.go` — PATCH update with field clearing support
7. `cmd/bundles_delete.go` — deletion with confirmation prompt

### Phase 2: Cross-Resource View

Prerequisite: bookmarks list command and output formatter exist (from bookmarks PRD).

1. `cmd/bundles_view.go` — list bookmarks matching bundle, delegating to bookmark list logic

## Known API Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No server-side bundle search/filter | Cannot filter bundles by name | Pipe `bundles list --json` to `jq` |
| No bulk delete | Must delete bundles one at a time | Script with a loop |
| Tag fields are strings not arrays | Multi-word tag names are ambiguous (`"my tag"` vs `"my"` + `"tag"`) | LinkDing tags should not contain spaces |

## Out of Scope

- Bundle alias (name is already short)
- Bulk bundle operations (multi-ID delete)
- Client-side bundle name filtering (`--query`)
- Bundle duplication/cloning
- Bookmarks, tags, assets, user, config — separate PRDs

## Open Questions & Risks

| # | Question/Risk | Status | Notes |
|---|---------------|--------|-------|
| 1 | Duplicate bundle name behavior | Open | API may allow duplicate names. Behavior will be documented during implementation |
| 2 | Multi-word tag names in space-separated fields | Low risk | LinkDing convention is single-word tags. If multi-word tags exist, behavior is undefined by the API |
| 3 | `bundles view` dependency on bookmarks implementation | Dependency | Phase 2 requires bookmarks list to be implemented first |

## Validation Checkpoints

| Checkpoint | Criteria |
|------------|----------|
| Phase 1 complete | All 5 CRUD commands work against a live LinkDing instance. `--json` works on all. Repeated tag flags produce correct space-separated strings. Update with empty string clears fields. Delete prompts for confirmation |
| Phase 2 complete | `bundles view` shows bookmarks matching the bundle. Pagination works. Output matches `bookmarks list` format. `--json` returns bookmark objects |
