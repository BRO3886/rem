# rem

A blazing fast CLI for macOS Reminders. Sub-200ms reads AND writes via EventKit, natural language dates, and import/export — all in a single binary.

**[Documentation](https://rem.sidv.dev)** | **[Architecture](https://rem.sidv.dev/docs/architecture/)** | **[go-eventkit](https://github.com/BRO3886/go-eventkit)**

## Features

- **Sub-200ms reads AND writes** — EventKit via cgo (go-eventkit), direct memory access, no IPC
- **Single binary** — EventKit compiled in via cgo, no helper processes
- **Natural language dates** — `tomorrow`, `next friday at 2pm`, `in 3 hours`, `eod`
- **20 commands** — full CRUD, search, stats, overdue, upcoming, interactive mode
- **Multiple output formats** — table, JSON, plain text
- **Native tags** — `#hashtag` in titles or `--tags` flag, stored as real Reminders.app tags
- **Shared list support** — full CRUD on shared lists, sharing state in `rem lists`, and moves across the shared-list boundary via copy (macOS has no true move there)
- **Import/Export** — JSON and CSV with full property round-trip (including tags)
- **Powered by [go-eventkit](https://github.com/BRO3886/go-eventkit)** — use the same library directly for programmatic Go access
- **Shell completions** — bash, zsh, fish

## Installation

### Homebrew

```bash
brew tap BRO3886/tap
brew install rem-cli
```

### Quick install (recommended)

```bash
curl -fsSL https://rem.sidv.dev/install | bash
```

Downloads the latest release, extracts, and installs to `~/.local/bin` (override with `INSTALL_DIR=...`). No sudo needed.

### Via Go

```bash
go install github.com/BRO3886/rem/cmd/rem@latest
```

Requires Go 1.21+ and Xcode Command Line Tools (cgo compiles EventKit bindings).

### Manual download

Download from [GitHub Releases](https://github.com/BRO3886/rem/releases):

```bash
# Apple Silicon
curl -LO https://github.com/BRO3886/rem/releases/latest/download/rem-darwin-arm64.tar.gz
tar xzf rem-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv rem ~/.local/bin/rem

# Intel
curl -LO https://github.com/BRO3886/rem/releases/latest/download/rem-darwin-amd64.tar.gz
tar xzf rem-darwin-amd64.tar.gz
mkdir -p ~/.local/bin && mv rem ~/.local/bin/rem
```

### Build from source

```bash
git clone https://github.com/BRO3886/rem.git
cd rem
make build
# Binary is at ./bin/rem
```

## Requirements

- macOS 13+ (uses EventKit + private ReminderKit bridge for all reads and writes via go-eventkit, AppleScript only for default list name query)
- Xcode Command Line Tools (for building from source — cgo/clang + framework headers)
- First run will prompt for Reminders app access in System Settings > Privacy & Security

## Quick Start

```bash
# List all reminder lists
rem lists --count

# Create a reminder with an alarm
rem add "Buy groceries" --list Personal --due tomorrow --priority high --remind-me 15m

# Create with tags (parsed from title + --tags flag)
rem add "Review PR #work #urgent" --tags "deploy"

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
rem add "Title" [--list LIST] [--due DATE] [--priority high|medium|low] [--notes TEXT] [--url URL] [-F/--flagged] [-t/--tags TAGS] [-r/--remind-me DURATION] [--repeat PATTERN]
rem add -i                          # Interactive creation

# List
rem list [--list LIST] [--incomplete] [--completed] [--flagged] [--due-before DATE] [--due-after DATE] [-o json|table|plain]
rem ls                              # Alias

# Show
rem show <id>                       # Full or partial ID
rem get <id> -o json

# Update
rem update <id> [-t/--title TEXT] [--due DATE] [--priority LEVEL] [--notes TEXT] [--url URL] [--add-tags TAGS] [--remove-tags TAGS] [-r/--remind-me DURATION] [--repeat PATTERN] [--list LIST]

# Complete / Uncomplete (support multiple IDs)
rem complete <id> [id2 id3...]
rem done <id>                       # Alias
rem uncomplete <id> [id2 id3...]

# Flag / Unflag (support multiple IDs)
rem flag <id> [id2 id3...]
rem unflag <id> [id2 id3...]

# Delete (supports multiple IDs)
rem delete <id> [id2 id3...]        # Asks for confirmation
rem rm <id> --force                 # Skip confirmation (-f / --yes / -y also work)

# Today — due and overdue reminders
rem today
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

### AI Agent Skills

```bash
rem skills install                          # Interactive picker (shows confirmation prompt)
rem skills install --agent claude           # Claude Code only
rem skills install --agent all             # All supported agents
rem skills install --dry-run              # Preview files without writing
rem skills status                          # Check installation status
rem skills uninstall                       # Remove the skill
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

Date parsing is powered by [`go-eventkit/dateparser`](https://github.com/BRO3886/go-eventkit):

| Input | Meaning |
|-------|---------|
| `now` | Current date and time |
| `today` | Today at 9:00 AM |
| `tomorrow` | Tomorrow at 9:00 AM |
| `next monday` | Next Monday at 9:00 AM |
| `monday 2pm` | Next Monday at 2:00 PM |
| `next friday at 2pm` | Next Friday at 2:00 PM |
| `in 2 days` | 2 days from now |
| `in 3 hours` | 3 hours from now |
| `5 days ago` | 5 days before now |
| `eod` / `end of day` | Today at 5:00 PM |
| `this week` | End of current week |
| `next week` | Next Monday at 9:00 AM |
| `next month` | 1st of next month at 9:00 AM |
| `mar 15` | March 15 at 9:00 AM |
| `5pm` | Today (or tomorrow) at 5:00 PM |
| `today 5pm` | Today at 5:00 PM |
| `2026-02-15` | February 15, 2026 |
| `2026-02-15 14:30` | February 15, 2026 at 2:30 PM |

## Go API

rem is powered by [**go-eventkit**](https://github.com/BRO3886/go-eventkit) — use it directly for programmatic access to macOS Reminders in your own Go programs:

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
│   ├── service/          # Service layer wrapping go-eventkit (AppleScript only for default list name)
│   ├── reminder/         # Domain models (Reminder, List, Priority)
│   ├── export/           # JSON & CSV import/export
│   ├── skills/           # Agent skill install/uninstall/status
│   ├── update/           # Background update check (GitHub releases)
│   └── ui/               # Table formatting, colored output
├── skills/rem-cli/       # Embedded agent skill files
├── website/              # Hugo documentation site
├── Makefile
├── LICENSE
└── README.md
```

**All reads and writes** — including reminder CRUD and list CRUD — go through `go-eventkit` (`github.com/BRO3886/go-eventkit`) — an Objective-C EventKit bridge compiled into the binary via cgo. Direct in-process access to the Reminders store, no IPC. All operations complete in under 200ms.

**Flagged, tag, and list-sharing operations** use the private ReminderKit bridge in go-eventkit — EventKit doesn't expose these properties, but `REMReminder.flagged`, `REMReminder.hashtags`, and `REMList.isShared` do. Tags degrade gracefully if the private API becomes unavailable. AppleScript is only used for the default list name query.

**Shared lists** work like any other list for creates, reads, updates, flags, tags, and deletes. Moving a reminder across a shared-list boundary is the one exception: macOS refuses a true move there (even Apple's own apps copy and delete behind the scenes), so rem does the same — the reminder is copied to the target with all fields intact, the original is deleted, and a warning with the new ID is printed to stderr.

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
- **No subtasks** — not exposed via EventKit
- **Tags, flagged, and list-sharing state use private API** — reads/writes go through Apple's private ReminderKit framework since EventKit doesn't expose these properties. If Apple changes the private API in a future macOS version, tags and flagged degrade gracefully (the reminder is still created/updated, a warning is printed — see [#44](https://github.com/BRO3886/rem/issues/44)) and lists simply report as unshared
- **No true move across shared-list boundaries** — macOS refuses it at the account level, so rem moves to/from shared lists by copy + delete; the reminder gets a new ID (warned on stderr). See [#50](https://github.com/BRO3886/rem/issues/50)
- **Immutable lists** cannot be renamed or deleted (system lists like Siri suggestions)

## License

MIT
