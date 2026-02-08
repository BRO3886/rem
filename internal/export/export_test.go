package export

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/BRO3886/rem/internal/reminder"
)

func sampleReminders() []*reminder.Reminder {
	dueDate := time.Date(2026, 2, 15, 14, 30, 0, 0, time.Local)
	return []*reminder.Reminder{
		{
			ID:       "test-id-1",
			Name:     "Buy groceries",
			Body:     "Milk, eggs, bread",
			ListName: "Personal",
			DueDate:  &dueDate,
			Priority: reminder.PriorityHigh,
			Flagged:  true,
		},
		{
			ID:       "test-id-2",
			Name:     "Review PR",
			Body:     "Check auth changes\n\nURL: https://github.com/example",
			ListName: "Work",
			Priority: reminder.PriorityNone,
			URL:      "https://github.com/example",
		},
	}
}

func TestExportJSON(t *testing.T) {
	var buf bytes.Buffer
	reminders := sampleReminders()

	err := ExportJSON(&buf, reminders)
	if err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Buy groceries") {
		t.Error("JSON output should contain reminder name")
	}
	if !strings.Contains(output, "test-id-1") {
		t.Error("JSON output should contain reminder ID")
	}
	if !strings.Contains(output, "high") {
		t.Error("JSON output should contain priority label")
	}
	if !strings.Contains(output, "https://github.com/example") {
		t.Error("JSON output should contain URL")
	}
}

func TestImportJSON(t *testing.T) {
	jsonData := `[
		{
			"name": "Test Reminder",
			"body": "Test notes",
			"list_name": "Personal",
			"priority": 1,
			"flagged": true,
			"due_date": "2026-02-15T14:30:00"
		}
	]`

	reminders, err := ImportJSON(strings.NewReader(jsonData))
	if err != nil {
		t.Fatalf("ImportJSON failed: %v", err)
	}

	if len(reminders) != 1 {
		t.Fatalf("expected 1 reminder, got %d", len(reminders))
	}

	r := reminders[0]
	if r.Name != "Test Reminder" {
		t.Errorf("expected name 'Test Reminder', got '%s'", r.Name)
	}
	if r.Body != "Test notes" {
		t.Errorf("expected body 'Test notes', got '%s'", r.Body)
	}
	if r.ListName != "Personal" {
		t.Errorf("expected list 'Personal', got '%s'", r.ListName)
	}
	if r.Priority != reminder.PriorityHigh {
		t.Errorf("expected priority high, got %v", r.Priority)
	}
	if !r.Flagged {
		t.Error("expected flagged to be true")
	}
	if r.DueDate == nil {
		t.Error("expected due date to be set")
	}
}

func TestExportCSV(t *testing.T) {
	var buf bytes.Buffer
	reminders := sampleReminders()

	err := ExportCSV(&buf, reminders)
	if err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	output := buf.String()

	// Check that output contains expected content
	// Note: Body fields with newlines result in quoted CSV fields spanning multiple lines
	if !strings.Contains(output, "id,name,body") {
		t.Error("CSV header should contain expected columns")
	}
	if !strings.Contains(output, "Buy groceries") {
		t.Error("CSV should contain first reminder name")
	}
	if !strings.Contains(output, "Review PR") {
		t.Error("CSV should contain second reminder name")
	}
}

func TestImportCSV(t *testing.T) {
	csvData := `name,body,list_name,priority,flagged,due_date
Test Reminder,Test notes,Personal,1,true,2026-02-15T14:30:00
Another Reminder,,Work,0,false,`

	reminders, err := ImportCSV(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("ImportCSV failed: %v", err)
	}

	if len(reminders) != 2 {
		t.Fatalf("expected 2 reminders, got %d", len(reminders))
	}

	r := reminders[0]
	if r.Name != "Test Reminder" {
		t.Errorf("expected name 'Test Reminder', got '%s'", r.Name)
	}
	if r.Priority != reminder.PriorityHigh {
		t.Errorf("expected priority high, got %v", r.Priority)
	}
	if !r.Flagged {
		t.Error("expected flagged to be true")
	}

	r2 := reminders[1]
	if r2.Name != "Another Reminder" {
		t.Errorf("expected name 'Another Reminder', got '%s'", r2.Name)
	}
	if r2.Flagged {
		t.Error("expected flagged to be false")
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	original := sampleReminders()

	// Export to JSON
	var jsonBuf bytes.Buffer
	if err := ExportJSON(&jsonBuf, original); err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Import from JSON
	imported, err := ImportJSON(&jsonBuf)
	if err != nil {
		t.Fatalf("ImportJSON failed: %v", err)
	}

	if len(imported) != len(original) {
		t.Fatalf("expected %d reminders, got %d", len(original), len(imported))
	}

	for i, orig := range original {
		imp := imported[i]
		if orig.Name != imp.Name {
			t.Errorf("reminder %d: name mismatch: %s != %s", i, orig.Name, imp.Name)
		}
		if orig.ListName != imp.ListName {
			t.Errorf("reminder %d: list name mismatch: %s != %s", i, orig.ListName, imp.ListName)
		}
		if orig.Priority != imp.Priority {
			t.Errorf("reminder %d: priority mismatch: %v != %v", i, orig.Priority, imp.Priority)
		}
	}
}

func TestSpecialCharacters(t *testing.T) {
	dueDate := time.Date(2026, 2, 15, 14, 30, 0, 0, time.Local)
	reminders := []*reminder.Reminder{
		{
			ID:       "test-special",
			Name:     `He said "hello" & goodbye`,
			Body:     "Line 1\nLine 2\nLine 3",
			ListName: "Test's List",
			DueDate:  &dueDate,
		},
	}

	// Test JSON round-trip with special characters
	var jsonBuf bytes.Buffer
	if err := ExportJSON(&jsonBuf, reminders); err != nil {
		t.Fatalf("ExportJSON with special chars failed: %v", err)
	}

	imported, err := ImportJSON(&jsonBuf)
	if err != nil {
		t.Fatalf("ImportJSON with special chars failed: %v", err)
	}

	if len(imported) != 1 {
		t.Fatalf("expected 1 reminder, got %d", len(imported))
	}

	if imported[0].Name != reminders[0].Name {
		t.Errorf("name with special chars: expected %q, got %q", reminders[0].Name, imported[0].Name)
	}

	// Test CSV round-trip with special characters
	var csvBuf bytes.Buffer
	if err := ExportCSV(&csvBuf, reminders); err != nil {
		t.Fatalf("ExportCSV with special chars failed: %v", err)
	}

	csvImported, err := ImportCSV(&csvBuf)
	if err != nil {
		t.Fatalf("ImportCSV with special chars failed: %v", err)
	}

	if len(csvImported) != 1 {
		t.Fatalf("expected 1 reminder, got %d", len(csvImported))
	}

	if csvImported[0].Name != reminders[0].Name {
		t.Errorf("CSV name with special chars: expected %q, got %q", reminders[0].Name, csvImported[0].Name)
	}
}
