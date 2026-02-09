---
title: "Go API"
description: "Use rem as a Go library — full CRUD, filtering, list management, and type-safe operations for macOS Reminders."
weight: 4
---

## Overview

rem exposes a public Go API in `pkg/client/` that wraps all the complexity of EventKit and AppleScript into clean, type-safe functions. Import it to manage macOS Reminders from your own Go programs.

```bash
go get github.com/BRO3886/rem
```

## Quick example

```go
package main

import (
    "fmt"
    "time"

    "github.com/BRO3886/rem/pkg/client"
)

func main() {
    c := client.New()

    // Create a reminder
    id, err := c.CreateReminder(&client.CreateReminderInput{
        Title:    "Deploy to production",
        ListName: "Work",
        DueDate:  time.Now().Add(24 * time.Hour),
        Priority: client.PriorityHigh,
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("Created:", id)

    // List incomplete reminders
    reminders, err := c.ListReminders(&client.ListOptions{
        ListName:   "Work",
        Incomplete: true,
    })
    if err != nil {
        panic(err)
    }
    for _, r := range reminders {
        fmt.Printf("%s  %s  (due: %s)\n", r.ID[:8], r.Name, r.DueDate)
    }

    // Mark as done
    c.CompleteReminder(id)
}
```

## Client

All operations start with creating a client:

```go
c := client.New()
```

The client manages the underlying EventKit store (for reads) and AppleScript executor (for writes).

## Creating reminders

```go
id, err := c.CreateReminder(&client.CreateReminderInput{
    Title:    "Buy groceries",
    ListName: "Personal",          // optional, uses default list
    DueDate:  time.Now().Add(48 * time.Hour),
    Priority: client.PriorityMedium,
    Notes:    "Milk, eggs, bread",
    URL:      "https://example.com",
    Flagged:  true,
})
```

All fields except `Title` are optional.

### Priority values

| Constant | Value | Meaning |
|----------|-------|---------|
| `client.PriorityNone` | 0 | No priority |
| `client.PriorityHigh` | 1 | High priority (1-4) |
| `client.PriorityMedium` | 5 | Medium priority |
| `client.PriorityLow` | 9 | Low priority (6-9) |

## Reading reminders

### Get a single reminder

```go
r, err := c.GetReminder("6ECEA745")  // full ID, UUID, or prefix
```

Supports full IDs, UUIDs, and prefix matching (case-insensitive).

### List with filters

```go
reminders, err := c.ListReminders(&client.ListOptions{
    ListName:  "Work",
    Incomplete: true,
    DueBefore: time.Now().Add(7 * 24 * time.Hour),
    Search:    "deploy",
})
```

All filter fields are optional. Pass `nil` to get all reminders.

| Field | Type | Description |
|-------|------|-------------|
| `ListName` | `string` | Filter by list |
| `Incomplete` | `bool` | Only incomplete |
| `Completed` | `bool` | Only completed |
| `Flagged` | `bool` | Only flagged |
| `DueBefore` | `time.Time` | Due before date |
| `DueAfter` | `time.Time` | Due after date |
| `Search` | `string` | Full-text search |

## Updating reminders

```go
err := c.UpdateReminder(id, &client.UpdateReminderInput{
    Priority: &client.PriorityMedium,
    Notes:    stringPtr("Updated notes"),
    DueDate:  &newDate,
})
```

Only specified fields are updated. Use pointers to distinguish between "not set" and "set to zero value".

## Status operations

```go
c.CompleteReminder(id)
c.UncompleteReminder(id)
c.FlagReminder(id)
c.UnflagReminder(id)
c.DeleteReminder(id)
```

## List management

```go
// Get all lists
lists, err := c.GetLists()

// Get default list name
name, err := c.DefaultListName()

// CRUD
c.CreateList("Projects")
c.RenameList("Projects", "Active Projects")
c.DeleteList("Old List")
```

## Context-aware variants

Every method has a `Context` variant for cancellation and timeout support:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

reminders, err := c.ListRemindersContext(ctx, opts)
```

## Reminder model

The `Reminder` struct returned by read operations:

```go
type Reminder struct {
    ID               string
    Name             string
    Body             string     // notes/body text
    ListName         string
    DueDate          time.Time
    RemindMeDate     time.Time
    CompletionDate   time.Time
    CreationDate     time.Time
    ModificationDate time.Time
    Priority         Priority
    Flagged          bool
    Completed        bool
    URL              string
}
```

## Error handling

All methods return standard Go errors. Common failure modes:

- **Permission denied**: macOS TCC hasn't granted Reminders access
- **Not found**: reminder ID doesn't match any reminder
- **AppleScript timeout**: write operation exceeded 30s (very rare)
