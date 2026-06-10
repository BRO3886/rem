package reminder

import (
	"fmt"
	"strings"
	"time"
)

// Priority represents the priority level of a reminder.
type Priority int

const (
	PriorityNone   Priority = 0
	PriorityHigh   Priority = 1
	PriorityMedium Priority = 5
	PriorityLow    Priority = 9
)

func (p Priority) String() string {
	switch {
	case p == PriorityNone:
		return "none"
	case p >= 1 && p <= 4:
		return "high"
	case p == 5:
		return "medium"
	case p >= 6 && p <= 9:
		return "low"
	default:
		return "none"
	}
}

// ParsePriority converts a string to a Priority value.
func ParsePriority(s string) Priority {
	switch s {
	case "high", "h", "1":
		return PriorityHigh
	case "medium", "med", "m", "5":
		return PriorityMedium
	case "low", "l", "9":
		return PriorityLow
	default:
		return PriorityNone
	}
}

// AlarmLocation is a geofence trigger attached to an alarm.
type AlarmLocation struct {
	Title     string
	Latitude  float64
	Longitude float64
	Radius    float64 // meters; 0 = system default
	Proximity string  // "enter" (arrive) or "leave"
}

// String returns a human-readable description of the location trigger.
func (l AlarmLocation) String() string {
	verb := "on arriving at"
	if l.Proximity == "leave" {
		verb = "on leaving"
	}
	name := l.Title
	if name == "" {
		name = fmt.Sprintf("%.4f,%.4f", l.Latitude, l.Longitude)
	}
	s := fmt.Sprintf("%s %s", verb, name)
	if l.Radius > 0 {
		s += fmt.Sprintf(" (within %.0fm)", l.Radius)
	}
	return s
}

// Alarm represents a reminder notification alert.
type Alarm struct {
	AbsoluteDate   *time.Time
	RelativeOffset time.Duration // negative = before due date
	Location       *AlarmLocation
}

// FormatAlarm returns a human-readable description of the alarm.
func (a Alarm) String() string {
	if a.Location != nil {
		return a.Location.String()
	}
	if a.AbsoluteDate != nil {
		return a.AbsoluteDate.Local().Format("Mon Jan 02, 2006 at 3:04 PM")
	}
	d := a.RelativeOffset
	if d == 0 {
		return "at time of event"
	}
	// RelativeOffset is negative (before due date)
	if d < 0 {
		d = -d
	}
	switch {
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute before"
		}
		return fmt.Sprintf("%d minutes before", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour before"
		}
		return fmt.Sprintf("%d hours before", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day before"
		}
		return fmt.Sprintf("%d days before", days)
	}
}

// RecurrenceFrequency defines how often a reminder repeats.
type RecurrenceFrequency int

const (
	FrequencyDaily   RecurrenceFrequency = 0
	FrequencyWeekly  RecurrenceFrequency = 1
	FrequencyMonthly RecurrenceFrequency = 2
	FrequencyYearly  RecurrenceFrequency = 3
)

// RecurrenceRule defines how a reminder repeats.
type RecurrenceRule struct {
	Frequency      RecurrenceFrequency
	Interval       int
	DaysOfWeek     []string // e.g., ["Mon", "Wed", "Fri"] — display names
	DaysOfWeekNums []int    // eventkit weekday numbers (1=Sun..7=Sat) — for creation
	DaysOfMonth    []int    // e.g., [1, 15] — for monthly rules
}

// FormatRecurrence returns a human-readable recurrence description.
func (r RecurrenceRule) String() string {
	freq := ""
	switch r.Frequency {
	case FrequencyDaily:
		if r.Interval == 1 {
			freq = "every day"
		} else {
			freq = fmt.Sprintf("every %d days", r.Interval)
		}
	case FrequencyWeekly:
		if r.Interval == 1 {
			freq = "every week"
		} else {
			freq = fmt.Sprintf("every %d weeks", r.Interval)
		}
	case FrequencyMonthly:
		if r.Interval == 1 {
			freq = "every month"
		} else {
			freq = fmt.Sprintf("every %d months", r.Interval)
		}
	case FrequencyYearly:
		if r.Interval == 1 {
			freq = "every year"
		} else {
			freq = fmt.Sprintf("every %d years", r.Interval)
		}
	}
	if len(r.DaysOfWeek) > 0 {
		freq += " on " + strings.Join(r.DaysOfWeek, ", ")
	} else if len(r.DaysOfMonth) > 0 {
		days := make([]string, len(r.DaysOfMonth))
		for i, d := range r.DaysOfMonth {
			days[i] = fmt.Sprintf("%d", d)
		}
		freq += " on day " + strings.Join(days, ", ")
	}
	return freq
}

// Reminder represents a single reminder item.
type Reminder struct {
	ID               string
	Name             string
	Body             string
	ListName         string
	DueDate          *time.Time
	AllDayDueDate    *time.Time
	RemindMeDate     *time.Time
	CompletionDate   *time.Time
	CreationDate     *time.Time
	ModificationDate *time.Time
	Priority         Priority
	Flagged          bool
	Completed        bool
	URL              string // native EventKit URL field (backwards compat: extracted from body if empty)
	Tags             []string
	Recurring        bool
	RecurrenceRules  []RecurrenceRule
	HasAlarms        bool
	Alarms           []Alarm
}

// List represents a Reminders list.
type List struct {
	ID    string
	Name  string
	Color string
	Count int // number of reminders in the list
	// IsShared is true if the list is shared with other participants.
	IsShared bool
	// SharedToMe is true if someone else shared the list with the current user.
	SharedToMe bool
	// IsOwnedByMe is true if the current user owns the list.
	IsOwnedByMe bool
}

// ListFilter specifies criteria for filtering reminders when listing.
type ListFilter struct {
	ListName    string
	Completed   *bool
	Flagged     *bool
	DueBefore   *time.Time
	DueAfter    *time.Time
	SearchQuery string
	Tags        []string
	PriorityMin *Priority
}
