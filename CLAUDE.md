# rem - CLI for macOS Reminders

## Non-Negotiables
- **Conventional Commits**: ALL commits MUST follow [Conventional Commits](https://www.conventionalcommits.org/). Format: `type(scope): description`. Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`. No exceptions.

## What is this?
Go CLI wrapping macOS Reminders via AppleScript (writes) and JXA (reads). Provides CRUD for reminders/lists, natural language dates, import/export, and a public Go API.

## Architecture
- `cmd/rem/commands/` - Cobra CLI commands (one file per command)
- `internal/applescript/` - JXA for reads, AppleScript for writes. `executor.go` runs `osascript`.
- `internal/reminder/` - Domain models: `Reminder`, `List`, `Priority`
- `internal/parser/` - Custom NL date parser (no external deps)
- `internal/export/` - JSON/CSV import/export
- `internal/ui/` - Table (`olekukonko/tablewriter` v1.x), plain, JSON output
- `pkg/client/` - Public Go API abstracting all AppleScript complexity

## Critical: AppleScript/JXA Rules
- **NEVER use AppleScript loops for reading multiple reminders** - times out even on 8 items
- **USE JXA bulk property access**: `list.reminders.name()` gets ALL names at once, instant for 90+ items
- **AppleScript is fine for single-item writes** (create/update/delete)
- Notes field is `body`, NOT `notes`
- No `url` property exists - URLs stored in `body` with `URL: ` prefix
- Priority: 0=none, 1-4=high, 5=medium, 6-9=low
- `due date` and `remind me date` are independent
- `delete list` may fail on some macOS versions
- JXA dates use `.toISOString()` (UTC) - convert to local time on parse

## Libraries
- `spf13/cobra` - CLI framework
- `olekukonko/tablewriter` v1.x - **new API**: `NewTable()`, `.Header()`, `.Append()`, `.Render()` (NOT the old `SetHeader`/`SetBorder` API)
- `fatih/color` - terminal colors
- `olekukonko/tablewriter/tw` - alignment constants (`tw.AlignLeft`)

## Build & Test
```bash
make build        # -> bin/rem
go test ./...     # unit tests (date parser, export, models)
make completions  # bash/zsh/fish
```

## Conventions
- Short IDs displayed as first 8 chars of full `x-apple-reminder://UUID` ID
- Prefix matching: users can pass partial IDs to any command
- All commands support `-o json|table|plain`
- `NO_COLOR` env var respected

## Journal
Engineering journals live in `journals/` dir. See `.claude/commands/journal.md` for the journaling command.
