# rem

A blazing fast CLI for macOS Reminders. Sub-200ms reads AND writes via EventKit, natural language dates, import/export, and a public Go API — all in a single binary.

**[Documentation](https://rem.sidv.dev)** | **[Architecture](https://rem.sidv.dev/docs/architecture/)** | **[Go API](https://rem.sidv.dev/docs/api/)**

## Features

- **Sub-200ms reads AND writes** — EventKit via cgo (go-eventkit), direct memory access, no IPC
- **Single binary** — EventKit compiled in via cgo, no helper processes
- **Natural language dates** — `tomorrow`, `next friday at 2pm`, `in 3 hours`, `eod`
- **19 commands** — full CRUD, search, stats, overdue, upcoming, interactive mode
- **Multiple output formats** — table, JSON, plain text
- **Import/Export** — JSON and CSV with full property round-trip
- **Public Go API** — powered by [go-eventkit](https://github.com/BRO3886/go-eventkit) for programmatic access
- **Shell completions** — bash, zsh, fish

## Installation

### From source

```bash
go install github.com/BRO3886/rem/cmd/rem@latest
```

### Build locally

```bash
git clone https://github.com/BRO3886/rem.git
cd rem
make build
# Binary is at ./bin/rem
```

## Requirements

- macOS 10.12+ (uses EventKit for reads and writes via go-eventkit, AppleScript for list CRUD only)
- Go 1.21+ (for building from source)
- Xcode Command Line Tools (cgo/clang + framework headers)
- First run will prompt for Reminders app access in System Settings > Privacy & Security

## Quick Start

```bash
# List all reminder lists
rem lists --count

# Create a reminder
rem add "Buy groceries" --list Personal --due tomorrow --priority high

# List incomplete reminders
rem list --list Work --incomplete

# Search reminders
rem search "meeting"

# Show reminder details
rem show <id>

# Complete a reminder
rem complete <id>

# Show statistics
rem stats
```

## Commands

### Reminders

```bash
# Create
rem add "Title" [--list LIST] [--due DATE] [--priority high|medium|low] [--notes TEXT] [--url URL] [--flagged]
rem add -i                          # Interactive creation

# List
rem list [--list LIST] [--incomplete] [--completed] [--flagged] [--due-before DATE] [--due-after DATE] [-o json|table|plain]
rem ls                              # Alias

# Show
rem show <id>                       # Full or partial ID
rem get <id> -o json

# Update
rem update <id> [--name TEXT] [--due DATE] [--priority LEVEL] [--notes TEXT] [--url URL]

# Complete / Uncomplete
rem complete <id>
rem done <id>                       # Alias
rem uncomplete <id>

# Flag / Unflag
rem flag <id>
rem unflag <id>

# Delete
rem delete <id>                     # Asks for confirmation
rem rm <id> --force                 # Skip confirmation
```

### Lists

```bash
# View all lists
rem lists
rem lists --count                   # Show reminder counts

# Create a list
rem list-mgmt create "My List"
rem lm new "Shopping"               # Alias

# Rename a list
rem list-mgmt rename "Old Name" "New Name"

# Delete a list
rem list-mgmt delete "Name"         # Asks for confirmation
rem lm rm "Name" --force
```

### Search & Analytics

```bash
rem search "query" [--list LIST] [--incomplete]
rem stats                           # Overall statistics
rem overdue                         # Overdue reminders
rem upcoming [--days 7]             # Upcoming due dates
```

### Import / Export

```bash
# Export
rem export --list Work --format json > work.json
rem export --format csv --output-file reminders.csv
rem export --incomplete --format json

# Import
rem import work.json
rem import reminders.csv --list "Imported"
rem import --dry-run data.json      # Preview without creating
```

### Interactive Mode

```bash
rem interactive                     # Full interactive menu
rem i                               # Alias
rem add -i                          # Interactive add
```

### Output Formats

All list/show commands support `--output` (`-o`):

```bash
rem list -o table                   # Default, formatted table
rem list -o json                    # Machine-readable JSON
rem list -o plain                   # Simple text
rem list -o json | jq '.[].name'   # Pipe to jq
```

Color output respects `NO_COLOR`:
```bash
NO_COLOR=1 rem list
rem list --no-color
```

### Shell Completions

```bash
# Bash
rem completion bash > /usr/local/etc/bash_completion.d/rem

# Zsh
rem completion zsh > "${fpath[1]}/_rem"

# Fish
rem completion fish > ~/.config/fish/completions/rem.fish
```

## Date Parsing

rem supports natural language dates:

| Input | Meaning |
|-------|---------|
| `today` | Today at 9:00 AM |
| `tomorrow` | Tomorrow at 9:00 AM |
| `next monday` | Next Monday at 9:00 AM |
| `next friday at 2pm` | Next Friday at 2:00 PM |
| `in 2 days` | 2 days from now |
| `in 3 hours` | 3 hours from now |
| `in 30 minutes` | 30 minutes from now |
| `eod` / `end of day` | Today at 5:00 PM |
| `next week` | 7 days from now |
| `next month` | 1 month from now |
| `5pm` | Today (or tomorrow) at 5:00 PM |
| `2026-02-15` | February 15, 2026 |
| `2026-02-15 14:30` | February 15, 2026 at 2:30 PM |

## Public Go API

For programmatic access to macOS Reminders, use [**go-eventkit**](https://github.com/BRO3886/go-eventkit) directly — the same library that powers rem and [cal](https://github.com/BRO3886/cal):

```bash
go get github.com/BRO3886/go-eventkit
```

```go
package main

import (
    "fmt"
    "time"

    "github.com/BRO3886/go-eventkit/reminders"
)

func main() {
    client, err := reminders.New()
    if err != nil {
        panic(err)
    }

    // Create a reminder
    due := time.Now().Add(24 * time.Hour)
    r, err := client.CreateReminder(reminders.CreateReminderInput{
        Title:    "Buy groceries",
        ListName: "Personal",
        DueDate:  &due,
        Priority: reminders.PriorityHigh,
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("Created:", r.ID)

    // List incomplete reminders
    items, _ := client.Reminders(
        reminders.WithList("Personal"),
        reminders.WithCompleted(false),
    )
    for _, item := range items {
        fmt.Printf("- %s (due: %v)\n", item.Title, item.DueDate)
    }

    // Complete a reminder
    client.CompleteReminder(r.ID)

    // Get all lists
    lists, _ := client.Lists()
    for _, l := range lists {
        fmt.Printf("%s (%d reminders)\n", l.Title, l.Count)
    }
}
```

See the [go-eventkit README](https://github.com/BRO3886/go-eventkit) for the full API reference.

## Architecture

```
rem/
├── cmd/rem/              # CLI entry point
│   ├── main.go
│   └── commands/         # Cobra command definitions
├── internal/
│   ├── service/          # Service layer wrapping go-eventkit + AppleScript for list CRUD
│   ├── reminder/         # Domain models (Reminder, List, Priority)
│   ├── parser/           # Natural language date parsing
│   ├── export/           # JSON & CSV import/export
│   └── ui/               # Table formatting, colored output
├── website/              # Hugo documentation site
├── Makefile
├── LICENSE
└── README.md
```

**Reads and writes** go through `go-eventkit` (`github.com/BRO3886/go-eventkit`) — an Objective-C EventKit bridge compiled into the binary via cgo. Direct in-process access to the Reminders store, no IPC. All operations complete in under 200ms.

**List CRUD and flagged** use AppleScript via `osascript` — go-eventkit does not support list create/rename/delete, and EventKit doesn't expose the flagged property.

## Performance

Tested with 224 reminders across 12 lists:

| Command | Time |
|---------|------|
| `rem lists` | 0.12s |
| `rem list` (all 224) | 0.13s |
| `rem show` (by prefix) | 0.11s |
| `rem search` | 0.11s |
| `rem stats` | 0.17s |

See [Performance docs](https://rem.sidv.dev/docs/performance/) for the full optimization story (JXA at 60s → EventKit at 0.13s).

## Known Limitations

- **macOS only** — requires EventKit framework and osascript
- **No tags/subtasks** — not exposed via EventKit
- **`--flagged` filter is slow** (~3-4s) — EventKit doesn't expose `flagged`, falls back to JXA
- **List CRUD uses AppleScript** — go-eventkit doesn't support list create/rename/delete
- **List deletion** may fail on some macOS versions

## License

MIT
