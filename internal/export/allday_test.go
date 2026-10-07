package export

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/BRO3886/rem/internal/reminder"
)

func TestJSONAllDayRoundTrip(t *testing.T) {
	due := time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local)
	in := []*reminder.Reminder{
		{Name: "all-day", DueDate: &due, AllDay: true},
		{Name: "timed", DueDate: &due},
	}

	var buf bytes.Buffer
	if err := ExportJSON(&buf, in); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(buf.String(), `"due_all_day": true`); got != 1 {
		t.Errorf("expected due_all_day on exactly one reminder, got %d:\n%s", got, buf.String())
	}

	out, err := ImportJSON(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !out[0].AllDay || out[1].AllDay {
		t.Errorf("AllDay after round trip = %v, %v; want true, false", out[0].AllDay, out[1].AllDay)
	}
}

func TestCSVAllDayRoundTrip(t *testing.T) {
	due := time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local)
	in := []*reminder.Reminder{
		{Name: "all-day", DueDate: &due, AllDay: true},
		{Name: "timed", DueDate: &due},
	}

	var buf bytes.Buffer
	if err := ExportCSV(&buf, in); err != nil {
		t.Fatal(err)
	}
	out, err := ImportCSV(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !out[0].AllDay || out[1].AllDay {
		t.Errorf("AllDay after round trip = %v, %v; want true, false", out[0].AllDay, out[1].AllDay)
	}
}

func TestCSVImportWithoutAllDayColumn(t *testing.T) {
	out, err := ImportCSV(strings.NewReader("name,due_date\nold,2026-09-30T00:00:00\n"))
	if err != nil {
		t.Fatal(err)
	}
	if out[0].DueDate == nil || out[0].AllDay {
		t.Errorf("old CSV should import a timed due, got DueDate=%v AllDay=%v", out[0].DueDate, out[0].AllDay)
	}
}
