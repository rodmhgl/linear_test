# Code Style and Conventions for ldctl

## Go Code Style
- **Go Version**: 1.25 (latest features available)
- **Module Path**: github.com/rodmhgl/ldctl
- **Formatting**: Use `gofmt` and `goimports` (enforced by Makefile)
- **Line Length**: Maximum 120 characters (enforced by golines)
- **Local Imports**: Group after 3rd-party packages (github.com/rodmhgl/ldctl)

## Naming Conventions
- **Packages**: Lowercase, single word preferred (e.g., `api`, `config`, `models`)
- **Files**: Snake_case for test files (e.g., `bookmark_test.go`)
- **Functions/Methods**: CamelCase (exported) or camelCase (unexported)
- **Variables**: camelCase for local, CamelCase for exported
- **Constants**: CamelCase or ALL_CAPS for grouped constants
- **Interfaces**: End with -er suffix when possible (e.g., `Reader`, `Writer`)

## Project Structure
```
cmd/           # Cobra commands (one file per command)
internal/      # Internal packages
  api/         # LinkDing API client
  config/      # Configuration loading
  models/      # Data structures
  export/      # Import/export logic
  version/     # Version information
```

## Documentation
- All exported functions, types, and packages must have godoc comments
- Comments should be complete sentences ending with periods
- Package comments should be at the top of the doc.go or main file

## Error Handling
- Always check errors, never ignore with `_`
- Wrap errors with context using fmt.Errorf with %w verb
- Sentinel errors should be prefixed with Err
- Error types should be suffixed with Error

## Testing
- Test files end with `_test.go`
- Use table-driven tests where appropriate
- Use testify for assertions
- Use separate `_test` package for black-box testing
- Mock HTTP calls for API client tests

## Linting Rules (golangci-lint)
- Extensive linting configuration in `.golangci.yml`
- Enforces security checks (gosec)
- Checks for code complexity (cyclop, gocognit)
- Enforces error handling (errcheck)
- No global variables (gochecknoglobals)
- No init functions (gochecknoinits)

## Git Hooks (lefthook)
Pre-commit:
- gofmt (auto-fixes)
- go vet
- golangci-lint
- go mod tidy
- unit tests
- gitleaks (secret scanning)

## DO NOT
- Add database/local caching
- Add interactive/TUI mode
- Add browser integration
- Use third-party HTTP clients (stdlib is sufficient)