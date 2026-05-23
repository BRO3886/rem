//go:build darwin

package service

import (
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/reminders"
	"github.com/BRO3886/rem/internal/reminder"
)

func TestFromEventKitReminder(t *testing.T) {
	now := time.Now()
	due := now.Add(24 * time.Hour)
	created := now.Add(-48 * time.Hour)

	r := &reminders.Reminder{
		ID:        "ABC-123",
		Title:     "Buy groceries",
		Notes:     "Milk, eggs, bread",
		List:      "Shopping",
		DueDate:   &due,
		CreatedAt: &created,
		Priority:  reminders.PriorityHigh,
		Completed: false,
		Flagged:   false,
		URL:       "https://example.com",
		Tags:      []string{"work", "urgent"},
	}

	result := fromEventKitReminder(r)

	if result.ID != "ABC-123" {
		t.Errorf("ID = %q, want %q", result.ID, "ABC-123")
	}
	if result.Name != "Buy groceries" {
		t.Errorf("Name = %q, want %q", result.Name, "Buy groceries")
	}
	if result.Body != "Milk, eggs, bread" {
		t.Errorf("Body = %q, want %q", result.Body, "Milk, eggs, bread")
	}
	if result.ListName != "Shopping" {
		t.Errorf("ListName = %q, want %q", result.ListName, "Shopping")
	}
	if result.DueDate == nil || !result.DueDate.Equal(due) {
		t.Errorf("DueDate = %v, want %v", result.DueDate, due)
	}
	if result.CreationDate == nil || !result.CreationDate.Equal(created) {
		t.Errorf("CreationDate = %v, want %v", result.CreationDate, created)
	}
	if result.Priority != reminder.PriorityHigh {
		t.Errorf("Priority = %d, want %d", result.Priority, reminder.PriorityHigh)
	}
	if result.Completed {
		t.Error("Completed = true, want false")
	}
	if result.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", result.URL, "https://example.com")
	}
	if len(result.Tags) != 2 || result.Tags[0] != "work" || result.Tags[1] != "urgent" {
		t.Errorf("Tags = %v, want [work urgent]", result.Tags)
	}
}

func TestFromEventKitReminderURLExtraction(t *testing.T) {
	// When URL is empty but notes contain a URL, it should be extracted
	r := &reminders.Reminder{
		ID:    "DEF-456",
		Title: "Check website",
		Notes: "Some notes\n\nURL: https://example.com/page",
	}

	result := fromEventKitReminder(r)

	if result.URL != "https://example.com/page" {
		t.Errorf("URL = %q, want %q", result.URL, "https://example.com/page")
	}
}

func TestFromEventKitReminderNativeURLTakesPrecedence(t *testing.T) {
	// When go-eventkit provides a URL, don't extract from notes
	r := &reminders.Reminder{
		ID:    "GHI-789",
		Title: "Check website",
		Notes: "URL: https://old.example.com",
		URL:   "https://native.example.com",
	}

	result := fromEventKitReminder(r)

	if result.URL != "https://native.example.com" {
		t.Errorf("URL = %q, want %q", result.URL, "https://native.example.com")
	}
}

func TestFromEventKitReminderNilDates(t *testing.T) {
	r := &reminders.Reminder{
		ID:    "JKL-012",
		Title: "No dates",
	}

	result := fromEventKitReminder(r)

	if result.DueDate != nil {
		t.Errorf("DueDate = %v, want nil", result.DueDate)
	}
	if result.RemindMeDate != nil {
		t.Errorf("RemindMeDate = %v, want nil", result.RemindMeDate)
	}
	if result.CompletionDate != nil {
		t.Errorf("CompletionDate = %v, want nil", result.CompletionDate)
	}
}

func timePtr(t time.Time) *time.Time { return &t }

