---
title: "Architecture"
description: "How rem achieves sub-200ms operations with a single binary — go-eventkit, EventKit, and the design decisions behind them."
weight: 3
---

## Overview

rem uses **go-eventkit** (`github.com/BRO3886/go-eventkit`) for both reads and writes via EventKit's cgo bridge, with AppleScript only for list CRUD and flagged operations that EventKit doesn't support.

<div class="arch-diagram">
  <div class="arch-layer">
    <div class="arch-box arch-full">
      <span class="arch-label">CLI Layer</span>
      <span class="arch-detail">cmd/rem/commands/*.go</span>
    </div>
  </div>
  <div class="arch-arrow">&#8595;</div>
  <div class="arch-layer">
    <div class="arch-box arch-full">
      <span class="arch-label">Service Layer</span>
      <span class="arch-detail">internal/service/</span>
    </div>
  </div>
  <div class="arch-arrow">&#8595;</div>
  <div class="arch-layer arch-split">
    <div class="arch-box arch-read">
      <span class="arch-badge arch-badge-fast">&#60; 200ms</span>
      <span class="arch-label">EventKit Path</span>
      <span class="arch-sublabel">go-eventkit (cgo)</span>
      <span class="arch-detail">Reads + Writes</span>
      <span class="arch-file">reminders, lists, CRUD</span>
    </div>
    <div class="arch-box arch-write">
      <span class="arch-badge arch-badge-write">~0.5s</span>
      <span class="arch-label">AppleScript Path</span>
      <span class="arch-sublabel">osascript fallback</span>
      <span class="arch-detail">internal/service/</span>
      <span class="arch-file">list CRUD, flagged</span>
    </div>
  </div>
  <div class="arch-arrow">&#8595;</div>
  <div class="arch-layer">
    <div class="arch-box arch-full arch-system">
      <span class="arch-label">macOS Frameworks</span>
      <div class="arch-tags">
        <span class="arch-tag">EventKit</span>
        <span class="arch-tag">Foundation</span>
        <span class="arch-tag">osascript</span>
      </div>
    </div>
  </div>
</div>

## Main path: go-eventkit

All reminder read and write operations go through `go-eventkit` (`github.com/BRO3886/go-eventkit/reminders`), which provides native EventKit bindings via cgo + Objective-C.

### How it works

1. rem creates a `reminders.Client` via `reminders.New()` — this requests TCC authorization
2. Read operations (e.g., `client.Reminders(opts...)`) call into cgo → Objective-C → EventKit
3. Write operations (e.g., `client.CreateReminder(input)`) go through the same path
4. Results are serialized as JSON strings across the cgo boundary and parsed into Go types
5. The entire round-trip completes in under 200ms for both reads and writes

### Key implementation details

**Store initialization** happens once via `dispatch_once` inside go-eventkit:

```objc
static EKEventStore *store = nil;
static dispatch_once_t onceToken;
dispatch_once(&onceToken, ^{
    store = [[EKEventStore alloc] init];
    // Request TCC authorization
});
```

**ARC is mandatory.** go-eventkit's cgo CFLAGS include `-fobjc-arc`. Without ARC, objects created inside completion handlers are released prematurely, causing silent empty results or crashes.

### Why not JXA or AppleScript for reads?

JXA (JavaScript for Automation) was rem's original read layer. Each property access is an Apple Event — a cross-process IPC call to the Reminders app. For 224 reminders with 11 properties, that's thousands of IPC calls serialized through a single pipe. Result: **42-60 seconds**.

EventKit is an in-process framework — direct memory access to the reminder store with no IPC. Result: **0.13 seconds** for the same dataset. That's a **462x speedup**.

## AppleScript fallback

Three operations still use AppleScript via `osascript`:

1. **List create/rename/delete** — go-eventkit doesn't support list CRUD operations
2. **Flag/unflag reminders** — EventKit doesn't expose the `flagged` property
3. **Default list name** — not exposed by go-eventkit

```applescript
tell application "Reminders"
    set newList to make new list with properties {name:"My List"}
    return id of newList
end tell
```

### The flagged exception

EventKit's `EKReminder` does not expose a `flagged` property. When the `--flagged` filter is active, rem falls back to JXA to fetch flagged reminder IDs. This is the only remaining slow path (~3-4 seconds) but is rarely used. Flag/unflag write operations use AppleScript.

## Single binary

go-eventkit's Objective-C code compiles directly into the Go binary via cgo. `go build` detects the `.m` files, invokes Clang to compile the Objective-C, and links the EventKit and Foundation frameworks. The result is a single binary with no external dependencies.

This means `go install github.com/BRO3886/rem/cmd/rem@latest` works out of the box — no separate compilation step, no helper binaries to distribute.

## Project structure

```
internal/
├── service/               # Service layer (go-eventkit + AppleScript fallback)
│   ├── executor.go        # Runs osascript (list CRUD, flagged)
│   ├── reminders.go       # ReminderService wrapping go-eventkit
│   ├── lists.go           # ListService wrapping go-eventkit + AppleScript
│   └── parser.go          # URL extraction from notes
│
├── reminder/              # Domain models
│   └── model.go           # Reminder, List, Priority types
│
├── parser/                # Natural language date parser
│   └── date.go            # 20+ patterns, no external deps
│
├── export/                # Import/export
│   ├── json.go            # JSON format
│   └── csv.go             # CSV format
│
└── ui/                    # Terminal output
    └── output.go          # Table, JSON, plain formatters
```

## Dependencies

rem uses four external Go dependencies:

| Package | Purpose |
|---------|---------|
| `BRO3886/go-eventkit` | Native EventKit bindings (cgo + ObjC, reads AND writes) |
| `spf13/cobra` | CLI framework (commands, flags, help) |
| `olekukonko/tablewriter` | Terminal table formatting |
| `fatih/color` | Terminal colors |

System frameworks linked via cgo (through go-eventkit):

| Framework | Purpose |
|-----------|---------|
| `EventKit` | macOS native reminder store access |
| `Foundation` | Objective-C runtime and utilities |

## Design decisions

### go-eventkit as a standalone library

The EventKit bridge was extracted from rem into a standalone Go library (`github.com/BRO3886/go-eventkit`). This provides:
- **Reusability** — other Go projects can use EventKit without rem
- **Separation of concerns** — rem is a thin CLI wrapper, go-eventkit handles all cgo/EventKit complexity
- **Calendar support** — go-eventkit also supports Calendar/Events, which rem doesn't use

### Custom date parser

Instead of using an external NL date library, rem includes a custom parser in `internal/parser/`. It handles 20+ patterns in ~250 lines of Go with deterministic behavior and no locale surprises.

### Prefix-matched IDs

Reminder IDs are UUIDs in the format `x-apple-reminder://UUID`. rem strips the prefix and displays only the first 8 characters. Users can pass any unique prefix to commands — matching is case-insensitive.
