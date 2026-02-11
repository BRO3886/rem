---
title: "Go API"
description: "Use go-eventkit for programmatic access to macOS Reminders — the same library that powers rem and cal."
weight: 4
---

## Overview

For programmatic Go access to macOS Reminders, use [**go-eventkit**](https://github.com/BRO3886/go-eventkit) directly — the same library that powers rem and [cal](https://github.com/BRO3886/cal).

```bash
go get github.com/BRO3886/go-eventkit
```

## Quick example

```go
package main

import (
    "fmt"
    "time"

    "github.com/BRO3886/go-eventkit/reminders"
)

func main() {
    client, err := reminders.New()
    if err != nil {
        panic(err)
    }

    // Create a reminder
    due := time.Now().Add(24 * time.Hour)
    r, err := client.CreateReminder(reminders.CreateReminderInput{
        Title:    "Deploy to production",
        ListName: "Work",
        DueDate:  &due,
        Priority: reminders.PriorityHigh,
    })
    if err != nil {
        panic(err)
    }
    fmt.Println("Created:", r.ID)

    // List incomplete reminders
    items, err := client.Reminders(
        reminders.WithList("Work"),
        reminders.WithCompleted(false),
    )
    if err != nil {
        panic(err)
    }
    for _, item := range items {
        fmt.Printf("%s  %s  (due: %v)\n", item.ID[:8], item.Title, item.DueDate)
    }

    // Mark as done
    client.CompleteReminder(r.ID)
}
```

## Client

All operations start with creating a client:

```go
client, err := reminders.New()
if err != nil {
    log.Fatal(err)
}
```

`New()` requests TCC (Transparency, Consent, and Control) access to Reminders. Returns an error if access is denied or the platform is unsupported.

## Creating reminders

```go
due := time.Now().Add(48 * time.Hour)
r, err := client.CreateReminder(reminders.CreateReminderInput{
    Title:    "Buy groceries",
    ListName: "Personal",          // optional, uses default list
    DueDate:  &due,
    Priority: reminders.PriorityMedium,
    Notes:    "Milk, eggs, bread",
    URL:      "https://example.com",
})
```

All fields except `Title` are optional.

### Priority values

| Constant | Value | Meaning |
|----------|-------|---------|
| `reminders.PriorityNone` | 0 | No priority |
| `reminders.PriorityHigh` | 1 | High priority (1-4) |
| `reminders.PriorityMedium` | 5 | Medium priority |
| `reminders.PriorityLow` | 9 | Low priority (6-9) |

## Reading reminders

### Get a single reminder

```go
r, err := client.Reminder("6ECEA745")  // full ID, UUID, or prefix
```

Supports full IDs, UUIDs, and prefix matching (case-insensitive).

### List with filters

```go
items, err := client.Reminders(
    reminders.WithList("Work"),
    reminders.WithCompleted(false),
    reminders.WithDueBefore(deadline),
    reminders.WithSearch("deploy"),
)
```

All filter options are optional. Call with no options to get all reminders.

| Option | Description |
|--------|-------------|
| `WithList(name)` | Filter by list name |
| `WithListID(id)` | Filter by list ID |
| `WithCompleted(bool)` | Filter by completion status |
| `WithDueBefore(time)` | Due before date |
| `WithDueAfter(time)` | Due after date |
| `WithSearch(query)` | Full-text search |

## Updating reminders

```go
newTitle := "Updated title"
newPriority := reminders.PriorityMedium
r, err := client.UpdateReminder(id, reminders.UpdateReminderInput{
    Title:    &newTitle,       // nil = don't change
    Priority: &newPriority,   // nil = don't change
    ClearDueDate: true,       // explicitly clear due date
})
```

Only non-nil pointer fields are modified.

## Status operations

```go
r, err := client.CompleteReminder(id)
r, err := client.UncompleteReminder(id)
err := client.DeleteReminder(id)
```

## List management

```go
// Get all lists
lists, err := client.Lists()
for _, l := range lists {
    fmt.Printf("%s (%d reminders, source: %s)\n", l.Title, l.Count, l.Source)
}

// Create a list
list, err := client.CreateList(reminders.CreateListInput{
    Title:  "Shopping",
    Source: "iCloud",    // required — use Lists() to discover sources
    Color:  "#FF6961",   // optional
})

// Rename a list
newTitle := "Groceries"
list, err = client.UpdateList(list.ID, reminders.UpdateListInput{
    Title: &newTitle,
})

// Delete a list
err = client.DeleteList(list.ID)
```

Immutable system lists (e.g., Siri suggestions) return `reminders.ErrImmutable` on update/delete.

## Reminder model

The `Reminder` struct returned by read operations:

```go
type Reminder struct {
    ID             string
    Title          string
    Notes          string
    List           string      // list display name
    ListID         string
    DueDate        *time.Time
    RemindMeDate   *time.Time
    CompletionDate *time.Time
    CreatedAt      *time.Time
    ModifiedAt     *time.Time
    Priority       Priority
    Completed      bool
    Flagged        bool        // always false (EventKit limitation)
    URL            string      // native URL field
    Recurring      bool
    HasAlarms      bool
}
```

## Error handling

All methods return standard Go errors. Use `errors.Is` for sentinel errors:

```go
if errors.Is(err, reminders.ErrAccessDenied) {
    // macOS TCC hasn't granted Reminders access
}
if errors.Is(err, reminders.ErrNotFound) {
    // reminder ID doesn't match any reminder
}
if errors.Is(err, reminders.ErrUnsupported) {
    // running on non-macOS
}
if errors.Is(err, reminders.ErrImmutable) {
    // list is immutable (system list)
}
```

## Learn more

See the [go-eventkit README](https://github.com/BRO3886/go-eventkit) for the full API reference, including calendar/events support.
