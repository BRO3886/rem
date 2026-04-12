---
name: rem
description: Manages macOS Reminders from the terminal using the rem CLI. Creates, lists, updates, completes, deletes, searches, and exports reminders and lists. Supports natural language due dates, filtering, import/export, and multiple output formats. Use when the user wants to interact with Apple Reminders via command line, automate reminder workflows, or build scripts around macOS Reminders.
metadata:
  author: BRO3886
  version: "0.7.0"
compatibility: Requires macOS with Reminders.app. Requires Xcode Command Line Tools for building from source.
---

# rem — CLI for macOS Reminders

A Go CLI that wraps macOS Reminders. Sub-200ms reads AND writes via cgo + EventKit. Single binary, no dependencies at runtime.

## Installation

```bash
# macOS (recommended)
curl -fsSL https://rem.sidv.dev/install | bash

# Or via Go
go install github.com/BRO3886/rem/cmd/rem@latest
```

Install this skill into your agent:

```bash
# Claude Code or Codex
rem skills install

# OpenClaw
rem skills install --agent openclaw
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
| `rem today` | Show today's due and overdue reminders |
| `rem overdue` | Show overdue reminders |
| `rem upcoming` | Show reminders due in next N days (default: 7) |

### Import/Export

| Command | Description |
|---------|-------------|
| `rem export` | Export to JSON or CSV |
| `rem import <file>` | Import from JSON or CSV file |

### Skills & Other

| Command | Description |
|---------|-------------|
| `rem skills install` | Install rem skill for AI agents |
| `rem skills uninstall` | Uninstall rem skill from AI agents |
| `rem skills status` | Show skill installation status |
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

### Notifications and alarms

When a reminder is created with `--due`, rem auto-attaches an alarm at the due time so the system actually fires a notification — matching Apple Reminders.app default behavior. This replaces an earlier footgun where `rem add --due X` produced a silent reminder with no alert.

- **Default**: `rem add "X" --due tomorrow` — notifies at the due time
- **Custom offset**: `rem add "X" --due tomorrow --remind-me 15m` — notifies 15 minutes before
- **Absolute time**: `rem add "X" --due tomorrow --remind-me "tomorrow at 8am"` — notifies at the absolute time
- **Silent**: `rem add "X" --due tomorrow --silent` — sets a due date but no alert (checklist-style)

Do NOT pass `--remind-me 0m` just to "enable notifications" — that's now the default when `--due` is set.

### URL field

URLs set via `--url` are stored in the real Reminders.app URL field (the one that shows in the UI with a link card), not in the notes. This requires go-eventkit v0.5.0+, which writes via the native `REMURLAttachment` path under the hood.

- `rem add "X" --url https://github.com/...` — URL shows in Reminders.app with link preview
- `rem update <id> --url https://other.com` — replaces the URL
- `rem update <id> --url ""` — clears the URL

Old reminders with URLs stored in the notes body (`URL: https://...`) still read correctly as a backward compatibility fallback.

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

For programmatic access, use [`go-eventkit`](https://github.com/BRO3886/go-eventkit) directly:

```go
import "github.com/BRO3886/go-eventkit/reminders"

client, _ := reminders.New()
r, _ := client.CreateReminder(reminders.CreateReminderInput{
    Title:    "Buy milk",
    ListName: "Shopping",
    Priority: reminders.PriorityHigh,
})
items, _ := client.Reminders(reminders.WithCompleted(false))
```

See [go-eventkit docs](https://github.com/BRO3886/go-eventkit) for the full API surface.

## Limitations

- **macOS only** — requires EventKit framework and `osascript`
- **No tags or subtasks** — not exposed by EventKit/AppleScript
- **Recurrence via `--repeat`** — supports daily, weekly (with day selection), monthly (with day-of-month selection), yearly. Use `--repeat none` to clear
- **`--flagged` filter is slower** (~3-4s) — falls back to JXA since EventKit doesn't expose flagged
- **List deletion** may fail on some macOS versions
