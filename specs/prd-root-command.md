# PRD: Root Command and Global Flags for ldctl

## Executive Summary

This PRD defines the root command behavior and global flags for the ldctl CLI tool. It establishes standards for help text formatting, global flag behavior, output conventions, and how global flags interact with all subcommands. This specification serves as the foundation that all other command PRDs reference for consistency.

## Problem Statement

The ldctl project currently lacks specification for:
1. Root command behavior when no subcommand is provided
2. Global flags available across all commands
3. Help text formatting standards
4. Flag precedence and interaction rules
5. Output routing conventions (stdout vs stderr)
6. Consistent formatting for human-readable vs machine-readable output

Without these standards, the CLI will have:
- Inconsistent user experience across commands
- Unpredictable flag behavior
- Non-standard help formatting
- Confusion about output destinations
- Difficulty in scripting and automation

## Goals

1. **Define root command behavior** - Clear action when no subcommand provided
2. **Standardize global flags** - Consistent flags available everywhere
3. **Establish help conventions** - Uniform help text formatting
4. **Specify flag interactions** - How global flags work together
5. **Define output standards** - When to use stdout vs stderr

### Success Metrics

| Metric | Target |
|--------|--------|
| Global flag availability | 100% of commands support all global flags |
| Help text consistency | All commands follow formatting standards |
| Output routing accuracy | 100% data to stdout, diagnostics to stderr |
| Flag interaction correctness | No conflicting flag combinations allowed |
| JSON output validity | All --json output is valid, parseable JSON |

## User Stories

### Story 1: User Explores CLI
**As a** new ldctl user
**I want to** run `ldctl` without arguments
**So that** I can see available commands and learn how to use the tool

**Acceptance Criteria:**
- `ldctl` with no args shows help text
- Help includes list of available commands
- Help shows global flags
- Version information is displayed
- Examples are provided

### Story 2: User Needs JSON Output
**As a** script author
**I want to** use `--json` flag with any command
**So that** I can parse output programmatically

**Acceptance Criteria:**
- `--json` available on all commands
- Output is valid JSON
- Errors also output as JSON
- Pretty-printed for readability

### Story 3: User Wants Quiet Operation
**As a** automation developer
**I want to** suppress non-essential output
**So that** I can use ldctl in scripts cleanly

**Acceptance Criteria:**
- `--quiet` suppresses progress messages
- Only essential output remains
- Errors still displayed
- Exit codes unchanged

## Functional Requirements

### REQ-001: Root Command Behavior [MUST]
When `ldctl` is invoked without subcommands, it shall display help text.

**Behavior:**
```bash
$ ldctl
ldctl - LinkDing CLI client (version 1.2.3)

A command-line interface for managing bookmarks, tags, and assets
in your LinkDing instance.

Usage:
  ldctl [command]

Available Commands:
  config      Manage LinkDing configuration
  bookmarks   Manage bookmarks (alias: bm)
  tags        Manage tags
  bundles     Manage bookmark bundles
  assets      Manage bookmark assets
  user        View user profile
  version     Show version information
  help        Help about any command

Global Flags:
  --json      Output in JSON format
  --quiet     Suppress non-essential output
  --verbose   Show detailed diagnostic output
  --version   Show version information
  --help      Show help for command

Use "ldctl [command] --help" for more information about a command.

Examples:
  # Initialize configuration
  ldctl config init
  
  # List bookmarks
  ldctl bookmarks list
  
  # Add a bookmark
  ldctl bookmarks add https://example.com --tags "example,demo"

For more information, visit: https://github.com/rodmhgl/ldctl
```

### REQ-002: Global JSON Flag [MUST]
A `--json` flag shall be available globally to enable JSON output format.

**Specification:**
- **Flag**: `--json`
- **Type**: Boolean
- **Default**: false
- **Availability**: All commands and subcommands
- **Behavior**: 
  - Changes output format to JSON
  - Errors also output as JSON to stderr
  - Pretty-printed with 2-space indentation
  - Always valid JSON (even if empty)

**Examples:**
```bash
# Human readable (default)
$ ldctl bookmarks get 123
ID:          123
URL:         https://example.com
Title:       Example Site
Tags:        demo, example

# JSON output
$ ldctl bookmarks get 123 --json
{
  "id": 123,
  "url": "https://example.com",
  "title": "Example Site",
  "tag_names": ["demo", "example"]
}
```

