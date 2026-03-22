# PRD: Error Handling and Exit Codes for ldctl

## Executive Summary

This PRD defines the comprehensive error handling strategy for the ldctl CLI tool, including standardized exit codes, error message formats, and structured error output for JSON mode. This specification resolves the existing inconsistency between AGENTS.md (which defines three exit codes) and the individual feature PRDs (which only mention two), establishing AGENTS.md's three-code system as the authoritative specification.

## Problem Statement

Currently, there is an inconsistency in the documentation:
- **AGENTS.md** specifies: "Exit codes: 0=success, 1=error, 2=config error"
- **Feature PRDs** only mention: "Exit code 0 on success, 1 on error"

Additionally, the project lacks:
1. Standardized error message formats
2. Consistent error categorization
3. Structured error output for JSON mode
4. Clear guidelines on when to use which exit code

This inconsistency and lack of standardization will lead to:
- Unpredictable behavior for scripts relying on exit codes
- Inconsistent error reporting across commands
- Difficulty in debugging and supporting users
- Poor user experience with unhelpful error messages

## Goals

1. **Resolve the exit code inconsistency** - Establish authoritative exit code specification
2. **Standardize error messages** - Consistent format across all commands
3. **Define error categories** - Clear mapping of error types to exit codes
4. **Enable structured errors** - JSON mode error output for automation
5. **Provide actionable errors** - Include resolution hints in error messages

### Success Metrics

| Metric | Target |
|--------|--------|
| Exit code consistency | 100% of commands follow the three-code system |
| Error message format compliance | 100% of errors follow standard format |
| JSON error structure validity | All JSON errors are valid and consistent |
| Error resolution hints | 80% of errors include actionable next steps |
| Test coverage for error paths | > 90% coverage of error scenarios |

## Exit Code Resolution

**DECISION**: Adopt AGENTS.md's three-code system as the authoritative specification.

### Rationale
1. **Config errors are special** - They prevent any API operations from succeeding
2. **Scripts benefit from distinction** - Can detect config issues vs operational failures
3. **Industry precedent** - Many CLI tools use 2 for usage/config errors (e.g., grep, curl)
4. **Already documented** - AGENTS.md already specifies this behavior
5. **Backward compatible** - Scripts checking for non-zero still work

### Exit Code Mapping

| Exit Code | Category | When Used | Examples |
|-----------|----------|-----------|----------|
| **0** | Success | All successful operations | Command completed, even if no results |
| **1** | General Error | API errors, validation, I/O failures | 404 not found, invalid input, file write failed |
| **2** | Configuration Error | Config/auth/connectivity issues | Missing config, auth failed, can't reach API |

### Detailed Exit Code Usage

**Exit Code 0 (Success):**
- Command executed successfully
- Operation completed as requested
- Empty result sets (e.g., no bookmarks found)
- Help text displayed
- Version information shown

