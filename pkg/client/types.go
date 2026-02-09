package client

import (
	"time"

	"github.com/BRO3886/rem/internal/reminder"
)

// Priority represents the priority level of a reminder.
type Priority = reminder.Priority

const (
	PriorityNone   = reminder.PriorityNone
	PriorityHigh   = reminder.PriorityHigh
	PriorityMedium = reminder.PriorityMedium
	PriorityLow    = reminder.PriorityLow
)

// Reminder represents a reminder item with all its properties.
type Reminder struct {
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	Notes            string     `json:"notes,omitempty"`
	ListName         string     `json:"list_name"`
	DueDate          *time.Time `json:"due_date,omitempty"`
	RemindMeDate     *time.Time `json:"remind_me_date,omitempty"`
	CompletionDate   *time.Time `json:"completion_date,omitempty"`
	CreationDate     *time.Time `json:"creation_date,omitempty"`
	ModificationDate *time.Time `json:"modification_date,omitempty"`
	Priority         Priority   `json:"priority"`
	Flagged          bool       `json:"flagged"`
	Completed        bool       `json:"completed"`
	URL              string     `json:"url,omitempty"`
}

// List represents a Reminders list.
type List struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
	Count int    `json:"count"`
}

// CreateReminderInput holds the parameters for creating a reminder.
type CreateReminderInput struct {
	Title        string
	Notes        string
	ListName     string
	DueDate      *time.Time
	RemindMeDate *time.Time
	Priority     Priority
	Flagged      bool
	URL          string
}

// UpdateReminderInput holds the parameters for updating a reminder.
type UpdateReminderInput struct {
	Title        *string
	Notes        *string
	DueDate      *time.Time
	ClearDueDate bool
	RemindMeDate *time.Time
	Priority     *Priority
	Flagged      *bool
}

// ListOptions holds filtering options for listing reminders.
type ListOptions struct {
	ListName   string
	Incomplete bool
	Completed  bool
	Flagged    bool
	DueBefore  *time.Time
	DueAfter   *time.Time
	Search     string
}