### REQ-003: Global Quiet Flag [SHOULD]
A `--quiet` flag shall suppress non-essential output for scripting.

**Specification:**
- **Flag**: `--quiet`, `-q`
- **Type**: Boolean
- **Default**: false
- **Suppresses**:
  - Progress indicators
  - Confirmation messages (unless critical)
  - Informational messages
  - Warnings (unless critical)
- **Still Shows**:
  - Requested data output
  - Error messages
  - Critical warnings

**Examples:**
```bash
# Normal output
$ ldctl bookmarks export --format json
Fetching bookmarks... [1/5 pages]
Fetching bookmarks... [2/5 pages]
...
Exported 483 bookmarks to bookmarks.json

# Quiet mode
$ ldctl bookmarks export --format json --quiet
# (only creates file, no progress messages)
```

### REQ-004: Global Verbose Flag [SHOULD]
A `--verbose` flag shall enable detailed diagnostic output.

**Specification:**
- **Flag**: `--verbose`, `-v`
- **Type**: Boolean (future: count for levels)
- **Default**: false
- **Shows**:
  - HTTP request/response details
  - Configuration loading steps
  - Timing information
  - Detailed error context
  - File operations
- **Output**: Always to stderr

**Examples:**
```bash
$ ldctl bookmarks list --verbose
[DEBUG] Loading config from /home/user/.config/ldctl/config.toml
[DEBUG] Config loaded: url=https://linkding.example.com
[DEBUG] GET https://linkding.example.com/api/bookmarks/?limit=100&offset=0
[DEBUG] Response: 200 OK (234ms)
[DEBUG] Received 100 bookmarks, total count: 483
...
```

### REQ-005: Global Version Flag [MUST]
A `--version` flag shall display version information and exit.

**Specification:**
- **Flag**: `--version`
- **Type**: Boolean
- **Behavior**:
  - Shows version info and exits immediately
  - Works from any command context
  - Exits with code 0
  - Takes precedence over other operations

**Example:**
```bash
$ ldctl --version
ldctl version 1.2.3 (commit a1b2c3d, built 2025-01-27T10:30:00Z, go1.25.0)

$ ldctl bookmarks list --version
ldctl version 1.2.3 (commit a1b2c3d, built 2025-01-27T10:30:00Z, go1.25.0)
```

### REQ-006: Global Help Flag [MUST]
A `--help` flag shall display context-sensitive help.

**Specification:**
- **Flag**: `--help`, `-h`
- **Type**: Boolean
- **Behavior**:
  - Shows help for current command
  - Available on all commands
  - Exits with code 0
  - Takes precedence over operations

### REQ-007: Flag Mutual Exclusivity [MUST]
Certain global flags shall be mutually exclusive.

**Rules:**
- `--quiet` and `--verbose` are mutually exclusive
- `--help` and `--version` short-circuit execution
- `--json` is compatible with all other flags

**Error Example:**
```bash
$ ldctl bookmarks list --quiet --verbose
Error: cannot use --quiet and --verbose together
```

### REQ-008: Flag Precedence [MUST]
Global flags shall have defined precedence order.

**Precedence (highest to lowest):**
1. `--help` - Shows help and exits
2. `--version` - Shows version and exits
3. `--json` - Affects output format
4. `--quiet` or `--verbose` - Affects output verbosity

### REQ-009: Output Routing [MUST]
Output shall be consistently routed to stdout or stderr.

**stdout (data):**
- Command results
- Requested information
- JSON output
- Help text
- Version info

**stderr (diagnostics):**
- Error messages
- Warning messages
- Progress indicators
- Verbose/debug output
- Confirmations/prompts

### REQ-010: JSON Output Standards [MUST]
JSON output shall follow consistent standards.

**Standards:**
- Always valid JSON
- Pretty-printed (2-space indent)
- Consistent key naming (snake_case)
- Nulls represented as `null`, not omitted
- Empty arrays as `[]`, not omitted
- Timestamps in ISO 8601 format

**Example:**
```json
{
  "id": 123,
  "url": "https://example.com",
  "title": "Example",
  "description": null,
  "tags": [],
  "created_at": "2025-01-27T10:30:00Z"
}
```

