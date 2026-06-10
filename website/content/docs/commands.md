---
title: "Commands"
description: "Complete reference for every rem CLI command: rem add, rem list, rem complete, rem delete, rem update, rem search, rem stats, rem today, rem overdue, rem upcoming, rem export, rem import, rem interactive, rem lists, rem list-mgmt, rem flag, rem skills, and more."
keywords:
  - rem CLI commands reference
  - rem add reminder terminal
  - rem list macOS
  - rem complete reminder
  - rem delete reminder
  - rem search reminders
  - rem export JSON CSV
  - rem stats macOS
weight: 2
sitemap:
  priority: 0.9
  changefreq: weekly
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
rem add "Review PR" --due "friday at 2pm" --remind-me 30m   # notify 30 min before
rem add "Checklist item" --due tomorrow --silent            # due date, no notification
rem add "Ship demo" --url https://github.com/BRO3886/rem/pull/24
rem add "Review PR #work #urgent" --tags "deploy"          # native tags
```

**Aliases:** `create`, `new`

**Notifications.** When `--due` is set, rem auto-attaches an alarm at the due time so the system actually fires a notification — matching Apple Reminders.app default behavior. Use `--remind-me` to override the timing (e.g. `15m` for 15 minutes before), or `--silent` to suppress the auto-alarm entirely.

**URLs.** `--url` writes to the real Reminders.app URL field (not the notes body), so URLs show up with Apple's native link card rendering in the Reminders.app UI.

**Tags.** `#hashtags` in the title are automatically parsed and stored as native Reminders.app tags. You can also pass tags via the `--tags` flag. Tags use the private ReminderKit API — if Apple changes it in a future macOS version, the reminder is still created successfully (tags just won't be set).

| Flag | Description |
|------|-------------|
| `-l, --list` | Target list (default: system default list) |
| `-d, --due` | Due date (natural language or standard format) |
| `-p, --priority` | Priority: `high`, `medium`, `low`, `none` |
| `-n, --notes` | Notes/body text |
| `-u, --url` | URL (shows in Reminders.app URL field) |
| `-f, --flagged` | Flag the reminder |
| `--tags` | Comma-separated native tags (e.g. `work,urgent`) |
| `--remind-me` | Custom alarm: duration before due (`15m`, `1h`, `2d`) or absolute time |
| `--silent` | Don't auto-attach an alarm when `--due` is set |
| `--repeat` | Set recurrence: `daily`, `weekly`, `monthly`, `yearly`, or `weekly on mon,wed,fri` |
| `-i, --interactive` | Step-by-step interactive creation |

The interactive form (`rem add -i`) has no `--silent` option. Reminders created interactively with a due date always get an auto-alarm. To create a silent reminder via the interactive flow, follow up with `rem update <id> --remind-me none` to clear the alarm.

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

Show full details for a single reminder, including alarms and recurrence rules.

```bash
rem show 6ECE
```

**Aliases:** `get`

Pass the full ID, UUID, or any unique prefix. IDs are case-insensitive and prefix-matched.

### `rem update`

Update an existing reminder.

```bash
rem update 6ECE --priority medium --due "next friday"
rem update 6ECE --url https://github.com/org/repo/pull/42
rem update 6ECE --url ""   # clear URL
rem update 6ECE --add-tags "work,urgent"    # add tags
rem update 6ECE --remove-tags "urgent"      # remove tags
```

**Aliases:** `edit`

`--url` writes to the native Reminders.app URL field (not the notes body). Pass an empty string (`--url ""`) to clear.

`--add-tags` and `--remove-tags` accept comma-separated tag names. Tags in `--name` are also parsed as hashtags. Tags use the private ReminderKit API and degrade gracefully if unavailable.

`--list` moves the reminder. Plain moves are native and keep the reminder's ID. Moving to or from a **shared list** is different: macOS has no true move across that boundary (Apple's own apps copy and delete behind the scenes), so rem copies the reminder — all fields preserved, including completed state — deletes the original, and prints a warning with the **new ID** on stderr. Because the ID changes, rem **asks for confirmation** before a shared-list move; pass `--force`/`-y` to skip the prompt (required in scripts — non-interactive runs refuse without it, same as `rem delete`). Re-resolve the ID after such a move.

