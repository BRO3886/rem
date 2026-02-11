//go:build darwin

package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit/reminders"
	"github.com/BRO3886/rem/internal/reminder"
)

// ReminderService provides operations for reminders using go-eventkit for
// reads and writes, with AppleScript fallback for flagged operations.
type ReminderService struct {
	client *reminders.Client
	exec   *Executor
}

// NewReminderService creates a new ReminderService.
func NewReminderService(client *reminders.Client, exec *Executor) *ReminderService {
	return &ReminderService{client: client, exec: exec}
}

// CreateReminder creates a new reminder and returns its ID.
func (s *ReminderService) CreateReminder(r *reminder.Reminder) (string, error) {
	if r.Name == "" {
		return "", fmt.Errorf("reminder name is required")
	}

	input := reminders.CreateReminderInput{
		Title:    r.Name,
		Notes:    r.Body,
		ListName: r.ListName,
		DueDate:  r.DueDate,
		Priority: reminders.Priority(r.Priority),
	}

	if r.RemindMeDate != nil {
		input.RemindMeDate = r.RemindMeDate
	}

	if r.URL != "" {
		input.URL = r.URL
	}

	created, err := s.client.CreateReminder(input)
	if err != nil {
		return "", fmt.Errorf("failed to create reminder: %w", err)
	}

	// Set flagged via AppleScript if needed (EventKit can't set flagged)
	if r.Flagged {
		_ = s.FlagReminder(created.ID)
	}

	return created.ID, nil
}

// GetReminder retrieves a single reminder by ID or ID prefix.
func (s *ReminderService) GetReminder(id string) (*reminder.Reminder, error) {
	r, err := s.client.Reminder(id)
	if err != nil {
		return nil, fmt.Errorf("reminder not found: %s", id)
	}
	return fromEventKitReminder(r), nil
}

// ListReminders returns reminders matching the given filter.
func (s *ReminderService) ListReminders(filter *reminder.ListFilter) ([]*reminder.Reminder, error) {
	var opts []reminders.ListOption

	if filter != nil {
		if filter.ListName != "" {
			opts = append(opts, reminders.WithList(filter.ListName))
		}
		if filter.Completed != nil {
			opts = append(opts, reminders.WithCompleted(*filter.Completed))
		}
		if filter.SearchQuery != "" {
			opts = append(opts, reminders.WithSearch(filter.SearchQuery))
		}
		if filter.DueBefore != nil {
			opts = append(opts, reminders.WithDueBefore(*filter.DueBefore))
		}
		if filter.DueAfter != nil {
			opts = append(opts, reminders.WithDueAfter(*filter.DueAfter))
		}
	}

	ekReminders, err := s.client.Reminders(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}

	// Apply flagged filter — EventKit doesn't expose flagged, so we need
	// JXA fallback when --flagged is active.
	needsFlagged := filter != nil && filter.Flagged != nil && *filter.Flagged

	var flaggedIDs map[string]bool
	if needsFlagged {
		flaggedIDs, err = s.fetchFlaggedIDs()
		if err != nil {
			return nil, err
		}
	}

	result := make([]*reminder.Reminder, 0, len(ekReminders))
	for i := range ekReminders {
		r := fromEventKitReminder(&ekReminders[i])

		if needsFlagged {
			r.Flagged = flaggedIDs[r.ID]
			if !r.Flagged {
				continue
			}
		}

		result = append(result, r)
	}

	return result, nil
}

// fetchFlaggedIDs uses JXA to get the set of flagged reminder IDs.
func (s *ReminderService) fetchFlaggedIDs() (map[string]bool, error) {
	script := `
const app = Application('Reminders');
const lists = app.lists();
const result = [];
for (const list of lists) {
	const n = list.reminders.length;
	if (n === 0) continue;
	const ids = list.reminders.id();
	const flagged = list.reminders.flagged();
	for (let i = 0; i < n; i++) {
		if (flagged[i]) result.push(ids[i]);
	}
}
JSON.stringify(result);`

	output, err := s.exec.RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch flagged status: %w", err)
	}

	var ids []string
	if output != "" && output != "[]" {
		if err := json.Unmarshal([]byte(output), &ids); err != nil {
			return nil, fmt.Errorf("failed to parse flagged IDs: %w", err)
		}
	}

	m := make(map[string]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m, nil
}

