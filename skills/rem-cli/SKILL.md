---
name: rem
description: Create, list, update, complete, and search macOS Reminders via the rem CLI. Use when the user wants to manage Apple Reminders from the terminal, automate reminder workflows, reference reminders in shell scripts, or schedule anything on macOS.
allowed-tools: Bash(rem *) Bash(date *)
argument-hint: "[natural language request]"
---

# rem — macOS Reminders from the terminal

rem is a single-binary Go CLI that reads and writes the Apple Reminders database in under 200ms via EventKit (cgo). Every `rem` invocation is fast and safe to run.

## Current time (do not guess, use these values)

- Today: !`date +"%A, %B %-d, %Y"`
- ISO date: !`date +"%Y-%m-%d"`
- Local time: !`date +"%H:%M %Z"`
- Day of week: !`date +"%A"`
- Tomorrow (ISO): !`date -v+1d +"%Y-%m-%d"`
- End of week (Friday ISO): !`date -v+fri +"%Y-%m-%d"`

**When the user says relative dates like "tomorrow", "next friday", "end of week", or "last tuesday", resolve them against the values above before calling rem.** Don't ask the user "what is today" and don't guess from training data.

rem's `--due` flag accepts natural language directly (see [references/dates.md](references/dates.md)), so in many cases you can pass the user's phrase through verbatim. But if the user asks something ambiguous ("what was I supposed to do yesterday?"), use the ISO date above and query rem with `--due-before` / `--due-after` against an exact date.

## When to use this skill

Use rem any time the user wants to:

- Add, list, search, complete, or delete macOS Reminders
- Summarize "what's due today" / "what's overdue" / "what's coming up"
- Capture quick tasks from natural language ("remind me to call dentist tomorrow")
- Move reminders between lists, reprioritize, or add URLs and notes
- Export reminders to JSON / CSV, or import them back
- Manage reminder lists (create, rename, delete)

Do NOT use rem for:

- Calendar events (use `ical` if available, or AppleScript with Calendar.app)
- Other Apple apps (Notes, Messages, Mail)
- Non-macOS platforms (rem is macOS-only; fail gracefully if on Linux)

## Quick decision tree

| User intent | Command to reach for |
|---|---|
| "what's on my plate" / "today" / "what's overdue" | `rem today`, `rem overdue`, `rem upcoming --days N` |
| "remind me to X (tomorrow / friday / etc)" | `rem add "X" --due <date>` |
| "remind me about X at Y time" | `rem add "X" --due "<date> <time>"` — notification is automatic |
| "make that one silent / no notification" | add `--silent` to `rem add` |
| "mark X as done" | `rem complete <short-id>` |
| "undo that / mark X as not done" | `rem uncomplete <short-id>` |
| "delete X" | `rem delete <short-id>` (supports multiple IDs) |
| "find reminders about X" | `rem search "X"` |
| "flag / unflag X" | `rem flag <short-id>` or `rem unflag <short-id>` |
| "show me flagged stuff" | `rem list --flagged` |
| "move X to list Y" | `rem update <short-id> --list "Y"` |
| "what lists do I have" | `rem lists` (add `--count` for per-list totals) |
| "how many reminders total / stats" | `rem stats` |
| "export my Work list" | `rem export --list Work --format json --output-file work.json` |
| "import this CSV" | `rem import file.csv --dry-run` first, then without |

For full flag details on any command, load [references/commands.md](references/commands.md).

## Output formats (always set one when scripting)

Every read command supports `-o table|json|plain`. Default is `table` (colored ASCII). When piping to another tool, into a script, or when you need to parse the result, use `-o json`:

```bash
rem today -o json | jq '.[] | .name'
rem overdue -o json | jq 'length'
rem list --incomplete --list Work -o json
```

`NO_COLOR=1` disables colors. `REM_NO_UPDATE_CHECK=1` disables the background update check (set it when scripting to avoid stray output).

## Short IDs

rem displays the first 8 characters of each reminder's UUID as its "short ID" (e.g. `AB12CD34`). You can pass any unique prefix to commands — `rem complete AB1` works as long as it matches exactly one reminder. Prefer short IDs when showing reminder IDs back to the user.

## Notifications default to ON (important behavior change in v0.10.0)

When `--due` is set on `rem add`, rem auto-attaches an alarm at the due time by default. This matches Apple Reminders.app behavior. **Do NOT pass `--remind-me 0m` just to enable notifications — that's now the default.**

- `rem add "Review PR" --due tomorrow` → notifies at tomorrow 9 AM (default due time)
- `rem add "Review PR" --due "tomorrow 2pm" --remind-me 15m` → notifies 15 minutes before
- `rem add "Groceries" --due tomorrow --silent` → due date, no notification (checklist-style)

Full detail: [references/commands.md](references/commands.md) `rem add` section.

## URLs go in the native Reminders.app URL field

When the user wants a link attached to a reminder, pass it via `--url`. rem writes to the real Reminders.app URL field (not the notes body), so the link shows in the Reminders.app UI with Apple's native link card rendering. Clear a URL with `--url ""`.

```bash
rem add "Review spec" --due friday --url https://example.com/spec.pdf
rem update AB12 --url https://github.com/org/repo/pull/42
rem update AB12 --url ""    # clear
```

## Critical gotchas

1. **macOS only.** rem uses EventKit via cgo. Will fail on Linux — detect OS first if you're unsure.
2. **Priority mapping is non-linear.** `--priority high` = 1, `medium` = 5, `low` = 9. Don't pass raw integers from user input; use the words.
3. **`--due none` clears** the due date in `rem update`. Same for `--remind-me none` and `--repeat none`.
4. **The `--flagged` filter is slower (~3–4s)** because EventKit doesn't expose the flagged property, so rem falls back to JXA for that filter only. All other operations are sub-200ms.
5. **`rem delete` prompts by default.** Pass `--force` / `--yes` / `-y` when scripting to skip the confirmation.
6. **`rem add -i` (interactive form) has no `--silent` equivalent.** If a user wants a silent reminder via the interactive flow, create it normally and then run `rem update <id> --remind-me none` to clear the alarm.
7. **Old reminders with URLs in the notes body** (`URL: https://...`) still read correctly as a backward-compat fallback. New reminders always use the native URL field.

## Reference files (load when needed)

- **[references/commands.md](references/commands.md)** — Complete flag reference for every rem command. Load when you need specific flag details, default values, or full usage examples.
- **[references/dates.md](references/dates.md)** — Natural language date grammar accepted by `--due`, `--due-before`, `--due-after`, and `--remind-me`. Load when constructing a date string from a user's phrasing.

## Common patterns

### Daily briefing
```bash
rem overdue -o plain
rem today -o plain
rem upcoming --days 3 -o plain
```

### Scripted cleanup
```bash
# Archive completed Work items to a JSON backup
rem export --list Work --format json --output-file work-$(date +%Y%m%d).json

# Count overdue items
rem overdue -o json | jq 'length'
```

### Quick capture from a user message
When the user says something like "remind me to call mom tomorrow at 5pm", parse the title, due date, and list (if mentioned) and run a single `rem add`:

```bash
rem add "Call mom" --due "tomorrow at 5pm" --list Personal
```

Always confirm the short ID of the created reminder back to the user so they can reference it later.
