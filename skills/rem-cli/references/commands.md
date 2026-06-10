# Complete Command Reference

## rem add

Create a new reminder.

```bash
rem add "Buy groceries" --list Personal --due tomorrow --priority high
rem add "Review PR" --due "next friday at 2pm" --url https://github.com/org/repo/pull/123
rem add "Call dentist" --notes "Ask about cleaning"
rem add "Meeting" --due "tomorrow at 10am" --remind-me 15m
rem add "Review PR #work #urgent" --tags "deploy"         # native tags from title + flag
rem add "Silent checklist item" --due tomorrow --silent   # due date, no notification
rem add -i   # Interactive mode
```

**Notifications.** When `--due` is set, rem auto-attaches an alarm at the due time so the system actually fires a notification — matching Apple Reminders.app default behavior. Override the timing with `--remind-me` (e.g. `--remind-me 15m` for 15 minutes before), or suppress the auto-alarm entirely with `--silent` for checklist-style reminders. Do NOT pass `--remind-me 0m` just to enable notifications — that's the default when `--due` is set.

**URLs.** `--url` writes to the real Reminders.app URL field (via the native `REMURLAttachment` path, requires go-eventkit v0.5.0+). URLs set this way show up with Apple's native link card rendering in the Reminders.app UI.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--list` | `-l` | Reminder list name | System default list |
| `--due` | `-d` | Due date (natural language or ISO) | None |
| `--priority` | `-p` | Priority: high, medium, low, none | none |
| `--notes` | `-n` | Notes/body text | Empty |
| `--url` | `-u` | URL to attach (shows in Reminders.app URL field) | None |
| `--flagged` | `-f` | Flag the reminder | false |
| `--tags` | — | Comma-separated native tags (e.g. `work,urgent`) | None |
| `--remind-me` | — | Custom alarm: duration before due (15m, 1h, 2d) or absolute time | Auto: at due time when `--due` is set |
| `--silent` | — | Don't auto-attach an alarm when `--due` is set | false |
| `--repeat` | — | Set recurrence: daily, weekly, monthly, yearly, or 'weekly on mon,wed,fri' | None |
| `--interactive` | `-i` | Create interactively | false |

Aliases: `create`, `new`

---

## rem list

List reminders with optional filters.

```bash
rem list
rem list --list Work --incomplete
rem list --due-before "2026-02-15" --output json
rem list --flagged
rem list --completed --list Personal
rem list --due-after today --due-before "next week"
rem list --search "groceries"
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--list` | `-l` | Filter by list name | All lists |
| `--incomplete` | — | Show only incomplete | false |
| `--completed` | — | Show only completed | false |
| `--flagged` | — | Show only flagged | false |
| `--due-before` | — | Due before this date | None |
| `--due-after` | — | Due after this date | None |
| `--search` | `-s` | Search title and notes | None |
| `--output` | `-o` | Output format: table, json, plain | table |

Aliases: `ls`

---

## rem show

Display full details of a specific reminder.