// UpdateReminder updates properties of an existing reminder.
func (s *ReminderService) UpdateReminder(id string, updates map[string]any) error {
	input := reminders.UpdateReminderInput{}
	var needsAppleScript bool
	var appleScriptUpdates map[string]any

	for key, value := range updates {
		switch key {
		case "name":
			v := value.(string)
			input.Title = &v
		case "body":
			v := value.(string)
			input.Notes = &v
		case "due_date":
			if value == nil {
				input.ClearDueDate = true
			} else {
				t := value.(time.Time)
				input.DueDate = &t
			}
		case "remind_me_date":
			if value == nil {
				// Clear remind me date by setting to zero time
				t := time.Time{}
				input.RemindMeDate = &t
			} else {
				t := value.(time.Time)
				input.RemindMeDate = &t
			}
		case "priority":
			p := reminders.Priority(value.(reminder.Priority))
			input.Priority = &p
		case "flagged":
			// EventKit can't set flagged — use AppleScript
			needsAppleScript = true
			if appleScriptUpdates == nil {
				appleScriptUpdates = make(map[string]any)
			}
			appleScriptUpdates["flagged"] = value
		case "completed":
			v := value.(bool)
			input.Completed = &v
		case "url":
			v := value.(string)
			input.URL = &v
		}
	}

	// Apply EventKit updates (all fields except flagged)
	hasEventKitUpdates := input.Title != nil || input.Notes != nil ||
		input.DueDate != nil || input.ClearDueDate ||
		input.RemindMeDate != nil || input.Priority != nil ||
		input.Completed != nil || input.URL != nil

	if hasEventKitUpdates {
		if _, err := s.client.UpdateReminder(id, input); err != nil {
			return fmt.Errorf("failed to update reminder: %w", err)
		}
	}

	// Apply AppleScript updates (flagged)
	if needsAppleScript {
		if err := s.updateViaAppleScript(id, appleScriptUpdates); err != nil {
			return fmt.Errorf("failed to update reminder flags: %w", err)
		}
	}

	return nil
}

// updateViaAppleScript updates reminder properties that EventKit doesn't support.
func (s *ReminderService) updateViaAppleScript(id string, updates map[string]any) error {
	var setStatements []string

	for key, value := range updates {
		switch key {
		case "flagged":
			if value.(bool) {
				setStatements = append(setStatements, `set flagged of r to true`)
			} else {
				setStatements = append(setStatements, `set flagged of r to false`)
			}
		}
	}

	if len(setStatements) == 0 {
		return nil
	}

	script := fmt.Sprintf(`tell application "Reminders"
	set r to first reminder whose id is "%s"
	%s
end tell`, EscapeString(id), strings.Join(setStatements, "\n\t"))

	_, err := s.exec.Run(script)
	return err
}

// DeleteReminder deletes a reminder by ID.
func (s *ReminderService) DeleteReminder(id string) error {
	if err := s.client.DeleteReminder(id); err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}
	return nil
}

// CompleteReminder marks a reminder as completed.
func (s *ReminderService) CompleteReminder(id string) error {
	if _, err := s.client.CompleteReminder(id); err != nil {
		return fmt.Errorf("failed to complete reminder: %w", err)
	}
	return nil
}

// UncompleteReminder marks a reminder as incomplete.
func (s *ReminderService) UncompleteReminder(id string) error {
	if _, err := s.client.UncompleteReminder(id); err != nil {
		return fmt.Errorf("failed to uncomplete reminder: %w", err)
	}
	return nil
}

// FlagReminder flags a reminder via AppleScript (EventKit doesn't support flagged).
func (s *ReminderService) FlagReminder(id string) error {
	return s.updateViaAppleScript(id, map[string]any{"flagged": true})
}

// UnflagReminder removes the flag from a reminder via AppleScript.
func (s *ReminderService) UnflagReminder(id string) error {
	return s.updateViaAppleScript(id, map[string]any{"flagged": false})
}

// fromEventKitReminder converts a go-eventkit Reminder to an internal Reminder.
func fromEventKitReminder(r *reminders.Reminder) *reminder.Reminder {
	result := &reminder.Reminder{
		ID:               r.ID,
		Name:             r.Title,
		Body:             r.Notes,
		ListName:         r.List,
		DueDate:          r.DueDate,
		RemindMeDate:     r.RemindMeDate,
		CompletionDate:   r.CompletionDate,
		CreationDate:     r.CreatedAt,
		ModificationDate: r.ModifiedAt,
		Priority:         reminder.Priority(r.Priority),
		Completed:        r.Completed,
		Flagged:          r.Flagged,
		URL:              r.URL,
	}

	// For backwards compatibility: if URL is empty but notes contain a URL, extract it
	if result.URL == "" && result.Body != "" {
		result.URL = extractURL(result.Body)
	}

	return result
}
