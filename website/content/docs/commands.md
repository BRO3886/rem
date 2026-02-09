---
title: "Commands"
description: "Complete reference for all rem CLI commands — create, list, update, delete, search, and more."
weight: 2
---

## Global flags

These flags work with every command:

| Flag | Description |
|------|-------------|
| `-o, --output` | Output format: `table`, `json`, or `plain` |
| `--no-color` | Disable colored output |
| `-h, --help` | Show help for any command |

## Reminders

### `rem add`

Create a new reminder.

```bash
rem add "Buy groceries" --due tomorrow --priority high --list Personal
```

**Aliases:** `create`, `new`

| Flag | Description |
|------|-------------|
| `-l, --list` | Target list (default: system default list) |
| `-d, --due` | Due date (natural language or standard format) |
| `-p, --priority` | Priority: `high`, `medium`, `low`, `none` |
| `-n, --notes` | Notes/body text |
| `-u, --url` | URL to attach (stored in body) |
| `-f, --flagged` | Flag the reminder |
| `-i, --interactive` | Step-by-step interactive creation |

### `rem list`

List reminders with optional filters.

```bash
rem list --list Work --incomplete --due-before 2026-03-01
```

**Aliases:** `ls`

| Flag | Description |
|------|-------------|
| `-l, --list` | Filter by list name |
| `--incomplete` | Show only incomplete reminders |
| `--completed` | Show only completed reminders |
| `--flagged` | Show only flagged reminders |
| `--due-before` | Reminders due before this date |
| `--due-after` | Reminders due after this date |
| `-s, --search` | Full-text search in title and notes |

### `rem show`

Show full details for a single reminder.

```bash
rem show 6ECE
```

**Aliases:** `get`

Pass the full ID, UUID, or any unique prefix. IDs are case-insensitive and prefix-matched.

### `rem update`

Update an existing reminder.

```bash
rem update 6ECE --priority medium --due "next friday"
```

**Aliases:** `edit`

| Flag | Description |
|------|-------------|
| `--name` | New title |
| `-d, --due` | New due date |
| `-p, --priority` | New priority |
| `-n, --notes` | New notes |
| `-u, --url` | New URL |
| `--flagged` | Set flag: `true` or `false` |
| `-i, --interactive` | Interactive update |

### `rem complete`

Mark a reminder as completed.

```bash
rem complete 6ECE
```

**Aliases:** `done`

### `rem uncomplete`

Mark a completed reminder as incomplete.

```bash
rem uncomplete 6ECE
```

### `rem flag` / `rem unflag`

Toggle the flag on a reminder.

```bash
rem flag 6ECE
rem unflag 6ECE
```

### `rem delete`

Delete a reminder.

```bash
rem delete 6ECE
rem delete 6ECE --force    # skip confirmation
```

**Aliases:** `rm`, `remove`

| Flag | Description |
|------|-------------|
| `--force` | Skip the confirmation prompt |

## Search & Analytics

### `rem search`

Full-text search across title and notes.

```bash
rem search "quarterly review" --list Work --incomplete
```

| Flag | Description |
|------|-------------|
| `-l, --list` | Search within a specific list |
| `--incomplete` | Search only incomplete reminders |

### `rem stats`

Show overall statistics.

```bash
rem stats
```

Displays: total count, completed, incomplete, flagged, overdue, completion rate, and per-list breakdown.

### `rem overdue`

Show all overdue incomplete reminders.

```bash
rem overdue
```

### `rem upcoming`

Show reminders due in the next N days.

```bash
rem upcoming           # next 7 days (default)
rem upcoming --days 14 # next 14 days
```

| Flag | Description |
|------|-------------|
| `--days` | Number of days to look ahead (default: 7) |

## Lists

### `rem lists`

View all reminder lists.

```bash
rem lists
rem lists --count    # include reminder count per list
```

| Flag | Description |
|------|-------------|
| `-c, --count` | Show reminder count per list |

### `rem list-mgmt`

Create, rename, or delete lists.

**Aliases:** `lm`

```bash
rem lm create "Projects"
rem lm rename "Projects" "Active Projects"
rem lm delete "Old List" --force
```

> Note: List deletion may fail on some macOS versions due to AppleScript limitations.

## Import & Export

### `rem export`

Export reminders to JSON or CSV.

```bash
rem export --format json --output-file backup.json
rem export --list Work --format csv
rem export --incomplete --format json
```

| Flag | Description |
|------|-------------|
| `-l, --list` | Export from a specific list |
| `--format` | `json` or `csv` (default: json) |
| `--output-file` | File path (default: stdout) |
| `--incomplete` | Export only incomplete reminders |

### `rem import`

Import reminders from JSON or CSV.

```bash
rem import backup.json
rem import data.csv --list "Imported" --dry-run
```

| Flag | Description |
|------|-------------|
| `-l, --list` | Import into this list (overrides source list) |
| `--dry-run` | Preview what would be created without creating |

## Interactive Mode

### `rem interactive`

Launch a full interactive menu.

```bash
rem i
```

**Aliases:** `i`

The menu includes:
1. Create reminder (guided prompts)
2. List reminders
3. Complete a reminder
4. Delete a reminder
5. View all lists
6. Create a new list

You can also use `-i` with `add` and `update` for interactive field entry.

## Date Formats

rem's built-in parser understands these patterns:

| Pattern | Example |
|---------|---------|
| Relative | `in 2 days`, `in 3 hours`, `in 30 minutes` |
| Named | `today`, `tomorrow`, `yesterday` |
| Day of week | `next monday`, `next friday` |
| Special | `eod` (5pm today), `eow` (Friday 5pm), `next week`, `next month` |
| Compound | `next friday at 2pm`, `tomorrow at 3:30pm` |
| Time only | `5pm`, `17:00`, `3:30pm` |
| ISO 8601 | `2026-02-15`, `2026-02-15T14:30:00` |
| US format | `02/15/2026`, `2/15` |
| Named month | `Feb 15`, `February 15, 2026` |

When a standalone time is given (like `5pm`), it uses today if the time hasn't passed yet, otherwise tomorrow.
