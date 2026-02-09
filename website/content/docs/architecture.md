---
title: "Architecture"
description: "How rem achieves sub-200ms reads with a single binary — cgo, EventKit, AppleScript, and the design decisions behind them."
weight: 3
---

## Overview

rem uses a **split architecture**: EventKit (via cgo) for reads, AppleScript for writes. This gives the best of both worlds — instant reads with direct memory access and simple writes with AppleScript's property syntax.

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
      <span class="arch-label">Public Go API</span>
      <span class="arch-detail">pkg/client/client.go</span>
    </div>
  </div>
  <div class="arch-arrow">&#8595;</div>
  <div class="arch-layer arch-split">
    <div class="arch-box arch-read">
      <span class="arch-badge arch-badge-fast">&#60; 200ms</span>
      <span class="arch-label">Read Path</span>
      <span class="arch-sublabel">EventKit + cgo</span>
      <span class="arch-detail">internal/eventkit/</span>
      <span class="arch-file">eventkit_darwin.m</span>
      <span class="arch-file">eventkit.go</span>
    </div>
    <div class="arch-box arch-write">
      <span class="arch-badge arch-badge-write">~0.5s</span>
      <span class="arch-label">Write Path</span>
      <span class="arch-sublabel">AppleScript</span>
      <span class="arch-detail">internal/applescript/</span>
      <span class="arch-file">executor.go</span>
      <span class="arch-file">reminders.go</span>
      <span class="arch-file">lists.go</span>
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

## Read path: EventKit via cgo

All read operations go through `internal/eventkit/`, which embeds Objective-C code directly into the Go binary via cgo.

### How it works

1. A Go function (e.g., `eventkit.FetchReminders()`) calls a C function via cgo
2. The C function is implemented in Objective-C (`eventkit_darwin.m`)
3. The ObjC code creates an `EKEventStore`, queries the EventKit framework
4. Results are serialized as a JSON string and returned as `char*`
5. Go converts the string and parses JSON into domain objects

The entire round-trip — from Go, through cgo, into EventKit, back through JSON parsing — completes in under 200ms.

### Key implementation details

**Store initialization** happens once via `dispatch_once`:

```objc
static EKEventStore *store = nil;
static dispatch_once_t onceToken;
dispatch_once(&onceToken, ^{
    store = [[EKEventStore alloc] init];
    // Request TCC authorization
});
```

**Synchronous fetching** uses `dispatch_semaphore` since EventKit's fetch API is completion-based:

```objc
dispatch_semaphore_t sema = dispatch_semaphore_create(0);
[store fetchRemindersMatchingPredicate:pred
    completion:^(NSArray<EKReminder *> *reminders) {
        // serialize to JSON
        dispatch_semaphore_signal(sema);
    }];
dispatch_semaphore_wait(sema, DISPATCH_TIME_FOREVER);
```

**ARC is mandatory.** The cgo CFLAGS include `-fobjc-arc`. Without ARC, objects created inside completion handlers are released prematurely, causing silent empty results or crashes.

### Why not JXA or AppleScript for reads?

JXA (JavaScript for Automation) was rem's original read layer. Each property access is an Apple Event — a cross-process IPC call to the Reminders app. For 224 reminders with 11 properties, that's thousands of IPC calls serialized through a single pipe. Result: **42-60 seconds**.

EventKit is an in-process framework — direct memory access to the reminder store with no IPC. Result: **0.13 seconds** for the same dataset. That's a **462x speedup**.

## Write path: AppleScript

Create, update, and delete operations use AppleScript executed via `osascript`:

```applescript
tell application "Reminders"
    set newReminder to make new reminder in list "Work" ¬
        with properties {name:"Ship v2.0", due date:date "2026-02-14", priority:1}
    return id of newReminder
end tell
```

### Why AppleScript for writes?

- **Simpler syntax**: `with properties {...}` sets everything in one call
- **EventKit writes** require a more verbose save/commit cycle
- **Write operations are single-item**: the ~0.5s overhead of `osascript` is acceptable for one reminder at a time

## The flagged exception

EventKit's `EKReminder` does not expose a `flagged` property. When the `--flagged` filter is active, rem falls back to JXA to fetch flagged reminder IDs. This is the only remaining slow path (~3-4 seconds) but is rarely used.

## Single binary

The EventKit bridge compiles directly into the Go binary via cgo. `go build` detects the `.m` file, invokes Clang to compile the Objective-C, and links the EventKit and Foundation frameworks. The result is a single binary with no external dependencies.

This means `go install github.com/BRO3886/rem/cmd/rem@latest` works out of the box — no separate compilation step, no helper binaries to distribute.

## Project structure

```
internal/
├── eventkit/              # cgo + ObjC EventKit bridge
│   ├── eventkit_darwin.h  # C header (3 functions)
│   ├── eventkit_darwin.m  # ObjC implementation (~190 lines)
│   └── eventkit.go        # Go wrapper with cgo directives
│
├── applescript/           # AppleScript executor + service
│   ├── executor.go        # Runs osascript with 30s timeout
│   ├── reminders.go       # CRUD operations (reads → eventkit)
│   ├── lists.go           # List operations (reads → eventkit)
│   └── parser.go          # JSON parsing from EventKit responses
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

rem uses only three external Go dependencies:

| Package | Purpose |
|---------|---------|
| `spf13/cobra` | CLI framework (commands, flags, help) |
| `olekukonko/tablewriter` | Terminal table formatting |
| `fatih/color` | Terminal colors |

System frameworks linked via cgo:

| Framework | Purpose |
|-----------|---------|
| `EventKit` | macOS native reminder store access |
| `Foundation` | Objective-C runtime and utilities |

## Design decisions

### JSON as the cgo bridge format

The Objective-C code returns JSON strings rather than passing struct fields across the cgo boundary. This keeps the cgo interface to just 3 functions (`ek_fetch_lists`, `ek_fetch_reminders`, `ek_get_reminder`) and reuses existing Go JSON parsing. The serialization cost is negligible — under 1ms for 224 reminders.

### Custom date parser

Instead of using an external NL date library, rem includes a custom parser in `internal/parser/`. It handles 20+ patterns in ~250 lines of Go with deterministic behavior and no locale surprises.

### Prefix-matched IDs

Reminder IDs are UUIDs in the format `x-apple-reminder://UUID`. rem strips the prefix and displays only the first 8 characters. Users can pass any unique prefix to commands — matching is case-insensitive.
