# rem - CLI for macOS Reminders

## Non-Negotiables
- **Conventional Commits**: ALL commits MUST follow [Conventional Commits](https://www.conventionalcommits.org/). Format: `type(scope): description`. Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`. No exceptions.

## What is this?
Go CLI wrapping macOS Reminders. Uses a compiled Swift EventKit helper for fast reads (<200ms) and AppleScript for writes. Provides CRUD for reminders/lists, natural language dates, import/export, and a public Go API.

## Architecture
- `cmd/rem/commands/` - Cobra CLI commands (one file per command)
- `internal/swift/helper.swift` - Compiled Swift binary using EventKit for all reads (lists, reminders, get). ~100KB, <200ms for any query.
- `internal/applescript/` - Go layer: `executor.go` runs the Swift helper for reads, `osascript` for writes. `reminders.go` and `lists.go` parse helper JSON output.
- `internal/reminder/` - Domain models: `Reminder`, `List`, `Priority`
- `internal/parser/` - Custom NL date parser (no external deps)
- `internal/export/` - JSON/CSV import/export
- `internal/ui/` - Table (`olekukonko/tablewriter` v1.x), plain, JSON output
- `pkg/client/` - Public Go API abstracting all complexity

## Critical: Architecture Rules
- **ALL reads go through `bin/reminders-helper`** (Swift EventKit) - instant for any number of reminders
- **Writes use AppleScript** (create/update/delete) - single-item operations, AppleScript syntax is simpler
- **EventKit doesn't expose `flagged`** - JXA fallback only used when `--flagged` filter is active
- **NEVER use JXA/AppleScript for reads** - osascript from Go exec.Command is 10-50x slower than EventKit
- Notes field is `body`, NOT `notes` (in AppleScript)
- No `url` property exists - URLs stored in `body` with `URL: ` prefix
- Priority: 0=none, 1-4=high, 5=medium, 6-9=low
- `due date` and `remind me date` are independent
- `delete list` may fail on some macOS versions

## Libraries
- `spf13/cobra` - CLI framework
- `olekukonko/tablewriter` v1.x - **new API**: `NewTable()`, `.Header()`, `.Append()`, `.Render()` (NOT the old `SetHeader`/`SetBorder` API)
- `fatih/color` - terminal colors
- `olekukonko/tablewriter/tw` - alignment constants (`tw.AlignLeft`)

## Build & Test
```bash
make build        # -> bin/rem + bin/reminders-helper
make build-helper # Swift helper only
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
