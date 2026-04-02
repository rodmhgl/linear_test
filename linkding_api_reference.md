# LinkDing REST API Reference

## Agent Instructions

This document provides a complete reference for the LinkDing REST API. Use this when implementing CLI commands for `ldctl`. All endpoints are relative to the LinkDing instance base URL (e.g., `https://linkding.example.com`).

---

## Authentication

**Method:** Token-based authentication via HTTP header

**Header Format:**

```
Authorization: Token <api_token>
```

**Token Acquisition:** Users obtain their API token from the LinkDing Settings page in the web UI.

**Error Response (401 Unauthorized):** Returned when token is missing, invalid, or expired.

---

## Common Patterns

### Pagination

All list endpoints support pagination with consistent parameters:

- `limit` (integer): Maximum results to return. Default: `100`
- `offset` (integer): Starting index for results (0-based)

Paginated responses include:

```json
{
  "count": 123,           // Total number of items
  "next": "url|null",     // URL for next page, null if no more
  "previous": "url|null", // URL for previous page, null if first
  "results": [...]        // Array of items
}
```

### HTTP Methods

- `GET` - Retrieve resources
- `POST` - Create resources
- `PUT` - Full update (all required fields must be provided)
- `PATCH` - Partial update (only provided fields are modified)
- `DELETE` - Remove resources

### Response Codes

- `200 OK` - Successful GET, PUT, PATCH
- `201 Created` - Successful POST
- `204 No Content` - Successful DELETE
- `400 Bad Request` - Invalid request body or parameters
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Resource does not exist

---

## Resources

## 1. Bookmarks

Base path: `/api/bookmarks/`

### Bookmark Object Schema

```json
{
  "id": 1,                                    // integer, read-only
  "url": "https://example.com",               // string, required
  "title": "Example title",                   // string, optional (auto-scraped if empty)
  "description": "Example description",       // string, optional (auto-scraped if empty)
  "notes": "Example notes",                   // string, optional, private notes
  "web_archive_snapshot_url": "https://...",  // string, read-only, Internet Archive URL
  "favicon_url": "http://...",                // string, read-only
  "preview_image_url": "http://...",          // string, read-only
  "is_archived": false,                       // boolean, default: false
  "unread": false,                            // boolean, default: false
  "shared": false,                            // boolean, default: false
  "tag_names": ["tag1", "tag2"],              // array of strings
  "date_added": "2020-09-26T09:46:23.006313Z",    // ISO 8601, read-only
  "date_modified": "2020-09-26T16:01:14.275335Z"  // ISO 8601, read-only
}
```

### Endpoints

#### List Bookmarks