**Exit Code 1 (General Errors):**
- Resource not found (404)
- Validation errors (400)
- Server errors (500-599)
- File I/O errors (can't write output file)
- User cancelled operation (Ctrl+C, answered "no" to prompt)
- Invalid command arguments (wrong ID format, bad URL)
- Operation failed (bookmark already exists, etc.)

**Exit Code 2 (Configuration Errors):**
- No configuration file found
- Malformed configuration file
- Missing required config fields
- Authentication failure (401, 403)
- Network connectivity failure (can't reach LinkDing)
- Invalid base URL in config
- Permission errors on config file

## Error Message Format Standards

### Basic Structure

All error messages follow this format:

```
Error: <description>
<optional context>
<optional resolution>
```

### Guidelines

1. **Start with "Error:"** - Consistent prefix for parsing
2. **Lowercase descriptions** - Unless proper nouns
3. **Specific details** - Include relevant IDs, URLs, status codes
4. **Actionable hints** - Tell user how to fix the problem
5. **stderr output** - All errors go to stderr, not stdout

### Examples

**Configuration Error:**
```
Error: no configuration found
Run 'ldctl config init' to get started.
```

**Authentication Error:**
```
Error: authentication failed (401 Unauthorized)
Your API token may be invalid or expired.
Run 'ldctl config init' to reconfigure.
```

**Resource Not Found:**
```
Error: bookmark not found (404)
No bookmark exists with ID: 123
```

**Validation Error:**
```
Error: invalid URL format
URL must start with http:// or https://
Provided: "not-a-url"
```

**Network Error:**
```
Error: could not connect to LinkDing instance
Failed to reach: https://linkding.example.com
Check your network connection and instance URL.
```

### Message Templates

**Config Errors (Exit 2):**
```
Error: no configuration found
Run 'ldctl config init' to get started.

Error: config file is malformed
Invalid TOML at line 3: expected '=' but got ':'
Fix the config file or run 'ldctl config init --force' to recreate.

Error: authentication failed (401 Unauthorized)
Your API token may be invalid or expired.
Run 'ldctl config init' to reconfigure.

Error: could not connect to LinkDing instance
Failed to reach: {url}
Check your network connection and instance URL.
```

**API Errors (Exit 1):**
```
Error: bookmark not found (404)
No bookmark exists with ID: {id}

Error: validation failed (400 Bad Request)
{field}: {validation_message}

Error: server error (500 Internal Server Error)
The LinkDing server encountered an error.
Try again later or contact your administrator.

Error: bookmark already exists
A bookmark with URL '{url}' already exists (ID: {id})
Use 'ldctl bookmarks update {id}' to modify it.
```

**I/O Errors (Exit 1):**
```
Error: cannot write to file
Permission denied: {filepath}

Error: file not found
No such file: {filepath}

Error: output directory does not exist
Directory not found: {dirpath}
Create it first or choose a different directory.
```

**Operational Errors (Exit 1):**
```
Error: operation cancelled by user

Error: invalid bookmark ID
Expected numeric ID, got: {input}

Error: conflicting flags
Cannot use --{flag1} with --{flag2}
```

## Structured Error Output (JSON Mode)

When `--json` flag is present, errors are output as structured JSON to stderr.

### JSON Error Schema

```json
{
  "error": {
    "type": "error_type",
    "message": "Human-readable error description",
    "details": {
      // Optional: Additional structured information
    }
  }
}
```

### Error Types

Standard error types for the `type` field:

| Type | Exit Code | Description |
|------|-----------|-------------|
| `config_error` | 2 | Configuration file issues |
| `auth_error` | 2 | Authentication failures |
| `network_error` | 2 | Connectivity problems |
| `api_error` | 1 | LinkDing API errors |
| `validation_error` | 1 | Input validation failures |
| `io_error` | 1 | File I/O problems |
| `not_found` | 1 | Resource doesn't exist |
| `user_cancelled` | 1 | User cancelled operation |

### JSON Error Examples

**Configuration Error:**
```json
{
  "error": {
    "type": "config_error",
    "message": "no configuration found",
    "details": {
      "config_path": "/home/user/.config/ldctl/config.toml",
      "suggestion": "Run 'ldctl config init' to get started"
    }
  }
}
```

**Authentication Error:**
```json
{
  "error": {
    "type": "auth_error",
    "message": "authentication failed",
    "details": {
      "http_status": 401,
      "instance_url": "https://linkding.example.com"
    }
  }
}
```

**API Error:**
```json
{
  "error": {
    "type": "api_error",
    "message": "bookmark not found",
    "details": {
      "http_status": 404,
      "bookmark_id": 123
    }
  }
}
```

**Validation Error:**
```json
{
  "error": {
    "type": "validation_error",
    "message": "invalid URL format",
    "details": {
      "field": "url",
      "value": "not-a-url",
      "requirement": "must start with http:// or https://"
    }
  }
}
```

**Network Error:**
```json
{
  "error": {
    "type": "network_error",
    "message": "connection failed",
    "details": {
      "url": "https://linkding.example.com",
      "error": "dial tcp: lookup linkding.example.com: no such host"
    }
  }
}
```

## Error Type Taxonomy

### Configuration Errors (Exit Code 2)

**Triggers:**
- Missing config file when required
- Malformed config file (bad TOML syntax)
- Missing required fields (url or token)
- Invalid config values
- Config file permission issues
- Environment variable parsing errors

**Error Types:**
- `config_error` - General configuration issues
- `auth_error` - Authentication/authorization failures (401, 403)
- `network_error` - Can't reach LinkDing instance

### Operational Errors (Exit Code 1)

**Triggers:**
- Resource not found (404)
- Validation failures (400)
- Server errors (500-599)
- Rate limiting (429)
- Conflicts (409)
- I/O failures
- User cancellation

**Error Types:**
- `api_error` - LinkDing API returned an error
- `validation_error` - Input validation failed
- `io_error` - File system operations failed
- `not_found` - Requested resource doesn't exist
- `user_cancelled` - User chose to cancel

## Functional Requirements

### REQ-001: Three-Level Exit Codes [MUST]
The CLI shall use three exit codes: 0 (success), 1 (general error), 2 (config error).

### REQ-002: Consistent Error Format [MUST]
All error messages shall follow the standard format: "Error: <description>"

### REQ-003: stderr for Errors [MUST]
All error messages shall be written to stderr, not stdout.

### REQ-004: JSON Error Structure [MUST]
When --json flag is present, errors shall be output as structured JSON to stderr.

### REQ-005: Error Type Classification [MUST]
Every error shall be classified into a standard error type with consistent exit code mapping.

### REQ-006: Actionable Error Messages [SHOULD]
Error messages should include actionable resolution hints when possible.

### REQ-007: HTTP Status Preservation [SHOULD]
API errors should include the HTTP status code in the error message.

### REQ-008: Context in Errors [SHOULD]
Errors should include relevant context (IDs, URLs, filenames) for debugging.

### REQ-009: Validation Details [SHOULD]
Validation errors should specify which field failed and why.

### REQ-010: Network Error Details [SHOULD]
Network errors should include the URL that couldn't be reached.

### REQ-011: Exit Code Documentation [MUST]
Every command's help text shall document its exit codes.

### REQ-012: Error Wrapping [COULD]
Errors could preserve the chain of causation for debugging.

### REQ-013: Error Codes [COULD]
Future: Specific error codes (E001, E002) for documentation.

### REQ-014: Retry Hints [COULD]
Transient errors could suggest retry with exponential backoff.

## Non-Functional Requirements

### Consistency
- All commands use the same error format
- Exit codes are predictable across commands
- JSON errors have consistent structure

### Debuggability
- Errors include sufficient context
- Original error causes are preserved
- Stack traces available in debug mode

### Usability
- Error messages are clear and actionable
- Technical jargon is minimized
- Resolution steps are provided

### Performance
- Error handling adds < 10ms overhead
- JSON serialization is efficient
- No unnecessary string allocations

## Technical Considerations

### Implementation Architecture

**Error Package Structure:**
```
internal/
  errors/
    errors.go       # Error types and constructors
    codes.go        # Exit code constants
    format.go       # Error formatting functions
    json.go         # JSON error serialization
```

**Error Type Definition:**
```go
package errors

type Type string

const (
    ConfigError     Type = "config_error"
    AuthError       Type = "auth_error"
    NetworkError    Type = "network_error"
    APIError        Type = "api_error"
    ValidationError Type = "validation_error"
    IOError         Type = "io_error"
    NotFound        Type = "not_found"
    UserCancelled   Type = "user_cancelled"
)

type Error struct {
    Type    Type
    Message string
    Details map[string]interface{}
    Cause   error
}

func (e Error) Error() string {
    return e.Message
}

func (e Error) ExitCode() int {
    switch e.Type {
    case ConfigError, AuthError, NetworkError:
        return 2
    default:
        return 1
    }
}
```

### Cobra Integration

```go
// In root command
cmd.SilenceErrors = true  // We handle errors ourselves
cmd.SilenceUsage = true   // Don't show usage on errors

// In command execution
if err := cmd.Execute(); err != nil {
    var ldctlErr *errors.Error
    if errors.As(err, &ldctlErr) {
        if *jsonFlag {
            printJSONError(ldctlErr)
        } else {
            printHumanError(ldctlErr)
        }
        os.Exit(ldctlErr.ExitCode())
    } else {
        // Fallback for unexpected errors
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### HTTP Response Mapping

```go
func mapHTTPError(resp *http.Response) *errors.Error {
    switch resp.StatusCode {
    case 401, 403:
        return &errors.Error{
            Type:    errors.AuthError,
            Message: fmt.Sprintf("authentication failed (%d %s)", 
                resp.StatusCode, http.StatusText(resp.StatusCode)),
            Details: map[string]interface{}{
                "http_status": resp.StatusCode,
            },
        }
    case 404:
        return &errors.Error{
            Type:    errors.NotFound,
            Message: "resource not found",
            Details: map[string]interface{}{
                "http_status": 404,
            },
        }
    case 400:
        return &errors.Error{
            Type:    errors.ValidationError,
            Message: "validation failed",
            Details: map[string]interface{}{
                "http_status": 400,
            },
        }
    case 429:
        return &errors.Error{
            Type:    errors.APIError,
            Message: "rate limit exceeded",
            Details: map[string]interface{}{
                "http_status": 429,
                "retry_after": resp.Header.Get("Retry-After"),
            },
        }
    case 500, 502, 503, 504:
        return &errors.Error{
            Type:    errors.APIError,
            Message: fmt.Sprintf("server error (%d %s)", 
                resp.StatusCode, http.StatusText(resp.StatusCode)),
            Details: map[string]interface{}{
                "http_status": resp.StatusCode,
            },
        }
    default:
        return &errors.Error{
            Type:    errors.APIError,
            Message: fmt.Sprintf("unexpected response (%d %s)", 
                resp.StatusCode, http.StatusText(resp.StatusCode)),
            Details: map[string]interface{}{
                "http_status": resp.StatusCode,
            },
        }
    }
}
```

### Constructor Helpers

```go
// Convenience constructors for common errors
func NewConfigNotFound(path string) *Error {
    return &Error{
        Type:    ConfigError,
        Message: "no configuration found",
        Details: map[string]interface{}{
            "config_path": path,
            "suggestion":  "Run 'ldctl config init' to get started",
        },
    }
}

func NewAuthFailed(url string, status int) *Error {
    return &Error{
        Type:    AuthError,
        Message: "authentication failed",
        Details: map[string]interface{}{
            "http_status":  status,
            "instance_url": url,
        },
    }
}

func NewNotFound(resource string, id interface{}) *Error {
    return &Error{
        Type:    NotFound,
        Message: fmt.Sprintf("%s not found", resource),
        Details: map[string]interface{}{
            "resource": resource,
            "id":       id,
        },
    }
}
```

## Migration from Existing PRDs

### Required Updates

All existing PRDs need to be updated to reflect the three-code system:

**Current Text in PRDs:**
```
Exit code 0 on success, 1 on any error
```

**Updated Text:**
```
Exit codes:
- 0: Success
- 1: Operational errors (API errors, not found, validation)
- 2: Configuration errors (missing config, auth failure, connectivity)
```

### Command-Specific Updates

Each command documentation should specify its exit codes:

```
## Exit Codes

- 0: Command completed successfully
- 1: Command failed (see error message for details)
- 2: Configuration or authentication error

Examples:
- Bookmark not found: exit 1
- Invalid URL format: exit 1
- Missing config file: exit 2
- Authentication failed: exit 2
```

### No Breaking Changes

This change is backward-compatible:
- Scripts checking for non-zero exit still work
- Scripts specifically checking for 1 still catch most errors
- Only new scripts can leverage the distinction

## Implementation Roadmap

### Phase 1: Error Infrastructure (Day 1-2)
1. Create `internal/errors/` package
2. Define error types and exit codes
3. Implement error constructors
4. Add JSON serialization
5. Write unit tests

### Phase 2: Config Commands (Day 3-4)
1. Update config commands to use new errors
2. Ensure exit code 2 for config errors
3. Add actionable error messages
4. Test all error paths

### Phase 3: All Other Commands (Week 2)
1. Update each command group
2. Map HTTP responses to error types
3. Add context to all errors
4. Ensure JSON mode works

## Testing Strategy

### Unit Tests
- Test each error constructor
- Verify exit code mapping
- Test JSON serialization
- Test error formatting

### Integration Tests
- Test exit codes for each command
- Verify stderr vs stdout
- Test JSON error output
- Test error messages include context

### Manual Testing Checklist
- [ ] Missing config file → exit 2
- [ ] Invalid token → exit 2
- [ ] Network down → exit 2
- [ ] Bookmark not found → exit 1
- [ ] Invalid URL → exit 1
- [ ] Server error → exit 1
- [ ] Success → exit 0
- [ ] --json errors are valid JSON
- [ ] Errors go to stderr
- [ ] Resolution hints present

## Out of Scope

This PRD does not cover:
- Internationalization (i18n) of error messages
- Error telemetry or reporting
- Retry logic (separate PRD)
- Rate limit handling (separate PRD)
- Colored output for errors
- Error recovery mechanisms
- Undo operations after errors

## Open Questions

1. **Should we add error codes (E001, E002)?**
   - Decision: Not in v1, consider for v2 if needed for documentation

2. **Should network timeouts be exit 1 or 2?**
   - Decision: Exit 2 (config-related, like connectivity)

3. **Should --debug flag show stack traces?**
   - Decision: Yes, but implement in logging PRD

4. **Should we wrap errors with github.com/pkg/errors?**
   - Decision: Use standard library errors with custom type

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Inconsistent implementation | Poor UX | Code review checklist |
| Missing error handling | Panics | Unit tests for error paths |
| Wrong exit codes | Script failures | Integration tests |
| Unhelpful messages | User frustration | Error message review |

## Dependencies

- Cobra CLI framework (error handling hooks)
- Go 1.13+ (errors.As, errors.Is)
- JSON encoding/json package

## Validation Checkpoints

### Phase 1 Complete
- [ ] Error package created and tested
- [ ] All error types defined
- [ ] Exit code mapping correct
- [ ] JSON serialization works

### Phase 2 Complete
- [ ] Config commands use new errors
- [ ] Exit code 2 for config errors
- [ ] Actionable messages present
- [ ] Tests pass

### Phase 3 Complete
- [ ] All commands updated
- [ ] HTTP errors mapped correctly
- [ ] JSON mode works everywhere
- [ ] Documentation updated

## Acceptance Criteria

- [ ] Three exit codes implemented: 0, 1, 2
- [ ] All errors follow standard format
- [ ] JSON errors are valid and consistent
- [ ] Error messages include context
- [ ] Resolution hints provided where applicable
- [ ] All errors go to stderr
- [ ] Exit codes are predictable
- [ ] Tests cover all error paths
- [ ] Documentation updated
- [ ] PRDs updated to reflect three-code system

## References

- [AGENTS.md Exit Codes](/home/rodst/Repos/ldctl/AGENTS.md)
- [POSIX Exit Codes](https://www.gnu.org/software/libc/manual/html_node/Exit-Status.html)
- [Cobra Error Handling](https://github.com/spf13/cobra#error-handling)
- [Go Error Handling](https://go.dev/blog/go1.13-errors)
- [JSON API Error Format](https://jsonapi.org/format/#errors)

## Appendix: Exit Code Comparison

| Tool | Success | General Error | Config/Usage Error |
|------|---------|--------------|-------------------|
| ldctl (this PRD) | 0 | 1 | 2 |
| grep | 0 | 2 | 2 |
| curl | 0 | various | 2 |
| git | 0 | 1 | 128 |
| docker | 0 | 1 | 125 |
| kubectl | 0 | 1 | 1 |

Most tools distinguish between operational and configuration/usage errors, validating our three-code approach.