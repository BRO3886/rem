//go:build darwin

package service

import (
	"testing"

	"github.com/BRO3886/go-eventkit/reminders"
)

func TestFromEventKitList(t *testing.T) {
	l := &reminders.List{
		ID:    "list-123",
		Title: "Shopping",
		Color: "#FF6961",
		Count: 5,
	}

	result := fromEventKitList(l)

	if result.ID != "list-123" {
		t.Errorf("ID = %q, want %q", result.ID, "list-123")
	}
	if result.Name != "Shopping" {
		t.Errorf("Name = %q, want %q", result.Name, "Shopping")
	}
	if result.Color != "#FF6961" {
		t.Errorf("Color = %q, want %q", result.Color, "#FF6961")
	}
	if result.Count != 5 {
		t.Errorf("Count = %d, want %d", result.Count, 5)
	}
}

func TestFromEventKitListEmptyFields(t *testing.T) {
	l := &reminders.List{
		ID:    "list-456",
		Title: "Default",
	}

	result := fromEventKitList(l)

	if result.Name != "Default" {
		t.Errorf("Name = %q, want %q", result.Name, "Default")
	}
	if result.Color != "" {
		t.Errorf("Color = %q, want empty", result.Color)
	}
	if result.Count != 0 {
		t.Errorf("Count = %d, want 0", result.Count)
	}
}
