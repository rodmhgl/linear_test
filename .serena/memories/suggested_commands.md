# Development Commands for ldctl

## Build Commands
- `make build` - Build the binary with version information
- `make install` - Build and install binary to GOPATH/bin
- `make clean` - Remove build artifacts

## Testing Commands
- `make test` - Run all unit tests
- `make coverage` - Run tests with coverage report
- `make integration-test` - Run integration tests (requires LinkDing instance)
- `go test ./...` - Run all tests directly
- `go test -v ./cmd/...` - Run tests for specific package

## Code Quality
- `make fmt` - Format Go code
- `make vet` - Run go vet
- `make lint` - Run golangci-lint (must be installed)
- `make check` - Run all quality checks (fmt, vet, lint, test)
- `golangci-lint run ./...` - Run linter directly

## Dependency Management
- `make deps` - Install project dependencies
- `make tidy` - Run go mod tidy
- `go mod init github.com/rodmhgl/ldctl` - Initialize Go module (if not exists)
- `go get <package>` - Add new dependency

## Development Tools
- `make install-tools` - Install development tools (golangci-lint, lefthook, gitleaks)
- `lefthook install` - Setup git hooks

## Git Commands
- `git add -A` - Stage all changes
- `git commit -m "message"` - Commit staged changes
- `git push` - Push to remote
- `git pull --rebase` - Pull and rebase from remote
- `git status` - Check repository status

## Running the Application
- `make run <args>` - Build and run with arguments
- `./ldctl` - Run the built binary
- `go run main.go` - Run without building

## System Commands (Linux)
- `ls -la` - List files with details
- `pwd` - Show current directory
- `cd <dir>` - Change directory
- `find . -name "*.go"` - Find Go files
- `grep -r "pattern" .` - Search for pattern in files