# rem - CLI for macOS Reminders

## Non-Negotiables
- **Conventional Commits**: ALL commits MUST follow [Conventional Commits](https://www.conventionalcommits.org/). Format: `type(scope): description`. Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`. No exceptions.

## What is this?
Go CLI wrapping macOS Reminders. Uses `go-eventkit` (cgo + Objective-C EventKit + private ReminderKit bridge) for fast reads AND writes (<200ms) as a single binary — including reminder CRUD, list CRUD, and flagged operations. AppleScript only for the default list name query. Provides CRUD for reminders/lists, natural language dates, and import/export. For programmatic Go access, use `go-eventkit` directly.

## Architecture
- `cmd/rem/commands/` - Cobra CLI commands (one file per command); `huh_helpers.go` has shared interactive utilities
- `internal/service/` - Service layer: `reminders.go` and `lists.go` wrap `go-eventkit` client. `executor.go` runs `osascript` for the default list name query only.
- `internal/reminder/` - Domain models: `Reminder`, `List`, `Priority`
- `internal/export/` - JSON/CSV import/export
- `internal/skills/` - Agent skill install/uninstall/status logic
- `internal/update/` - Background update check (GitHub releases, 24h cache)
- `internal/ui/` - Table (`olekukonko/tablewriter` v1.x), plain, JSON output
- `skills/rem-cli/` - Embedded agent skill files (go:embed'd into binary)

## Critical: Architecture Rules
- **ALL reads AND writes go through `go-eventkit`** (`github.com/BRO3886/go-eventkit/reminders`) — in-process EventKit via cgo, <200ms
- **Single binary** — go-eventkit's cgo code compiled into the binary
- **AppleScript only for**: default list name query (nothing else)
- **Flagged is read/written via go-eventkit** using the private ReminderKit bridge (same mechanism as URL attachments). `EKReminder.flagged` doesn't exist, but `REMReminder.flagged` does. Read via KVC (`valueForKey:`, not `isFlagged` selector — it's a dynamic property). Write via `REMSaveRequest.updateReminder:` change item, then `changeItem.flaggedContext.setFlagged:`, then `saveSynchronouslyWithError:`. Refetch the EKReminder after save so the returned dict reflects the new value. See journal 011. **Graceful degradation (#44):** like tags, flagged is set via a separate `UpdateReminder` call after the core create/update — if the private API is unavailable, rem warns on stderr and the reminder still succeeds. `FlagReminder`/`UnflagReminder` delegate to `UpdateReminder(id, {"flagged":bool})` so all four flagged paths (add/update/flag/unflag) share one degradation policy: a private-API write failure warns + exits 0, a genuine error (not found) stays fatal. Requires go-eventkit ≥ v0.10.0 (v0.9.0 swallowed the flagged write failure silently). See journal 014.
- **go-eventkit field names**: `Title` (not `Name`), `Notes` (not `Body`), `List` (not `ListName`), native `URL` field
- **URL field goes to `REMURLAttachment`, NOT `EKCalendarItem.URL`**: `EKCalendarItem.URL` is a public property that's disconnected from the Reminders.app UI (Apple bug/limitation). go-eventkit v0.5.0+ writes to the real UI-visible URL field via a private ReminderKit bridge (`REMSaveRequest` → `attachmentContext.setURLAttachmentWithURL:`), guarded by `respondsToSelector:` with fallback to `EKCalendarItem.URL`. See journal 009.
- **List CRUD via go-eventkit**: `CreateList` (auto-discovers source), `UpdateList` (ID-based), `DeleteList` (ID-based). Immutable lists are rejected.
- **List moves (#50)**: `UpdateReminderInput.ListName` moves a reminder natively, ID-stable. Plain moves use public EventKit (`calendar` reassignment); when a shared list is on either end — which public EventKit rejects with reminderkit error `-3002` — go-eventkit ≥ v0.11.0 reparents via the private ReminderKit bridge (change item `listID` written by KVC), with a rescue reparent if a public save fails unexpectedly. No rem-side handling needed.
- **Tags use the private ReminderKit hashtag API** (`REMHashtag`, `REMReminderHashtagContext`). Title `#hashtag` parsing happens in rem only (not go-eventkit). Tags are written as a separate `UpdateReminder` call after the core create/update — if the private API is unavailable, rem warns on stderr but the reminder still succeeds. Pure-numeric fragments like `#42` are filtered out by `isNumeric()` to avoid treating issue references as tags.
- Priority: 0=none, 1-4=high, 5=medium, 6-9=low
- `due date` and `remind me date` are independent
- **`rem add --due` auto-attaches a zero-offset alarm by default** (matches Apple Reminders.app). Use `--silent` to suppress. The alarm decision logic lives in `buildAlarms()` in `filters.go` — both CLI and interactive add paths call it. Do NOT pass `--remind-me 0m` to enable — it's the default.

## Libraries
- `BRO3886/go-eventkit` - **EventKit bindings** (cgo + ObjC, reads AND writes) + `dateparser` package for NL date parsing
- `spf13/cobra` - CLI framework
- `olekukonko/tablewriter` v1.x - **new API**: `NewTable()`, `.Header()`, `.Append()`, `.Render()` (NOT the old `SetHeader`/`SetBorder` API)
- `fatih/color` - terminal colors
- `olekukonko/tablewriter/tw` - alignment constants (`tw.AlignLeft`)
- `charmbracelet/huh` - interactive forms (multi-select, select, input, confirm) for all `-i` modes and skills

## Build & Test
```bash
make build        # -> bin/rem (single binary, includes EventKit via cgo)
make release      # -> bin/rem-darwin-arm64.tar.gz (for GitHub Releases upload)
go test ./...     # unit tests (date parser, export, models)
make completions  # bash/zsh/fish
```

## Releasing
Release steps in order — do not skip or reorder:
1. `git push` — push all commits to main **first**. Never tag unpushed commits.
2. `git tag vX.Y.Z` — tag after push so the tag points to a commit already on remote main
3. `git push origin vX.Y.Z` — push the tag explicitly
4. `make release` — builds `bin/rem-darwin-arm64.tar.gz` with correct version embedded via `git describe --tags`
5. Verify: `./bin/rem version` must show `vX.Y.Z` (not a dirty describe like `v0.5.0-2-gae75da9`)
6. `gh release create vX.Y.Z bin/rem-darwin-arm64.tar.gz`

- **CRITICAL: Push before tag.** Tagging an unpushed commit then running `gh release create` pushes the tag + that commit but leaves `main` behind on remote — release binary is built from code not reachable from main.
- **CRITICAL: Tag before build.** `make release` uses `git describe --tags` to embed the version. Tag first or the binary gets a dirty version string.
- **Always use `make release`** — produces a `.tar.gz` that preserves execute permissions (HTTP downloads strip +x from raw binaries)
- No CI release workflow — macOS runners are too expensive for free tier

## Agent Skills
- `rem skills install/uninstall/status` manages embedded agent skill files
- Skills are `go:embed`'d from `skills/rem-cli/` into the binary
- `internal/skills/` handles install/uninstall logic, `internal/update/` handles background update checks
- Background update check runs in a goroutine during `PersistentPreRun`, prints notice in `PersistentPostRun`
- `REM_NO_UPDATE_CHECK=1` disables update checks; also skipped for dev builds, json output, non-TTY, meta commands

## Conventions
- Short IDs displayed as first 8 chars of full `x-apple-reminder://UUID` ID
- Prefix matching: users can pass partial IDs to any command
- **Multi-ID mutations**: `complete`, `uncomplete`, `flag`, `unflag`, `delete` all accept multiple IDs. Shared helper `runBatchAction` in `cmd/rem/commands/batch.go` resolves all IDs before mutating (fail-fast on a bad ID), continues past per-item errors, exits non-zero if any failed
- All commands support `-o json|table|plain`
- `NO_COLOR` env var respected
- **Interactive mode (`-i`)**: All mutation commands (add, complete, delete, flag, unflag, update, list-mgmt create/rename/delete) support `-i` for huh-based interactive forms. Shared helpers in `huh_helpers.go`. Theme: `ThemeCatppuccin()`. All handle `ErrUserAborted` gracefully.

## Website & Hosting
- Documentation site: `rem.sidv.dev` (Hugo on Cloudflare Pages)
- Source: `website/` dir, deployed via `.github/workflows/deploy.yml`
- Install script served at `rem.sidv.dev/install` (from `website/static/install`)
- Domain: `sidv.dev` (owned by user, managed on Cloudflare)

## Journal
Engineering journals live in `journals/` dir. See `.claude/commands/journal.md` for the journaling command.
