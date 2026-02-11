# Public Go API Reference

The `pkg/client` package has been removed. For programmatic Go access to macOS Reminders, use [`go-eventkit`](https://github.com/BRO3886/go-eventkit) directly:

```bash
go get github.com/BRO3886/go-eventkit
```

Import path: `github.com/BRO3886/go-eventkit/reminders`

## Client

```go
client, err := reminders.New()
if err != nil {
    log.Fatal(err) // ErrAccessDenied or ErrUnsupported
}
```

## Types

### Priority

```go
reminders.PriorityNone   // 0
reminders.PriorityHigh   // 1
reminders.PriorityMedium // 5
reminders.PriorityLow    // 9
```

### Reminder

```go
type Reminder struct {
    ID             string
    Title          string
    Notes          string
    List           string      // List display name
    ListID         string
    DueDate        *time.Time
    RemindMeDate   *time.Time
    CompletionDate *time.Time
    CreatedAt      *time.Time
    ModifiedAt     *time.Time
    Priority       Priority
    Completed      bool
    Flagged        bool        // Always false (EventKit limitation)
    URL            string
    Recurring      bool
    RecurrenceRules []eventkit.RecurrenceRule
    HasAlarms      bool
    Alarms         []Alarm
}
```

### List

```go
type List struct {
    ID       string
    Title    string
    Color    string
    Source   string
    Count    int
    ReadOnly bool
}
```

## Reminder Operations

### Create

```go
due := time.Now().Add(24 * time.Hour)
r, err := client.CreateReminder(reminders.CreateReminderInput{
    Title:    "Buy milk",
    ListName: "Shopping",        // Optional, uses default list if empty
    DueDate:  &due,              // Optional
    Priority: reminders.PriorityHigh,
    Notes:    "Whole milk",      // Optional
    URL:      "https://...",     // Optional, native URL field
})
// r.ID is the reminder's unique identifier
```

### Retrieve

```go
r, err := client.Reminder("abc12345")  // Accepts prefix match
```

### List with Filters

```go
items, err := client.Reminders(
    reminders.WithList("Work"),
    reminders.WithCompleted(false),
    reminders.WithDueBefore(deadline),
    reminders.WithSearch("meeting"),
)
```

### Update

```go
newTitle := "Updated title"
r, err := client.UpdateReminder("abc12345", reminders.UpdateReminderInput{
    Title:        &newTitle,     // nil = don't change
    ClearDueDate: true,          // Explicitly clear due date
})
```

### Delete

```go
err := client.DeleteReminder("abc12345")
```

### Complete / Uncomplete

```go
r, err := client.CompleteReminder("abc12345")
r, err := client.UncompleteReminder("abc12345")
```

## List Operations

```go
// Get all lists
lists, err := client.Lists()

// Create a list (Source is required — discover from existing lists)
list, err := client.CreateList(reminders.CreateListInput{
    Title:  "Shopping",
    Source: "iCloud",    // Required
    Color:  "#FF6961",   // Optional
})

// Rename a list
newTitle := "Groceries"
list, err = client.UpdateList(list.ID, reminders.UpdateListInput{
    Title: &newTitle,
})

// Delete a list
err = client.DeleteList(list.ID)
// Returns ErrImmutable for system lists
```

## Notes

- `DueDate` and `RemindMeDate` are independent fields
- `Flagged` is always false — Apple's EventKit doesn't expose this property
- URL is a native field in go-eventkit (no more `URL: ` prefix hack in notes)
- ID prefix matching works the same as the CLI
