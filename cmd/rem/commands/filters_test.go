package commands

import (
	"testing"
	"time"

	"github.com/BRO3886/rem/internal/reminder"
)

func TestCompleteFilter(t *testing.T) {
	t.Run("complete shows incomplete reminders", func(t *testing.T) {
		f := completeFilter(false, "", false)
		if f.Completed == nil || *f.Completed != false {
			t.Errorf("expected Completed=false, got %v", f.Completed)
		}
	})

	t.Run("uncomplete shows completed reminders", func(t *testing.T) {
		f := completeFilter(true, "", false)
		if f.Completed == nil || *f.Completed != true {
			t.Errorf("expected Completed=true, got %v", f.Completed)
		}
	})

	t.Run("list filter applied", func(t *testing.T) {
		f := completeFilter(false, "Work", false)
		if f.ListName != "Work" {
			t.Errorf("expected ListName=Work, got %q", f.ListName)
		}
	})

	t.Run("empty list not set", func(t *testing.T) {
		f := completeFilter(false, "", false)
		if f.ListName != "" {
			t.Errorf("expected empty ListName, got %q", f.ListName)
		}
	})

	t.Run("flagged applied on complete", func(t *testing.T) {
		f := completeFilter(false, "", true)
		if f.Flagged == nil || *f.Flagged != true {
			t.Errorf("expected Flagged=true, got %v", f.Flagged)
		}
	})

	t.Run("flagged ignored on uncomplete", func(t *testing.T) {
		f := completeFilter(true, "", true)
		if f.Flagged != nil {
			t.Errorf("expected Flagged=nil on uncomplete, got %v", *f.Flagged)
		}
	})
}

func TestDeleteFilter(t *testing.T) {
	t.Run("shows incomplete reminders", func(t *testing.T) {
		f := deleteFilter("", false)
		if f.Completed == nil || *f.Completed != false {
			t.Errorf("expected Completed=false, got %v", f.Completed)
		}
	})

	t.Run("list filter applied", func(t *testing.T) {
		f := deleteFilter("Personal", false)
		if f.ListName != "Personal" {
			t.Errorf("expected ListName=Personal, got %q", f.ListName)
		}
	})

	t.Run("flagged filter applied", func(t *testing.T) {
		f := deleteFilter("", true)
		if f.Flagged == nil || *f.Flagged != true {
			t.Errorf("expected Flagged=true, got %v", f.Flagged)
		}
	})

	t.Run("flagged not set when false", func(t *testing.T) {
		f := deleteFilter("", false)
		if f.Flagged != nil {
			t.Errorf("expected Flagged=nil, got %v", *f.Flagged)
		}
	})
}

func TestFlagFilter(t *testing.T) {
	t.Run("shows incomplete reminders", func(t *testing.T) {
		f := flagFilter("")
		if f.Completed == nil || *f.Completed != false {
			t.Errorf("expected Completed=false, got %v", f.Completed)
		}
	})

	t.Run("list filter applied", func(t *testing.T) {
		f := flagFilter("Work")
		if f.ListName != "Work" {
			t.Errorf("expected ListName=Work, got %q", f.ListName)
		}
	})

	t.Run("no flagged filter", func(t *testing.T) {
		f := flagFilter("")
		if f.Flagged != nil {
			t.Errorf("expected Flagged=nil, got %v", *f.Flagged)
		}
	})
}

