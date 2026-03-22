# PRD: ldctl Assets Feature

## Executive Summary

ldctl needs an assets command group to manage file attachments (HTML snapshots and user uploads) associated with LinkDing bookmarks. Assets are nested under bookmarks in the API, so every operation requires a parent bookmark ID. This PRD covers five commands (list, get, download, upload, delete) plus a cross-bookmark listing mode that aggregates assets across all bookmarks.

## Problem Statement

**User pain point:** LinkDing users who use HTML snapshots or file uploads have no CLI access to these assets. Downloading snapshots for offline archiving, uploading files to bookmarks, or auditing asset usage across bookmarks all require the web UI.

**Business impact:** Without asset management, ldctl can't serve users who rely on LinkDing's archival features — a key differentiator for self-hosted bookmark managers. Automation workflows (scheduled snapshot downloads, bulk uploads) are impossible.

## Goals & Success Metrics

| Goal | Metric | Target |
|------|--------|--------|
| Full API coverage | Asset endpoints implemented | 5/5 endpoints |
| CLI-native experience | Commands follow Unix conventions | All commands support --json, exit codes |
| Cross-bookmark visibility | List assets across all bookmarks | `--all` flag on list |
| Safe downloads | Status-aware download with pre-check | Status check before download attempt |

## User Stories

### US-001: List Assets for a Bookmark

**As a** terminal user, **I want to** list all assets attached to a bookmark **so that I can** see what snapshots and uploads exist.

**Acceptance criteria:**

- `ldctl assets list <bookmark-id>` displays assets in compact table format
- Table columns: ID, Type, Status, Content-Type, Name, Created
- `--json` outputs raw JSON array
- Non-zero exit code on API errors or invalid bookmark ID

### US-002: List Assets Across All Bookmarks

**As a** terminal user, **I want to** list assets across all my bookmarks **so that I can** audit snapshot coverage or find uploaded files.

**Acceptance criteria:**

- `ldctl assets list --all` iterates all bookmarks and aggregates their assets
- Table adds Bookmark ID and Bookmark Title columns
- Progress indicator on stderr during multi-bookmark fetch
- `--json` outputs flat JSON array with bookmark context included
- `--limit` and `--offset` apply to bookmark pagination (controls which bookmarks are scanned)

### US-003: View Single Asset Metadata

**As a** terminal user, **I want to** view an asset's metadata **so that I can** check its status, type, and content-type before downloading.

**Acceptance criteria:**

- `ldctl assets get <bookmark-id> <asset-id>` displays asset in key-value format
- Fields: ID, Bookmark, Type, Status, Content-Type, Name, Created
- `--json` outputs raw JSON object
- Exit code 1 on 404

### US-004: Download Asset

**As a** terminal user, **I want to** download an asset file **so that I can** save snapshots locally or retrieve uploaded files.

**Acceptance criteria:**

- `ldctl assets download <bookmark-id> <asset-id>` saves file to current directory
- Default filename: `asset-{bookmark_id}-{asset_id}.{ext}` where ext is derived from content-type
- `--output <path>` writes to a specific file path
- `--output-dir <dir>` writes auto-named file into specified directory
- `--output` and `--output-dir` are mutually exclusive
- Before downloading, fetches asset metadata via GET to check status
- If status is not `complete`: displays message ("Asset #X is not available (status: pending)") and exits 1
- `--force` skips the status check and attempts download regardless
- Displays "Saved: <filepath>" on success
- Exit code 1 on error (404, status check failure, write failure)

### US-005: Upload Asset

**As a** terminal user, **I want to** upload a file to a bookmark **so that I can** attach documents, PDFs, or other files.

**Acceptance criteria:**

- `ldctl assets upload <bookmark-id> <file>` uploads file via multipart/form-data
- Content-Type inferred from file extension using Go's `mime.TypeByExtension`
- Falls back to `application/octet-stream` if extension is unknown
- Displays created asset metadata on success
- Exit code 1 if file doesn't exist, bookmark not found, or upload fails

### US-006: Delete Asset

