# Natural Language Date Reference

The `--due` flag on `rem add` and `rem update` accepts natural language dates. All parsing is built-in with no external dependencies.

## Keywords

| Input | Resolves to |
|-------|-------------|
| `today` | Today at 9:00 AM |
| `tomorrow` | Tomorrow at 9:00 AM |
| `yesterday` | Yesterday at 9:00 AM |
| `next week` | 7 days from now at 9:00 AM |
| `next month` | 30 days from now at 9:00 AM |
| `eod` / `end of day` | Today at 5:00 PM |
| `eow` / `end of week` | Next Friday at 5:00 PM |

## Relative Time

Pattern: `in <number> <unit>`

| Input | Example result |
|-------|---------------|
| `in 5 minutes` / `in 5 mins` | 5 minutes from now |
| `in 3 hours` / `in 3 hrs` | 3 hours from now |
| `in 2 days` | 2 days from now |
| `in 1 week` | 7 days from now |
| `in 2 months` | 2 months from now |

## Weekday Patterns

| Input | Resolves to |
|-------|-------------|
| `next monday` | Next Monday at 9:00 AM |
| `next friday` | Next Friday at 9:00 AM |
| `next monday at 2pm` | Next Monday at 2:00 PM |
| `next fri at 3:30pm` | Next Friday at 3:30 PM |

Supports full names and 3-letter abbreviations: `sun`, `mon`, `tue`, `wed`, `thu`, `fri`, `sat`.

## Keyword + Time

| Input | Resolves to |
|-------|-------------|
| `today at 5pm` | Today at 5:00 PM |
| `tomorrow at 3:30pm` | Tomorrow at 3:30 PM |
| `yesterday at 9am` | Yesterday at 9:00 AM |

## Time Only

| Input | Resolves to |
|-------|-------------|
| `5pm` | Today at 5:00 PM (or tomorrow if already past) |
| `9am` | Today at 9:00 AM (or tomorrow if already past) |
| `5:30pm` | Today at 5:30 PM (or tomorrow if already past) |
| `17:00` | Today at 5:00 PM (or tomorrow if already past) |

## Standard Date Formats

| Format | Example |
|--------|---------|
| ISO date | `2026-02-15` |
| ISO datetime | `2026-02-15 14:30` |
| ISO with T | `2026-02-15T14:30` |
| 12-hour | `2026-02-15 2:30PM` |
| US format | `02/15/2026` |
| US with time | `02/15/2026 14:30` |
| Month name | `Feb 15, 2026` / `February 15, 2026` |
| Day-first | `15 Feb 2026` |

## Clearing a Due Date

Use `none` to clear an existing due date:

```bash
rem update abc12345 --due none
```
