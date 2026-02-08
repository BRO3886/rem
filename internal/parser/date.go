package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseDate parses a natural language or formatted date string into a time.Time.
func ParseDate(input string) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	now := time.Now()

	// Try standard formats first
	if t, err := tryStandardFormats(input); err == nil {
		return t, nil
	}

	// Try natural language parsing
	lower := strings.ToLower(input)

	// Handle "today", "tomorrow", "yesterday"
	switch {
	case lower == "today":
		return todayAt(now, 9, 0), nil
	case lower == "tomorrow":
		return todayAt(now.AddDate(0, 0, 1), 9, 0), nil
	case lower == "yesterday":
		return todayAt(now.AddDate(0, 0, -1), 9, 0), nil
	}

	// Handle "in X hours/minutes/days/weeks"
	if t, err := parseRelative(lower, now); err == nil {
		return t, nil
	}

	// Handle "next monday", "next tuesday", etc.
	if t, err := parseNextWeekday(lower, now); err == nil {
		return t, nil
	}

	// Handle "today at 5pm", "tomorrow at 3:30pm"
	if t, err := parseDateWithTime(lower, now); err == nil {
		return t, nil
	}

	// Handle standalone time like "5pm", "17:00", "3:30pm"
	if t, err := parseTimeOnly(lower, now); err == nil {
		return t, nil
	}

	// Handle "next week", "next month"
	switch {
	case lower == "next week":
		return todayAt(now.AddDate(0, 0, 7), 9, 0), nil
	case lower == "next month":
		return todayAt(now.AddDate(0, 1, 0), 9, 0), nil
	case lower == "end of day", lower == "eod":
		return todayAt(now, 17, 0), nil
	case lower == "end of week", lower == "eow":
		daysUntilFriday := (5 - int(now.Weekday()) + 7) % 7
		if daysUntilFriday == 0 {
			daysUntilFriday = 7
		}
		return todayAt(now.AddDate(0, 0, daysUntilFriday), 17, 0), nil
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %q", input)
}

func tryStandardFormats(input string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 3:04PM",
		"2006-01-02 3:04pm",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"01/02/2006",
		"01/02/2006 15:04",
		"01/02/2006 3:04PM",
		"Jan 2, 2006",
		"Jan 2, 2006 3:04PM",
		"Jan 2, 2006 15:04",
		"January 2, 2006",
		"January 2, 2006 3:04PM",
		"January 2, 2006 15:04",
		"2 Jan 2006",
		"02 Jan 2006",
	}

	for _, f := range formats {
		t, err := time.ParseInLocation(f, input, time.Now().Location())
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("no standard format matched")
}

var relativePattern = regexp.MustCompile(`^in\s+(\d+)\s+(minute|minutes|min|mins|hour|hours|hr|hrs|day|days|week|weeks|month|months)$`)

func parseRelative(input string, now time.Time) (time.Time, error) {
	matches := relativePattern.FindStringSubmatch(input)
	if matches == nil {
		return time.Time{}, fmt.Errorf("not a relative date")
	}

	amount, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch {
	case strings.HasPrefix(unit, "min"):
		return now.Add(time.Duration(amount) * time.Minute), nil
	case strings.HasPrefix(unit, "hour"), strings.HasPrefix(unit, "hr"):
		return now.Add(time.Duration(amount) * time.Hour), nil
	case strings.HasPrefix(unit, "day"):
		return now.AddDate(0, 0, amount), nil
	case strings.HasPrefix(unit, "week"):
		return now.AddDate(0, 0, amount*7), nil
	case strings.HasPrefix(unit, "month"):
		return now.AddDate(0, amount, 0), nil
	}

	return time.Time{}, fmt.Errorf("unknown unit: %s", unit)
}

var weekdays = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
	"sun":       time.Sunday,
	"mon":       time.Monday,
	"tue":       time.Tuesday,
	"wed":       time.Wednesday,
	"thu":       time.Thursday,
	"fri":       time.Friday,
	"sat":       time.Saturday,
}

