---
title: "Getting Started"
description: "Install rem — a CLI for macOS Reminders — and start managing your reminders from the terminal in under a minute. Covers install script, go install, shell completions, and AI agent skills."
keywords:
  - install rem CLI
  - macOS Reminders terminal setup
  - rem install script
  - rem go install
  - rem shell completions
  - rem AI agent skills
weight: 1
sitemap:
  priority: 0.9
  changefreq: weekly
---

## Requirements

- **macOS 10.12+** (Sierra or later)
- **Go 1.21+** (for building from source)
- **Xcode Command Line Tools** (provides clang and framework headers for cgo)

On first run, macOS will prompt you to grant Reminders access in System Settings > Privacy & Security > Reminders.

## Installation

### Homebrew

```bash
brew tap BRO3886/tap
brew install rem-cli
```

Note: the formula is named `rem-cli` to avoid a conflict with an older `rem` package in homebrew-core. The binary installs as `rem`.

### Via `go install`

```bash
go install github.com/BRO3886/rem/cmd/rem@latest
```

This compiles the single binary with EventKit support via cgo — no separate helper binaries needed.

### From source

```bash
git clone https://github.com/BRO3886/rem.git
cd rem
make build
```

The compiled binary will be at `bin/rem`. To install system-wide:

```bash
make install
```

## Quick start

### List your reminders

```bash
rem list
```

### Create a reminder

```bash
rem add "Review pull request" --due tomorrow --priority high --list Work
```

When `--due` is set, rem automatically attaches an alarm at the due time so you actually get a notification. Use `--remind-me 15m` for a custom offset before the due time, or `--silent` to skip the alarm entirely.

### Use natural language dates

```bash
rem add "Team standup" --due "next monday at 9am"
rem add "Submit report" --due "in 3 hours"
rem add "Friday wrap-up" --due "eod friday"
```

### Complete a reminder

Use the short ID shown in the list output:

```bash
rem complete 6ECE
```

Short IDs are prefix-matched — you only need enough characters to be unique.

### View your stats

```bash
rem stats
```

### Search across all lists

```bash
rem search "pull request"
```

## Output formats

Every command supports three output formats:

```bash
rem list -o table    # colored ASCII table (default)
rem list -o json     # machine-readable JSON
rem list -o plain    # simple text, one per line
```

The `NO_COLOR` environment variable is respected for colorless output.

## Shell completions

Generate completions for your shell:

```bash
# Bash
rem completion bash > /usr/local/etc/bash_completion.d/rem

# Zsh
rem completion zsh > "${fpath[1]}/_rem"

# Fish
rem completion fish > ~/.config/fish/completions/rem.fish
```

## AI Agent Skills

rem ships with an embedded skill that teaches AI coding agents (Claude Code, Codex CLI, OpenClaw, etc.) how to use it effectively. Install it with:

```bash
rem skills install
```

This copies the skill files to the agent's skill directory (e.g. `~/.claude/skills/rem-cli/`). The skill includes command references, date parsing docs, and usage examples. On a TTY you'll see a brief explanation and a confirmation prompt before anything is written — use `--dry-run` to preview the exact files first.

```bash
rem skills status      # Check installation status
rem skills uninstall   # Remove the skill
```

You can also target specific agents:

```bash
rem skills install --agent claude    # Claude Code only
rem skills install --agent codex     # Codex CLI only
rem skills install --agent openclaw  # OpenClaw only
rem skills install --agent all       # All supported agents
```

## What's next

- [Commands](/docs/commands/) — Full reference for all 20 commands
- [Architecture](/docs/architecture/) — How rem works under the hood
- [Go API](/docs/api/) — Use rem as a library in your Go programs
