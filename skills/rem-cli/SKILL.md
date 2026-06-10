---
name: rem-cli
description: Create, list, update, complete, tag, and search macOS Reminders via the rem CLI. Use when the user wants to manage Apple Reminders from the terminal, automate reminder workflows, reference reminders in shell scripts, or schedule anything on macOS.
license: MIT
compatibility: Requires macOS with the rem CLI installed (https://rem.sidv.dev)
allowed-tools: Bash(rem *) Bash(echo *)
argument-hint: "[natural language request]"
metadata:
  author: BRO3886
  homepage: https://github.com/BRO3886/rem
---

# rem — macOS Reminders from the terminal

rem is a single-binary Go CLI that reads and writes the Apple Reminders database in under 200ms via EventKit (cgo). Every `rem` invocation is fast and safe to run.

## Current time (do not guess, use these values)

- Today: !`date +"%A, %B %-d, %Y"`
- ISO date: !`date +"%Y-%m-%d"`
- Local time: !`date +"%H:%M %Z"`
- Day of week: !`date +"%A"`
- Tomorrow (ISO): !`date -v+1d +"%Y-%m-%d"`
- Next Friday (ISO, always forward): !`date -v+1d -v+fri +"%Y-%m-%d"`

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

## Quick decision tree

| User intent | Command to reach for |
|---|---|
| "what's on my plate" / "today" / "what's overdue" | `rem today`, `rem overdue`, `rem upcoming --days N` |
| "remind me to X (tomorrow / friday / etc)" | `rem add "X" --due <date>` |
| "remind me about X at Y time" | `rem add "X" --due "<date> <time>"` — notification is automatic |
| "remind me 15 minutes before" | add `--remind-me 15m` to `rem add` |
| "make that one silent / no notification" | add `--silent` to `rem add` (not available in `-i` mode, see gotchas) |
| "tell me more about X" / "show that reminder" | `rem show <short-id>` |
| "mark X as done" / "mark these done" | `rem complete <short-id>...` (supports multiple IDs) |
| "undo that / mark X as not done" | `rem uncomplete <short-id>...` (supports multiple IDs) |
| "delete X" | `rem delete <short-id>...` (supports multiple IDs) |
| "find reminders about X" | `rem search "X"` |
| "flag / unflag X" | `rem flag <short-id>...` or `rem unflag <short-id>...` (supports multiple IDs) |
| "show me flagged stuff" | `rem list --flagged` |
| "move X to list Y" | `rem update <short-id> --list "Y"` (if Y or source is shared: confirm with user first, then `-f` — see gotchas) |
| "change priority to high" | `rem update <short-id> --priority high` |
| "add notes to X" | `rem update <short-id> --notes "..."` |
| "tag this as work" / "add tags" | `rem add "Task #work"` or `rem update <short-id> --add-tags "work,urgent"` |
| "remove the urgent tag" | `rem update <short-id> --remove-tags "urgent"` |
| "make it repeat weekly / monthly" | `rem add ... --repeat weekly` or `--repeat "weekly on mon,wed,fri"` |
| "remind me when I get to / leave X" | `rem add "..." --location "lat,lng"` (+ `--on-leave` for departure) — see location section |
| "clear the due date on X" | `rem update <short-id> --due none` |
| "what lists do I have" | `rem lists` (add `--count` for per-list totals) |
| "how many reminders total / stats" | `rem stats` |
| "export my Work list" | `rem export --list Work --format json --output-file work.json` |
| "import this CSV" | `rem import file.csv --dry-run` first, then without |

For full flag details on any command, load [references/commands.md](references/commands.md).

## Batch everything into one tool call

Every `rem` invocation is <200ms, so the expensive part is YOUR round trip, not the command. Two rules:

1. **Multiple IDs, one command.** `complete`, `uncomplete`, `flag`, `unflag`, and `delete` all accept multiple IDs: `rem complete AB12 CD34 EF56`. Never loop one ID per call.
2. **Multiple commands, one Bash call.** When answering one question needs several rem reads (or independent writes), chain them with `;` and header markers in a single Bash invocation — never run them as separate tool calls. See the daily briefing pattern below.

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

## Notifications default to ON

When `--due` is set on `rem add`, rem auto-attaches an alarm at the due time. This matches Apple Reminders.app behavior. **Do NOT pass `--remind-me 0m` to enable notifications — that's already the default when `--due` is set.**

- `rem add "Review PR" --due tomorrow` → notifies at tomorrow 9 AM (default due time)
- `rem add "Review PR" --due "tomorrow 2pm" --remind-me 15m` → notifies 15 minutes before
- `rem add "Groceries" --due tomorrow --silent` → due date, no notification (checklist-style)

Full detail: [references/commands.md](references/commands.md) `rem add` section.

## Location reminders take coordinates, not addresses

`--location "lat,lng"` attaches a geofence trigger that fires on **arrival by default**; add `--on-leave` for departure, `--radius <meters>` to widen the fence (0 = system minimum). rem does NO geocoding — you must supply decimal coordinates:

- Well-known places (landmarks, chains' flagship stores, cities): use coordinates you know.
- Personal places ("the office", "mom's house", "my gym"): **ask the user** for the address or coordinates — never guess.

```bash
rem add "Buy milk" --location "37.3318,-122.0312" --radius 200          # on arrival
rem update AB12 --location "37.7749,-122.4194" --on-leave               # on departure
rem update AB12 --location none                                         # remove geofence only
```

A reminder can have both a due date and a location. `--remind-me` and `--location` manage separate alarm buckets — clearing one never touches the other.

## URLs go in the native Reminders.app URL field

When the user wants a link attached to a reminder, pass it via `--url`. rem writes to the real Reminders.app URL field (not the notes body), so the link shows in the Reminders.app UI with Apple's native link card rendering. Clear a URL with `--url ""`.

```bash
rem add "Review spec" --due friday --url https://example.com/spec.pdf
rem update AB12 --url https://github.com/org/repo/pull/42
rem update AB12 --url ""    # clear
```

## Critical gotchas

1. **macOS only.** rem uses EventKit via cgo. Will fail on Linux — detect OS first if you're unsure.
2. **Priority uses word forms, not raw numbers.** Pass `--priority high|medium|low|none`. rem does accept raw ints, but Apple's 1–9 scale is inverted (1 = highest, 9 = lowest), so the words are safer when relaying from user input.
3. **`--due none` clears** the due date in `rem update`. Same for `--remind-me none` and `--repeat none`.
4. **Flagged and tags use private API.** Both go through Apple's private ReminderKit framework since EventKit doesn't expose these properties. Sub-200ms like everything else, but may break on future macOS versions. Both degrade gracefully — if the private API is unavailable, the reminder is still created/updated (just without the flag/tags) and a warning is printed on stderr. This applies to `flag`/`unflag` too, which behave exactly like `update --flagged`.
5. **Tags from title are additive.** `#hashtags` in the title are parsed and stored as native Reminders.app tags. They stay in the title text AND become tag objects. Pure numbers like `#42` are ignored (treated as issue references, not tags).
6. **`rem delete` prompts by default.** Pass `--force` / `--yes` / `-f` / `-y` when scripting to skip the confirmation.
7. **`rem add -i` (interactive form) has no `--silent` equivalent.** If a user wants a silent reminder via the interactive flow, create it then clear the alarm in the same Bash call: `id=$(rem add "Task" --due tomorrow -o json | jq -r '.id'); rem update "$id" --remind-me none`.
8. **Location alarms save even without Location Services, but never fire.** rem writes the geofence via public EventKit regardless; the notification only fires if Location Services is enabled for Reminders on the device watching the fence (usually the user's iPhone). If a user reports a location reminder "not working", that's the first thing to check — not rem.
9. **Old reminders with URLs in the notes body** (`URL: https://...`) still read correctly as a backward-compat fallback. New reminders always use the native URL field.
10. **Moving to/from a shared list changes the reminder's ID — and rem prompts before doing it.** macOS has no true move across a shared-list boundary, so rem copies the reminder (all fields preserved, including completed state) and deletes the original. Without `-y`, `rem update --list` blocks on a confirmation you cannot answer (TTY prompt; non-TTY it errors). Procedure for a move involving a shared list:
   1. Detect: `rem lists -o json | jq -r '.[] | select(.IsShared) | .Name'` — if neither source nor target list is in that output, move normally, no flag needed.
   2. If one is shared: tell the user the reminder will be recreated with a new ID and confirm with them — do NOT silently pass `-y`; the prompt exists to protect their data on a list other people see.
   3. Run with the flag once confirmed: `rem update <id> --list "Shared List" -f`.
   4. Re-resolve the ID: the old short ID is dead. Find the new one with `rem list --list "Shared List" -o json` (the stderr warning also prints it).

## Reference files (load when needed)

- **[references/commands.md](references/commands.md)** — Complete flag reference for every rem command. Load when you need specific flag details, default values, or full usage examples.
- **[references/dates.md](references/dates.md)** — Natural language date grammar accepted by `--due`, `--due-before`, `--due-after`, and `--remind-me`. Load when constructing a date string from a user's phrasing.

## Common patterns

### Daily briefing — one tool call, not three
```bash
echo "== OVERDUE =="; rem overdue -o plain; echo "== TODAY =="; rem today -o plain; echo "== NEXT 3 DAYS =="; rem upcoming --days 3 -o plain
```

Chain with plain `;` (no `{ }` subshell grouping, no `$(...)` substitution) — permission allowlists check each `;`-separated segment, and `rem`/`echo` are pre-approved by this skill.

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