func parseNextWeekday(input string, now time.Time) (time.Time, error) {
	// Match "next monday", "next tuesday at 2pm", etc.
	parts := strings.Fields(input)
	if len(parts) < 2 || parts[0] != "next" {
		return time.Time{}, fmt.Errorf("not a next weekday expression")
	}

	dayName := parts[1]
	targetDay, ok := weekdays[dayName]
	if !ok {
		return time.Time{}, fmt.Errorf("unknown weekday: %s", dayName)
	}

	daysAhead := int(targetDay) - int(now.Weekday())
	if daysAhead <= 0 {
		daysAhead += 7
	}

	targetDate := now.AddDate(0, 0, daysAhead)
	result := todayAt(targetDate, 9, 0) // default 9 AM

	// Check for "at" time specification
	if len(parts) >= 4 && parts[2] == "at" {
		timeStr := strings.Join(parts[3:], " ")
		hour, min, err := parseTimeStr(timeStr)
		if err == nil {
			result = todayAt(targetDate, hour, min)
		}
	}

	return result, nil
}

func parseDateWithTime(input string, now time.Time) (time.Time, error) {
	// Handle "today at 5pm", "tomorrow at 3:30pm"
	parts := strings.SplitN(input, " at ", 2)
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("not a date with time")
	}

	datePart := strings.TrimSpace(parts[0])
	timePart := strings.TrimSpace(parts[1])

	var baseDate time.Time
	switch datePart {
	case "today":
		baseDate = now
	case "tomorrow":
		baseDate = now.AddDate(0, 0, 1)
	case "yesterday":
		baseDate = now.AddDate(0, 0, -1)
	default:
		return time.Time{}, fmt.Errorf("unknown date base: %s", datePart)
	}

	hour, min, err := parseTimeStr(timePart)
	if err != nil {
		return time.Time{}, err
	}

	return todayAt(baseDate, hour, min), nil
}

func parseTimeOnly(input string, now time.Time) (time.Time, error) {
	hour, min, err := parseTimeStr(input)
	if err != nil {
		return time.Time{}, err
	}

	result := todayAt(now, hour, min)
	// If the time has already passed today, schedule for tomorrow
	if result.Before(now) {
		result = result.AddDate(0, 0, 1)
	}

	return result, nil
}

var timePatterns = []struct {
	re     *regexp.Regexp
	parser func([]string) (int, int, error)
}{
	{
		// "5pm", "5PM", "5 pm"
		re: regexp.MustCompile(`^(\d{1,2})\s*(am|pm|AM|PM)$`),
		parser: func(m []string) (int, int, error) {
			h, _ := strconv.Atoi(m[1])
			return convertTo24(h, 0, strings.ToLower(m[2]))
		},
	},
	{
		// "5:30pm", "5:30 PM"
		re: regexp.MustCompile(`^(\d{1,2}):(\d{2})\s*(am|pm|AM|PM)$`),
		parser: func(m []string) (int, int, error) {
			h, _ := strconv.Atoi(m[1])
			min, _ := strconv.Atoi(m[2])
			return convertTo24(h, min, strings.ToLower(m[3]))
		},
	},
	{
		// "17:00", "9:30"
		re: regexp.MustCompile(`^(\d{1,2}):(\d{2})$`),
		parser: func(m []string) (int, int, error) {
			h, _ := strconv.Atoi(m[1])
			min, _ := strconv.Atoi(m[2])
			if h > 23 || min > 59 {
				return 0, 0, fmt.Errorf("invalid time")
			}
			return h, min, nil
		},
	},
}

func parseTimeStr(s string) (int, int, error) {
	s = strings.TrimSpace(s)
	for _, p := range timePatterns {
		matches := p.re.FindStringSubmatch(s)
		if matches != nil {
			return p.parser(matches)
		}
	}
	return 0, 0, fmt.Errorf("unable to parse time: %q", s)
}

func convertTo24(hour, min int, period string) (int, int, error) {
	if hour < 1 || hour > 12 {
		return 0, 0, fmt.Errorf("invalid hour: %d", hour)
	}
	if period == "am" {
		if hour == 12 {
			hour = 0
		}
	} else {
		if hour != 12 {
			hour += 12
		}
	}
	return hour, min, nil
}

func todayAt(base time.Time, hour, min int) time.Time {
	return time.Date(base.Year(), base.Month(), base.Day(), hour, min, 0, 0, base.Location())
}
