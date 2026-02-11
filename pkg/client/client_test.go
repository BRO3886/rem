//go:build darwin

package client

import (
	"testing"
	"time"

	"github.com/BRO3886/rem/internal/reminder"
)

func TestToPublicReminder(t *testing.T) {
	now := time.Now()
	due := now.Add(24 * time.Hour)

	r := &reminder.Reminder{
		ID:       "test-id",
		Name:     "Test Reminder",
		Body:     "Some notes",
		ListName: "Personal",
		DueDate:  &due,
		Priority: reminder.PriorityHigh,
		Flagged:  true,
		URL:      "https://example.com",
	}

	pub := toPublicReminder(r)

	if pub.ID != "test-id" {
		t.Errorf("ID = %q, want %q", pub.ID, "test-id")
	}
	if pub.Title != "Test Reminder" {
		t.Errorf("Title = %q, want %q", pub.Title, "Test Reminder")
	}
	if pub.Notes != "Some notes" {
		t.Errorf("Notes = %q, want %q", pub.Notes, "Some notes")
	}
	if pub.ListName != "Personal" {
		t.Errorf("ListName = %q, want %q", pub.ListName, "Personal")
	}
	if pub.DueDate == nil || !pub.DueDate.Equal(due) {
		t.Errorf("DueDate = %v, want %v", pub.DueDate, due)
	}
	if pub.Priority != PriorityHigh {
		t.Errorf("Priority = %d, want %d", pub.Priority, PriorityHigh)
	}
	if !pub.Flagged {
		t.Error("Flagged = false, want true")
	}
	if pub.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", pub.URL, "https://example.com")
	}
}

func TestToPublicList(t *testing.T) {
	l := &reminder.List{
		ID:    "list-id",
		Name:  "Work",
		Color: "#0000FF",
		Count: 10,
	}

	pub := toPublicList(l)

	if pub.ID != "list-id" {
		t.Errorf("ID = %q, want %q", pub.ID, "list-id")
	}
	if pub.Name != "Work" {
		t.Errorf("Name = %q, want %q", pub.Name, "Work")
	}
	if pub.Color != "#0000FF" {
		t.Errorf("Color = %q, want %q", pub.Color, "#0000FF")
	}
	if pub.Count != 10 {
		t.Errorf("Count = %d, want %d", pub.Count, 10)
	}
}
