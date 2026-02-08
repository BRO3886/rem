package applescript

import (
	"encoding/json"
	"fmt"

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

type jxaList struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Count int    `json:"count"`
}

// GetLists returns all reminder lists.
func (s *ListService) GetLists() ([]*reminder.List, error) {
	script := `
const app = Application('Reminders');
const lists = app.lists();
const result = lists.map(l => {
	const props = l.properties();
	return {
		id: props.id,
		name: props.name,
		color: props.color || '',
		count: l.reminders.length,
	};
});
JSON.stringify(result);`

	output, err := s.exec.RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get lists: %w", err)
	}

	if output == "" || output == "[]" {
		return []*reminder.List{}, nil
	}

	var jxaLists []jxaList
	if err := json.Unmarshal([]byte(output), &jxaLists); err != nil {
		return nil, fmt.Errorf("failed to parse lists: %w", err)
	}

	lists := make([]*reminder.List, 0, len(jxaLists))
	for _, jl := range jxaLists {
		lists = append(lists, &reminder.List{
			ID:    jl.ID,
			Name:  jl.Name,
			Color: jl.Color,
			Count: jl.Count,
		})
	}

	return lists, nil
}

// GetList returns a single list by name.
func (s *ListService) GetList(name string) (*reminder.List, error) {
	script := fmt.Sprintf(`
const app = Application('Reminders');
const l = app.lists.byName('%s');
const props = l.properties();
JSON.stringify({
	id: props.id,
	name: props.name,
	color: props.color || '',
	count: l.reminders.length,
});`, EscapeJXA(name))

	output, err := s.exec.RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get list '%s': %w", name, err)
	}

	var jl jxaList
	if err := json.Unmarshal([]byte(output), &jl); err != nil {
		return nil, fmt.Errorf("failed to parse list: %w", err)
	}

	return &reminder.List{
		ID:    jl.ID,
		Name:  jl.Name,
		Color: jl.Color,
		Count: jl.Count,
	}, nil
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
