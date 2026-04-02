# PRD: ldctl Tags Feature

## Executive Summary

ldctl needs a tags command group to list, retrieve, and create tags in LinkDing. Tags are a simple resource (id, name, date_added) with only three API endpoints and no update or delete support. The CLI mirrors the API surface directly with no client-side enrichment.

## Problem Statement

**User pain point:** Users managing large bookmark collections need to discover existing tags and explicitly create new ones without switching to the browser. While tags are implicitly created via bookmark commands, explicit tag management is needed for discovery and scripting workflows.

**Dependency:** Requires `config.Load()` from the config subsystem (separate PRD).

## Goals & Success Metrics

| Goal | Metric | Target |
|------|--------|--------|
| Full API coverage | Tag endpoints implemented | 3/3 endpoints |
| Consistency | Commands follow patterns from bookmarks PRD | Same output format, flags, exit codes |
| Scriptability | Machine-readable output | All commands support `--json` |

## User Stories

### US-001: List Tags

**As a** terminal user, **I want to** list all my tags **so that I can** discover existing tags before creating bookmarks.

**Acceptance criteria:**

- `ldctl tags list` returns paginated tag list in key-value format
- `--limit` and `--offset` provide manual pagination
- `--all` auto-fetches all pages sequentially
- `--json` outputs raw JSON array
- Tags separated by blank lines in default output
- Exit code 0 on success, 1 on API error

### US-002: View Single Tag

**As a** terminal user, **I want to** retrieve a tag by ID **so that I can** see its full details.

**Acceptance criteria:**

- `ldctl tags get <id>` displays tag in key-value format
- `--json` outputs raw JSON object
- Exit code 1 on 404

### US-003: Create Tag

**As a** terminal user, **I want to** explicitly create a tag **so that I can** set up tag taxonomy before adding bookmarks.

**Acceptance criteria:**

- `ldctl tags create <name>` creates a tag with the given name
- Displays created tag on success
- `--json` outputs raw JSON object
- Fires POST directly, surfaces API response on duplicate names (no pre-check)
- `--help` includes note: "Tags are also created implicitly when used in bookmark tag_names arrays"
- Exit code 0 on success, 1 on error

## Functional Requirements

| ID | Priority | Requirement |
|----|----------|-------------|
| REQ-001 | Must | `tags list` with pagination (`--limit`, `--offset`, `--all`) |
| REQ-002 | Must | `tags get <id>` retrieves single tag |
| REQ-003 | Must | `tags create <name>` creates tag via POST |
| REQ-004 | Must | All commands support `--json` flag |
| REQ-005 | Must | Key-value list output format consistent with bookmarks |
| REQ-006 | Must | `tags create` help text notes implicit creation via bookmark commands |

## Non-Functional Requirements

| Category | Requirement |
|----------|-------------|
| Consistency | Output format, exit codes, and flag conventions match bookmarks PRD |
| Error handling | Meaningful error messages on stderr. Non-zero exit codes on failure |
| Exit codes | 0 = success, 1 = error (API error, not found) |
| Output | stdout for data, stderr for errors. Allows clean piping |

## Technical Considerations

### Tag Object

```json
{
  "id": 1,
  "name": "example",
  "date_added": "2020-09-26T09:46:23.006313Z"
}
```

### Architecture

```
cmd/
  tags.go              # tags subcommand group
  tags_list.go         # tags list
  tags_get.go          # tags get
  tags_create.go       # tags create
internal/
  api/
    tags.go            # tag API methods (List, Get, Create)
```

### Display Format

Single tag:

```
ID:    1
Name:  example
Added: 2020-09-26 09:46:23
```

`tags list` separates entries with blank lines. `--json` outputs raw API response.

### API Endpoints

| Command | Method | Endpoint | Parameters |
|---------|--------|----------|------------|
| `tags list` | GET | `/api/tags/` | `limit`, `offset` |
| `tags get` | GET | `/api/tags/<id>/` | — |
| `tags create` | POST | `/api/tags/` | `name` (required) |

## Implementation Roadmap

### Phase 1: Core Commands

Prerequisite: config subsystem and API client exist (from config and bookmarks PRDs).

1. `internal/api/tags.go` — List, Get, Create methods on existing API client
2. `cmd/tags.go` — tags subcommand group
3. `cmd/tags_list.go` — list with pagination
4. `cmd/tags_get.go` — single tag retrieval
5. `cmd/tags_create.go` — tag creation with implicit-creation help note

## Known API Limitations

| Limitation | Impact | Workaround |
|------------|--------|------------|
| No delete endpoint | Typo tags cannot be removed via CLI | Must use LinkDing web UI |
| No update/rename endpoint | Tags cannot be renamed via CLI | Create new tag, re-tag bookmarks manually, orphaned tag remains |
| No usage counts | Cannot show how many bookmarks use a tag | Would require fetching all bookmarks per tag — not practical |
| No server-side search | `tags list` cannot filter by name on the server | Pipe to `grep` for client-side filtering |

## Out of Scope

- Client-side name filtering (`--query`)
- Tag usage/bookmark count enrichment
- Tag delete or rename workarounds
- Tag alias (command is already short)
- Bundles, bookmarks, assets, user, config — separate PRDs

## Open Questions & Risks

| # | Question/Risk | Status | Notes |
|---|---------------|--------|-------|
| 1 | Duplicate tag creation behavior | Open | API may return existing tag or error on duplicate name. Behavior will be documented during implementation |
| 2 | Tag name validation | Low risk | API likely handles validation; CLI passes through |

## Validation Checkpoints

| Checkpoint | Criteria |
|------------|----------|
| Phase 1 complete | All 3 commands work against a live LinkDing instance. `--json` works on all. `tags list --all` fetches multiple pages. `tags create` help text includes implicit creation note |