```bash
rem show abc12345
rem show abc12345 --output json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output format: table, json, plain | table |

Aliases: `get`

---

## rem update

Update properties of an existing reminder.

```bash
rem update abc12345 --due "next monday"
rem update abc12345 --notes "Updated notes" --priority medium
rem update abc12345 --name "New title"
rem update abc12345 --due none    # Clear due date
rem update abc12345 --flagged true
rem update abc12345 --list "Work"  # Move to a different list
rem update abc12345 --remind-me 15m
rem update abc12345 --url https://github.com/org/repo/pull/42
rem update abc12345 --url ""      # Clear URL
rem update abc12345 --add-tags "work,urgent"    # Add tags
rem update abc12345 --remove-tags "urgent"      # Remove tags
rem update abc12345 -i            # Interactive mode
```

**Shared lists.** Moving a reminder to or from a shared list has no true move on macOS: rem copies the reminder to the target and deletes the original, so it gets a **new ID** (rem warns on stderr and prints the new ID). rem **prompts for confirmation** before such a move — pass `--force`/`-y`/`--yes` to skip it (required when scripting; non-interactive runs error out without it). Re-resolve the ID after such a move. Plain moves keep the ID and never prompt.

**URLs.** `--url` writes to the native Reminders.app URL field (not notes/body). Pass `--url ""` to clear the URL.

**Tags.** `--add-tags` and `--remove-tags` accept comma-separated tag names. Tags in `--name` are also parsed (e.g. `--name "Task #work"` adds the `work` tag). Tags use the private ReminderKit API and degrade gracefully if unavailable.

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--name` | — | New title | — |
| `--list` | `-l` | Move reminder to a different list | — |
| `--force` / `--yes` | `-y` | Skip the shared-list move confirmation | false |
| `--due` | `-d` | New due date (use `none` to clear) | — |
| `--notes` | `-n` | New notes/body | — |
| `--priority` | `-p` | New priority: high, medium, low, none | — |
| `--url` | `-u` | New URL (empty string to clear) | — |
| `--flagged` | — | Set flagged: true/false | — |
| `--add-tags` | — | Add comma-separated native tags | — |
| `--remove-tags` | — | Remove comma-separated native tags | — |
| `--remind-me` | — | Set alarm: duration (15m, 1h, 2d), 'none' to clear | — |
| `--repeat` | — | Set recurrence: daily, weekly, monthly, yearly, 'none' to clear | — |
| `--interactive` | `-i` | Update interactively | false |

Aliases: `edit`

---

## rem delete

Delete one or more reminders. Supports multiple IDs for batch deletion. Prompts for confirmation unless `--force` is used.

```bash
rem delete abc12345
rem delete abc12345 def67890 --force
rem rm abc12345 --force
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--force` / `--yes` | `-y` | Skip confirmation | false |

Aliases: `rm`, `remove`

---

## rem complete

Mark one or more reminders as complete.

```bash
rem complete abc12345
rem complete abc12345 def67890
rem done abc12345
```

Aliases: `done`

---

## rem uncomplete

Mark one or more reminders as incomplete.

```bash
rem uncomplete abc12345
rem uncomplete abc12345 def67890
```

---

## rem flag

Flag one or more reminders.

```bash
rem flag abc12345
rem flag abc12345 def67890
```

---

## rem unflag

Remove flag from one or more reminders.

```bash
rem unflag abc12345
rem unflag abc12345 def67890
```

---

## rem lists

Show all reminder lists.

```bash
rem lists
rem lists --count
rem lists --output json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--count` | `-c` | Show reminder count per list | false |
| `--output` | `-o` | Output format: table, json, plain | table |

---

## rem list-mgmt create

Create a new reminder list.

```bash
rem list-mgmt create "My List"
rem lm new "Shopping"
```

Aliases: `lm new`

---

## rem list-mgmt rename

Rename a reminder list.

```bash
rem list-mgmt rename "Old Name" "New Name"
```

---

## rem list-mgmt delete

Delete a reminder list and all its reminders. Prompts for confirmation.

```bash
rem list-mgmt delete "My List"
rem lm rm "My List" --force
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--force` / `--yes` | `-y` | Skip confirmation | false |

Aliases: `lm rm`

---

## rem search

Search reminders by title and notes.

```bash
rem search "groceries"
rem search "meeting" --list Work --incomplete
rem search "report" -o json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--list` | `-l` | Search within a specific list | All lists |
| `--incomplete` | — | Search only incomplete | false |
| `--output` | `-o` | Output format: table, json, plain | table |

---

## rem stats

Show reminder statistics.

