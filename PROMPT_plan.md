<context>
You are analyzing the LinkDing CLI project to produce an implementation plan. You are operating in planning-only mode — no code, no files (except the plan), no commits.

This is a gap analysis task: compare what the specs require against what currently exists in the repository, then output a prioritized, dependency-ordered task list.
</context>

<task>
Produce a single file: IMPLEMENTATION_PLAN.md

Steps:

1. Read AGENTS.md — extract all project constraints, architecture decisions, and conventions
2. Read every file in `specs/` — extract all required features, behaviors, and acceptance criteria
3. Examine all existing source code in the repository — catalog what is already implemented
4. Identify every gap between specs and current implementation
5. Write IMPLEMENTATION_PLAN.md containing a prioritized task list covering all gaps
</task>

<output_format>
Each task in IMPLEMENTATION_PLAN.md MUST use this exact format:

```markdown
- [ ] **P0** | Task Title | ~complexity
  - Acceptance: specific, testable pass/fail criteria derived from specs
  - Files: list of files to create or modify
  - Depends on: task title(s) this blocks on, or "none"
```

Priority levels:

- **P0**: Foundational/blocking — nothing else can start without these (project scaffolding, API client, config)
- **P1**: Core functionality — primary features that define the CLI
- **P2**: Enhanced features — secondary capabilities that extend core
- **P3**: Polish/optional — nice-to-haves, UX improvements, edge cases

Complexity estimates:

- ~small: < 100 lines changed
- ~medium: 100–300 lines changed
- ~large: 300+ lines changed

Ordering rules:

1. All P0 tasks first
2. Within each priority level, order by dependency chain (API client before CRUD, CRUD before import/export)
3. Group related tasks together (all bookmark CRUD tasks adjacent, all tag operations adjacent)
</output_format>

<constraints>
FORBIDDEN ACTIONS — violating any of these fails the task:
- Do NOT create, modify, or delete any file except IMPLEMENTATION_PLAN.md
- Do NOT write any implementation code
- Do NOT run any build, test, or install commands
- Do NOT make any git commits
- Do NOT add tasks that are not traceable to a gap between specs and current state

REQUIRED:

- Every task MUST trace back to a specific spec requirement or a missing foundational component
- Acceptance criteria MUST be concrete enough that another agent can verify pass/fail
- If a spec is ambiguous, note the ambiguity in the task description — do not guess
</constraints>

<done>
The task is complete when IMPLEMENTATION_PLAN.md exists, contains all identified gaps as tasks in the specified format, and no implementation has been performed.

When finished: <promise>PLANNED</promise>
</done>

