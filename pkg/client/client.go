//go:build darwin

// Package client provides a public Go API for interacting with macOS Reminders.
//
// This package uses go-eventkit for fast, native EventKit access and provides
// a clean, idiomatic Go interface for CRUD operations on reminders and lists.
//
// Usage:
//
//	c, err := client.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a reminder
//	id, err := c.CreateReminder(&client.CreateReminderInput{
//	    Title:    "Buy groceries",
//	    ListName: "Personal",
//	    DueDate:  time.Now().Add(24 * time.Hour),
//	    Priority: client.PriorityHigh,
//	})
//
//	// List reminders
//	reminders, err := c.ListReminders(&client.ListOptions{
//	    ListName:   "Personal",
//	    Incomplete: true,
//	})
//
//	// Complete a reminder
//	err = c.CompleteReminder(id)
package client

import (
	"context"

	"github.com/BRO3886/go-eventkit/reminders"
	"github.com/BRO3886/rem/internal/service"
	"github.com/BRO3886/rem/internal/reminder"
)

// Client provides methods for interacting with macOS Reminders.
type Client struct {
	reminderSvc *service.ReminderService
	listSvc     *service.ListService
}

// New creates a new Reminders client.
// Returns an error if Reminders access is denied or unavailable.
func New() (*Client, error) {
	ekClient, err := reminders.New()
	if err != nil {
		return nil, err
	}
	exec := service.NewExecutor()
	return &Client{
		reminderSvc: service.NewReminderService(ekClient, exec),
		listSvc:     service.NewListService(ekClient, exec),
	}, nil
}

// CreateReminder creates a new reminder and returns its ID.
func (c *Client) CreateReminder(input *CreateReminderInput) (string, error) {
	return c.CreateReminderContext(context.Background(), input)
}

// CreateReminderContext creates a new reminder with context support.
func (c *Client) CreateReminderContext(_ context.Context, input *CreateReminderInput) (string, error) {
	r := &reminder.Reminder{
		Name:         input.Title,
		Body:         input.Notes,
		ListName:     input.ListName,
		DueDate:      input.DueDate,
		RemindMeDate: input.RemindMeDate,
		Priority:     input.Priority,
		Flagged:      input.Flagged,
		URL:          input.URL,
	}
	return c.reminderSvc.CreateReminder(r)
}

// GetReminder retrieves a reminder by its ID.
func (c *Client) GetReminder(id string) (*Reminder, error) {
	r, err := c.reminderSvc.GetReminder(id)
	if err != nil {
		return nil, err
	}
	return toPublicReminder(r), nil
}

// ListReminders returns reminders matching the given options.
func (c *Client) ListReminders(opts *ListOptions) ([]*Reminder, error) {
	filter := &reminder.ListFilter{}
	if opts != nil {
		filter.ListName = opts.ListName
		filter.SearchQuery = opts.Search
		filter.DueBefore = opts.DueBefore
		filter.DueAfter = opts.DueAfter
		if opts.Incomplete {
			v := false
			filter.Completed = &v
		}
		if opts.Completed {
			v := true
			filter.Completed = &v
		}
		if opts.Flagged {
			v := true
			filter.Flagged = &v
		}
	}

	items, err := c.reminderSvc.ListReminders(filter)
	if err != nil {
		return nil, err
	}

	result := make([]*Reminder, 0, len(items))
	for _, r := range items {
		result = append(result, toPublicReminder(r))
	}
	return result, nil
}

// UpdateReminder updates an existing reminder.
func (c *Client) UpdateReminder(id string, input *UpdateReminderInput) error {
	updates := make(map[string]any)

	if input.Title != nil {
		updates["name"] = *input.Title
	}
	if input.Notes != nil {
		updates["body"] = *input.Notes
	}
	if input.ClearDueDate {
		updates["due_date"] = nil
	} else if input.DueDate != nil {
		updates["due_date"] = *input.DueDate
	}
	if input.RemindMeDate != nil {
		updates["remind_me_date"] = *input.RemindMeDate
	}
	if input.Priority != nil {
		updates["priority"] = *input.Priority
	}
	if input.Flagged != nil {
		updates["flagged"] = *input.Flagged
	}

	return c.reminderSvc.UpdateReminder(id, updates)
}

// DeleteReminder deletes a reminder by its ID.
func (c *Client) DeleteReminder(id string) error {
	return c.reminderSvc.DeleteReminder(id)
}

// CompleteReminder marks a reminder as completed.
func (c *Client) CompleteReminder(id string) error {
	return c.reminderSvc.CompleteReminder(id)
}

// UncompleteReminder marks a reminder as incomplete.
func (c *Client) UncompleteReminder(id string) error {
	return c.reminderSvc.UncompleteReminder(id)
}

// FlagReminder flags a reminder.
func (c *Client) FlagReminder(id string) error {
	return c.reminderSvc.FlagReminder(id)
}

// UnflagReminder removes the flag from a reminder.
func (c *Client) UnflagReminder(id string) error {
	return c.reminderSvc.UnflagReminder(id)
}

// GetLists returns all reminder lists.
func (c *Client) GetLists() ([]*List, error) {
	lists, err := c.listSvc.GetLists()
	if err != nil {
		return nil, err
	}
	result := make([]*List, 0, len(lists))
	for _, l := range lists {
		result = append(result, toPublicList(l))
	}
	return result, nil
}

// CreateList creates a new reminder list.
func (c *Client) CreateList(name string) (*List, error) {
	l, err := c.listSvc.CreateList(name)
	if err != nil {
		return nil, err
	}
	return toPublicList(l), nil
}

// RenameList renames a reminder list.
func (c *Client) RenameList(oldName, newName string) error {
	return c.listSvc.RenameList(oldName, newName)
}

// DeleteList deletes a reminder list.
func (c *Client) DeleteList(name string) error {
	return c.listSvc.DeleteList(name)
}

// DefaultListName returns the name of the default reminder list.
func (c *Client) DefaultListName() (string, error) {
	return c.listSvc.GetDefaultListName()
}

func toPublicReminder(r *reminder.Reminder) *Reminder {
	return &Reminder{
		ID:               r.ID,
		Title:            r.Name,
		Notes:            r.Body,
		ListName:         r.ListName,
		DueDate:          r.DueDate,
		RemindMeDate:     r.RemindMeDate,
		CompletionDate:   r.CompletionDate,
		CreationDate:     r.CreationDate,
		ModificationDate: r.ModificationDate,
		Priority:         r.Priority,
		Flagged:          r.Flagged,
		Completed:        r.Completed,
		URL:              r.URL,
	}
}

func toPublicList(l *reminder.List) *List {
	return &List{
		ID:    l.ID,
		Name:  l.Name,
		Color: l.Color,
		Count: l.Count,
	}
}
