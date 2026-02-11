---
title: "Performance"
description: "How rem went from 60 seconds to 130 milliseconds — the optimization journey from JXA to EventKit via cgo."
weight: 5
---

## The numbers

Every read command completes in under 200ms. Tested with 224 reminders across 12 lists.

| Command | Time |
|---------|------|
| `rem lists` | 0.12s |
| `rem list` (all 224 reminders) | 0.13s |
| `rem list --incomplete` | 0.11s |
| `rem show` (by prefix) | 0.11s |
| `rem search "query"` | 0.11s |
| `rem stats` | 0.17s |
| `rem overdue` | 0.12s |
| `rem upcoming` | 0.12s |
| `rem export --format json` | 0.13s |

All write operations — including reminder CRUD and list CRUD — go through EventKit via go-eventkit, completing in under 200ms. Only flagged operations still use AppleScript (~0.5s).

## The optimization journey

rem went through four performance stages, each achieving an order-of-magnitude improvement.

### Stage 1: AppleScript loops (unusable)

The first attempt used AppleScript's `repeat with r in theReminders` to iterate through reminders. Even 8 reminders caused a 30+ second timeout. AppleScript per-element property access is catastrophically slow.

### Stage 2: JXA bulk access (slow but functional)

Switched to JXA's columnar access pattern: `list.reminders.name()` returns all names in a single Apple Event. This was functional but still painfully slow.

**Why it was slow:** Each Apple Event is a cross-process IPC call to the Reminders app. For 11 properties across 4 lists, that's 44 IPC calls — each taking 3-4 seconds. The Reminders app serializes all incoming requests.

| Command | JXA Time |
|---------|----------|
| `rem lists` | 8.3s |
| `rem list` (all) | ~60s |
| `rem show` | ~5s |
| `rem search` | ~60s |
| `rem stats` | ~68s |

Concurrent `osascript` processes didn't help — Reminders.app serializes all Apple Events internally.

### Stage 3: Swift EventKit helper (fast, two binaries)

Built a compiled Swift binary using the EventKit framework. EventKit is an in-process framework — direct memory access, no IPC. All reads dropped to under 250ms.

The tradeoff: the build produced two binaries (`rem` + `reminders-helper`), complicating installation and distribution.

### Stage 4: cgo + Objective-C (fast, single binary)

Replaced the Swift helper with an Objective-C file compiled directly into the Go binary via cgo. Same performance as Stage 3, but produces a single binary.

**The key insight:** `go build` automatically compiles `.m` (Objective-C) files when cgo is enabled. No separate compilation step. `go install` just works.

### Stage 5: go-eventkit (fast reads AND writes)

Extracted the EventKit bridge into a standalone library (`github.com/BRO3886/go-eventkit`) and extended it to support writes (create, update, delete, complete/uncomplete) — all through EventKit. This eliminated AppleScript for reminder operations entirely, bringing write times from ~0.5s to under 200ms.

### Stage 6: go-eventkit list CRUD (full EventKit coverage)

go-eventkit v0.2.1 added list CRUD support (create, rename/recolor, delete). rem now uses EventKit for all list operations too, eliminating the last AppleScript dependency for CRUD. AppleScript is now only used for flagged operations (EventKit limitation) and default list name queries.

## Before vs after

| Command | Before (JXA) | After (EventKit) | Speedup |
|---------|-------------|-------------------|---------|
| `rem lists` | 8.3s | 0.12s | **69x** |
| `rem list` (all 224) | ~60s | 0.13s | **462x** |
| `rem list --incomplete` | ~42s | 0.11s | **382x** |
| `rem show` (prefix) | ~5s | 0.11s | **45x** |
| `rem search` | ~60s | 0.11s | **545x** |
| `rem stats` | ~68s | 0.17s | **400x** |
| `rem export json` | ~60s | 0.12s | **500x** |

## Where the time goes

For a typical read operation (~130ms):

| Phase | Time |
|-------|------|
| Binary startup | ~5ms |
| cgo function call | <1ms |
| EventKit query | 100-170ms |
| JSON serialization (ObjC side) | <1ms |
| JSON parsing (Go side) | <5ms |
| Terminal output | <5ms |

EventKit is the bottleneck — and it's the fastest possible path. There's no further optimization to be done for reads on macOS.

## Why EventKit is fast

EventKit is an in-process framework. When you call `fetchRemindersMatchingPredicate:`, it reads directly from the local reminder store (a SQLite database) without any IPC or process boundary crossing.

JXA/AppleScript, by contrast, sends Apple Events to the Reminders.app process. Each event is serialized, sent over Mach IPC, deserialized, processed, and the result sent back the same way. For bulk operations, this overhead compounds dramatically.

## Writes are now fast too

Since the migration to go-eventkit, all write operations — including reminder CRUD and list CRUD — go through EventKit and complete in under 200ms. Only flagged operations still use AppleScript (~0.5s).

## Known slow path

The `--flagged` filter falls back to JXA because EventKit's `EKReminder` doesn't expose a `flagged` property. This takes ~3-4 seconds. All other reads use EventKit and complete in under 200ms.
