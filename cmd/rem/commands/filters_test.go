package commands

import (
	"testing"

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
