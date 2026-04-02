# Task Completion Workflow

## After Code Changes
1. **Run quality gates**:
   - `make fmt` - Format code
   - `make vet` - Run go vet
   - `make lint` - Run linter (if golangci-lint installed)
   - `make test` - Run unit tests
   - `make check` - Run all checks at once

2. **Commit changes**:
   - `git add -A` - Stage all changes
   - `git commit -m "feat: description"` - Commit with descriptive message
   - Commit message format: `type: description` (feat, fix, docs, refactor, test, chore)

3. **Push to remote**:
   - `git pull --rebase` - Get latest changes
   - `git push` - Push to remote repository
   - `git status` - Verify "up to date with origin"

## End of Session (Landing the Plane)
**MANDATORY WORKFLOW** (from AGENTS.md):

1. **File issues for remaining work** - Create issues for anything needing follow-up
2. **Run quality gates** - Tests, linters, builds (if code changed)
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** (MANDATORY):
   ```bash
   git pull --rebase
   bd sync  # if applicable
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

## Critical Rules
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

## Integration Testing
When integration tests are needed:
- Ensure LinkDing instance is available
- Set environment variables for API URL and token
- Run: `make integration-test` or `go test -tags=integration ./...`

## Version Tagging
After successful builds and tests:
- If no tags exist, start at 0.0.0
- Increment patch version (e.g., 0.0.1, 0.0.2)
- Create tag: `git tag v0.0.X`
- Push tag: `git push --tags`