```bash
rem stats
rem stats -o json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output format: plain, json | plain |

Output includes: total, completed, incomplete, flagged, overdue counts, completion rate, list count, and per-list breakdown.

---

## rem today

Show today's due and overdue reminders (incomplete with due date up to end of today).

```bash
rem today
rem today --list Work
rem today -o json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--list` | `-l` | Filter by list name | All lists |
| `--output` | `-o` | Output format: table, json, plain | table |

---

## rem overdue

Show overdue reminders (incomplete with due date in the past).

```bash
rem overdue
rem overdue --list Work
rem overdue -o json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--list` | `-l` | Filter by list name | All lists |
| `--output` | `-o` | Output format: table, json, plain | table |

---

## rem upcoming

Show upcoming reminders (due in the next N days).

```bash
rem upcoming
rem upcoming --days 14
rem upcoming --list Work
rem upcoming -o json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--days` | — | Number of days to look ahead | 7 |
| `--list` | `-l` | Filter by list name | All lists |
| `--output` | `-o` | Output format: table, json, plain | table |

---

## rem export

Export reminders to JSON or CSV.

```bash
rem export > all.json
rem export --list Work --format json > work.json
rem export --format csv --output-file reminders.csv
rem export --incomplete --format json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--list` | `-l` | Export from a specific list | All lists |
| `--format` | — | Export format: json, csv | json |
| `--output-file` | — | Output file path | stdout |
| `--incomplete` | — | Export only incomplete | false |

---

## rem import

Import reminders from a JSON or CSV file.

```bash
rem import work.json
rem import reminders.csv --list "Imported"
rem import --dry-run data.json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--list` | `-l` | Import all into this list | Original list names |
| `--dry-run` | — | Preview without creating | false |

---

## rem interactive

Launch interactive menu-driven session.

```bash
rem interactive
rem i
```

Menu options: create reminder, list reminders, complete reminder, delete reminder, list all lists, create list, quit.

---

## rem completion

Generate shell completion scripts.

```bash
rem completion bash > /usr/local/etc/bash_completion.d/rem
rem completion zsh > "${fpath[1]}/_rem"
rem completion fish > ~/.config/fish/completions/rem.fish
```

---

## rem skills install

Install the rem agent skill for AI coding agents.

```bash
rem skills install                          # Interactive picker
rem skills install --agent claude           # Install for Claude Code only
rem skills install --agent codex            # Install for Codex CLI only
rem skills install --agent openclaw         # Install for OpenClaw only
rem skills install --agent all              # Install for all agents
```

| Flag | Description | Default |
|------|-------------|---------|
| `--agent` | Agent target: claude, codex, openclaw, or all | Interactive picker |

Supported targets:
- `claude`   → `~/.claude/skills/rem-cli/`    (Claude Code, Copilot, Cursor, OpenCode, Augment)
- `codex`    → `~/.agents/skills/rem-cli/`    (Codex CLI, Copilot, Windsurf, OpenCode, Augment)
- `openclaw` → `~/.openclaw/skills/rem-cli/`  (OpenClaw)

---

## rem skills uninstall

Remove the rem agent skill from AI coding agents.

```bash
rem skills uninstall                          # Interactive picker
rem skills uninstall --agent claude           # Uninstall from Claude Code only
rem skills uninstall --agent openclaw         # Uninstall from OpenClaw only
rem skills uninstall --agent all              # Uninstall from all agents
```

| Flag | Description | Default |
|------|-------------|---------|
| `--agent` | Agent target: claude, codex, openclaw, or all | Interactive picker |

---

## rem skills status

Show the installation status of the rem skill across all supported agents.

```bash
rem skills status
```

Displays installed version, location, and whether the skill is outdated compared to the binary.

---

## rem version

Print version information.

```bash
rem version
```

---

## Global Behavior

- All read commands accept `-o` / `--output` for format selection (table, json, plain)
- `NO_COLOR=1` environment variable disables color output
- `REM_NO_UPDATE_CHECK=1` environment variable disables the background update check
- ID arguments accept prefix matches — pass any unique prefix of a short ID