```
GET /api/bookmarks/
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `q` | string | Search query (uses same logic as web UI search) |
| `limit` | integer | Max results, default 100 |
| `offset` | integer | Starting index |
| `modified_since` | ISO 8601 datetime | Filter: bookmarks modified after this date |
| `added_since` | ISO 8601 datetime | Filter: bookmarks added after this date |
| `bundle` | integer | Filter: only bookmarks matching bundle ID |

**Response:** Paginated list of bookmark objects

---

#### List Archived Bookmarks

```
GET /api/bookmarks/archived/
```

**Parameters:** Same as List Bookmarks

**Response:** Paginated list of archived bookmark objects

**Note:** This is a separate endpoint from the main list - archived bookmarks are NOT included in `/api/bookmarks/` results.

---

#### Retrieve Single Bookmark

```
GET /api/bookmarks/<id>/
```

**Path Parameters:**

- `id` (integer): Bookmark ID

**Response:** Single bookmark object

---

#### Check URL

```
GET /api/bookmarks/check/?url=<encoded_url>
```

**Query Parameters:**

- `url` (string, required): URL-encoded URL to check

**Response:**

```json
{
  "bookmark": { ... } | null,  // Existing bookmark if URL is bookmarked, else null
  "metadata": {
    "title": "Scraped website title",
    "description": "Scraped website description"
  },
  "auto_tags": ["tag1", "tag2"]  // Tags that would be auto-assigned
}
```

**Use Cases:**

- Check if URL already exists before creating
- Preview what metadata would be scraped
- See which auto-tags would be applied

---

#### Create Bookmark

```
POST /api/bookmarks/
```

**Request Body:**

```json
{
  "url": "https://example.com",      // required
  "title": "Example title",          // optional
  "description": "Example description", // optional
  "notes": "Example notes",          // optional
  "is_archived": false,              // optional, default: false
  "unread": false,                   // optional, default: false
  "shared": false,                   // optional, default: false
  "tag_names": ["tag1", "tag2"]      // optional
}
```

**Query Parameters:**

- `disable_scraping` (flag): If present, disables automatic title/description scraping

**Behavior Notes:**

- If `title` or `description` are empty/omitted, LinkDing auto-scrapes them from the URL
- **IMPORTANT:** If URL already exists, this SILENTLY UPDATES the existing bookmark instead of creating new. No error is returned. Use `/check` endpoint first if you need to detect duplicates.
- Setting `is_archived: true` saves directly to archive

**Response:** Created/updated bookmark object

---

#### Update Bookmark (Full)

```
PUT /api/bookmarks/<id>/
```

**Path Parameters:**

- `id` (integer): Bookmark ID

**Request Body:** Same as Create, but `url` is required at minimum

**Behavior:** Fields not provided are NOT modified (contrary to typical PUT semantics)

**Error:** Returns error if provided URL is already used by a different bookmark

---

#### Update Bookmark (Partial)

```
PATCH /api/bookmarks/<id>/
```

**Path Parameters:**

- `id` (integer): Bookmark ID

**Request Body:** Any subset of bookmark fields

**Behavior:** Only provided fields are updated

---

#### Archive Bookmark

```
POST /api/bookmarks/<id>/archive/
```

**Path Parameters:**

- `id` (integer): Bookmark ID

**Request Body:** None

**Response:** Success confirmation (likely 204 No Content)

---

#### Unarchive Bookmark

```
POST /api/bookmarks/<id>/unarchive/
```

**Path Parameters:**

- `id` (integer): Bookmark ID

**Request Body:** None

**Response:** Success confirmation

---

#### Delete Bookmark

```
DELETE /api/bookmarks/<id>/
```

**Path Parameters:**

- `id` (integer): Bookmark ID

**Response:** 204 No Content

---

## 2. Bookmark Assets

Base path: `/api/bookmarks/<bookmark_id>/assets/`

Assets are file attachments associated with bookmarks (snapshots, uploads).

### Asset Object Schema

```json
{
  "id": 1,                                    // integer, read-only
  "bookmark": 1,                              // integer, parent bookmark ID
  "asset_type": "snapshot" | "upload",        // string
  "date_created": "2023-10-01T12:00:00Z",     // ISO 8601, read-only
  "content_type": "text/html",                // MIME type
  "display_name": "HTML snapshot from 10/01/2023", // string
  "status": "complete" | "pending" | "failed" // string
}
```

### Asset Types

- `snapshot`: HTML snapshots created by LinkDing
- `upload`: User-uploaded files

### Endpoints

#### List Assets

```
GET /api/bookmarks/<bookmark_id>/assets/
```

**Path Parameters:**

- `bookmark_id` (integer): Parent bookmark ID

**Response:** Paginated list of asset objects

---

#### Retrieve Asset

```
GET /api/bookmarks/<bookmark_id>/assets/<id>/
```

**Path Parameters:**

- `bookmark_id` (integer): Parent bookmark ID
- `id` (integer): Asset ID

**Response:** Single asset object

---

#### Download Asset

```
GET /api/bookmarks/<bookmark_id>/assets/<id>/download/
```

**Path Parameters:**

- `bookmark_id` (integer): Parent bookmark ID
- `id` (integer): Asset ID

**Response:** Binary file content with appropriate Content-Type header

---

#### Upload Asset

```
POST /api/bookmarks/<bookmark_id>/assets/upload/
```

**Path Parameters:**

- `bookmark_id` (integer): Parent bookmark ID

**Request:** `multipart/form-data` with field named `file`

**Response:** Created asset object

---

#### Delete Asset

```
DELETE /api/bookmarks/<bookmark_id>/assets/<id>/
```

**Path Parameters:**

- `bookmark_id` (integer): Parent bookmark ID
- `id` (integer): Asset ID

**Response:** 204 No Content

---

## 3. Tags

Base path: `/api/tags/`

### Tag Object Schema

```json
{
  "id": 1,                                    // integer, read-only
  "name": "example",                          // string, required
  "date_added": "2020-09-26T09:46:23.006313Z" // ISO 8601, read-only
}
```

### Endpoints

#### List Tags

```
GET /api/tags/
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Max results, default 100 |
| `offset` | integer | Starting index |

