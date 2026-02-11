# rem - CLI for macOS Reminders

## Non-Negotiables
- **Conventional Commits**: ALL commits MUST follow [Conventional Commits](https://www.conventionalcommits.org/). Format: `type(scope): description`. Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`. No exceptions.

## What is this?
Go CLI wrapping macOS Reminders. Uses `go-eventkit` (cgo + Objective-C EventKit) for fast reads AND writes (<200ms) as a single binary. AppleScript only for list CRUD and flagged operations. Provides CRUD for reminders/lists, natural language dates, import/export, and a public Go API.

## Architecture
- `cmd/rem/commands/` - Cobra CLI commands (one file per command)
- `internal/service/` - Service layer: `reminders.go` and `lists.go` wrap `go-eventkit` client. `executor.go` runs `osascript` for list CRUD and flagged operations only.
- `internal/reminder/` - Domain models: `Reminder`, `List`, `Priority`
- `internal/parser/` - Custom NL date parser (no external deps)
- `internal/export/` - JSON/CSV import/export
- `internal/ui/` - Table (`olekukonko/tablewriter` v1.x), plain, JSON output
- `pkg/client/` - Public Go API abstracting all complexity

## Critical: Architecture Rules
- **ALL reads AND writes go through `go-eventkit`** (`github.com/BRO3886/go-eventkit/reminders`) — in-process EventKit via cgo, <200ms
- **Single binary** — go-eventkit's cgo code compiled into the binary
- **AppleScript only for**: list create/rename/delete (go-eventkit doesn't support list CRUD), flagged operations (EventKit doesn't expose flagged), default list name query
- **EventKit doesn't expose `flagged`** - JXA fallback only used when `--flagged` filter is active, AppleScript for flag/unflag writes
- **go-eventkit field names**: `Title` (not `Name`), `Notes` (not `Body`), `List` (not `ListName`), native `URL` field
- Priority: 0=none, 1-4=high, 5=medium, 6-9=low
- `due date` and `remind me date` are independent
- `delete list` may fail on some macOS versions

## Libraries
- `BRO3886/go-eventkit` - **EventKit bindings** (cgo + ObjC, reads AND writes)
- `spf13/cobra` - CLI framework
- `olekukonko/tablewriter` v1.x - **new API**: `NewTable()`, `.Header()`, `.Append()`, `.Render()` (NOT the old `SetHeader`/`SetBorder` API)
- `fatih/color` - terminal colors
- `olekukonko/tablewriter/tw` - alignment constants (`tw.AlignLeft`)

## Build & Test
```bash
make build        # -> bin/rem (single binary, includes EventKit via cgo)
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
