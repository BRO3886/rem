//go:build darwin

package applescript

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/rem/internal/eventkit"
	"github.com/BRO3886/rem/internal/reminder"
)

// ReminderService provides operations for reminders using EventKit for reads
// and AppleScript for writes.
type ReminderService struct {
	exec *Executor
}

// NewReminderService creates a new ReminderService.
func NewReminderService(exec *Executor) *ReminderService {
	return &ReminderService{exec: exec}
}

// CreateReminder creates a new reminder and returns its ID.
func (s *ReminderService) CreateReminder(r *reminder.Reminder) (string, error) {
	if r.Name == "" {
		return "", fmt.Errorf("reminder name is required")
	}

	listName := r.ListName
	if listName == "" {
		defaultList, err := s.exec.Run(`tell application "Reminders" to get name of default list`)
		if err != nil {
			return "", fmt.Errorf("failed to get default list: %w", err)
		}
		listName = defaultList
	}

	var dateSetup strings.Builder
	props := []string{fmt.Sprintf(`name:"%s"`, EscapeString(r.Name))}

	body := r.Body
	if r.URL != "" {
		if body != "" {
			body = body + "\n\nURL: " + r.URL
		} else {
			body = "URL: " + r.URL
		}
	}
	if body != "" {
		props = append(props, fmt.Sprintf(`body:"%s"`, EscapeString(body)))
	}

	if r.DueDate != nil {
		dateSetup.WriteString(FormatDateInline("dueDate", *r.DueDate) + "\n")
		props = append(props, "due date:dueDate")
	}

	if r.AllDayDueDate != nil {
		dateSetup.WriteString(FormatDateInline("alldayDate", *r.AllDayDueDate) + "\n")
		props = append(props, "allday due date:alldayDate")
	}

	if r.RemindMeDate != nil {
		dateSetup.WriteString(FormatDateInline("remindDate", *r.RemindMeDate) + "\n")
		props = append(props, "remind me date:remindDate")
	}

	if r.Priority != reminder.PriorityNone {
		props = append(props, fmt.Sprintf("priority:%d", r.Priority))
	}

	if r.Flagged {
		props = append(props, "flagged:true")
	}

	script := fmt.Sprintf(`%s
tell application "Reminders"
	tell list "%s"
		set newReminder to make new reminder at end with properties {%s}
		return id of newReminder
	end tell
end tell`, dateSetup.String(), EscapeString(listName), strings.Join(props, ", "))

	id, err := s.exec.Run(script)
	if err != nil {
		return "", fmt.Errorf("failed to create reminder: %w", err)
	}

	return id, nil
}

// helperReminder is the JSON structure returned by the EventKit bridge.
type helperReminder struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Body           *string `json:"body"`
	ListName       string  `json:"listName"`
	Completed      bool    `json:"completed"`
	Flagged        bool    `json:"flagged"`
	Priority       int     `json:"priority"`
	DueDate        *string `json:"dueDate"`
	RemindMeDate   *string `json:"remindMeDate"`
	CompletionDate *string `json:"completionDate"`
	CreationDate   *string `json:"creationDate"`
	ModDate        *string `json:"modDate"`
}

func helperToReminder(h *helperReminder) *reminder.Reminder {
	r := &reminder.Reminder{
		ID:        h.ID,
		Name:      h.Name,
		ListName:  h.ListName,
		Completed: h.Completed,
		Flagged:   h.Flagged,
		Priority:  reminder.Priority(h.Priority),
	}

	if h.Body != nil {
		r.Body = *h.Body
	}

	if h.DueDate != nil {
		r.DueDate = parseISODate(*h.DueDate)
	}
	if h.RemindMeDate != nil {
		r.RemindMeDate = parseISODate(*h.RemindMeDate)
	}
	if h.CompletionDate != nil {
		r.CompletionDate = parseISODate(*h.CompletionDate)
	}
	if h.CreationDate != nil {
		r.CreationDate = parseISODate(*h.CreationDate)
	}
	if h.ModDate != nil {
		r.ModificationDate = parseISODate(*h.ModDate)
	}

	r.URL = extractURL(r.Body)
	return r
}

func parseISODate(s string) *time.Time {
	if s == "" || s == "null" {
		return nil
	}
	// Try RFC3339 with fractional seconds first (ISO8601 output)
	t, err := time.Parse("2006-01-02T15:04:05.000Z", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return nil
		}
	}
	local := t.Local()
	return &local
}

// GetReminder retrieves a single reminder by ID or ID prefix via EventKit.
func (s *ReminderService) GetReminder(id string) (*reminder.Reminder, error) {
	output, err := eventkit.GetReminder(id)
	if err != nil {
		return nil, fmt.Errorf("reminder not found: %s", id)
	}

	if output == "" || output == "null" {
		return nil, fmt.Errorf("reminder not found: %s", id)
	}

	var hr helperReminder
	if err := json.Unmarshal([]byte(output), &hr); err != nil {
		return nil, fmt.Errorf("failed to parse reminder: %w", err)
	}

	return helperToReminder(&hr), nil
}