### REQ-011: Help Text Format [MUST]
Help text shall follow a consistent template.

**Template:**
```
<description>

Usage:
  <command> [flags]
  <command> [command]

Available Commands:
  <command>   <description>

Flags:
  <flags and descriptions>

Global Flags:
  <inherited global flags>

Examples:
  <practical examples>

Use "<command> [subcommand] --help" for more information
```

### REQ-012: Exit Code Standards [MUST]
Commands shall use consistent exit codes (per prd-error-handling.md).

**Exit Codes:**
- 0: Success
- 1: General error (API error, not found, etc.)
- 2: Usage error (bad flags, missing args, config error)

## Non-Functional Requirements

### Performance
- Root command help display < 50ms
- Flag parsing overhead < 10ms
- No network calls for help/version

### Usability
- Help text fits in 80-column terminal
- Examples use realistic scenarios
- Error messages are actionable
- Consistent terminology

### Compatibility
- POSIX-compliant flag syntax
- Works in all common shells
- UTF-8 output support
- Windows terminal compatibility

## Technical Considerations

### Implementation with Cobra

**Root Command Setup:**
```go
// cmd/root.go
package cmd

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "github.com/rodmhgl/ldctl/internal/version"
)

var (
    jsonFlag    bool
    quietFlag   bool
    verboseFlag bool
    versionFlag bool
)

var rootCmd = &cobra.Command{
    Use:   "ldctl",
    Short: "LinkDing CLI client",
    Long: `ldctl - LinkDing CLI client (version ` + version.Version + `)

A command-line interface for managing bookmarks, tags, and assets
in your LinkDing instance.`,
    
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Handle version flag globally
        if versionFlag {
            fmt.Println(version.String())
            os.Exit(0)
        }
        
        // Check mutual exclusivity
        if quietFlag && verboseFlag {
            return fmt.Errorf("cannot use --quiet and --verbose together")
        }
        
        // Load configuration (except for certain commands)
        if cmd.Name() != "config" && cmd.Name() != "version" {
            if err := loadConfig(); err != nil {
                return err
            }
        }
        
        return nil
    },
    
    // Show help when no subcommand
    Run: func(cmd *cobra.Command, args []string) {
        cmd.Help()
    },
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, 
        "Output in JSON format")
    rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false,
        "Suppress non-essential output")
    rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false,
        "Show detailed diagnostic output")
    rootCmd.PersistentFlags().BoolVar(&versionFlag, "version", false,
        "Show version information")
    
    // Configure help
    rootCmd.SetHelpTemplate(customHelpTemplate)
    
    // Disable default completion command
    rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        handleError(err)
        os.Exit(1)
    }
}
```

### Global Flag Access in Subcommands

```go
// Any subcommand can access global flags
func runListCommand(cmd *cobra.Command, args []string) error {
    // Access global flags from root
    if jsonFlag {
        return outputJSON(bookmarks)
    }
    
    if !quietFlag {
        showProgress("Fetching bookmarks...")
    }
    
    if verboseFlag {
        log.Printf("DEBUG: Fetching from %s", apiURL)
    }
    
    // ... rest of command
}
```

### Output Helper Functions

```go
// internal/output/output.go
package output

import (
    "encoding/json"
    "fmt"
    "os"
)

// PrintJSON outputs data as JSON to stdout
func PrintJSON(data interface{}) error {
    encoder := json.NewEncoder(os.Stdout)
    encoder.SetIndent("", "  ")
    return encoder.Encode(data)
}

// PrintData outputs data to stdout (for piping)
func PrintData(format string, args ...interface{}) {
    fmt.Fprintf(os.Stdout, format, args...)
}

// PrintError outputs to stderr
func PrintError(format string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, "Error: "+format, args...)
}

// PrintProgress outputs to stderr if not quiet
func PrintProgress(message string) {
    if !quietFlag {
        fmt.Fprintln(os.Stderr, message)
    }
}

// PrintVerbose outputs to stderr if verbose
func PrintVerbose(format string, args ...interface{}) {
    if verboseFlag {
        fmt.Fprintf(os.Stderr, "[DEBUG] "+format, args...)
    }
}
```