func TestParseRecurrence(t *testing.T) {
	tests := []struct {
		input       string
		wantErr     bool
		frequency   reminder.RecurrenceFrequency
		interval    int
		daysOfWeek  []int
		daysOfMonth []int
		desc        string
	}{
		// Simple keywords
		{"daily", false, reminder.FrequencyDaily, 1, nil, nil, "daily keyword"},
		{"every day", false, reminder.FrequencyDaily, 1, nil, nil, "every day"},
		{"weekly", false, reminder.FrequencyWeekly, 1, nil, nil, "weekly keyword"},
		{"every week", false, reminder.FrequencyWeekly, 1, nil, nil, "every week"},
		{"monthly", false, reminder.FrequencyMonthly, 1, nil, nil, "monthly keyword"},
		{"every month", false, reminder.FrequencyMonthly, 1, nil, nil, "every month"},
		{"yearly", false, reminder.FrequencyYearly, 1, nil, nil, "yearly keyword"},
		{"every year", false, reminder.FrequencyYearly, 1, nil, nil, "every year"},

		// Every N units
		{"every 2 days", false, reminder.FrequencyDaily, 2, nil, nil, "every 2 days"},
		{"every 3 weeks", false, reminder.FrequencyWeekly, 3, nil, nil, "every 3 weeks"},
		{"every 2 months", false, reminder.FrequencyMonthly, 2, nil, nil, "every 2 months"},
		{"every 5 years", false, reminder.FrequencyYearly, 5, nil, nil, "every 5 years"},
		{"every 1 day", false, reminder.FrequencyDaily, 1, nil, nil, "every 1 day (singular)"},
		{"every 1 week", false, reminder.FrequencyWeekly, 1, nil, nil, "every 1 week (singular)"},

		// Weekly with days
		{"weekly on mon", false, reminder.FrequencyWeekly, 1, []int{2}, nil, "weekly on monday"},
		{"weekly on mon,wed,fri", false, reminder.FrequencyWeekly, 1, []int{2, 4, 6}, nil, "weekly on MWF"},
		{"every 2 weeks on tue,thu", false, reminder.FrequencyWeekly, 2, []int{3, 5}, nil, "biweekly on tue/thu"},
		{"weekly on sun", false, reminder.FrequencyWeekly, 1, []int{1}, nil, "weekly on sunday"},
		{"weekly on sat", false, reminder.FrequencyWeekly, 1, []int{7}, nil, "weekly on saturday"},
		{"weekly on monday,friday", false, reminder.FrequencyWeekly, 1, []int{2, 6}, nil, "weekly full day names"},

		// Monthly with days of month
		{"monthly on 1", false, reminder.FrequencyMonthly, 1, nil, []int{1}, "monthly on 1st"},
		{"monthly on 1,15", false, reminder.FrequencyMonthly, 1, nil, []int{1, 15}, "monthly on 1st and 15th"},
		{"every 2 months on 10", false, reminder.FrequencyMonthly, 2, nil, []int{10}, "bimonthly on 10th"},

		// Case insensitivity
		{"DAILY", false, reminder.FrequencyDaily, 1, nil, nil, "uppercase DAILY"},
		{"Weekly On Mon,Fri", false, reminder.FrequencyWeekly, 1, []int{2, 6}, nil, "mixed case"},

		// Whitespace handling
		{"  daily  ", false, reminder.FrequencyDaily, 1, nil, nil, "extra whitespace"},

		// Error cases
		{"", true, 0, 0, nil, nil, "empty string"},
		{"biweekly", true, 0, 0, nil, nil, "unknown keyword"},
		{"every potato", true, 0, 0, nil, nil, "invalid unit"},
		{"weekly on xyz", true, 0, 0, nil, nil, "invalid day name"},
		{"monthly on 0", true, 0, 0, nil, nil, "day 0 invalid"},
		{"monthly on 32", true, 0, 0, nil, nil, "day 32 invalid"},
		{"monthly on abc", true, 0, 0, nil, nil, "non-numeric day of month"},
		{"daily on mon", true, 0, 0, nil, nil, "on clause not supported for daily"},
		{"yearly on mon", true, 0, 0, nil, nil, "on clause not supported for yearly"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			rule, err := parseRecurrence(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
				return
			}
			if rule.Frequency != tt.frequency {
				t.Errorf("frequency: got %d, want %d", rule.Frequency, tt.frequency)
			}
			if rule.Interval != tt.interval {
				t.Errorf("interval: got %d, want %d", rule.Interval, tt.interval)
			}
			if len(rule.DaysOfWeekNums) != len(tt.daysOfWeek) {
				t.Errorf("days of week: got %v, want %v", rule.DaysOfWeekNums, tt.daysOfWeek)
			} else {
				for i, d := range rule.DaysOfWeekNums {
					if d != tt.daysOfWeek[i] {
						t.Errorf("day of week[%d]: got %d, want %d", i, d, tt.daysOfWeek[i])
					}
				}
			}
			if len(rule.DaysOfMonth) != len(tt.daysOfMonth) {
				t.Errorf("days of month: got %v, want %v", rule.DaysOfMonth, tt.daysOfMonth)
			} else {
				for i, d := range rule.DaysOfMonth {
					if d != tt.daysOfMonth[i] {
						t.Errorf("day of month[%d]: got %d, want %d", i, d, tt.daysOfMonth[i])
					}
				}
			}
		})
	}
}