// ListReminders returns reminders matching the given filter via EventKit.
func (s *ReminderService) ListReminders(filter *reminder.ListFilter) ([]*reminder.Reminder, error) {
	var listName, completed, search, dueBefore, dueAfter string

	if filter != nil {
		listName = filter.ListName
		if filter.Completed != nil {
			if *filter.Completed {
				completed = "true"
			} else {
				completed = "false"
			}
		}
		search = filter.SearchQuery
		if filter.DueBefore != nil {
			dueBefore = filter.DueBefore.UTC().Format("2006-01-02T15:04:05.000Z")
		}
		if filter.DueAfter != nil {
			dueAfter = filter.DueAfter.UTC().Format("2006-01-02T15:04:05.000Z")
		}
	}

	output, err := eventkit.FetchReminders(listName, completed, search, dueBefore, dueAfter)
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}

	if output == "" || output == "[]" {
		return []*reminder.Reminder{}, nil
	}

	var hrs []helperReminder
	if err := json.Unmarshal([]byte(output), &hrs); err != nil {
		return nil, fmt.Errorf("failed to parse reminders: %w", err)
	}

	// Apply flagged filter in Go since EventKit doesn't expose flagged.
	// For --flagged filter, we need JXA to get the actual flagged status.
	needsFlagged := filter != nil && filter.Flagged != nil && *filter.Flagged

	var flaggedIDs map[string]bool
	if needsFlagged {
		flaggedIDs, err = s.fetchFlaggedIDs()
		if err != nil {
			return nil, err
		}
	}

	reminders := make([]*reminder.Reminder, 0, len(hrs))
	for i := range hrs {
		r := helperToReminder(&hrs[i])

		if needsFlagged {
			r.Flagged = flaggedIDs[r.ID]
			if !r.Flagged {
				continue
			}
		}

		reminders = append(reminders, r)
	}

	return reminders, nil
}

// fetchFlaggedIDs uses JXA to get the set of flagged reminder IDs.
// This is needed because EventKit doesn't expose the flagged property.
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
	var dateSetup strings.Builder
	var setStatements []string

	for key, value := range updates {
		switch key {
		case "name":
			setStatements = append(setStatements, fmt.Sprintf(`set name of r to "%s"`, EscapeString(value.(string))))
		case "body":
			setStatements = append(setStatements, fmt.Sprintf(`set body of r to "%s"`, EscapeString(value.(string))))
		case "due_date":
			if value == nil {
				setStatements = append(setStatements, `set due date of r to missing value`)
			} else {
				t := value.(time.Time)
				dateSetup.WriteString(FormatDateInline("newDueDate", t) + "\n")
				setStatements = append(setStatements, `set due date of r to newDueDate`)
			}
		case "remind_me_date":
			if value == nil {
				setStatements = append(setStatements, `set remind me date of r to missing value`)
			} else {
				t := value.(time.Time)
				dateSetup.WriteString(FormatDateInline("newRemindDate", t) + "\n")
				setStatements = append(setStatements, `set remind me date of r to newRemindDate`)
			}
		case "priority":
			setStatements = append(setStatements, fmt.Sprintf(`set priority of r to %d`, value.(reminder.Priority)))
		case "flagged":
			if value.(bool) {
				setStatements = append(setStatements, `set flagged of r to true`)
			} else {
				setStatements = append(setStatements, `set flagged of r to false`)
			}
		case "completed":
			if value.(bool) {
				setStatements = append(setStatements, `set completed of r to true`)
			} else {
				setStatements = append(setStatements, `set completed of r to false`)
			}
		}
	}

	if len(setStatements) == 0 {
		return nil
	}

	script := fmt.Sprintf(`%s
tell application "Reminders"
	set r to first reminder whose id is "%s"
	%s
end tell`, dateSetup.String(), EscapeString(id), strings.Join(setStatements, "\n\t"))

	_, err := s.exec.Run(script)
	if err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}

	return nil
}

// DeleteReminder deletes a reminder by ID.
func (s *ReminderService) DeleteReminder(id string) error {
	script := fmt.Sprintf(`tell application "Reminders"
	delete (first reminder whose id is "%s")
end tell`, EscapeString(id))

	_, err := s.exec.Run(script)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}

	return nil
}

// CompleteReminder marks a reminder as completed.
func (s *ReminderService) CompleteReminder(id string) error {
	return s.UpdateReminder(id, map[string]any{"completed": true})
}

// UncompleteReminder marks a reminder as incomplete.
func (s *ReminderService) UncompleteReminder(id string) error {
	return s.UpdateReminder(id, map[string]any{"completed": false})
}

// FlagReminder flags a reminder.
func (s *ReminderService) FlagReminder(id string) error {
	return s.UpdateReminder(id, map[string]any{"flagged": true})
}

// UnflagReminder removes the flag from a reminder.
func (s *ReminderService) UnflagReminder(id string) error {
	return s.UpdateReminder(id, map[string]any{"flagged": false})
}
