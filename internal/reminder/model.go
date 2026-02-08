package reminder

import "time"

// Priority represents the priority level of a reminder.
type Priority int

const (
	PriorityNone   Priority = 0
	PriorityHigh   Priority = 1
	PriorityMedium Priority = 5
	PriorityLow    Priority = 9
)

func (p Priority) String() string {
	switch {
	case p == PriorityNone:
		return "none"
	case p >= 1 && p <= 4:
		return "high"
	case p == 5:
		return "medium"
	case p >= 6 && p <= 9:
		return "low"
	default:
		return "none"
	}
}

// ParsePriority converts a string to a Priority value.
func ParsePriority(s string) Priority {
	switch s {
	case "high", "h", "1":
		return PriorityHigh
	case "medium", "med", "m", "5":
		return PriorityMedium
	case "low", "l", "9":
		return PriorityLow
	default:
		return PriorityNone
	}
}

// Reminder represents a single reminder item.
type Reminder struct {
	ID               string
	Name             string
	Body             string
	ListName         string
	DueDate          *time.Time
	AllDayDueDate    *time.Time
	RemindMeDate     *time.Time
	CompletionDate   *time.Time
	CreationDate     *time.Time
	ModificationDate *time.Time
	Priority         Priority
	Flagged          bool
	Completed        bool
	URL              string // stored in body, extracted for convenience
}

// List represents a Reminders list.
type List struct {
	ID    string
	Name  string
	Color string
	Count int // number of reminders in the list
}

// ListFilter specifies criteria for filtering reminders when listing.
type ListFilter struct {
	ListName     string
	Completed    *bool
	Flagged      *bool
	DueBefore    *time.Time
	DueAfter     *time.Time
	SearchQuery  string
	PriorityMin  *Priority
}
