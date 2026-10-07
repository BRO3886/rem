package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/BRO3886/rem/internal/reminder"
)

func TestParseDueAt(t *testing.T) {
	// Tuesday, Sep 29 2026, 10:15 local.
	now := time.Date(2026, 9, 29, 10, 15, 0, 0, time.Local)

	tests := []struct {
		name         string
		input        string
		forceAllDay  bool
		preferAllDay bool
		want         time.Time
		wantAllDay   bool
		wantErr      string
	}{
		{name: "ISO date is all-day", input: "2026-09-30",
			want: time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local), wantAllDay: true},
		{name: "US date is all-day", input: "09/30/2026",
			want: time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local), wantAllDay: true},
		{name: "month name date is all-day", input: "Sep 30, 2026",
			want: time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local), wantAllDay: true},
		{name: "ISO date with time is timed", input: "2026-09-30 14:00",
			want: time.Date(2026, 9, 30, 14, 0, 0, 0, time.Local)},
		{name: "explicit midnight stays timed", input: "2026-09-30T00:00:00",
			want: time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local)},
		{name: "relative day keeps 9am default", input: "tomorrow",
			want: time.Date(2026, 9, 30, 9, 0, 0, 0, time.Local)},
		{name: "relative day with time is timed", input: "tomorrow at 2pm",
			want: time.Date(2026, 9, 30, 14, 0, 0, 0, time.Local)},

		{name: "--all-day on relative day", input: "tomorrow", forceAllDay: true,
			want: time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local), wantAllDay: true},
		{name: "--all-day on weekday", input: "friday", forceAllDay: true,
			want: time.Date(2026, 10, 2, 0, 0, 0, 0, time.Local), wantAllDay: true},
		{name: "--all-day on ISO date", input: "2026-09-30", forceAllDay: true,
			want: time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local), wantAllDay: true},
		{name: "--all-day rejects a time", input: "tomorrow at 2pm", forceAllDay: true,
			wantErr: "--all-day"},
		{name: "--all-day rejects a relative time", input: "in 2 hours", forceAllDay: true,
			wantErr: "--all-day"},

		{name: "all-day reminder stays all-day on a date-only update", input: "tomorrow", preferAllDay: true,
			want: time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local), wantAllDay: true},
		{name: "all-day reminder becomes timed when a time is given", input: "tomorrow at 2pm", preferAllDay: true,
			want: time.Date(2026, 9, 30, 14, 0, 0, 0, time.Local)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, allDay, err := parseDueAt(tt.input, now, tt.forceAllDay, tt.preferAllDay)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("due = %v, want %v", got, tt.want)
			}
			if allDay != tt.wantAllDay {
				t.Errorf("allDay = %v, want %v", allDay, tt.wantAllDay)
			}
		})
	}
}

func TestDropDueTimeAlarms(t *testing.T) {
	abs := time.Date(2026, 9, 30, 8, 0, 0, 0, time.Local)
	loc := &reminder.AlarmLocation{Latitude: 1, Longitude: 2, Proximity: "enter"}
	alarms := []reminder.Alarm{
		{RelativeOffset: 0},                 // auto alarm at the due time: dropped
		{RelativeOffset: -15 * time.Minute}, // explicit offset: kept
		{AbsoluteDate: &abs},                // absolute: kept
		{Location: loc},                     // geofence: kept
	}

	got, dropped := dropDueTimeAlarms(alarms)
	if !dropped {
		t.Error("dropped = false, want true")
	}
	if len(got) != 3 || got[0].RelativeOffset != -15*time.Minute || got[1].AbsoluteDate == nil || got[2].Location == nil {
		t.Errorf("kept = %+v", got)
	}

	if _, dropped := dropDueTimeAlarms(got); dropped {
		t.Error("no zero-offset alarm left, dropped should be false")
	}
}