**Response:** Paginated list of tag objects

---

#### Retrieve Tag

```
GET /api/tags/<id>/
```

**Path Parameters:**

- `id` (integer): Tag ID

**Response:** Single tag object

---

#### Create Tag

```
POST /api/tags/
```

**Request Body:**

```json
{
  "name": "example"  // required
}
```

**Response:** Created tag object

**Note:** Tags can also be created implicitly by including new tag names in bookmark `tag_names` arrays.

---

## 4. Bundles

Base path: `/api/bundles/`

Bundles are saved search filters/views for organizing bookmarks.

### Bundle Object Schema

```json
{
  "id": 1,                                    // integer, read-only
  "name": "Work Resources",                   // string, required
  "search": "productivity tools",             // string, search terms
  "any_tags": "work productivity",            // string, space-separated tag names (OR logic)
  "all_tags": "",                             // string, space-separated tag names (AND logic)
  "excluded_tags": "personal",                // string, space-separated tag names to exclude
  "order": 0,                                 // integer, display order
  "date_created": "2020-09-26T09:46:23.006313Z",   // ISO 8601, read-only
  "date_modified": "2020-09-26T16:01:14.275335Z"   // ISO 8601, read-only
}
```

### Bundle Filter Logic

- `search`: Free-text search terms
- `any_tags`: Bookmark matches if it has ANY of these tags (OR)
- `all_tags`: Bookmark matches only if it has ALL of these tags (AND)
- `excluded_tags`: Bookmark excluded if it has ANY of these tags

### Endpoints

#### List Bundles

```
GET /api/bundles/
```

**Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Max results, default 100 |
| `offset` | integer | Starting index |

**Response:** Paginated list of bundle objects

---

#### Retrieve Bundle

```
GET /api/bundles/<id>/
```

**Path Parameters:**

- `id` (integer): Bundle ID

**Response:** Single bundle object

---

#### Create Bundle

```
POST /api/bundles/
```

**Request Body:**

```json
{
  "name": "My Bundle",           // required
  "search": "search terms",      // optional
  "any_tags": "tag1 tag2",       // optional, space-separated
  "all_tags": "required-tag",    // optional, space-separated
  "excluded_tags": "excluded",   // optional, space-separated
  "order": 5                     // optional, auto-assigned if omitted
}
```

**Response:** Created bundle object

---

#### Update Bundle (Full)

```
PUT /api/bundles/<id>/
```

**Path Parameters:**

- `id` (integer): Bundle ID

**Request Body:** All writable fields should be provided

---

#### Update Bundle (Partial)

```
PATCH /api/bundles/<id>/
```

**Path Parameters:**

- `id` (integer): Bundle ID

**Request Body:** Any subset of writable fields

---

#### Delete Bundle

```
DELETE /api/bundles/<id>/
```

**Path Parameters:**

- `id` (integer): Bundle ID

**Response:** 204 No Content

---

## 5. User

Base path: `/api/user/`

### Endpoints

#### Get User Profile

```
GET /api/user/profile/
```

**Response:**

```json
{
  "theme": "auto" | "dark" | "light",
  "bookmark_date_display": "relative" | "absolute",
  "bookmark_link_target": "_blank" | "_self",
  "web_archive_integration": "enabled" | "disabled",
  "tag_search": "lax" | "strict",
  "enable_sharing": true,
  "enable_public_sharing": true,
  "enable_favicons": false,
  "display_url": false,
  "permanent_notes": false,
  "search_preferences": {
    "sort": "title_asc" | "title_desc" | "date_added_asc" | "date_added_desc" | ...,
    "shared": "on" | "off",
    "unread": "on" | "off"
  }
}
```

**Use Cases:**

- Verify connection works (use as health check)
- Get user preferences for CLI display formatting

---

## API Quirks and Implementation Notes

### Critical Behaviors

1. **Duplicate URL Handling on Create:** POST to `/api/bookmarks/` with an existing URL silently updates instead of returning an error. Always use `/api/bookmarks/check/` first if you need to detect duplicates.

2. **PUT vs PATCH:** Despite being PUT, the bookmark update endpoint doesn't require all fields - unprovided fields are preserved. This is non-standard REST behavior.

3. **Archived Bookmarks Separation:** Archived bookmarks have their own endpoint (`/archived/`). They don't appear in the main list.

