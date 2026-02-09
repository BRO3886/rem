# Session 3 — Single Binary & Pure Go Research

## Problem Statement

The current `rem` architecture has two pain points:

1. **Two binaries**: `bin/rem` (Go CLI) + `bin/reminders-helper` (compiled Swift EventKit binary). Users need both, and the Go executor locates the helper via `os.Executable()` dir or PATH — fragile.
2. **`pkg/client/` is not pure Go**: The public Go API at `pkg/client/` shells out to the Swift helper. Anyone doing `go get github.com/BRO3886/rem/pkg/client` gets a library that doesn't work without separately compiling and installing the Swift binary. This defeats the purpose of a public API.
3. **CLI calling CLI**: Go spawning a subprocess, capturing stdout, parsing JSON — functional but not the cleanest architecture. Adds ~5ms process spawn overhead per call.

## Options Explored

We evaluated **12 distinct approaches** across three categories:

- **A.** Go ↔ Swift/ObjC bindings (single binary via linking)
- **B.** Embedding the Swift binary (single binary via bundling)
- **C.** Pure Go alternatives (no Swift at all)

---

## A. Go ↔ Swift/Objective-C Bindings

### A1. cgo + Objective-C `.m` File (⭐ RECOMMENDED for single binary)

Place an Objective-C `.m` file in the same Go package. Since Go 1.3, `go build` automatically compiles `.m` files when cgo is enabled. The ObjC code calls EventKit directly, returns JSON as a C string, and Go calls it via cgo.

```
internal/eventkit/
    eventkit.go          // Go: cgo directives, C.ek_fetch_lists() calls
    eventkit_darwin.m    // ObjC: EKEventStore, fetchReminders, JSON output
    eventkit_darwin.h    // C header: ek_fetch_lists(), ek_fetch_reminders(), etc.
```

**How async EventKit works**: `dispatch_semaphore` in the ObjC layer blocks while the completion handler fires on a background GCD queue. Exactly what our Swift helper already does. No NSRunLoop needed — EventKit dispatches callbacks on background queues, not the main thread.

