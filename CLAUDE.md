# rem - CLI for macOS Reminders

## Non-Negotiables
- **Conventional Commits**: ALL commits MUST follow [Conventional Commits](https://www.conventionalcommits.org/). Format: `type(scope): description`. Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`. No exceptions.

## What is this?
Go CLI wrapping macOS Reminders. Uses `go-eventkit` (cgo + Objective-C EventKit) for fast reads AND writes (<200ms) as a single binary — including reminder CRUD and list CRUD. AppleScript only for flagged operations and default list name query. Provides CRUD for reminders/lists, natural language dates, and import/export. For programmatic Go access, use `go-eventkit` directly.

## Architecture
- `cmd/rem/commands/` - Cobra CLI commands (one file per command)
- `internal/service/` - Service layer: `reminders.go` and `lists.go` wrap `go-eventkit` client. `executor.go` runs `osascript` for flagged operations and default list name query only.
- `internal/reminder/` - Domain models: `Reminder`, `List`, `Priority`
- `internal/parser/` - Custom NL date parser (no external deps)
- `internal/export/` - JSON/CSV import/export
- `internal/ui/` - Table (`olekukonko/tablewriter` v1.x), plain, JSON output

## Critical: Architecture Rules
- **ALL reads AND writes go through `go-eventkit`** (`github.com/BRO3886/go-eventkit/reminders`) — in-process EventKit via cgo, <200ms
- **Single binary** — go-eventkit's cgo code compiled into the binary
- **AppleScript only for**: flagged operations (EventKit doesn't expose flagged), default list name query
- **EventKit doesn't expose `flagged`** - JXA fallback only used when `--flagged` filter is active, AppleScript for flag/unflag writes
- **go-eventkit field names**: `Title` (not `Name`), `Notes` (not `Body`), `List` (not `ListName`), native `URL` field
- **List CRUD via go-eventkit**: `CreateList` (auto-discovers source), `UpdateList` (ID-based), `DeleteList` (ID-based). Immutable lists are rejected.
- Priority: 0=none, 1-4=high, 5=medium, 6-9=low
- `due date` and `remind me date` are independent

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

## Website & Hosting
- Documentation site: `rem.sidv.dev` (Hugo on Cloudflare Pages)
- Source: `website/` dir, deployed via `.github/workflows/deploy.yml`
- Install script served at `rem.sidv.dev/install` (from `website/static/install`)
- Domain: `sidv.dev` (owned by user, managed on Cloudflare)

## Journal
Engineering journals live in `journals/` dir. See `.claude/commands/journal.md` for the journaling command.
