//go:build !darwin

package eventkit

import "fmt"

var errUnsupported = fmt.Errorf("eventkit: only supported on macOS (darwin)")

// FetchLists returns a JSON string containing all reminder lists.
func FetchLists() (string, error) { return "", errUnsupported }

// FetchReminders returns a JSON string of reminders matching the given filters.
func FetchReminders(listName, completedFilter, searchQuery, dueBefore, dueAfter string) (string, error) {
	return "", errUnsupported
}

// GetReminder returns a JSON string for a single reminder by ID or ID prefix.
func GetReminder(targetID string) (string, error) { return "", errUnsupported }
