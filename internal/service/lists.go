//go:build darwin

package service

import (
	"fmt"

	"github.com/BRO3886/go-eventkit/reminders"
	"github.com/BRO3886/rem/internal/reminder"
)

// ListService provides operations for reminder lists.
// Uses go-eventkit for all operations (reads and writes).
// AppleScript is only used for querying the default list name.
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

// findListByName looks up a list by name and returns the go-eventkit List.
func (s *ListService) findListByName(name string) (*reminders.List, error) {
	ekLists, err := s.client.Lists()
	if err != nil {
		return nil, fmt.Errorf("failed to get lists: %w", err)
	}

	for _, l := range ekLists {
		if l.Title == name {
			return &l, nil
		}
	}

	return nil, fmt.Errorf("list not found: %s", name)
}

// defaultSource discovers the default source name from existing lists.
// Falls back to "iCloud" if no lists exist.
func (s *ListService) defaultSource() (string, error) {
	ekLists, err := s.client.Lists()
	if err != nil {
		return "", fmt.Errorf("failed to get lists: %w", err)
	}

	for _, l := range ekLists {
		if l.Source != "" {
			return l.Source, nil
		}
	}

	return "iCloud", nil
}

// CreateList creates a new reminder list via go-eventkit.
// The list is created in the default source (discovered from existing lists).
func (s *ListService) CreateList(name string) (*reminder.List, error) {
	if name == "" {
		return nil, fmt.Errorf("list name is required")
	}

	source, err := s.defaultSource()
	if err != nil {
		return nil, err
	}

	created, err := s.client.CreateList(reminders.CreateListInput{
		Title:  name,
		Source: source,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create list: %w", err)
	}

	return fromEventKitList(created), nil
}

// RenameList renames an existing list via go-eventkit.
func (s *ListService) RenameList(oldName, newName string) error {
	ekList, err := s.findListByName(oldName)
	if err != nil {
		return err
	}

	if ekList.ReadOnly {
		return fmt.Errorf("cannot rename list '%s': list is immutable", oldName)
	}

	_, err = s.client.UpdateList(ekList.ID, reminders.UpdateListInput{
		Title: &newName,
	})
	if err != nil {
		return fmt.Errorf("failed to rename list: %w", err)
	}

	return nil
}

// DeleteList deletes a list by name via go-eventkit.
func (s *ListService) DeleteList(name string) error {
	ekList, err := s.findListByName(name)
	if err != nil {
		return err
	}

	if ekList.ReadOnly {
		return fmt.Errorf("cannot delete list '%s': list is immutable", name)
	}

	if err := s.client.DeleteList(ekList.ID); err != nil {
		return fmt.Errorf("failed to delete list: %w", err)
	}

	return nil
}

// GetDefaultListName returns the name of the default reminder list via AppleScript.
// EventKit does not expose which list is the "default" list, so AppleScript is used.
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