4. **Tag Creation:** Tags are created implicitly when used in bookmark `tag_names`. Explicit POST to `/api/tags/` is optional.

5. **Bundle Tag Syntax:** Bundle tag fields (`any_tags`, `all_tags`, `excluded_tags`) use space-separated strings, not arrays.

### Search Query Syntax (for `q` parameter)

The `q` parameter supports the same syntax as the web UI:

- Simple terms: `kubernetes tutorial`
- Tag filter: `#devops` or `#kubernetes`
- Exclude tag: `!#archived`
- Unread only: `!unread`
- Title search: `title:kubernetes`
- Description search: `description:tutorial`
- Notes search: `notes:important`
- URL search: `url:github.com`

Multiple filters can be combined: `#devops kubernetes !#old`

### Date Formats

All dates use ISO 8601 format: `YYYY-MM-DDTHH:MM:SS.ffffffZ`

Example: `2025-01-24T15:30:00.000000Z`

---

## Endpoint Summary Table

| Resource | Method | Endpoint | Description |
|----------|--------|----------|-------------|
| Bookmarks | GET | `/api/bookmarks/` | List bookmarks |
| Bookmarks | GET | `/api/bookmarks/archived/` | List archived bookmarks |
| Bookmarks | GET | `/api/bookmarks/<id>/` | Get single bookmark |
| Bookmarks | GET | `/api/bookmarks/check/?url=` | Check if URL bookmarked |
| Bookmarks | POST | `/api/bookmarks/` | Create bookmark |
| Bookmarks | PUT | `/api/bookmarks/<id>/` | Update bookmark (full) |
| Bookmarks | PATCH | `/api/bookmarks/<id>/` | Update bookmark (partial) |
| Bookmarks | POST | `/api/bookmarks/<id>/archive/` | Archive bookmark |
| Bookmarks | POST | `/api/bookmarks/<id>/unarchive/` | Unarchive bookmark |
| Bookmarks | DELETE | `/api/bookmarks/<id>/` | Delete bookmark |
| Assets | GET | `/api/bookmarks/<bid>/assets/` | List assets |
| Assets | GET | `/api/bookmarks/<bid>/assets/<id>/` | Get single asset |
| Assets | GET | `/api/bookmarks/<bid>/assets/<id>/download/` | Download asset file |
| Assets | POST | `/api/bookmarks/<bid>/assets/upload/` | Upload asset |
| Assets | DELETE | `/api/bookmarks/<bid>/assets/<id>/` | Delete asset |
| Tags | GET | `/api/tags/` | List tags |
| Tags | GET | `/api/tags/<id>/` | Get single tag |
| Tags | POST | `/api/tags/` | Create tag |
| Bundles | GET | `/api/bundles/` | List bundles |
| Bundles | GET | `/api/bundles/<id>/` | Get single bundle |
| Bundles | POST | `/api/bundles/` | Create bundle |
| Bundles | PUT | `/api/bundles/<id>/` | Update bundle (full) |
| Bundles | PATCH | `/api/bundles/<id>/` | Update bundle (partial) |
| Bundles | DELETE | `/api/bundles/<id>/` | Delete bundle |
| User | GET | `/api/user/profile/` | Get user profile |

---

## CLI Command Mapping Suggestions

Based on the API structure, suggested CLI command hierarchy:

```
ldctl
├── config
│   ├── init
│   ├── show
│   └── test
├── bookmarks (alias: bm)
│   ├── list [--archived] [--query] [--limit] [--offset]
│   ├── get <id>
│   ├── add <url> [--title] [--description] [--tags] [--archived] [--unread]
│   ├── check <url>
│   ├── update <id> [--title] [--description] [--tags] [--archived] [--unread]
│   ├── archive <id>
│   ├── unarchive <id>
│   └── delete <id> [--force]
├── assets
│   ├── list <bookmark-id>
│   ├── download <bookmark-id> <asset-id> [--output]
│   ├── upload <bookmark-id> <file>
│   └── delete <bookmark-id> <asset-id>
├── tags
│   ├── list
│   ├── get <id>
│   └── create <name>
├── bundles
│   ├── list
│   ├── get <id>
│   ├── create <name> [--search] [--any-tags] [--all-tags] [--exclude-tags]
│   ├── update <id> [options]
│   └── delete <id>
└── user
    └── profile
```