func TestSortReminders(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-2 * time.Hour)
	later := now.Add(2 * time.Hour)

	tests := []struct {
		name      string
		input     []*reminder.Reminder
		wantOrder []string // expected Name order after sort
	}{
		{
			name: "sort by due date ascending",
			input: []*reminder.Reminder{
				{Name: "later", DueDate: timePtr(later)},
				{Name: "earlier", DueDate: timePtr(earlier)},
				{Name: "now", DueDate: timePtr(now)},
			},
			wantOrder: []string{"earlier", "now", "later"},
		},
		{
			name: "nil due dates sort last",
			input: []*reminder.Reminder{
				{Name: "no-due", DueDate: nil},
				{Name: "has-due", DueDate: timePtr(now)},
			},
			wantOrder: []string{"has-due", "no-due"},
		},
		{
			name: "both nil due dates - priority tiebreaker",
			input: []*reminder.Reminder{
				{Name: "low", DueDate: nil, Priority: reminder.PriorityLow},
				{Name: "high", DueDate: nil, Priority: reminder.PriorityHigh},
			},
			wantOrder: []string{"high", "low"},
		},
		{
			name: "same due date - priority tiebreaker",
			input: []*reminder.Reminder{
				{Name: "medium", DueDate: timePtr(now), Priority: reminder.PriorityMedium},
				{Name: "high", DueDate: timePtr(now), Priority: reminder.PriorityHigh},
				{Name: "low", DueDate: timePtr(now), Priority: reminder.PriorityLow},
			},
			wantOrder: []string{"high", "medium", "low"},
		},
		{
			name: "priority none sorts last in tiebreaker",
			input: []*reminder.Reminder{
				{Name: "none", DueDate: timePtr(now), Priority: reminder.PriorityNone},
				{Name: "low", DueDate: timePtr(now), Priority: reminder.PriorityLow},
			},
			wantOrder: []string{"low", "none"},
		},
		{
			name: "both nil due and both priority none - stable",
			input: []*reminder.Reminder{
				{Name: "a", DueDate: nil, Priority: reminder.PriorityNone},
				{Name: "b", DueDate: nil, Priority: reminder.PriorityNone},
			},
			wantOrder: []string{"a", "b"},
		},
		{
			name:      "empty slice",
			input:     []*reminder.Reminder{},
			wantOrder: []string{},
		},
		{
			name: "single element",
			input: []*reminder.Reminder{
				{Name: "only"},
			},
			wantOrder: []string{"only"},
		},
		{
			name: "already sorted",
			input: []*reminder.Reminder{
				{Name: "first", DueDate: timePtr(earlier), Priority: reminder.PriorityHigh},
				{Name: "second", DueDate: timePtr(now), Priority: reminder.PriorityMedium},
				{Name: "third", DueDate: timePtr(later), Priority: reminder.PriorityLow},
			},
			wantOrder: []string{"first", "second", "third"},
		},
		{
			name: "mixed nil and non-nil with priorities",
			input: []*reminder.Reminder{
				{Name: "nil-none", DueDate: nil, Priority: reminder.PriorityNone},
				{Name: "later-high", DueDate: timePtr(later), Priority: reminder.PriorityHigh},
				{Name: "nil-high", DueDate: nil, Priority: reminder.PriorityHigh},
				{Name: "earlier-low", DueDate: timePtr(earlier), Priority: reminder.PriorityLow},
				{Name: "earlier-high", DueDate: timePtr(earlier), Priority: reminder.PriorityHigh},
			},
			wantOrder: []string{"earlier-high", "earlier-low", "later-high", "nil-high", "nil-none"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortReminders(tt.input)
			for i, r := range tt.input {
				if i >= len(tt.wantOrder) {
					t.Errorf("unexpected extra element at index %d: %s", i, r.Name)
					continue
				}
				if r.Name != tt.wantOrder[i] {
					t.Errorf("index %d: got %q, want %q", i, r.Name, tt.wantOrder[i])
				}
			}
		})
	}
}

func TestFromEventKitReminderPriorityMapping(t *testing.T) {
	tests := []struct {
		name     string
		input    reminders.Priority
		expected reminder.Priority
	}{
		{"none", reminders.PriorityNone, reminder.PriorityNone},
		{"high", reminders.PriorityHigh, reminder.PriorityHigh},
		{"medium", reminders.PriorityMedium, reminder.PriorityMedium},
		{"low", reminders.PriorityLow, reminder.PriorityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reminders.Reminder{
				ID:       "test",
				Title:    "test",
				Priority: tt.input,
			}
			result := fromEventKitReminder(r)
			if result.Priority != tt.expected {
				t.Errorf("Priority = %d, want %d", result.Priority, tt.expected)
			}
		})
	}
}
