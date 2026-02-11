//go:build darwin

package service

import (
	"fmt"

	"github.com/BRO3886/go-eventkit/reminders"
	"github.com/BRO3886/rem/internal/reminder"
)

// ListService provides operations for reminder lists.
// Uses go-eventkit for reads, AppleScript for writes (create/rename/delete).
type ListService struct {
	client *reminders.Client
	exec   *Executor
}

// NewListService creates a new ListService.
func NewListService(client *reminders.Client, exec *Executor) *ListService {
	return &ListService{client: client, exec: exec}
}

// GetLists returns all reminder lists via go-eventkit.
func (s *ListService) GetLists() ([]*reminder.List, error) {
	ekLists, err := s.client.Lists()
	if err != nil {
		return nil, fmt.Errorf("failed to get lists: %w", err)
	}

	lists := make([]*reminder.List, 0, len(ekLists))
	for _, l := range ekLists {
		lists = append(lists, fromEventKitList(&l))
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

// CreateList creates a new reminder list via AppleScript.
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

// RenameList renames an existing list via AppleScript.
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

// DeleteList deletes a list by name via AppleScript.
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

// GetDefaultListName returns the name of the default reminder list via AppleScript.
func (s *ListService) GetDefaultListName() (string, error) {
	output, err := s.exec.Run(`tell application "Reminders" to get name of default list`)
	if err != nil {
		return "", fmt.Errorf("failed to get default list: %w", err)
	}
	return output, nil
}

// fromEventKitList converts a go-eventkit List to an internal List.
func fromEventKitList(l *reminders.List) *reminder.List {
	return &reminder.List{
		ID:    l.ID,
		Name:  l.Title,
		Color: l.Color,
		Count: l.Count,
	}
}
