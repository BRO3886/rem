# Complete Command Reference

## rem add

Create a new reminder.

```bash
rem add "Buy groceries" --list Personal --due tomorrow --priority high
rem add "Review PR" --due "next friday at 2pm" --url https://github.com/org/repo/pull/123
rem add "Call dentist" --notes "Ask about cleaning"
rem add -i   # Interactive mode
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--list` | `-l` | Reminder list name | System default list |
| `--due` | `-d` | Due date (natural language or ISO) | None |
| `--priority` | `-p` | Priority: high, medium, low, none | none |
| `--notes` | `-n` | Notes/body text | Empty |
| `--url` | `-u` | URL to attach (stored in body) | None |
| `--flagged` | `-f` | Flag the reminder | false |
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
rem update abc12345 -i            # Interactive mode
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--name` | — | New title | — |
| `--due` | `-d` | New due date (use `none` to clear) | — |
| `--notes` | `-n` | New notes/body | — |
| `--priority` | `-p` | New priority: high, medium, low, none | — |
| `--url` | `-u` | New URL | — |
| `--flagged` | — | Set flagged: true/false | — |
| `--interactive` | `-i` | Update interactively | false |

Aliases: `edit`

---

## rem delete

Delete a reminder. Prompts for confirmation unless `--force` is used.

```bash
rem delete abc12345
rem rm abc12345 --force
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--force` | — | Skip confirmation | false |

Aliases: `rm`, `remove`

---

## rem complete

Mark a reminder as complete.

```bash
rem complete abc12345
rem done abc12345
```

Aliases: `done`

---

## rem uncomplete

Mark a reminder as incomplete.

```bash
rem uncomplete abc12345
```

---

## rem flag

Flag a reminder.

```bash
rem flag abc12345
```

---

## rem unflag

Remove flag from a reminder.

```bash
rem unflag abc12345
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
| `--force` | — | Skip confirmation | false |

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

## rem overdue

Show overdue reminders (incomplete with due date in the past).

```bash
rem overdue
rem overdue -o json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--output` | `-o` | Output format: table, json, plain | table |

---

## rem upcoming

Show upcoming reminders (due in the next N days).

```bash
rem upcoming
rem upcoming --days 14
rem upcoming -o json
```

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--days` | — | Number of days to look ahead | 7 |
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

## rem version

Print version information.

```bash
rem version
```

---

## Global Behavior

- All read commands accept `-o` / `--output` for format selection (table, json, plain)
- `NO_COLOR=1` environment variable disables color output
- ID arguments accept prefix matches — pass any unique prefix of a short ID
