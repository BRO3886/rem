---
title: "Go API"
description: "Use rem as a Go library — full CRUD, filtering, list management, and type-safe operations for macOS Reminders."
weight: 4
---

## Overview

rem exposes a public Go API in `pkg/client/` that wraps `go-eventkit` and provides clean, type-safe functions for managing macOS Reminders from your own Go programs.

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
    c, err := client.New()
    if err != nil {
        panic(err)
    }

    // Create a reminder
    due := time.Now().Add(24 * time.Hour)
    id, err := c.CreateReminder(&client.CreateReminderInput{
        Title:    "Deploy to production",
        ListName: "Work",
        DueDate:  &due,
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
        fmt.Printf("%s  %s  (due: %v)\n", r.ID[:8], r.Title, r.DueDate)
    }

    // Mark as done
    c.CompleteReminder(id)
}
```

## Client

All operations start with creating a client:

```go
c, err := client.New()
if err != nil {
    log.Fatal(err)
}
```

The client initializes go-eventkit for EventKit access. `New()` returns an error if Reminders access is denied or the platform is unsupported.

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

`CreateReminder` has a `Context` variant for cancellation and timeout support:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

id, err := c.CreateReminderContext(ctx, input)
```

## Reminder model

The `Reminder` struct returned by read operations:

```go
type Reminder struct {
    ID               string
    Title            string
    Notes            string      // notes/body text
    ListName         string
    DueDate          *time.Time
    RemindMeDate     *time.Time
    CompletionDate   *time.Time
    CreationDate     *time.Time
    ModificationDate *time.Time
    Priority         Priority
    Flagged          bool
    Completed        bool
    URL              string
}
```

## Error handling

All methods return standard Go errors. Common failure modes:

- **Permission denied**: macOS TCC hasn't granted Reminders access (`reminders.ErrAccessDenied`)
- **Not found**: reminder ID doesn't match any reminder (`reminders.ErrNotFound`)
- **Unsupported platform**: running on non-macOS (`reminders.ErrUnsupported`)
