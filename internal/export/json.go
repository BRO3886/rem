package export

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/BRO3886/rem/internal/reminder"
)

// JSONReminder is the JSON-serializable representation of a reminder.
type JSONReminder struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Body             string  `json:"body,omitempty"`
	ListName         string  `json:"list_name"`
	DueDate          *string `json:"due_date,omitempty"`
	RemindMeDate     *string `json:"remind_me_date,omitempty"`
	CompletionDate   *string `json:"completion_date,omitempty"`
	CreationDate     *string `json:"creation_date,omitempty"`
	ModificationDate *string `json:"modification_date,omitempty"`
	Priority         int     `json:"priority"`
	PriorityLabel    string  `json:"priority_label"`
	Flagged          bool    `json:"flagged"`
	Completed        bool    `json:"completed"`
	URL              string  `json:"url,omitempty"`
}

const timeFormat = "2006-01-02T15:04:05"

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(timeFormat)
	return &s
}

// ToJSON converts a reminder to its JSON representation.
func ToJSON(r *reminder.Reminder) JSONReminder {
	return JSONReminder{
		ID:               r.ID,
		Name:             r.Name,
		Body:             r.Body,
		ListName:         r.ListName,
		DueDate:          formatTimePtr(r.DueDate),
		RemindMeDate:     formatTimePtr(r.RemindMeDate),
		CompletionDate:   formatTimePtr(r.CompletionDate),
		CreationDate:     formatTimePtr(r.CreationDate),
		ModificationDate: formatTimePtr(r.ModificationDate),
		Priority:         int(r.Priority),
		PriorityLabel:    r.Priority.String(),
		Flagged:          r.Flagged,
		Completed:        r.Completed,
		URL:              r.URL,
	}
}

// ExportJSON writes reminders as JSON to the writer.
func ExportJSON(w io.Writer, reminders []*reminder.Reminder) error {
	jsonReminders := make([]JSONReminder, 0, len(reminders))
	for _, r := range reminders {
		jsonReminders = append(jsonReminders, ToJSON(r))
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonReminders)
}

// ImportJSON reads reminders from a JSON reader.
func ImportJSON(r io.Reader) ([]*reminder.Reminder, error) {
	var jsonReminders []JSONReminder
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&jsonReminders); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	reminders := make([]*reminder.Reminder, 0, len(jsonReminders))
	for _, jr := range jsonReminders {
		rem := &reminder.Reminder{
			Name:      jr.Name,
			Body:      jr.Body,
			ListName:  jr.ListName,
			Priority:  reminder.Priority(jr.Priority),
			Flagged:   jr.Flagged,
			Completed: jr.Completed,
			URL:       jr.URL,
		}

		if jr.DueDate != nil {
			t, err := time.ParseInLocation(timeFormat, *jr.DueDate, time.Now().Location())
			if err == nil {
				rem.DueDate = &t
			}
		}
		if jr.RemindMeDate != nil {
			t, err := time.ParseInLocation(timeFormat, *jr.RemindMeDate, time.Now().Location())
			if err == nil {
				rem.RemindMeDate = &t
			}
		}

		reminders = append(reminders, rem)
	}

	return reminders, nil
}