### Help Template Customization

```go
var customHelpTemplate = `{{.Long}}

Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
```

## Implementation Roadmap

### Phase 1: Root Command Setup (Day 1)
1. Create `cmd/root.go` with Cobra setup
2. Define all global flags
3. Implement flag mutual exclusivity
4. Setup help template
5. Write tests for root command

### Phase 2: Output Infrastructure (Day 2)
1. Create `internal/output/` package
2. Implement output routing helpers
3. Add JSON formatting utilities
4. Setup verbose/quiet helpers
5. Test output functions

### Phase 3: Integration (Day 3)
1. Wire global flags to all subcommands
2. Test flag inheritance
3. Verify help text formatting
4. Ensure consistent exit codes
5. Update documentation

## Testing Strategy

### Unit Tests
- Test flag parsing
- Test mutual exclusivity
- Test help text generation
- Test version display
- Test output routing

### Integration Tests
- Global flags work on all subcommands
- Help accessible everywhere
- Version flag short-circuits
- JSON output is valid
- Exit codes are correct

### Manual Testing Checklist
- [ ] `ldctl` shows help
- [ ] `ldctl --version` shows version
- [ ] `ldctl --help` shows help
- [ ] `ldctl bookmarks --json` works
- [ ] `ldctl bookmarks list --quiet` suppresses output
- [ ] `ldctl bookmarks list --verbose` shows debug info
- [ ] `ldctl --quiet --verbose` shows error
- [ ] JSON output is pretty-printed
- [ ] Errors go to stderr
- [ ] Data goes to stdout

## Documentation Requirements

### README.md
- Document all global flags
- Show examples with global flags
- Explain output modes
- Describe exit codes

### Man Page
- Include global flags section
- Document flag interactions
- Provide examples

### Help Text
- Every command must have help
- Include practical examples
- Reference related commands
- Show available flags

## Out of Scope

This PRD does not cover:
- Shell completion (separate PRD)
- Configuration file for default flags
- Command aliases beyond documented ones
- Interactive mode flags
- Color output flags
- Pager integration
- Plugins or extensions

## Open Questions

1. **Should --verbose support levels (-v, -vv, -vvv)?**
   - Decision: Start with boolean, add levels if needed

2. **Should we add --dry-run flag globally?**
   - Decision: No, add per-command where relevant

3. **Should --no-config flag exist globally?**
   - Decision: Not in v1, config is required

4. **Should we support short flag combining (-qv)?**
   - Decision: Let Cobra handle it by default

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Flag conflicts with subcommands | Confusion | Reserve global flag names |
| Inconsistent flag behavior | Poor UX | Central flag handling |
| Output routing errors | Broken pipes | Clear stdout/stderr rules |
| Help text too long | Poor readability | Template with pagination |

## Dependencies

- Cobra CLI framework
- Go 1.21+ (for structured logging)
- Version package (from prd-version.md)
- Error package (from prd-error-handling.md)

## Validation Checkpoints

### Phase 1 Complete
- [ ] Root command shows help
- [ ] All global flags defined
- [ ] Flag validation works
- [ ] Help template customized

### Phase 2 Complete
- [ ] Output helpers created
- [ ] JSON formatting works
- [ ] Verbose/quiet implemented
- [ ] Routing correct

### Phase 3 Complete
- [ ] All subcommands integrated
- [ ] Flags work everywhere
- [ ] Documentation updated
- [ ] Tests pass

## Acceptance Criteria

- [ ] `ldctl` without args shows help
- [ ] All global flags available on all commands
- [ ] `--json` outputs valid JSON everywhere
- [ ] `--quiet` suppresses non-essential output
- [ ] `--verbose` shows diagnostic information
- [ ] `--version` works from any command
- [ ] `--help` shows context-sensitive help
- [ ] `--quiet` and `--verbose` are mutually exclusive
- [ ] stdout receives data only
- [ ] stderr receives diagnostics only
- [ ] Exit codes follow standards
- [ ] Help text follows template
- [ ] Tests cover all scenarios
- [ ] Documentation complete

## References

- [Cobra Documentation](https://cobra.dev/)
- [POSIX Utility Conventions](https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap12.html)
- [GNU Coding Standards](https://www.gnu.org/prep/standards/standards.html#Command_002dLine-Interfaces)
- [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46)
- [prd-error-handling.md](./prd-error-handling.md)
- [prd-version.md](./prd-version.md)

## Appendix: Example Command Outputs

### Default Help Output
```
$ ldctl
ldctl - LinkDing CLI client (version 1.2.3)

