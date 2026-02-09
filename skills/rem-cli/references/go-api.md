# Public Go API Reference

Import path: `github.com/BRO3886/rem/pkg/client`

## Client

```go
c := client.New()
```

All methods return errors. Context-aware variants (e.g. `CreateReminderContext`) are also available.

## Types

### Priority

```go
client.PriorityNone   // 0
client.PriorityHigh   // 1
client.PriorityMedium // 5
client.PriorityLow    // 9
```

### Reminder

```go
type Reminder struct {
    ID               string
    Title            string     // Display name
    Notes            string     // Body text
    ListName         string
    DueDate          *time.Time
    RemindMeDate     *time.Time
    CompletionDate   *time.Time
    CreationDate     *time.Time
    ModificationDate *time.Time
    Priority         Priority
    Flagged          bool
    Completed        bool
    URL              string     // Extracted from Notes if present
}
```

### List

```go
type List struct {
    ID    string
    Name  string
    Color string
    Count int
}
```

## Reminder Operations

### Create

```go
id, err := c.CreateReminder(&client.CreateReminderInput{
    Title:    "Buy milk",
    ListName: "Shopping",        // Optional, uses default list if empty
    DueDate:  &dueTime,          // Optional
    Priority: client.PriorityHigh,
    Notes:    "Whole milk",      // Optional
    URL:      "https://...",     // Optional, stored in Notes with "URL: " prefix
    Flagged:  true,              // Optional
})
// id is the full x-apple-reminder://UUID string
```

### Retrieve

```go
reminder, err := c.GetReminder("abc12345")  // Accepts prefix match
```

### List with Filters

```go
reminders, err := c.ListReminders(&client.ListOptions{
    ListName:   "Work",       // Optional
    Incomplete: true,         // Optional
    Completed:  false,        // Optional
    Flagged:    false,        // Optional
    DueBefore:  &beforeTime,  // Optional
    DueAfter:   &afterTime,   // Optional
    Search:     "meeting",    // Optional
})
```

### Update

```go
newTitle := "Updated title"
newPriority := client.PriorityMedium

err := c.UpdateReminder("abc12345", &client.UpdateReminderInput{
    Title:        &newTitle,     // nil = don't change
    Priority:     &newPriority,  // nil = don't change
    ClearDueDate: true,          // Explicitly clear due date
})
```

### Delete

```go
err := c.DeleteReminder("abc12345")
```

### Complete / Uncomplete

```go
err := c.CompleteReminder("abc12345")
err := c.UncompleteReminder("abc12345")
```

### Flag / Unflag

```go
err := c.FlagReminder("abc12345")
err := c.UnflagReminder("abc12345")
```

## List Operations

```go
lists, err := c.GetLists()
list, err := c.CreateList("Shopping")
err := c.RenameList("Old Name", "New Name")
err := c.DeleteList("Shopping")
name, err := c.DefaultListName()
```

## Notes

- `DueDate` and `RemindMeDate` are independent fields
- URL is stored in the Notes/body field with a `URL: ` prefix (no native URL property)
- ID prefix matching works the same as the CLI
- The `Title`/`Notes` fields in the public API map to `Name`/`Body` internally
