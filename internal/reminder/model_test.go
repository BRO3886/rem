package reminder

import (
	"testing"
	"time"
)

func TestPriorityString(t *testing.T) {
	tests := []struct {
		p    Priority
		want string
	}{
		{PriorityNone, "none"},
		{PriorityHigh, "high"},
		{Priority(2), "high"},
		{Priority(3), "high"},
		{Priority(4), "high"},
		{PriorityMedium, "medium"},
		{PriorityLow, "low"},
		{Priority(6), "low"},
		{Priority(7), "low"},
		{Priority(8), "low"},
		{Priority(-1), "none"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.p.String()
			if got != tt.want {
				t.Errorf("Priority(%d).String() = %q, want %q", tt.p, got, tt.want)
			}
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input string
		want  Priority
	}{
		{"high", PriorityHigh},
		{"h", PriorityHigh},
		{"1", PriorityHigh},
		{"medium", PriorityMedium},
		{"med", PriorityMedium},
		{"m", PriorityMedium},
		{"5", PriorityMedium},
		{"low", PriorityLow},
		{"l", PriorityLow},
		{"9", PriorityLow},
		{"none", PriorityNone},
		{"", PriorityNone},
		{"invalid", PriorityNone},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParsePriority(tt.input)
			if got != tt.want {
				t.Errorf("ParsePriority(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestAlarmLocationString(t *testing.T) {
	tests := []struct {
		name string
		loc  AlarmLocation
		want string
	}{
		{
			"arrive with title and radius",
			AlarmLocation{Title: "Grocery Store", Latitude: 37.3318, Longitude: -122.0312, Radius: 200, Proximity: "enter"},
			"on arriving at Grocery Store (within 200m)",
		},
		{
			"leave without radius",
			AlarmLocation{Title: "Office", Latitude: 37.7749, Longitude: -122.4194, Proximity: "leave"},
			"on leaving Office",
		},
		{
			"no title falls back to coordinates",
			AlarmLocation{Latitude: 37.3318, Longitude: -122.0312, Proximity: "enter"},
			"on arriving at 37.3318,-122.0312",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.loc.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAlarmStringPrefersLocation(t *testing.T) {
	a := Alarm{
		RelativeOffset: 0,
		Location:       &AlarmLocation{Title: "Home", Proximity: "leave"},
	}
	if got := a.String(); got != "on leaving Home" {
		t.Errorf("String() = %q, want location description", got)
	}
}

func TestRecurrenceRuleStringWithEnd(t *testing.T) {
	endDate := time.Date(2026, 10, 10, 23, 59, 59, 0, time.Local)
	tests := []struct {
		rule RecurrenceRule
		want string
	}{
		{RecurrenceRule{Frequency: FrequencyDaily, Interval: 1}, "every day"},
		{RecurrenceRule{Frequency: FrequencyDaily, Interval: 1, End: &RecurrenceEnd{EndDate: &endDate}}, "every day, until Sat Oct 10, 2026"},
		{RecurrenceRule{Frequency: FrequencyWeekly, Interval: 1, DaysOfWeek: []string{"Mon", "Fri"}, End: &RecurrenceEnd{OccurrenceCount: 5}}, "every week on Mon, Fri, 5 times"},
		{RecurrenceRule{Frequency: FrequencyMonthly, Interval: 2, End: &RecurrenceEnd{OccurrenceCount: 1}}, "every 2 months, 1 time"},
	}
	for _, tt := range tests {
		if got := tt.rule.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}
