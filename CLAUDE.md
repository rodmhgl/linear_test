# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ldctl**  is a CLI tool for interacting with the [LinkDing](https://github.com/sissbruecker/linkding) self-hosted bookmark manager via its REST API. The project is in its initial phase — the API reference specification exists but no implementation code has been written yet.

## Current State

- `linkding_api_reference.md` — Complete LinkDing REST API reference with agent instructions, endpoint documentation, and suggested CLI command hierarchy
- No language, framework, or build system has been chosen yet

## Target CLI Structure

The CLI should implement this command hierarchy (from the API reference):

```
ldctl
├── config (init, show, test)
├── bookmarks (alias: bm) — list, get, add, check, update, archive, unarchive, delete
├── assets — list, download, upload, delete
├── tags — list, get, create
├── bundles — list, get, create, update, delete
└── user — profile
```

## API Quirks to Remember

These non-obvious behaviors are critical when implementing commands:

1. **Duplicate URL on create** — `POST /api/bookmarks/` silently updates an existing bookmark if the URL matches. Use `/api/bookmarks/check/` first to detect duplicates.
2. **PUT acts like PATCH** — The bookmark PUT endpoint preserves unprovided fields instead of clearing them.
3. **Archived bookmarks are separate** — They live at `/api/bookmarks/archived/` and are excluded from the main list endpoint.
4. **Implicit tag creation** — Tags are auto-created when referenced in bookmark `tag_names` arrays.
5. **Bundle tag fields are space-separated strings**, not arrays (e.g., `"any_tags": "work productivity"`).

## Authentication

Token-based via `Authorization: Token <api_token>` header. Tokens are obtained from the LinkDing web UI Settings page.
