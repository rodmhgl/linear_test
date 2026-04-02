# ldctl Implementation Plan

## Overview

This document tracks the implementation progress of the ldctl CLI tool for LinkDing bookmark management. Tasks are prioritized based on dependencies and core functionality requirements.

## Implementation Phases

### Phase 0: Project Setup

- [ ] Project specifications created (9 PRDs in specs/)
- [ ] Build infrastructure (Makefile, golangci.yml, lefthook.yml)
- [ ] Documentation structure (AGENTS.md, CLAUDE.md, API reference)

### Phase 1: Foundation

**Priority: CRITICAL - Required for all other features**

- [ ] Initialize Go module (go.mod with github.com/rodmhgl/ldctl)
- [ ] Create main.go entry point
- [ ] Implement internal/version package
  - [ ] Version variables with ldflags support
  - [ ] Display formatting functions
  - [ ] Unit tests
- [ ] Implement internal/errors package
  - [ ] Error types and constructors
  - [ ] Exit code mapping
  - [ ] JSON error formatting
  - [ ] Unit tests
- [ ] Implement cmd/root.go
  - [ ] Root command with help display
  - [ ] Global flags (--json, --quiet, --verbose, --version, --help)
  - [ ] Mutual exclusivity validation
  - [ ] Output helper functions
- [ ] Implement cmd/version.go
  - [ ] Version command with --short and --json support
  - [ ] Integration with root command

### Phase 2: Configuration Management ✅ [COMPLETED: 2026-01-28]

**Priority: HIGH - Required for API operations**

- [ ] Implement internal/config package
  - [ ] Config loading from file and env vars
  - [ ] Config file creation and validation
  - [ ] Permission checking (0600)
  - [ ] Platform-specific path resolution
  - [ ] Unit tests with mocked filesystem
- [ ] Implement cmd/config.go (group command)
- [ ] Implement cmd/config_init.go
  - [ ] Interactive and non-interactive modes
  - [ ] URL normalization
  - [ ] Token masking
  - [ ] Verification against API
- [ ] Implement cmd/config_show.go
  - [ ] Display with source labels
  - [ ] Token masking
  - [ ] JSON output
- [ ] Implement cmd/config_test.go
  - [ ] Stepwise diagnostics
  - [ ] JSON output for automation

### Phase 3: API Client Foundation [IN PROGRESS]

**Priority: HIGH - Required for all API operations**

- [ ] Implement internal/api/client.go
  - [ ] HTTP client with token auth
  - [ ] Base URL configuration
  - [ ] Request/response helpers
  - [ ] Error mapping to internal/errors types
  - [ ] Rate limiting detection
  - [ ] Unit tests with mocked HTTP
- [ ] Implement internal/models package
  - [ ] Bookmark struct
  - [ ] Tag struct
  - [ ] Bundle struct
  - [ ] Asset struct
  - [ ] User profile struct
  - [ ] JSON marshaling/unmarshaling

### Phase 4: Core Features - Bookmarks [PENDING]

**Priority: HIGH - Primary use case**

- [ ] Implement internal/api/bookmarks.go
  - [ ] ListBookmarks with pagination
  - [ ] GetBookmark
  - [ ] CreateBookmark
  - [ ] UpdateBookmark
  - [ ] DeleteBookmark
  - [ ] CheckBookmark
  - [ ] ArchiveBookmark
  - [ ] UnarchiveBookmark
  - [ ] Unit tests
- [ ] Implement cmd/bookmarks.go (group command with 'bm' alias)
- [ ] Implement bookmark subcommands:
  - [ ] cmd/bookmarks_list.go
  - [ ] cmd/bookmarks_get.go
  - [ ] cmd/bookmarks_add.go
  - [ ] cmd/bookmarks_check.go
  - [ ] cmd/bookmarks_update.go
  - [ ] cmd/bookmarks_archive.go
  - [ ] cmd/bookmarks_unarchive.go
  - [ ] cmd/bookmarks_delete.go
  - [ ] cmd/bookmarks_open.go
- [ ] Implement internal/export package
  - [ ] JSON export
  - [ ] CSV export
  - [ ] HTML export
  - [ ] Import from JSON/HTML
- [ ] Implement cmd/bookmarks_export.go
- [ ] Implement cmd/bookmarks_import.go

### Phase 5: Supporting Features - Tags [PENDING]

**Priority: MEDIUM**

- [ ] Implement internal/api/tags.go
  - [ ] ListTags
  - [ ] GetTag
  - [ ] CreateTag
  - [ ] Unit tests
- [ ] Implement cmd/tags.go (group command)
- [ ] Implement tag subcommands:
  - [ ] cmd/tags_list.go
  - [ ] cmd/tags_get.go
  - [ ] cmd/tags_create.go

### Phase 6: User Profile [PENDING]

**Priority: MEDIUM**

- [ ] Implement internal/api/user.go
  - [ ] GetProfile
  - [ ] Unit tests
- [ ] Implement cmd/user.go (group command)
- [ ] Implement cmd/user_profile.go

### Phase 7: Advanced Features - Assets [PENDING]

**Priority: LOW**

- [ ] Implement internal/api/assets.go
  - [ ] ListAssets
  - [ ] GetAsset
  - [ ] DownloadAsset
  - [ ] UploadAsset
  - [ ] DeleteAsset
  - [ ] Unit tests
- [ ] Implement cmd/assets.go (group command)
- [ ] Implement asset subcommands:
  - [ ] cmd/assets_list.go
  - [ ] cmd/assets_get.go
  - [ ] cmd/assets_download.go
  - [ ] cmd/assets_upload.go
  - [ ] cmd/assets_delete.go

### Phase 8: Advanced Features - Bundles [PENDING]

**Priority: LOW**

- [ ] Implement internal/api/bundles.go
  - [ ] ListBundles
  - [ ] GetBundle
  - [ ] CreateBundle
  - [ ] UpdateBundle
  - [ ] DeleteBundle
  - [ ] ViewBundle (bookmarks)
  - [ ] Unit tests
- [ ] Implement cmd/bundles.go (group command)
- [ ] Implement bundle subcommands:
  - [ ] cmd/bundles_list.go
  - [ ] cmd/bundles_get.go
  - [ ] cmd/bundles_create.go
  - [ ] cmd/bundles_update.go
  - [ ] cmd/bundles_delete.go
  - [ ] cmd/bundles_view.go

### Phase 9: Integration and Polish [PENDING]

**Priority: FINAL**

- [ ] Integration tests with mocked LinkDing API
- [ ] End-to-end tests (build tag: integration)
- [ ] Performance optimization
- [ ] Documentation generation
- [ ] Release preparation

## Current Status

**Last Updated:** 2026-01-28

**Active Phase:** Phase 3 - API Client Foundation

**Next Steps:**

1. Implement internal/api/client.go with HTTP client and token auth
2. Implement internal/models package with all API data structures
3. Add comprehensive unit tests with mocked HTTP responses

## Testing Strategy

Each phase includes:

- Unit tests for all packages
- Integration tests for commands
- Manual testing against LinkDing instance
- Code coverage target: 80%+

## Success Metrics

Per PRD specifications:

- Commands complete in < 2s (network-bound)
- Export/import handle 10,000+ bookmarks
- Memory-efficient streaming
- Clean piping support (stdout/stderr separation)

## Notes

- Following specification-driven development
- Implementing completely (no placeholders/stubs)
- Maintaining backward compatibility
- Using only stdlib for HTTP (no third-party clients)
- No local caching/database (LinkDing is source of truth)