| Criterion            | Assessment                                                                                                                                                      |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Single binary        | **Yes** — `go build` links everything into one executable                                                                                                       |
| Requires cgo         | Yes (`CGO_ENABLED=1`, default on macOS)                                                                                                                         |
| Build complexity     | **Low** — standard `go build`, no separate compile steps                                                                                                        |
| `go install` works   | **Yes** — `.m` files compiled automatically by cgo's Clang                                                                                                      |
| `pkg/client/` works  | **Yes** — library consumers just need Xcode CLT (any macOS dev has this)                                                                                        |
| Runtime overhead     | **~0** — direct function call, no process spawn                                                                                                                 |
| Rewrite effort       | Medium — translate 247 lines of Swift to Objective-C (straightforward, same EventKit API)                                                                       |
| Cross-compilation    | No (macOS only, but that's inherent to EventKit)                                                                                                                |
| Production precedent | High — used by [keybase/go-keychain](https://github.com/keybase/go-keychain), [common-fate/go-apple-security](https://github.com/common-fate/go-apple-security) |

**Tradeoffs**:

- ✅ Simplest build — just `go build`, no Makefile changes
- ✅ `go install` and `go get` work — `.m` files are compiled by cgo automatically
- ✅ Proven pattern in production Go projects
- ✅ Eliminates process spawn overhead (~5ms per call)
- ❌ Requires cgo (default on macOS, but prevents `CGO_ENABLED=0` builds)
- ❌ Must rewrite Swift to Objective-C (one-time effort, ~200 lines)
- ❌ Requires Xcode Command Line Tools for building

---

### A2. Swift Static Library via `@_cdecl` + cgo

Compile our existing `helper.swift` as a static library (`.a`) with `@_cdecl` C-callable exports, then link into Go via cgo.

```bash
swiftc -parse-as-library -emit-object -O eventkit_bridge.swift -o eventkit_bridge.o
ar rcs libeventkitbridge.a eventkit_bridge.o
# Go links via: #cgo LDFLAGS: -leventkitbridge -framework EventKit -lswiftCore
```

| Criterion          | Assessment                                                                                                    |
| ------------------ | ------------------------------------------------------------------------------------------------------------- |
| Single binary      | **Yes** — `.a` statically linked                                                                              |
| Requires cgo       | Yes                                                                                                           |
| Build complexity   | **High** — separate `swiftc` step, Swift runtime linking (`-lswiftCore`), fragile LDFLAGS                     |
| `go install` works | **No** — `go install` can't run `swiftc`                                                                      |
| Rewrite effort     | Low — keep Swift code, just add `@_cdecl` exports                                                             |
| Risk               | `@_cdecl` is officially unsupported (replaced by `@c` in Swift 6.3 via SE-0495, but not yet widely available) |

**Tradeoffs**:

- ✅ Keep existing Swift code
- ✅ Single binary output
- ❌ Complex multi-step build pipeline
- ❌ Swift runtime linkage fragile across Xcode versions
- ❌ `@_cdecl` is unsupported — `@c` (SE-0495) is the replacement but not yet stable
- ❌ `go install` broken — can't run `swiftc`
- ❌ Two runtimes (Go + Swift) in one process — largely untested territory

---

### A3. purego + ObjC Runtime (⭐ RECOMMENDED for pure Go)

[ebitengine/purego](https://github.com/ebitengine/purego) provides pure-Go bindings to the Objective-C runtime (`objc_msgSend`, class/selector lookup, block creation) — **no cgo required**. Call EventKit's ObjC API directly from Go.

```go
import "github.com/ebitengine/purego/objc"

// Load EventKit
purego.Dlopen("/System/Library/Frameworks/EventKit.framework/EventKit", purego.RTLD_GLOBAL)

// Create EKEventStore
store := objc.ID(objc.GetClass("EKEventStore")).Send(selAlloc).Send(selInit)

// Fetch reminders (async with Go channel)
ch := make(chan objc.ID)
block := objc.NewBlock(func(b objc.Block, reminders objc.ID) { ch <- reminders })
store.Send(selFetchReminders, predicate, block)
result := <-ch
```

| Criterion           | Assessment                                                                   |
| ------------------- | ---------------------------------------------------------------------------- |
| Single binary       | **Yes** — pure Go binary                                                     |
| Requires cgo        | **No** — `CGO_ENABLED=0` works                                               |
| Build complexity    | **Lowest** — standard `go build`, no Xcode headers needed                    |
| `go install` works  | **Yes** — no C compiler needed                                               |
| `pkg/client/` works | **Yes** — truly pure Go, any macOS user can `go get` it                      |
| Runtime overhead    | Negligible — `syscall.Syscall` for each `objc_msgSend`, ~ns overhead         |
| Rewrite effort      | **High** — manually bridge every ObjC class, selector, type conversion       |
| Code volume         | ~800-1500 lines to replicate 247 lines of Swift                              |
| Maturity            | purego: actively maintained (pushed Feb 2026), v0.9.1. Block support merged. |

**Tradeoffs**:

- ✅ **Pure Go** — no cgo, no Xcode dependency for building
- ✅ `go install` works perfectly — the dream for `pkg/client/`
- ✅ Single binary, smallest possible
- ✅ Fast compilation (no C compiler invocation)
- ✅ Actively maintained (ebitengine/purego pushed yesterday)
- ❌ **Very verbose** — every ObjC property access = manual selector + type conversion
- ❌ No compile-time type safety — typo in selector name = runtime crash
- ❌ Must manually bridge: NSString, NSArray, NSDate, NSDateComponents, EKReminder (15+ properties), EKCalendar, EKEventStore, NSPredicate
- ❌ ObjC block callbacks need careful goroutine/thread handling
- ❌ Framework must be loaded at runtime via `dlopen`

---

### A4. DarwinKit (progrium/darwinkit)

Auto-generated Go bindings for Apple frameworks. Currently supports 33 frameworks with 2,353 classes.

**EventKit is NOT supported.** No issues, PRs, or discussions requesting it. Adding it requires running DarwinKit's code generation pipeline — iterative, partially documented, and the project's last release was July 2024.

| Criterion        | Assessment                                                  |
| ---------------- | ----------------------------------------------------------- |
| EventKit support | **No**                                                      |
| Requires cgo     | Yes (cgo + libffi)                                          |
| Adding EventKit  | Possible but high effort — undocumented generation pipeline |
| Maintenance      | Slow — last release Jul 2024, last push Mar 2025            |

**Verdict**: Not viable unless you want to contribute EventKit bindings upstream. The generation pipeline is complex and the project's release cadence is slow.

DarwinKit's low-level `objc` package could be used manually (like purego), but it requires cgo — no advantage over approach A1.

---

## B. Embedding the Swift Binary

### B1. `go:embed` + Extract at Runtime

Embed the compiled `reminders-helper` (98KB) in the Go binary via `//go:embed`. Extract to `~/Library/Caches/rem/` on first run, hash-check to skip re-extraction.

| Criterion             | Assessment                                                                       |
| --------------------- | -------------------------------------------------------------------------------- |
| Single binary to user | **Yes** (helper extracted transparently)                                         |
| Binary size overhead  | +98KB (trivial)                                                                  |
| First-run cost        | ~5-15ms to write + codesign                                                      |
| `go install` works    | **No** — embedded file must exist at build time, `go install` can't run `swiftc` |
| Still two "binaries"  | Yes (architecturally) — just hidden. Still spawns subprocess.                    |

**Tradeoffs**:

- ✅ Simplest change from current architecture — just embed what we have
- ✅ Single file distribution
- ❌ Still spawning a subprocess per call (~5ms overhead)
- ❌ `go install` broken
- ❌ Temp file extraction — cache invalidation, code signing on Apple Silicon
- ❌ Doesn't fix `pkg/client/` problem

### B2. Append to Binary (maja42/ember, knadh/stuffbin)

Append helper bytes after the Go binary's Mach-O end. Read at runtime via self-inspection.

**Verdict**: Breaks code signing. Equivalent to B1 with extra complexity. Skip.

### B3. In-Memory Execution

Linux has `memfd_create` + `execveat` for executing from memory. **macOS has no equivalent** — SIP prevents executing from anonymous memory. All libraries (go-memexec) fall back to temp files on macOS.

**Verdict**: Not possible on macOS.

---

## C. Pure Go Alternatives (No Swift, No ObjC)

### C1. Direct SQLite Access

Reminders database is at `~/Library/Group Containers/group.com.apple.reminders/Container_v1/Stores/Data-<UUID>.sqlite`. Core Data schema with `ZREMCDREMINDER` table.

**Showstopper**: The group container is TCC-protected. Accessing it requires **Full Disk Access** granted to Terminal — a major security escalation that most users won't do. EventKit properly triggers the narrow "Reminders" TCC prompt instead.

Also: undocumented Core Data schema changes between macOS versions. Writing directly risks Core Data corruption.

**Verdict**: Dead end due to TCC/FDA requirement and schema instability.

### C2. CalDAV Protocol

Apple **dropped CalDAV support for Reminders** in macOS Catalina (2019). Upgraded reminders are invisible to CalDAV clients.

**Verdict**: Complete dead end.

### C3. CloudKit REST API

CloudKit Web Services only allow server-to-server access to **public** databases. Reminders are in the **private** container. No way to access them without a browser-based OAuth flow.

**Verdict**: Dead end for a CLI tool.

### C4. XPC / Mach Ports

Reminders XPC service is private/undocumented. All Go XPC libraries require cgo. Reverse-engineering the service would be extremely fragile.

**Verdict**: Not viable.

---

## Comparison Matrix

| Approach               | Single Binary | Pure Go  | `go install` | `pkg/client/` | Build Complexity | Rewrite Effort | Risk          |
| ---------------------- | ------------- | -------- | ------------ | ------------- | ---------------- | -------------- | ------------- |
| **A1. cgo + ObjC .m**  | ✅            | ❌ (cgo) | ✅           | ✅            | Low              | Medium         | Low           |
| **A2. Swift .a + cgo** | ✅            | ❌ (cgo) | ❌           | ❌            | High             | Low            | High          |
| **A3. purego + ObjC**  | ✅            | ✅       | ✅           | ✅            | Lowest           | High           | Medium        |
| **A4. DarwinKit**      | ✅            | ❌ (cgo) | ❌           | ❌            | High             | High           | High          |
| **B1. go:embed**       | ✅ (cached)   | ❌       | ❌           | ❌            | Low              | None           | Low           |
| **C1. SQLite**         | ✅            | ✅       | ✅           | ✅            | Low              | Medium         | **Very High** |
| **C2-C4**              | —             | —        | —            | —             | —                | —              | Dead ends     |

---

## Recommendations

### Best Overall: A1 (cgo + Objective-C `.m` file)

**Why**: Lowest risk, lowest build complexity, proven pattern, single binary, and `go install` just works. The only cost is rewriting 247 lines of Swift to Objective-C — a straightforward 1:1 translation since EventKit's API is identical in both languages.

This is what [keybase/go-keychain](https://github.com/keybase/go-keychain) and [common-fate/go-apple-security](https://github.com/common-fate/go-apple-security) do for their Apple framework access.

**The cgo requirement is not a real downside** — cgo is enabled by default on macOS, and this project is macOS-only by nature (EventKit is macOS-only).

### Best for Pure Go Purists: A3 (purego + ObjC runtime)

**Why**: The only approach that is truly pure Go (`CGO_ENABLED=0`), requires no Xcode, and makes `go install` / `pkg/client/` work for any macOS user with zero setup. The tradeoff is significant implementation effort (~800-1500 lines of manual ObjC runtime bridging) and no compile-time type safety.

### Pragmatic Short-Term: B1 (go:embed)

If the priority is just "one file to distribute" and the `pkg/client/` problem can wait, embedding the Swift helper is the smallest change. But it doesn't fix the architectural issues — still two binaries under the hood, still subprocess per call, `go install` still broken.

---

## Decision Framework

| Priority                               | Best Choice                       |
| -------------------------------------- | --------------------------------- |
| Ship quickly, single binary            | **B1** (go:embed)                 |
| Clean architecture, `go install` works | **A1** (cgo + ObjC)               |
| Pure Go, no cgo, maximum portability   | **A3** (purego)                   |
| Keep Swift code unchanged              | **A2** (Swift .a) — but high risk |

### What I'd Do

**A1 (cgo + ObjC)** as the next step. It's the pragmatic sweet spot:

- Single binary ✅
- `go install` works ✅
- `pkg/client/` works ✅
- Standard `go build` ✅
- Proven pattern ✅
- Medium effort (ObjC rewrite of 247 lines) ✅

Then **consider A3 (purego)** later if eliminating the cgo requirement becomes important — e.g., if you want `pkg/client/` to work without Xcode Command Line Tools installed.