**As a** terminal user, **I want to** delete an asset **so that I can** clean up snapshots or remove uploaded files.

**Acceptance criteria:**

- `ldctl assets delete <bookmark-id> <asset-id>` prompts for confirmation: `Delete asset #7 from bookmark #42? [y/N]`
- `--force` skips confirmation
- Displays confirmation message on success
- Exit code 0 on success, 1 on failure

## Functional Requirements

| ID | Priority | Requirement |
|----|----------|-------------|
| REQ-001 | Must | `assets list <bookmark-id>` with compact table output |
| REQ-002 | Must | `assets get <bookmark-id> <asset-id>` with key-value output |
| REQ-003 | Must | `assets download <bookmark-id> <asset-id>` with auto-naming |
| REQ-004 | Must | `assets upload <bookmark-id> <file>` with MIME inference |
| REQ-005 | Must | `assets delete <bookmark-id> <asset-id>` with confirmation prompt |
| REQ-006 | Must | All commands support `--json` flag |
| REQ-007 | Must | `--output` flag on download for explicit file path |
| REQ-008 | Must | `--output-dir` flag on download for target directory with auto-naming |
| REQ-009 | Must | `--output` and `--output-dir` are mutually exclusive |
| REQ-010 | Must | Download pre-checks asset status via GET before attempting download |
| REQ-011 | Must | `--force` on download skips status check |
| REQ-012 | Must | `--force` on delete skips confirmation prompt |
| REQ-013 | Should | `assets list --all` aggregates assets across all bookmarks |
| REQ-014 | Should | `--all` table includes Bookmark ID and Bookmark Title columns |
| REQ-015 | Should | Progress indicator on stderr during `--all` multi-bookmark fetch |

## Non-Functional Requirements

| Category | Requirement |
|----------|-------------|
| Performance | Single-asset commands complete in < 2s (network-bound) |
| Performance | `--all` listing handles 1000+ bookmarks without memory issues (stream per-bookmark) |
| Compatibility | Linux, macOS, Windows. Single static binary via `go build` |
| Error handling | Meaningful error messages on stderr. Non-zero exit codes on failure |
| Exit codes | 0 = success, 1 = error (API error, not found, write failure) |
| Output | stdout for data, stderr for progress/errors/confirmations. Allows clean piping |
| File safety | Download refuses to overwrite existing files unless `--force` is passed |

## Technical Considerations

### Architecture

```
cmd/
  assets.go              # assets subcommand group
  assets_list.go         # assets list
  assets_get.go          # assets get
  assets_download.go     # assets download
  assets_upload.go       # assets upload
  assets_delete.go       # assets delete
internal/
  api/
    assets.go            # asset API methods
  format/
    table.go             # compact table formatter (reusable)
```

### API Client Methods

```go
// internal/api/assets.go

func (c *Client) ListAssets(bookmarkID int) ([]Asset, error)
func (c *Client) GetAsset(bookmarkID, assetID int) (*Asset, error)
func (c *Client) DownloadAsset(bookmarkID, assetID int) (io.ReadCloser, string, error) // body, content-type, error
func (c *Client) UploadAsset(bookmarkID int, filename string, body io.Reader, contentType string) (*Asset, error)
func (c *Client) DeleteAsset(bookmarkID, assetID int) error
```

### Display Formats

**`assets list` default output (per-bookmark):**

```
ID   Type      Status    Content-Type  Name                              Created
3    snapshot  complete  text/html     HTML snapshot from 10/01/2023     2023-10-01
7    upload    complete  application/pdf  project-notes.pdf              2023-11-15
```

**`assets list --all` output (cross-bookmark):**

```
Bookmark  Title                  ID   Type      Status    Content-Type     Name                           Created
42        Example Article        3    snapshot  complete  text/html        HTML snapshot from 10/01/2023  2023-10-01
42        Example Article        7    upload    complete  application/pdf  project-notes.pdf              2023-11-15
89        Go Documentation       12   snapshot  pending   text/html        HTML snapshot from 11/20/2023  2023-11-20
```