func TestParseRecurrenceDisplayNames(t *testing.T) {
	// Verify DaysOfWeek display names are populated alongside DaysOfWeekNums
	rule, err := parseRecurrence("weekly on mon,wed,fri")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedNames := []string{"Mon", "Wed", "Fri"}
	if len(rule.DaysOfWeek) != len(expectedNames) {
		t.Fatalf("DaysOfWeek: got %v, want %v", rule.DaysOfWeek, expectedNames)
	}
	for i, name := range rule.DaysOfWeek {
		if name != expectedNames[i] {
			t.Errorf("DaysOfWeek[%d]: got %q, want %q", i, name, expectedNames[i])
		}
	}
}

func TestUnflagFilter(t *testing.T) {
	t.Run("shows incomplete reminders", func(t *testing.T) {
		f := unflagFilter("")
		if f.Completed == nil || *f.Completed != false {
			t.Errorf("expected Completed=false, got %v", f.Completed)
		}
	})

	t.Run("shows flagged reminders", func(t *testing.T) {
		f := unflagFilter("")
		if f.Flagged == nil || *f.Flagged != true {
			t.Errorf("expected Flagged=true, got %v", f.Flagged)
		}
	})

	t.Run("list filter applied", func(t *testing.T) {
		f := unflagFilter("Personal")
		if f.ListName != "Personal" {
			t.Errorf("expected ListName=Personal, got %q", f.ListName)
		}
	})
}

func TestTagsFromTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  []string
	}{
		{
			name:  "single tag",
			title: "Review PR #work",
			want:  []string{"work"},
		},
		{
			name:  "multiple tags",
			title: "Review PR #work #urgent",
			want:  []string{"work", "urgent"},
		},
		{
			name:  "dedupes case-insensitively",
			title: "Review PR #Work #work",
			want:  []string{"Work"},
		},
		{
			name:  "trims trailing punctuation",
			title: "Review PR #work, then deploy #urgent.",
			want:  []string{"work", "urgent"},
		},
		{
			name:  "ignores issue references",
			title: "Fix issue #42",
			want:  nil,
		},
		{
			name:  "ignores bare hash",
			title: "Look at # and # work",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tagsFromTitle(tt.title)
			if len(got) != len(tt.want) {
				t.Fatalf("tagsFromTitle(%q) = %v, want %v", tt.title, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("tagsFromTitle(%q) = %v, want %v", tt.title, got, tt.want)
				}
			}
		})
	}
}

func TestMergeTags(t *testing.T) {
	got := mergeTags([]string{"existing", "Work"}, []string{"#work", "new"}, []string{"existing"})
	want := []string{"Work", "new"}
	if len(got) != len(want) {
		t.Fatalf("mergeTags = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("mergeTags = %v, want %v", got, want)
		}
	}
}

func TestParseTagList(t *testing.T) {
	got := parseTagList("work, #urgent, ,Deep Work")
	want := []string{"work", "urgent", "Deep Work"}
	if len(got) != len(want) {
		t.Fatalf("parseTagList = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("parseTagList = %v, want %v", got, want)
		}
	}
}

