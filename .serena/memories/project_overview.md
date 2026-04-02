# ldctl Project Overview

## Project Purpose
ldctl is a command-line interface (CLI) tool for managing bookmarks in a LinkDing instance (self-hosted bookmark manager). It interacts with LinkDing's REST API to provide a terminal-based interface for bookmark management operations.

## Tech Stack
- **Language**: Go 1.25
- **CLI Framework**: cobra (for command structure)
- **HTTP Client**: net/http (standard library)
- **Configuration**: viper (for config file + environment variables)
- **Output**: tablewriter for formatted lists, JSON for scriptable output
- **Testing**: testify for test assertions
- **Module**: github.com/rodmhgl/ldctl

## Architecture
- Single binary with no external dependencies
- Configuration via ~/.config/ldctl/config.yaml or environment variables
- All commands support --json flag for scriptable output
- Exit codes: 0=success, 1=error, 2=config error

## Command Structure
```
ldctl
├── config (init, show, test)
├── bookmarks (alias: bm) — list, get, add, check, update, archive, unarchive, delete
├── assets — list, download, upload, delete
├── tags — list, get, create
├── bundles — list, get, create, update, delete
└── user — profile
```

## Authentication
Token-based authentication via `Authorization: Token <api_token>` header. Tokens are obtained from the LinkDing web UI Settings page.

## API Quirks to Remember
1. **Duplicate URL on create** — POST /api/bookmarks/ silently updates existing bookmark if URL matches
2. **PUT acts like PATCH** — Bookmark PUT endpoint preserves unprovided fields
3. **Archived bookmarks are separate** — They live at /api/bookmarks/archived/
4. **Implicit tag creation** — Tags auto-created when referenced in bookmark tag_names arrays
5. **Bundle tag fields are space-separated strings**, not arrays