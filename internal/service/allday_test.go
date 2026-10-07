//go:build darwin

package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/BRO3886/go-eventkit/reminders"
)

func TestFromEventKitReminderAllDay(t *testing.T) {
	due := time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local)
	r := fromEventKitReminder(&reminders.Reminder{ID: "A", DueDate: &due, DueDateAllDay: true})
	if !r.AllDay {
		t.Error("AllDay = false, want true")
	}

	r = fromEventKitReminder(&reminders.Reminder{ID: "A", DueDate: &due})
	if r.AllDay {
		t.Error("AllDay = true, want false")
	}
}

func TestAmbiguousIDFromEventKit(t *testing.T) {
	ekErr := &reminders.AmbiguousIDError{
		Prefix: "4",
		Candidates: []reminders.Reminder{
			{ID: "4AAAAAAA-1111", Title: "Buy milk", List: "Work"},
			{ID: "4BBBBBBB-2222", Title: "Call plumber", List: "Inbox"},
		},
	}

	err := ambiguousIDError(fmt.Errorf("wrapped: %w", ekErr))
	var amb *AmbiguousIDError
	if !errors.As(err, &amb) {
		t.Fatalf("expected *AmbiguousIDError, got %T: %v", err, err)
	}

	msg := err.Error()
	for _, want := range []string{`"4"`, "2 reminders", "4AAAAAAA", "Buy milk", "Work", "4BBBBBBB", "Call plumber", "Inbox"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message missing %q:\n%s", want, msg)
		}
	}

	if ambiguousIDError(errors.New("other")) != nil {
		t.Error("non-ambiguous errors should map to nil")
	}
}

func TestAmbiguousIDErrorTruncatesLongLists(t *testing.T) {
	amb := &AmbiguousIDError{Prefix: "A"}
	for i := 0; i < 25; i++ {
		amb.Candidates = append(amb.Candidates, fromEventKitReminder(&reminders.Reminder{
			ID: fmt.Sprintf("A%07d-X", i), Title: fmt.Sprintf("Item %d", i), List: "Inbox",
		}))
	}

	msg := amb.Error()
	if !strings.Contains(msg, "25 reminders") || !strings.Contains(msg, "and 15 more") {
		t.Errorf("expected a count and truncation note:\n%s", msg)
	}
	if strings.Contains(msg, "Item 10") {
		t.Errorf("expected at most 10 candidates listed:\n%s", msg)
	}
}
