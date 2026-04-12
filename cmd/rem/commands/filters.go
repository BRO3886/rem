package commands

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BRO3886/go-eventkit/dateparser"
	"github.com/BRO3886/rem/internal/reminder"
)

// parseDate wraps dateparser.ParseDate with rem's default options:
// bare dates at 9am, past times roll to tomorrow, eow skips today.
func parseDate(input string) (time.Time, error) {
	return dateparser.ParseDate(input,
		dateparser.WithDefaultHour(9),
		dateparser.WithSmartTimeRollover(),
		dateparser.WithEOWSkipToday(),
	)
}

// parseAlarm parses an alarm specification. Supports:
//   - Relative offsets: "0m", "15m", "1h", "2d" (before the due date)
//   - Absolute times: any input parseable by parseDate (e.g., "tomorrow at 9am")
//
// Note: a bare "0" (without a unit) is not accepted — use "0m" for an alarm
// at the due time. That's also the default when --due is set and --remind-me
// is not, so passing "0m" explicitly is rarely needed.
func parseAlarm(input string) (reminder.Alarm, error) {
	// Try relative offset first (e.g., "15m", "1h", "2d")
	if d, err := dateparser.ParseAlertDuration(input); err == nil {
		return reminder.Alarm{RelativeOffset: -d}, nil // negative = before due date
	}

	// Try absolute date
	t, err := parseDate(input)
	if err != nil {
		return reminder.Alarm{}, fmt.Errorf("invalid alarm: %q — use a duration (15m, 1h, 2d) or a date/time", input)
	}
	return reminder.Alarm{AbsoluteDate: &t}, nil
}

// buildAlarms decides which alarms to attach to a new reminder based on the
// --due, --remind-me, and --silent flag values. The decision rules are:
//
//  1. If --remind-me is explicitly set, always use that alarm (overrides
//     --silent; an explicit request from the user wins).
//  2. Else if --due is set AND --silent is not, attach a zero-offset alarm
//     (fire at the due time). This matches Apple Reminders.app's default
//     behavior when you create a reminder with a date in the UI.
//  3. Otherwise, no alarms.
//
// Extracted from the RunE closure in add.go so the decision logic is unit
// testable independent of Cobra plumbing and the service layer.
func buildAlarms(hasDueDate bool, remindMe string, silent bool) ([]reminder.Alarm, error) {
	if remindMe != "" {
		a, err := parseAlarm(remindMe)
		if err != nil {
			return nil, err
		}
		return []reminder.Alarm{a}, nil
	}
	if hasDueDate && !silent {
		return []reminder.Alarm{{RelativeOffset: 0}}, nil
	}
	return nil, nil
}

// weekdayNum maps day name strings to eventkit weekday numbers (1=Sun..7=Sat).
var weekdayNum = map[string]int{
	"sun": 1, "sunday": 1,
	"mon": 2, "monday": 2,
	"tue": 3, "tuesday": 3,
	"wed": 4, "wednesday": 4,
	"thu": 5, "thursday": 5,
	"fri": 6, "friday": 6,
	"sat": 7, "saturday": 7,
}

// weekdayName maps eventkit weekday numbers to short display names.
var weekdayName = map[int]string{
	1: "Sun", 2: "Mon", 3: "Tue", 4: "Wed", 5: "Thu", 6: "Fri", 7: "Sat",
}

var everyNPattern = regexp.MustCompile(`^every\s+(\d+)\s+(day|days|week|weeks|month|months|year|years)$`)

