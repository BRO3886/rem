---
title: "Architecture"
description: "How rem achieves sub-200ms macOS Reminders operations with a single Go binary — using go-eventkit to bridge EventKit via cgo, eliminating AppleScript IPC overhead entirely."
keywords:
  - rem architecture
  - go-eventkit EventKit cgo
  - macOS Reminders EventKit Go
  - rem single binary
  - EventKit via cgo Go
  - rem AppleScript vs EventKit
weight: 3
sitemap:
  priority: 0.7
  changefreq: monthly
---

## Overview

rem uses **go-eventkit** (`github.com/BRO3886/go-eventkit`) for all reads and writes — including reminder CRUD, list CRUD, and flagged operations — via EventKit's cgo bridge and a private ReminderKit bridge. AppleScript is only used for the default list name query.

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
      <span class="arch-file">default list name only</span>
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

## Private ReminderKit bridge

EventKit's `EKReminder` does not expose a `flagged` property or the real URL field visible in Reminders.app. go-eventkit bridges these through Apple's private `ReminderKit.framework`:

- **Flagged** — read via KVC (`valueForKey:@"flagged"` on `REMReminder`), write via `REMSaveRequest` → `flaggedContext.setFlagged:`
- **URL attachments** — write via `REMSaveRequest` → `attachmentContext.setURLAttachmentWithURL:`

Both operations complete in under 200ms, the same as all other EventKit operations. All private API calls are guarded by `respondsToSelector:` checks and will fail cleanly if Apple removes them in a future macOS version.

## AppleScript fallback

One operation still uses AppleScript via `osascript`:

1. **Default list name** — not exposed by EventKit or go-eventkit

## Single binary

go-eventkit's Objective-C code compiles directly into the Go binary via cgo. `go build` detects the `.m` files, invokes Clang to compile the Objective-C, and links the EventKit and Foundation frameworks. The result is a single binary with no external dependencies.

This means `go install github.com/BRO3886/rem/cmd/rem@latest` works out of the box — no separate compilation step, no helper binaries to distribute.

## Project structure

```
internal/
├── service/               # Service layer (go-eventkit + AppleScript for default list name only)
│   ├── executor.go        # Runs osascript (default list name query)
│   ├── reminders.go       # ReminderService wrapping go-eventkit
│   ├── lists.go           # ListService wrapping go-eventkit
│   └── parser.go          # Backward-compat URL extraction from notes body (fallback reader)
│
├── reminder/              # Domain models
│   └── model.go           # Reminder, List, Priority types
│
├── export/                # Import/export
│   ├── json.go            # JSON format
│   └── csv.go             # CSV format
│
└── ui/                    # Terminal output
    └── output.go          # Table, JSON, plain formatters
```

## Dependencies

rem uses five external Go dependencies:

| Package | Purpose |
|---------|---------|
| `BRO3886/go-eventkit` v0.8.0+ | Native EventKit bindings (cgo + ObjC, reads AND writes). Includes the private ReminderKit bridge for flagged state and URL attachments. |
| `spf13/cobra` | CLI framework (commands, flags, help) |
| `olekukonko/tablewriter` | Terminal table formatting |
| `fatih/color` | Terminal colors |
| `charmbracelet/huh` | Interactive forms for all `-i` modes |

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

### Date parsing via go-eventkit

Date parsing uses the shared `dateparser` package from `go-eventkit`. This provides 30+ patterns including `now`, `this week`, backward-relative (`5 days ago`), bare weekdays (`monday 2pm`), and month-day patterns (`mar 15`). Configured with `WithDefaultHour(9)` so bare dates resolve to 9 AM, and `WithSmartTimeRollover()` so past times like `5pm` roll to tomorrow.

### Prefix-matched IDs

Reminder IDs are UUIDs in the format `x-apple-reminder://UUID`. rem strips the prefix and displays only the first 8 characters. Users can pass any unique prefix to commands — matching is case-insensitive.
