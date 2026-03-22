# PRD: Version Management System for ldctl

## Executive Summary

This PRD defines the version management system for the ldctl CLI tool, including the version command, build metadata embedding, semantic versioning scheme, and version display formats. The version system provides users with comprehensive information about their ldctl installation and helps with debugging, support, and compatibility verification.

## Problem Statement

Users and support teams need to:
1. Identify the exact version of ldctl being used
2. Verify compatibility with LinkDing API versions
3. Report bugs with precise version information
4. Understand build provenance (when and from what source)

Without proper version management, debugging issues becomes difficult and users cannot determine if they're running the latest release.

## Goals

1. **Provide clear version identification** - Users can easily determine their ldctl version
2. **Include build metadata** - Show commit hash, build date, and Go version
3. **Support multiple output formats** - Human-readable and machine-parseable
4. **Follow semantic versioning** - Predictable version progression
5. **Enable version embedding** - Build process automatically injects version info

### Success Metrics

| Metric | Target |
|--------|--------|
| Version info availability | 100% of builds include version |
| Version display speed | < 100ms |
| Build reproducibility | Version + metadata uniquely identifies build |
| User understanding | Clear version format documentation |

## User Stories

### Story 1: User Checks Version
**As a** ldctl user
**I want to** check my installed version
**So that** I can verify I have the latest release or report bugs accurately

**Acceptance Criteria:**
- `ldctl version` displays version information
- `ldctl --version` works from any command context
- Output includes semantic version number
- Response time < 100ms

### Story 2: User Needs Short Version
**As a** script author
**I want to** get just the semantic version number
**So that** I can parse it in automation scripts

**Acceptance Criteria:**
- `ldctl version --short` outputs only "1.2.3"
- `ldctl --version` outputs full version line
- No additional text or formatting
- Exit code 0 on success

### Story 3: Support Team Needs Build Metadata
**As a** support team member
**I want to** see complete build information
**So that** I can identify the exact build and reproduce issues

**Acceptance Criteria:**
- Version output includes git commit hash
- Build date/time shown in UTC
- Go compiler version displayed
- JSON format available for parsing

## Functional Requirements

### REQ-001: Version Command [MUST]
The CLI shall provide a `version` subcommand that displays version information.

**Command Structure:**
```
ldctl version [flags]
```

**Flags:**
- `--short`: Display only the semantic version number
- `--json`: Output in JSON format

**Examples:**
```bash
# Default output
$ ldctl version
ldctl version 1.2.3 (commit a1b2c3d, built 2025-01-27T10:30:00Z, go1.25.0)

# Short output
$ ldctl version --short
1.2.3

# JSON output
$ ldctl version --json
{
  "version": "1.2.3",
  "commit": "a1b2c3d",
  "buildDate": "2025-01-27T10:30:00Z",
  "goVersion": "go1.25.0",
  "os": "linux",
  "arch": "amd64"
}
```

### REQ-002: Global Version Flag [MUST]
The root command shall support a global `--version` flag that displays version information from any command context.

**Behavior:**
- Available on root command and all subcommands
- Short-circuits command execution (version displayed, command not run)
- Displays single-line version format
- Exits with code 0

**Example:**
```bash
$ ldctl --version
ldctl version 1.2.3 (commit a1b2c3d, built 2025-01-27T10:30:00Z, go1.25.0)

$ ldctl bookmarks list --version
ldctl version 1.2.3 (commit a1b2c3d, built 2025-01-27T10:30:00Z, go1.25.0)
```

