//go:build darwin

package applescript

import (
	"encoding/json"
	"fmt"

	"github.com/BRO3886/rem/internal/eventkit"
	"github.com/BRO3886/rem/internal/reminder"
)

// ListService provides operations for reminder lists.
type ListService struct {
	exec *Executor
}

// NewListService creates a new ListService.
func NewListService(exec *Executor) *ListService {
	return &ListService{exec: exec}
}

type helperList struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// GetLists returns all reminder lists using EventKit via cgo.
func (s *ListService) GetLists() ([]*reminder.List, error) {
	output, err := eventkit.FetchLists()
	if err != nil {
		return nil, fmt.Errorf("failed to get lists: %w", err)
	}

	if output == "" || output == "[]" {
		return []*reminder.List{}, nil
	}

	var helperLists []helperList
	if err := json.Unmarshal([]byte(output), &helperLists); err != nil {
		return nil, fmt.Errorf("failed to parse lists: %w", err)
	}

	lists := make([]*reminder.List, 0, len(helperLists))
	for _, hl := range helperLists {
		lists = append(lists, &reminder.List{
			ID:    hl.ID,
			Name:  hl.Name,
			Count: hl.Count,
		})
	}

	return lists, nil
}

// GetList returns a single list by name.
func (s *ListService) GetList(name string) (*reminder.List, error) {
	lists, err := s.GetLists()
	if err != nil {
		return nil, err
	}

	for _, l := range lists {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, fmt.Errorf("list not found: %s", name)
}

// CreateList creates a new reminder list.
func (s *ListService) CreateList(name string) (*reminder.List, error) {
	if name == "" {
		return nil, fmt.Errorf("list name is required")
	}

	script := fmt.Sprintf(`tell application "Reminders"
	set newList to make new list with properties {name:"%s"}
	return id of newList
end tell`, EscapeString(name))

	id, err := s.exec.Run(script)
	if err != nil {
		return nil, fmt.Errorf("failed to create list: %w", err)
	}

	return &reminder.List{
		ID:   id,
		Name: name,
	}, nil
}

// RenameList renames an existing list.
func (s *ListService) RenameList(oldName, newName string) error {
	script := fmt.Sprintf(`tell application "Reminders"
	set name of list "%s" to "%s"
end tell`, EscapeString(oldName), EscapeString(newName))

	_, err := s.exec.Run(script)
	if err != nil {
		return fmt.Errorf("failed to rename list: %w", err)
	}

	return nil
}

// DeleteList deletes a list by name.
func (s *ListService) DeleteList(name string) error {
	script := fmt.Sprintf(`tell application "Reminders"
	delete list "%s"
end tell`, EscapeString(name))

	_, err := s.exec.Run(script)
	if err != nil {
		return fmt.Errorf("failed to delete list (this may not work on all macOS versions): %w", err)
	}

	return nil
}

// GetDefaultListName returns the name of the default reminder list.
func (s *ListService) GetDefaultListName() (string, error) {
	output, err := s.exec.Run(`tell application "Reminders" to get name of default list`)
	if err != nil {
		return "", fmt.Errorf("failed to get default list: %w", err)
	}
	return output, nil
}
