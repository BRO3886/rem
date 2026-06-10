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

// parseLocationAlarm builds a geofence alarm from the --location, --radius,
// and --on-arrive/--on-leave flag values. The location format is "lat,lng"
// (e.g. "37.3318,-122.0312"). Proximity defaults to arrive when neither flag
// is set; setting both is an error. Radius is in meters; 0 means the system
// default. The coordinate string doubles as the location title so
// Reminders.app has something to display.
func parseLocationAlarm(location string, radius float64, onArrive, onLeave bool) (reminder.Alarm, error) {
	if onArrive && onLeave {
		return reminder.Alarm{}, fmt.Errorf("--on-arrive and --on-leave are mutually exclusive")
	}

	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		return reminder.Alarm{}, fmt.Errorf("invalid location: %q — use \"lat,lng\" (e.g. \"37.3318,-122.0312\")", location)
	}
	lat, latErr := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, lngErr := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if latErr != nil || lngErr != nil {
		return reminder.Alarm{}, fmt.Errorf("invalid location: %q — use \"lat,lng\" (e.g. \"37.3318,-122.0312\")", location)
	}
	if lat < -90 || lat > 90 {
		return reminder.Alarm{}, fmt.Errorf("invalid latitude %v: must be between -90 and 90", lat)
	}
	if lng < -180 || lng > 180 {
		return reminder.Alarm{}, fmt.Errorf("invalid longitude %v: must be between -180 and 180", lng)
	}
	if radius < 0 {
		return reminder.Alarm{}, fmt.Errorf("invalid radius %v: must be >= 0 meters", radius)
	}

	proximity := "enter"
	if onLeave {
		proximity = "leave"
	}

	return reminder.Alarm{
		Location: &reminder.AlarmLocation{
			Title:     fmt.Sprintf("%.4f, %.4f", lat, lng),
			Latitude:  lat,
			Longitude: lng,
			Radius:    radius,
			Proximity: proximity,
		},
	}, nil
}

// splitAlarmsByTrigger separates a reminder's alarms into time-based and
// location-based buckets so --remind-me and --location can each replace only
// their own kind.
func splitAlarmsByTrigger(alarms []reminder.Alarm) (timeAlarms, locationAlarms []reminder.Alarm) {
	for _, a := range alarms {
		if a.Location != nil {
			locationAlarms = append(locationAlarms, a)
		} else {
			timeAlarms = append(timeAlarms, a)
		}
	}
	return timeAlarms, locationAlarms
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

func normalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	tag = strings.TrimPrefix(tag, "#")
	return strings.Trim(tag, ".,;:!?)]}\"'")
}

func tagsFromTitle(title string) []string {
	fields := strings.Fields(title)
	var tags []string
	seen := map[string]bool{}
	for _, field := range fields {
		if !strings.HasPrefix(field, "#") {
			continue
		}
		tag := normalizeTag(field)
		if tag == "" || isNumeric(tag) {
			continue
		}
		key := strings.ToLower(tag)
		if seen[key] {
			continue
		}
		seen[key] = true
		tags = append(tags, tag)
	}
	return tags
}

func parseTagList(input string) []string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	parts := strings.Split(input, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := normalizeTag(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func mergeTagUpdates(existing, titleTags []string, addComma string, removeComma string) []string {
	add := append([]string(nil), titleTags...)
	add = append(add, parseTagList(addComma)...)
	remove := parseTagList(removeComma)
	return mergeTags(existing, add, remove)
}

func mergeTagInputs(titleTags []string, commaSeparated string) []string {
	all := append([]string(nil), titleTags...)
	all = append(all, parseTagList(commaSeparated)...)
	return mergeTags(nil, all, nil)
}

// isNumeric filters out pure-number fragments like #42 (issue references).
// Tags containing any non-digit (e.g. #5min, #q4) pass through.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func mergeTags(existing, add, remove []string) []string {
	removed := make(map[string]bool, len(remove))
	for _, tag := range remove {
		tag = strings.ToLower(normalizeTag(tag))
		if tag != "" {
			removed[tag] = true
		}
	}

	seen := make(map[string]bool, len(existing)+len(add))
	var merged []string
	for _, tag := range existing {
		tag = normalizeTag(tag)
		key := strings.ToLower(tag)
		if tag == "" || removed[key] || seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, tag)
	}
	for _, tag := range add {
		tag = normalizeTag(tag)
		key := strings.ToLower(tag)
		if tag == "" || removed[key] || seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, tag)
	}
	return merged
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