**`assets get` default output:**

```
ID:           7
Bookmark:     42
Type:         upload
Status:       complete
Content-Type: application/pdf
Name:         project-notes.pdf
Created:      2023-11-15 09:30:00
```

### Download Filename Strategy

Default pattern: `asset-{bookmark_id}-{asset_id}.{ext}`

Extension derivation from content-type:

| Content-Type | Extension |
|--------------|-----------|
| `text/html` | `.html` |
| `application/pdf` | `.pdf` |
| `image/png` | `.png` |
| `image/jpeg` | `.jpg` |
| Unknown | `.bin` |

Use Go's `mime.ExtensionsByType` for the mapping, falling back to `.bin`.

### Status Check Flow (Download)

```
1. GET /api/bookmarks/<bid>/assets/<aid>/  →  asset metadata
2. If status != "complete" AND --force not set:
   - Print: "Asset #<aid> is not available (status: <status>)"
   - Exit 1
3. If status == "complete" OR --force set:
   - GET /api/bookmarks/<bid>/assets/<aid>/download/
   - Write to file
```

### Cross-Bookmark Listing Flow

```
1. Paginate GET /api/bookmarks/ (all pages)
2. For each bookmark, GET /api/bookmarks/<id>/assets/
3. Merge results, adding bookmark ID and title to each asset row
4. Print stderr progress: "Scanning bookmark 42/150..."
```

### Upload Content-Type Inference

```go
contentType := mime.TypeByExtension(filepath.Ext(filename))
if contentType == "" {
    contentType = "application/octet-stream"
}
```

### Overwrite Protection

Download refuses to write if target file already exists:

```
Error: file already exists: asset-42-7.html
Use --force to overwrite.
```

`--force` on download serves double duty: skips status check AND allows overwrite.

## Implementation Roadmap

### Phase 1: Core CRUD

Prerequisite: API client and config subsystem exist.

1. `internal/api/assets.go` — asset API methods (list, get, download, upload, delete)
2. `internal/format/table.go` — compact table formatter
3. `cmd/assets.go` — assets subcommand group
4. `assets list <bookmark-id>` with table output
5. `assets get <bookmark-id> <asset-id>` with key-value output
6. `assets download` with auto-naming, status check, `--output`, `--output-dir`, `--force`
7. `assets upload` with MIME inference
8. `assets delete` with confirmation + `--force`

### Phase 2: Cross-Bookmark

1. `assets list --all` with bookmark columns and progress indicator

## Out of Scope

- Bulk download (all assets for a bookmark in one command)
- Bulk upload (multiple files in one command)
- Bulk delete (multiple asset IDs in one command)
- Asset search/filter by type or content-type
- `--stdin` piping for asset IDs
- Asset rename/update metadata
- Automatic snapshot creation (server-side feature, not API-exposed)
- Asset content preview in terminal

## Open Questions & Risks

| # | Question/Risk | Status | Notes |
|---|---------------|--------|-------|
| 1 | Cross-bookmark listing performance with many bookmarks | Low risk | Stream per-bookmark, don't buffer all results. Progress indicator keeps user informed |
| 2 | `mime.ExtensionsByType` returns multiple extensions for some types | Low risk | Pick first result, fall back to `.bin` |
| 3 | Large file uploads may need progress indicator | Open | Go's multipart writer doesn't natively support progress. Could wrap the reader with a counting wrapper. Defer to implementation |
| 4 | `--force` flag overloaded (skip status check + allow overwrite) | Accepted | Single flag for "just do it" semantics. Simpler than two separate flags |

## Validation Checkpoints

| Checkpoint | Criteria |
|------------|----------|
| Phase 1 complete | All 5 commands work against a live LinkDing instance. Table and key-value output correct. Download creates files with correct content. Upload attaches file to bookmark. Delete prompts and removes. `--json` works on all commands |
| Phase 2 complete | `--all` listing scans all bookmarks and produces merged table with bookmark context. Progress indicator visible on stderr. Handles bookmarks with zero assets gracefully |
