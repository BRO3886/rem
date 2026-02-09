---
title: "Getting Started"
description: "Install rem and start managing your macOS Reminders from the terminal in under a minute."
weight: 1
---

## Requirements

- **macOS 10.12+** (Sierra or later)
- **Go 1.21+** (for building from source)
- **Xcode Command Line Tools** (provides clang and framework headers for cgo)

On first run, macOS will prompt you to grant Reminders access in System Settings > Privacy & Security > Reminders.

## Installation

### Via `go install` (recommended)

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

## What's next

- [Commands](/docs/commands/) — Full reference for all 19 commands
- [Architecture](/docs/architecture/) — How rem works under the hood
- [Go API](/docs/api/) — Use rem as a library in your Go programs