| Flag | Description |
|------|-------------|
| `--name` | New title |
| `-l, --list` | Move reminder to a different list |
| `-d, --due` | New due date |
| `-p, --priority` | New priority |
| `-n, --notes` | New notes |
| `-u, --url` | New URL (empty string to clear) |
| `--flagged` | Set flag: `true` or `false` |
| `--add-tags` | Add comma-separated native tags |
| `--remove-tags` | Remove comma-separated native tags |
| `--remind-me` | Set alarm: duration (`15m`, `1h`, `2d`), `none` to clear |
| `--repeat` | Set recurrence: `daily`, `weekly`, `monthly`, `yearly`, `none` to clear |
| `-y, --force`, `--yes` | Skip the shared-list move confirmation |
| `-i, --interactive` | Interactive update |

### `rem complete`

Mark one or more reminders as completed.

```bash
rem complete 6ECE
rem complete 6ECE A1B2 C3D4    # batch complete
```

**Aliases:** `done`

### `rem uncomplete`

Mark one or more completed reminders as incomplete.

```bash
rem uncomplete 6ECE
rem uncomplete 6ECE A1B2
```

### `rem flag` / `rem unflag`

Toggle the flag on one or more reminders.

```bash
rem flag 6ECE
rem flag 6ECE A1B2    # batch flag
rem unflag 6ECE
```

### `rem delete`

Delete one or more reminders. Supports multiple IDs for batch deletion.

```bash
rem delete 6ECE
rem delete 6ECE A1B2 --force    # batch delete, skip confirmation
```

**Aliases:** `rm`, `remove`

| Flag | Description |
|------|-------------|
| `--force` / `--yes` / `-y` | Skip the confirmation prompt |

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

### `rem today`

Show today's due and overdue reminders.

```bash
rem today
rem today --list Work
```

| Flag | Description |
|------|-------------|
| `-l, --list` | Filter by list name |

### `rem overdue`

Show all overdue incomplete reminders.

```bash
rem overdue
rem overdue --list Work
```

| Flag | Description |
|------|-------------|
| `-l, --list` | Filter by list name |

### `rem upcoming`

Show reminders due in the next N days.

```bash
rem upcoming           # next 7 days (default)
rem upcoming --days 14 # next 14 days
```

| Flag | Description |
|------|-------------|
| `--days` | Number of days to look ahead (default: 7) |
| `-l, --list` | Filter by list name |

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

Shared lists are marked — `(shared)` in table output, `[shared]` in plain. JSON output carries three booleans per list: `IsShared`, `SharedToMe` (someone shared it with you), and `IsOwnedByMe`.

### `rem list-mgmt`

Create, rename, or delete lists.

**Aliases:** `lm`

```bash
rem lm create "Projects"
rem lm rename "Projects" "Active Projects"
rem lm delete "Old List" --force
```

> Note: Immutable system lists (e.g. Siri suggestions) cannot be renamed or deleted.

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

## Skills

### `rem skills install`

Install the rem agent skill for AI coding agents.

```bash
rem skills install                         # Interactive picker
rem skills install --agent claude          # Claude Code only
rem skills install --agent openclaw        # OpenClaw only
rem skills install --agent all             # All agents
```

| Flag | Description |
|------|-------------|
| `--agent` | Target agent: `claude`, `codex`, `openclaw`, or `all` (default: interactive picker) |

### `rem skills uninstall`

Remove the rem agent skill.

```bash
rem skills uninstall --agent claude
```

| Flag | Description |
|------|-------------|
| `--agent` | Target agent: `claude`, `codex`, `openclaw`, or `all` (default: interactive picker) |

### `rem skills status`

Show skill installation status across all supported agents.

```bash
rem skills status
```

## Date Formats

Date parsing is powered by [`go-eventkit/dateparser`](https://github.com/BRO3886/go-eventkit). Supported patterns:

| Pattern | Example |
|---------|---------|
| Relative | `in 2 days`, `in 3 hours`, `in 30 minutes`, `5 days ago` |
| Named | `now`, `today`, `tomorrow`, `yesterday` |
| Day of week | `monday`, `next monday`, `next friday` |
| Special | `eod` (5pm today), `eow` (Friday 5pm), `this week`, `next week` (next Monday), `next month` (1st of next month) |
| Compound | `next friday at 2pm`, `tomorrow at 3:30pm`, `today 5pm`, `monday 2pm` |
| Month-day | `mar 15`, `21 march`, `december 31 11:59pm` |
| Time only | `5pm`, `17:00`, `3:30pm` |
| ISO 8601 | `2026-02-15`, `2026-02-15T14:30:00` |
| US format | `02/15/2026`, `2/15` |
| Named month | `Feb 15`, `February 15, 2026` |

When a standalone time is given (like `5pm`), it uses today if the time hasn't passed yet, otherwise tomorrow.