### REQ-003: Semantic Versioning [MUST]
Version numbers shall follow Semantic Versioning 2.0.0 (https://semver.org/).

**Format:** `MAJOR.MINOR.PATCH[-PRERELEASE][+BUILDMETA]`

**Examples:**
- `1.0.0` - First stable release
- `1.2.3` - Standard release
- `2.0.0-rc.1` - Release candidate
- `1.2.3-alpha+a1b2c3d` - Alpha with build metadata
- `0.1.0` - Pre-1.0 development version

### REQ-004: Build Metadata [MUST]
Version information shall include build metadata for debugging and support.

**Required Metadata:**
- **Version**: Semantic version string
- **Commit**: Git commit hash (short form, 7 characters)
- **Build Date**: RFC3339 timestamp in UTC
- **Go Version**: Go compiler version used

**Optional Metadata:**
- **OS**: Operating system (linux, darwin, windows)
- **Arch**: CPU architecture (amd64, arm64, etc.)
- **Dirty**: Boolean indicating uncommitted changes

### REQ-005: Version Embedding Strategy [MUST]
Version information shall be embedded at build time using Go linker flags.

**Implementation:**
```go
// internal/version/version.go
package version

var (
    // Set via ldflags at build time
    Version   = "dev"
    Commit    = "unknown"
    BuildDate = "unknown"
    GoVersion = runtime.Version()
)
```

**Build Command:**
```bash
go build -ldflags "\
  -X github.com/rodmhgl/ldctl/internal/version.Version=1.2.3 \
  -X github.com/rodmhgl/ldctl/internal/version.Commit=$(git rev-parse --short HEAD) \
  -X github.com/rodmhgl/ldctl/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  " -o ldctl main.go
```

### REQ-006: Development Builds [SHOULD]
Development builds (untagged) shall display "dev" as the version with current commit.

**Example:**
```bash
$ ./ldctl version  # Built from main branch, no tag
ldctl version dev (commit a1b2c3d, built 2025-01-27T10:30:00Z, go1.25.0)
```

### REQ-007: Version in Help Text [SHOULD]
The version shall be displayed in the root command help text.

**Example:**
```
ldctl - LinkDing CLI client (version 1.2.3)

Usage:
  ldctl [command]

Available Commands:
  ...
```

### REQ-008: Version File [COULD]
A VERSION file in the repository root could store the current version for automation.

**Format:**
```
1.2.3
```

**Usage:**
- Read by Makefile for builds
- Updated by release scripts
- Single source of truth

### REQ-009: Version Check Command [COULD]
Future enhancement: Check for newer versions available.

```bash
$ ldctl version check
Current version: 1.2.3
Latest version:  1.3.0
Update available! Run: brew upgrade ldctl
```

## Non-Functional Requirements

### Performance
- Version display must complete in < 100ms
- No network calls for basic version display
- Version check (future) may make network calls

### Compatibility
- Version format must be parseable by standard tools
- JSON output must be valid JSON
- Semantic version must be valid per semver spec

### Reliability
- Version command works without config file
- Version command works without network
- Version command works with minimal dependencies

## Technical Considerations

### Architecture

**File Structure:**
```
internal/
  version/
    version.go      # Version variables and display logic
    version_test.go # Unit tests
cmd/
  version.go        # Version command implementation
```

**Version Package:**
```go
package version

import (
    "encoding/json"
    "fmt"
    "runtime"
)

var (
    Version   = "dev"
    Commit    = "unknown"
    BuildDate = "unknown"
    GoVersion = runtime.Version()
)

type Info struct {
    Version   string `json:"version"`
    Commit    string `json:"commit"`
    BuildDate string `json:"buildDate"`
    GoVersion string `json:"goVersion"`
    OS        string `json:"os"`
    Arch      string `json:"arch"`
}

func Get() Info {
    return Info{
        Version:   Version,
        Commit:    Commit,
        BuildDate: BuildDate,
        GoVersion: GoVersion,
        OS:        runtime.GOOS,
        Arch:      runtime.GOARCH,
    }
}

func String() string {
    return fmt.Sprintf("ldctl version %s (commit %s, built %s, %s)",
        Version, Commit, BuildDate, GoVersion)
}
```

### Display Formats

**Human-Readable (default):**
```
ldctl version 1.2.3 (commit a1b2c3d, built 2025-01-27T10:30:00Z, go1.25.0)
```

**Short Format (--short):**
```
1.2.3
```

**JSON Format (--json):**
```json
{
  "version": "1.2.3",
  "commit": "a1b2c3d",
  "buildDate": "2025-01-27T10:30:00Z",
  "goVersion": "go1.25.0",
  "os": "linux",
  "arch": "amd64"
}
```

### Git Commit Hash Format
- Use short form (7 characters) for display
- Full hash available in JSON output (future)
- Handle missing git info gracefully ("unknown")

### Build Date Format
- RFC3339 format: `2025-01-27T10:30:00Z`
- Always UTC timezone
- Set at build time, not runtime

### Dirty Workspace Detection
Optional enhancement to detect uncommitted changes:
```bash
DIRTY=$(git diff --quiet || echo "-dirty")
VERSION="1.2.3${DIRTY}"
```

## Implementation Roadmap

### Phase 1: Core Version System (Week 1)
1. Create `internal/version/version.go` with variables
2. Implement `cmd/version.go` command
3. Add `--version` global flag to root
4. Write unit tests

### Phase 2: Build Integration (Week 1-2)
1. Update Makefile with version injection
2. Create build scripts with proper ldflags
3. Add VERSION file to repository
4. Test builds with version info

### Phase 3: CI/CD Integration (Week 2)
1. Update GitHub Actions to inject version
2. Configure GoReleaser with version tags
3. Test release builds
4. Document release process

## Version Release Process

### Development Builds
Commits to main branch produce development builds:
- Version: "dev"
- Commit: Current commit hash
- Build Date: Build timestamp

### Tagged Releases
Git tags trigger release builds:
```bash
git tag -a v1.2.3 -m "Release version 1.2.3"
git push origin v1.2.3
```

Results in:
- Version: "1.2.3"
- Commit: Tag commit hash
- Build Date: Build timestamp

### Pre-release Versions
For release candidates and betas:
```bash
git tag v2.0.0-rc.1
```

Results in:
- Version: "2.0.0-rc.1"

## Semantic Versioning Rules

### Version Bumping Guidelines

**MAJOR version (1.0.0 → 2.0.0):**
- Breaking changes to CLI interface
- Removal of commands or flags
- Changed behavior that breaks scripts
- Dropped support for LinkDing API versions

**MINOR version (1.2.0 → 1.3.0):**
- New commands or subcommands
- New flags (backward compatible)
- New features
- Performance improvements

**PATCH version (1.2.3 → 1.2.4):**
- Bug fixes
- Documentation updates
- Security fixes
- Dependency updates (non-breaking)

### Pre-1.0 Development
While version < 1.0.0:
- API may change between minor versions
- 0.x.y follows different rules
- 0.x updates may include breaking changes
- 0.x.y updates are patches/fixes

## Security Considerations

### Information Disclosure
Version information is not sensitive and can be displayed freely.

### Build Reproducibility
Including commit hash and build date aids in build verification and security auditing.

### No Phone Home
Version command must not make network calls or send telemetry.

## Testing Strategy

### Unit Tests
```go
func TestVersionString(t *testing.T) {
    version.Version = "1.2.3"
    version.Commit = "abc123d"
    version.BuildDate = "2025-01-27T10:30:00Z"
    
    expected := "ldctl version 1.2.3 (commit abc123d, built 2025-01-27T10:30:00Z, go1.25.0)"
    actual := version.String()
    assert.Equal(t, expected, actual)
}
```

### Integration Tests
1. `ldctl version` returns exit code 0
2. `ldctl --version` works from any subcommand
3. `ldctl version --json` produces valid JSON
4. `ldctl version --short` returns only version number

### Build Tests
1. Makefile correctly injects version
2. GoReleaser builds include version
3. Development builds show "dev"
4. Tagged builds show tag version

## Documentation Requirements

### README.md Updates
- Document how to check version
- Include version in installation verification
- Show version in bug report template

### Building Documentation
- Document ldflags for version injection
- Provide build examples
- Explain VERSION file usage

### User Documentation
- Help text mentions version command
- Examples show version checking
- Troubleshooting guide references version

## Out of Scope

This PRD does not cover:
- Automatic update checking (future enhancement)
- Version compatibility matrix with LinkDing
- Telemetry or usage analytics
- Version-specific configuration migration
- Beta/alpha channel management
- Version-based feature flags

## Open Questions

1. **Should we check for updates automatically?**
   - Decision: No, start with manual check only

2. **Include build machine info?**
   - Decision: No, keep metadata minimal

3. **Version in every command output?**
   - Decision: No, only on request

4. **Support version aliases (latest, stable)?**
   - Decision: Not in v1, consider for future

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Version not injected in builds | Users see "dev" version | Makefile validation, CI tests |
| Invalid semver format | Tools can't parse version | Semver validation in CI |
| Git info not available | Missing commit hash | Graceful fallback to "unknown" |
| Build date incorrect | Confusion about build age | Use UTC, automate timestamp |

## Dependencies

- Go build toolchain for ldflags
- Git for commit hash
- Make or build scripts
- GoReleaser for releases (optional)

## Acceptance Criteria

- [ ] `ldctl version` displays version info
- [ ] `ldctl --version` works globally
- [ ] `--short` flag shows only version number
- [ ] `--json` flag outputs valid JSON
- [ ] Build process injects version automatically
- [ ] Development builds show "dev" version
- [ ] Tagged releases show tag version
- [ ] Version info includes all required metadata
- [ ] Tests pass for all version formats
- [ ] Documentation updated with version info

## References

- [Semantic Versioning 2.0.0](https://semver.org/)
- [Go Build Ldflags](https://golang.org/cmd/link/)
- [GoReleaser](https://goreleaser.com/)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
- [Git Commit Hash](https://git-scm.com/docs/git-rev-parse)

## Appendix: Version Examples for ldctl

### When to bump MAJOR version (breaking changes):
```
2.0.0: Remove bookmarks export --format=netscape
2.0.0: Change config file from TOML to YAML
2.0.0: Rename 'bookmarks' command to 'bm' only
2.0.0: Change --json output structure
```

### When to bump MINOR version (new features):
```
1.1.0: Add bookmarks import command
1.2.0: Add shell completions support
1.3.0: Add rate limiting with retry
1.4.0: Add progress indicators
```

### When to bump PATCH version (fixes):
```
1.2.1: Fix authentication error handling
1.2.2: Fix pagination bug in bookmarks list
1.2.3: Update documentation
1.2.4: Security fix in config file permissions
```