// parseRecurrence parses a recurrence specification into a domain RecurrenceRule.
// Supported patterns:
//   - "daily", "every day", "every 2 days"
//   - "weekly", "every week", "every 2 weeks"
//   - "weekly on mon,wed,fri"
//   - "monthly", "every month", "every 2 months"
//   - "monthly on 1,15" (days of month)
//   - "yearly", "every year", "every 2 years"
func parseRecurrence(input string) (reminder.RecurrenceRule, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	// Split "on" clause: "weekly on mon,wed" → base="weekly", onClause="mon,wed"
	base, onClause, _ := strings.Cut(input, " on ")
	base = strings.TrimSpace(base)
	onClause = strings.TrimSpace(onClause)

	// Parse base frequency and interval
	var freq reminder.RecurrenceFrequency
	interval := 1

	switch base {
	case "daily", "every day":
		freq = reminder.FrequencyDaily
	case "weekly", "every week":
		freq = reminder.FrequencyWeekly
	case "monthly", "every month":
		freq = reminder.FrequencyMonthly
	case "yearly", "every year":
		freq = reminder.FrequencyYearly
	default:
		matches := everyNPattern.FindStringSubmatch(base)
		if matches == nil {
			return reminder.RecurrenceRule{}, fmt.Errorf("invalid recurrence: %q — use 'daily', 'weekly', 'monthly', 'yearly', or 'every N days/weeks/months/years'", input)
		}
		interval, _ = strconv.Atoi(matches[1])
		unit := matches[2]
		switch {
		case strings.HasPrefix(unit, "day"):
			freq = reminder.FrequencyDaily
		case strings.HasPrefix(unit, "week"):
			freq = reminder.FrequencyWeekly
		case strings.HasPrefix(unit, "month"):
			freq = reminder.FrequencyMonthly
		case strings.HasPrefix(unit, "year"):
			freq = reminder.FrequencyYearly
		}
	}

	rule := reminder.RecurrenceRule{
		Frequency: freq,
		Interval:  interval,
	}

	// Parse "on" clause
	if onClause != "" {
		parts := strings.Split(onClause, ",")
		switch freq {
		case reminder.FrequencyWeekly:
			for _, p := range parts {
				num, ok := weekdayNum[strings.TrimSpace(p)]
				if !ok {
					return reminder.RecurrenceRule{}, fmt.Errorf("unknown day: %q — use mon, tue, wed, thu, fri, sat, sun", strings.TrimSpace(p))
				}
				rule.DaysOfWeekNums = append(rule.DaysOfWeekNums, num)
				rule.DaysOfWeek = append(rule.DaysOfWeek, weekdayName[num])
			}
		case reminder.FrequencyMonthly:
			for _, p := range parts {
				d, err := strconv.Atoi(strings.TrimSpace(p))
				if err != nil || d < 1 || d > 31 {
					return reminder.RecurrenceRule{}, fmt.Errorf("invalid day of month: %q — use 1-31", strings.TrimSpace(p))
				}
				rule.DaysOfMonth = append(rule.DaysOfMonth, d)
			}
		default:
			return reminder.RecurrenceRule{}, fmt.Errorf("'on' clause not supported for %s frequency", base)
		}
	}

	return rule, nil
}

// completeFilter builds the filter for the complete/uncomplete interactive flow.
// When uncomplete is false (completing), it shows incomplete reminders (Completed=false).
// When uncomplete is true (uncompleting), it shows completed reminders (Completed=true).
func completeFilter(uncomplete bool, listName string, flagged bool) *reminder.ListFilter {
	completed := uncomplete
	filter := &reminder.ListFilter{Completed: &completed}
	if listName != "" {
		filter.ListName = listName
	}
	if !uncomplete && flagged {
		f := true
		filter.Flagged = &f
	}
	return filter
}

// deleteFilter builds the filter for the delete interactive flow.
// Always shows incomplete reminders.
func deleteFilter(listName string, flagged bool) *reminder.ListFilter {
	incomplete := false
	filter := &reminder.ListFilter{Completed: &incomplete}
	if listName != "" {
		filter.ListName = listName
	}
	if flagged {
		f := true
		filter.Flagged = &f
	}
	return filter
}

// flagFilter builds the filter for the flag interactive flow.
// Shows incomplete reminders, optionally filtered by list.
func flagFilter(listName string) *reminder.ListFilter {
	incomplete := false
	filter := &reminder.ListFilter{Completed: &incomplete}
	if listName != "" {
		filter.ListName = listName
	}
	return filter
}

// unflagFilter builds the filter for the unflag interactive flow.
// Shows incomplete AND flagged reminders, optionally filtered by list.
func unflagFilter(listName string) *reminder.ListFilter {
	flagged := true
	incomplete := false
	filter := &reminder.ListFilter{Completed: &incomplete, Flagged: &flagged}
	if listName != "" {
		filter.ListName = listName
	}
	return filter
}
