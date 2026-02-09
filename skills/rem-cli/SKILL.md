---
name: rem-cli
description: Manages macOS Reminders from the terminal using the rem CLI. Creates, lists, updates, completes, deletes, searches, and exports reminders and lists. Supports natural language due dates, filtering, import/export, and multiple output formats. Use when the user wants to interact with Apple Reminders via command line, automate reminder workflows, or build scripts around macOS Reminders.
metadata:
  author: sidv
  version: "1.0"
compatibility: Requires macOS with Reminders.app. Requires Xcode Command Line Tools for building from source.
---

# rem — CLI for macOS Reminders

A Go CLI that wraps macOS Reminders. Sub-200ms reads via cgo + EventKit. Single binary, no dependencies at runtime.

## Installation

```bash
go install github.com/sidv/rem/cmd/rem@latest
```

Or build from source:

```bash
git clone <repo-url> && cd rem
make build    # produces bin/rem
```

## Quick Start

```bash
# See all lists with reminder counts
rem lists --count

# Add a reminder with natural language date
rem add "Buy groceries" --list Personal --due tomorrow --priority high

# List incomplete reminders in a list
rem list --list Work --incomplete

# Search across all reminders
rem search "meeting"

# Complete a reminder by short ID
rem complete abc12345

# View stats
rem stats
```

## Command Reference

### Reminder CRUD

| Command | Aliases | Description |
|---------|---------|-------------|
| `rem add` | `create`, `new` | Create a reminder |
| `rem list` | `ls` | List reminders with filters |
| `rem show` | `get` | Show full details of one reminder |
| `rem update` | `edit` | Update reminder properties |
| `rem delete` | `rm`, `remove` | Delete a reminder |
| `rem complete` | `done` | Mark reminder complete |
| `rem uncomplete` | — | Mark reminder incomplete |
| `rem flag` | — | Flag a reminder |
| `rem unflag` | — | Remove flag |

### List Management

| Command | Aliases | Description |
|---------|---------|-------------|
| `rem lists` | — | Show all lists |
| `rem list-mgmt create` | `lm new` | Create a list |
| `rem list-mgmt rename` | — | Rename a list |
| `rem list-mgmt delete` | `lm rm` | Delete a list |

### Search & Analytics

| Command | Description |
|---------|-------------|
| `rem search <query>` | Search title and notes |
| `rem stats` | Show statistics and per-list breakdown |
| `rem overdue` | Show overdue reminders |
| `rem upcoming` | Show reminders due in next N days (default: 7) |

### Import/Export

| Command | Description |
|---------|-------------|
| `rem export` | Export to JSON or CSV |
| `rem import <file>` | Import from JSON or CSV file |

### Other

| Command | Description |
|---------|-------------|
| `rem interactive` / `rem i` | Interactive menu-driven mode |
| `rem version` | Print version |
| `rem completion` | Generate shell completions (bash/zsh/fish) |

For full flag details on every command, see [references/commands.md](references/commands.md).

## Key Concepts

### Short IDs

Reminders have UUIDs like `x-apple-reminder://AB12CD34-...`. The CLI displays the first 8 characters as a short ID (`AB12CD34`). You can pass any unique prefix to commands — `rem complete AB1` works if it matches exactly one reminder.

### Natural Language Dates

The `--due` flag accepts natural language:

```bash
rem add "Call dentist" --due tomorrow
rem add "Submit report" --due "next friday at 2pm"
rem add "Quick task" --due "in 30 minutes"
rem add "Wrap up" --due eod
```

Supported patterns: `today`, `tomorrow`, `next monday`, `in 3 hours`, `eod`, `eow`, `5pm`, `2026-02-15`, and more. See [references/dates.md](references/dates.md) for the full list.

### Priority Levels

| Level | Flag value | AppleScript value |
|-------|-----------|-------------------|
| High | `--priority high` | 1 (range 1-4) |
| Medium | `--priority medium` | 5 |
| Low | `--priority low` | 9 (range 6-9) |
| None | `--priority none` | 0 |

### Output Formats

All read commands support `-o` / `--output`:

- **table** (default) — formatted table with borders
- **json** — machine-readable JSON
- **plain** — simple text, one item per line

The `NO_COLOR` environment variable is respected.

### URL Storage

macOS Reminders has no native URL field. rem stores URLs in the notes/body field with a `URL: ` prefix and extracts them for display.

## Common Workflows

### Daily review
```bash
rem overdue                          # Check what's past due
rem upcoming --days 1                # See today's reminders
rem list --list Work --incomplete    # Focus on work items
```

### Batch operations with JSON
```bash
rem export --list Work --format json > backup.json
rem import backup.json --list "Work Archive"
```

### Scripting with JSON output
```bash
# Get overdue count
rem overdue -o json | jq 'length'

# List all incomplete reminder titles
rem list --incomplete -o json | jq -r '.[].name'
```

## Public Go API

For programmatic access, import `pkg/client`:

```go
import "github.com/sidv/rem/pkg/client"

c := client.New()
id, _ := c.CreateReminder(&client.CreateReminderInput{
    Title:    "Buy milk",
    ListName: "Shopping",
    Priority: client.PriorityHigh,
})
reminders, _ := c.ListReminders(&client.ListOptions{Incomplete: true})
```

See [references/go-api.md](references/go-api.md) for the full API surface.

## Limitations

- **macOS only** — requires EventKit framework and `osascript`
- **No tags, subtasks, or recurrence** — not exposed by EventKit/AppleScript
- **`--flagged` filter is slower** (~3-4s) — falls back to JXA since EventKit doesn't expose flagged
- **List deletion** may fail on some macOS versions
