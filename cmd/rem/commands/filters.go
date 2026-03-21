package commands

import (
	"fmt"
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
//   - Relative offsets: "15m", "1h", "2d", "0" (at time of event)
//   - Absolute times: any input parseable by parseDate (e.g., "tomorrow at 9am")
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
