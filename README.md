# rem

A powerful CLI tool for macOS Reminders app. Manage your reminders and lists from the terminal with full CRUD operations, natural language date parsing, import/export, and a clean table UI.

## Features

- **Full CRUD** for reminders and lists via AppleScript/JXA
- **Natural language dates**: `tomorrow`, `next friday at 2pm`, `in 3 hours`, `eod`
- **Multiple output formats**: table, JSON, plain text
- **Import/Export**: JSON and CSV with full property preservation
- **Search**: full-text search across titles and notes
- **Interactive mode**: step-by-step reminder creation wizard
- **Statistics**: completion rates, overdue counts, per-list breakdown
- **Public Go API**: `pkg/client` package for programmatic access
- **Shell completions**: bash, zsh, fish

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

- macOS (uses AppleScript/JXA to interact with the Reminders app)
- Go 1.21+ (for building from source)
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

The `pkg/client` package provides a clean Go API for programmatic access:

```go
package main

import (
    "fmt"
    "time"

    "github.com/BRO3886/rem/pkg/client"
)

func main() {
    c := client.New()

    // Create a reminder
    due := time.Now().Add(24 * time.Hour)
    id, err := c.CreateReminder(&client.CreateReminderInput{
        Title:    "Buy groceries",
        ListName: "Personal",
        DueDate:  &due,
        Priority: client.PriorityHigh,
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("Created:", id)

    // List incomplete reminders
    reminders, _ := c.ListReminders(&client.ListOptions{
        ListName:   "Personal",
        Incomplete: true,
    })
    for _, r := range reminders {
        fmt.Printf("- %s (due: %v)\n", r.Title, r.DueDate)
    }

    // Complete a reminder
    _ = c.CompleteReminder(id)

    // Manage lists
    lists, _ := c.GetLists()
    for _, l := range lists {
        fmt.Printf("%s (%d reminders)\n", l.Name, l.Count)
    }
}
```

### API Methods

| Method | Description |
|--------|-------------|
| `New()` | Create a new client |
| `CreateReminder(input)` | Create a reminder, returns ID |
| `GetReminder(id)` | Get a reminder by ID |
| `ListReminders(opts)` | List reminders with filters |
| `UpdateReminder(id, input)` | Update reminder properties |
| `DeleteReminder(id)` | Delete a reminder |
| `CompleteReminder(id)` | Mark as complete |
| `UncompleteReminder(id)` | Mark as incomplete |
| `FlagReminder(id)` | Flag a reminder |
| `UnflagReminder(id)` | Remove flag |
| `GetLists()` | Get all lists |
| `CreateList(name)` | Create a list |
| `RenameList(old, new)` | Rename a list |
| `DeleteList(name)` | Delete a list |
| `DefaultListName()` | Get default list name |

## Architecture

```
rem/
├── cmd/rem/              # CLI entry point
│   ├── main.go
│   └── commands/         # Cobra command definitions
├── internal/
│   ├── applescript/      # AppleScript/JXA execution & templates
│   ├── reminder/         # Domain models (Reminder, List, Priority)
│   ├── parser/           # Natural language date parsing
│   ├── export/           # JSON & CSV import/export
│   └── ui/               # Table formatting, colored output
├── pkg/client/           # Public Go API
├── Makefile
├── LICENSE
└── README.md
```

The tool uses **JXA (JavaScript for Automation)** for bulk read operations (listing reminders/lists) which is significantly faster than traditional AppleScript loops, and **AppleScript** for write operations (create, update, delete) where the syntax is simpler.

## Known Limitations

- **macOS only**: Requires the Reminders app and osascript
- **No URL property**: Reminders AppleScript API has no URL field; URLs are stored in the notes/body field
- **No tags**: Tags are not exposed via the AppleScript/JXA API
- **No subtasks**: Sub-reminders are invisible to AppleScript
- **No recurrence**: Recurring reminders cannot be set via the scripting API
- **List deletion**: `delete list` may fail on some macOS versions
- **No move operation**: Moving reminders between lists requires delete + recreate

## License

MIT