A command-line interface for managing bookmarks, tags, and assets
in your LinkDing instance.

Usage:
  ldctl [command]

Available Commands:
  config      Manage LinkDing configuration
  bookmarks   Manage bookmarks (alias: bm)
  tags        Manage tags
  bundles     Manage bookmark bundles
  assets      Manage bookmark assets
  user        View user profile
  version     Show version information
  help        Help about any command

Global Flags:
  --json      Output in JSON format
  --quiet     Suppress non-essential output
  --verbose   Show detailed diagnostic output
  --version   Show version information
  --help      Show help for command

Examples:
  # Initialize configuration
  ldctl config init
  
  # List recent bookmarks
  ldctl bookmarks list --limit 10
  
  # Search bookmarks
  ldctl bookmarks list --query "#important !#archived"
  
  # Add a bookmark with tags
  ldctl bookmarks add https://example.com --tags "work,reference"

Use "ldctl [command] --help" for more information about a command.

For more information, visit: https://github.com/rodmhgl/ldctl
```

### Subcommand Help
```
$ ldctl bookmarks --help
Manage bookmarks in your LinkDing instance

Usage:
  ldctl bookmarks [command]

Aliases:
  bookmarks, bm

Available Commands:
  list        List bookmarks
  get         Get a single bookmark
  add         Add a new bookmark
  check       Check if URL exists
  update      Update a bookmark
  archive     Archive bookmarks
  unarchive   Unarchive bookmarks
  delete      Delete bookmarks
  open        Open bookmark in browser
  export      Export bookmarks
  import      Import bookmarks

Flags:
  -h, --help   help for bookmarks

Global Flags:
  --json      Output in JSON format
  --quiet     Suppress non-essential output
  --verbose   Show detailed diagnostic output

Use "ldctl bookmarks [command] --help" for more information about a command.
```

### JSON Output Example
```
$ ldctl bookmarks get 123 --json
{
  "id": 123,
  "url": "https://example.com",
  "title": "Example Website",
  "description": "An example website for demonstration",
  "notes": "",
  "web_archive_snapshot_url": "",
  "favicon_url": "https://example.com/favicon.ico",
  "preview_image_url": null,
  "is_archived": false,
  "unread": false,
  "shared": false,
  "tag_names": [
    "example",
    "demo"
  ],
  "date_added": "2025-01-27T10:30:00Z",
  "date_modified": "2025-01-27T10:30:00Z"
}
```

### Verbose Output Example
```
$ ldctl bookmarks list --limit 1 --verbose
[DEBUG] Loading config from /home/user/.config/ldctl/config.toml
[DEBUG] Config loaded successfully
[DEBUG] Base URL: https://linkding.example.com
[DEBUG] Token: abc***xyz (masked)
[DEBUG] Building request: GET /api/bookmarks/?limit=1
[DEBUG] Request headers:
[DEBUG]   Authorization: Token ****
[DEBUG]   Accept: application/json
[DEBUG]   User-Agent: ldctl/1.2.3
[DEBUG] Sending request...
[DEBUG] Response received: 200 OK (145ms)
[DEBUG] Response headers:
[DEBUG]   Content-Type: application/json
[DEBUG]   X-Total-Count: 483
[DEBUG] Parsing response body...
[DEBUG] Found 1 bookmarks (total: 483)

ID: 123
URL: https://example.com
Title: Example Website
Tags: example, demo
Added: 2025-01-27 10:30:00
```

### Quiet Mode Example
```
$ ldctl bookmarks export bookmarks.json
Fetching bookmarks... [1/5 pages]
Fetching bookmarks... [2/5 pages]
Fetching bookmarks... [3/5 pages]
Fetching bookmarks... [4/5 pages]
Fetching bookmarks... [5/5 pages]
Exported 483 bookmarks to bookmarks.json

$ ldctl bookmarks export bookmarks.json --quiet
$ # (no output, just creates file)
```