func TestMergeTagInputs(t *testing.T) {
	got := mergeTagInputs([]string{"title"}, "bulk, #extra")
	want := []string{"title", "bulk", "extra"}
	if len(got) != len(want) {
		t.Fatalf("mergeTagInputs = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("mergeTagInputs = %v, want %v", got, want)
		}
	}
}

func TestMergeTagUpdates(t *testing.T) {
	got := mergeTagUpdates([]string{"existing", "keep"}, []string{"title"}, "bulk,extra", "existing")
	want := []string{"keep", "title", "bulk", "extra"}
	if len(got) != len(want) {
		t.Fatalf("mergeTagUpdates = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("mergeTagUpdates = %v, want %v", got, want)
		}
	}
}

func TestBuildAlarms(t *testing.T) {
	t.Run("no due date, no remind-me: no alarms", func(t *testing.T) {
		alarms, err := buildAlarms(false, "", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if alarms != nil {
			t.Errorf("expected nil alarms, got %v", alarms)
		}
	})

	t.Run("due date, no remind-me, not silent: auto-attach zero offset", func(t *testing.T) {
		alarms, err := buildAlarms(true, "", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(alarms) != 1 {
			t.Fatalf("expected 1 alarm, got %d", len(alarms))
		}
		if alarms[0].RelativeOffset != 0 {
			t.Errorf("expected zero offset, got %v", alarms[0].RelativeOffset)
		}
		if alarms[0].AbsoluteDate != nil {
			t.Errorf("expected no absolute date, got %v", alarms[0].AbsoluteDate)
		}
	})

	t.Run("due date with --silent: no alarms", func(t *testing.T) {
		alarms, err := buildAlarms(true, "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if alarms != nil {
			t.Errorf("expected nil alarms when --silent, got %v", alarms)
		}
	})

	t.Run("remind-me wins over silent (explicit user intent)", func(t *testing.T) {
		alarms, err := buildAlarms(true, "15m", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(alarms) != 1 {
			t.Fatalf("expected 1 alarm, got %d", len(alarms))
		}
		// 15m before due → -15m offset
		if alarms[0].RelativeOffset != -15*time.Minute {
			t.Errorf("expected -15m offset, got %v", alarms[0].RelativeOffset)
		}
	})

	t.Run("remind-me 0m matches auto-attach behavior", func(t *testing.T) {
		auto, _ := buildAlarms(true, "", false)
		explicit, _ := buildAlarms(true, "0m", false)
		if len(auto) != 1 || len(explicit) != 1 {
			t.Fatalf("both should produce 1 alarm: auto=%v explicit=%v", auto, explicit)
		}
		if auto[0].RelativeOffset != explicit[0].RelativeOffset {
			t.Errorf("auto offset=%v, explicit 0m offset=%v — should match",
				auto[0].RelativeOffset, explicit[0].RelativeOffset)
		}
	})

	t.Run("remind-me without due date still works", func(t *testing.T) {
		// Edge case: user passes --remind-me but no --due. The remind-me
		// alarm should still attach (the service layer will reject it if
		// the combination is invalid at the EventKit level).
		alarms, err := buildAlarms(false, "1h", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(alarms) != 1 {
			t.Fatalf("expected 1 alarm, got %d", len(alarms))
		}
		if alarms[0].RelativeOffset != -1*time.Hour {
			t.Errorf("expected -1h offset, got %v", alarms[0].RelativeOffset)
		}
	})

	t.Run("invalid remind-me surfaces the parse error", func(t *testing.T) {
		_, err := buildAlarms(true, "not a duration or date", false)
		if err == nil {
			t.Fatal("expected error for invalid remind-me")
		}
	})

	t.Run("absolute remind-me time", func(t *testing.T) {
		alarms, err := buildAlarms(true, "tomorrow at 9am", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(alarms) != 1 {
			t.Fatalf("expected 1 alarm, got %d", len(alarms))
		}
		if alarms[0].AbsoluteDate == nil {
			t.Errorf("expected absolute date alarm, got %+v", alarms[0])
		}
		if alarms[0].RelativeOffset != 0 {
			t.Errorf("absolute alarm should not have a relative offset, got %v", alarms[0].RelativeOffset)
		}
	})

	// Silence the unused import warning if reminder package stops being referenced.
	_ = reminder.Alarm{}
}

func TestParseLocationAlarm(t *testing.T) {
	t.Run("valid coordinates default to arrive", func(t *testing.T) {
		a, err := parseLocationAlarm("37.3318,-122.0312", 0, false, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Location == nil {
			t.Fatal("Location should not be nil")
		}
		if a.Location.Latitude != 37.3318 || a.Location.Longitude != -122.0312 {
			t.Errorf("coords = %f,%f", a.Location.Latitude, a.Location.Longitude)
		}
		if a.Location.Proximity != "enter" {
			t.Errorf("Proximity = %q, want enter (default)", a.Location.Proximity)
		}
		if a.Location.Radius != 0 {
			t.Errorf("Radius = %f, want 0 (system default)", a.Location.Radius)
		}
		if a.Location.Title == "" {
			t.Error("Title should default to the coordinate string")
		}
	})

	t.Run("on-leave and radius", func(t *testing.T) {
		a, err := parseLocationAlarm(" 37.33 , -122.03 ", 200, false, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Location.Proximity != "leave" {
			t.Errorf("Proximity = %q, want leave", a.Location.Proximity)
		}
		if a.Location.Radius != 200 {
			t.Errorf("Radius = %f, want 200", a.Location.Radius)
		}
	})

	t.Run("explicit on-arrive", func(t *testing.T) {
		a, err := parseLocationAlarm("0,0", 0, true, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Location.Proximity != "enter" {
			t.Errorf("Proximity = %q, want enter", a.Location.Proximity)
		}
	})

	t.Run("errors", func(t *testing.T) {
		cases := []struct {
			name     string
			location string
			radius   float64
			arrive   bool
			leave    bool
		}{
			{"both proximity flags", "1,2", 0, true, true},
			{"missing longitude", "37.33", 0, false, false},
			{"too many parts", "1,2,3", 0, false, false},
			{"non-numeric", "north,south", 0, false, false},
			{"latitude out of range", "91,0", 0, false, false},
			{"longitude out of range", "0,181", 0, false, false},
			{"negative radius", "1,2", -5, false, false},
			{"empty", "", 0, false, false},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				if _, err := parseLocationAlarm(tc.location, tc.radius, tc.arrive, tc.leave); err == nil {
					t.Errorf("parseLocationAlarm(%q, %v, %v, %v) should error", tc.location, tc.radius, tc.arrive, tc.leave)
				}
			})
		}
	})
}

func TestSplitAlarmsByTrigger(t *testing.T) {
	loc := &reminder.AlarmLocation{Latitude: 1, Longitude: 2, Proximity: "enter"}
	alarms := []reminder.Alarm{
		{RelativeOffset: 0},
		{Location: loc},
		{RelativeOffset: -15 * time.Minute},
	}

	timeAlarms, locationAlarms := splitAlarmsByTrigger(alarms)
	if len(timeAlarms) != 2 {
		t.Errorf("timeAlarms = %d, want 2", len(timeAlarms))
	}
	if len(locationAlarms) != 1 {
		t.Errorf("locationAlarms = %d, want 1", len(locationAlarms))
	}
	if locationAlarms[0].Location != loc {
		t.Error("location alarm not preserved")
	}

	timeAlarms, locationAlarms = splitAlarmsByTrigger(nil)
	if timeAlarms != nil || locationAlarms != nil {
		t.Error("nil input should produce nil buckets")
	}
}
