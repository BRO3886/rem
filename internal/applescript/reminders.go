package applescript

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/BRO3886/rem/internal/reminder"
)

// ReminderService provides operations for reminders using AppleScript and JXA.
type ReminderService struct {
	exec *Executor
}

// NewReminderService creates a new ReminderService.
func NewReminderService(exec *Executor) *ReminderService {
	return &ReminderService{exec: exec}
}

// CreateReminder creates a new reminder and returns its ID.
func (s *ReminderService) CreateReminder(r *reminder.Reminder) (string, error) {
	if r.Name == "" {
		return "", fmt.Errorf("reminder name is required")
	}

	listName := r.ListName
	if listName == "" {
		defaultList, err := s.exec.Run(`tell application "Reminders" to get name of default list`)
		if err != nil {
			return "", fmt.Errorf("failed to get default list: %w", err)
		}
		listName = defaultList
	}

	var dateSetup strings.Builder
	props := []string{fmt.Sprintf(`name:"%s"`, EscapeString(r.Name))}

	body := r.Body
	if r.URL != "" {
		if body != "" {
			body = body + "\n\nURL: " + r.URL
		} else {
			body = "URL: " + r.URL
		}
	}
	if body != "" {
		props = append(props, fmt.Sprintf(`body:"%s"`, EscapeString(body)))
	}

	if r.DueDate != nil {
		dateSetup.WriteString(FormatDateInline("dueDate", *r.DueDate) + "\n")
		props = append(props, "due date:dueDate")
	}

	if r.AllDayDueDate != nil {
		dateSetup.WriteString(FormatDateInline("alldayDate", *r.AllDayDueDate) + "\n")
		props = append(props, "allday due date:alldayDate")
	}

	if r.RemindMeDate != nil {
		dateSetup.WriteString(FormatDateInline("remindDate", *r.RemindMeDate) + "\n")
		props = append(props, "remind me date:remindDate")
	}

	if r.Priority != reminder.PriorityNone {
		props = append(props, fmt.Sprintf("priority:%d", r.Priority))
	}

	if r.Flagged {
		props = append(props, "flagged:true")
	}

	script := fmt.Sprintf(`%s
tell application "Reminders"
	tell list "%s"
		set newReminder to make new reminder at end with properties {%s}
		return id of newReminder
	end tell
end tell`, dateSetup.String(), EscapeString(listName), strings.Join(props, ", "))

	id, err := s.exec.Run(script)
	if err != nil {
		return "", fmt.Errorf("failed to create reminder: %w", err)
	}

	return id, nil
}

// jxaReminder is the JSON structure returned by JXA scripts.
type jxaReminder struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Body           *string `json:"body"`
	ListName       string  `json:"listName"`
	Completed      bool    `json:"completed"`
	Flagged        bool    `json:"flagged"`
	Priority       int     `json:"priority"`
	DueDate        *string `json:"dueDate"`
	RemindMeDate   *string `json:"remindMeDate"`
	CompletionDate *string `json:"completionDate"`
	CreationDate   *string `json:"creationDate"`
	ModDate        *string `json:"modDate"`
}

func jxaToReminder(j *jxaReminder) *reminder.Reminder {
	r := &reminder.Reminder{
		ID:        j.ID,
		Name:      j.Name,
		ListName:  j.ListName,
		Completed: j.Completed,
		Flagged:   j.Flagged,
		Priority:  reminder.Priority(j.Priority),
	}

	if j.Body != nil {
		r.Body = *j.Body
	}

	if j.DueDate != nil {
		r.DueDate = parseISODate(*j.DueDate)
	}
	if j.RemindMeDate != nil {
		r.RemindMeDate = parseISODate(*j.RemindMeDate)
	}
	if j.CompletionDate != nil {
		r.CompletionDate = parseISODate(*j.CompletionDate)
	}
	if j.CreationDate != nil {
		r.CreationDate = parseISODate(*j.CreationDate)
	}
	if j.ModDate != nil {
		r.ModificationDate = parseISODate(*j.ModDate)
	}

	r.URL = extractURL(r.Body)
	return r
}

func parseISODate(s string) *time.Time {
	if s == "" || s == "null" {
		return nil
	}
	// JXA Date.toISOString() returns UTC ISO 8601
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05.000Z", s)
		if err != nil {
			return nil
		}
	}
	local := t.Local()
	return &local
}

// GetReminder retrieves a single reminder by ID.
func (s *ReminderService) GetReminder(id string) (*reminder.Reminder, error) {
	script := fmt.Sprintf(`
const app = Application('Reminders');
const r = app.reminders.byId('%s');
const props = r.properties();
const container = r.container();
JSON.stringify({
	id: props.id,
	name: props.name,
	body: props.body || null,
	listName: container.name(),
	completed: props.completed,
	flagged: props.flagged,
	priority: props.priority,
	dueDate: props.dueDate ? props.dueDate.toISOString() : null,
	remindMeDate: props.remindMeDate ? props.remindMeDate.toISOString() : null,
	completionDate: props.completionDate ? props.completionDate.toISOString() : null,
	creationDate: props.creationDate ? props.creationDate.toISOString() : null,
	modDate: props.modificationDate ? props.modificationDate.toISOString() : null,
});`, EscapeJXA(id))

	output, err := s.exec.RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("failed to get reminder: %w", err)
	}

	var jr jxaReminder
	if err := json.Unmarshal([]byte(output), &jr); err != nil {
		return nil, fmt.Errorf("failed to parse reminder: %w", err)
	}

	return jxaToReminder(&jr), nil
}

// jxaBulkResult is the columnar JSON structure returned by the bulk JXA list script.
type jxaBulkResult struct {
	IDs            []string  `json:"ids"`
	Names          []string  `json:"names"`
	Bodies         []*string `json:"bodies"`
	ListNames      []string  `json:"listNames"`
	Completed      []bool    `json:"completed"`
	Flagged        []bool    `json:"flagged"`
	Priorities     []int     `json:"priorities"`
	DueDates       []*string `json:"dueDates"`
	RemindMeDates  []*string `json:"remindMeDates"`
	CompletionDates []*string `json:"completionDates"`
	CreationDates  []*string `json:"creationDates"`
	ModDates       []*string `json:"modDates"`
}

// ListReminders returns reminders matching the given filter.
func (s *ReminderService) ListReminders(filter *reminder.ListFilter) ([]*reminder.Reminder, error) {
	// Determine source - specific list or all lists
	useSpecificList := filter != nil && filter.ListName != ""

	var script string
	if useSpecificList {
		// For a specific list, use bulk property access on the list's reminders collection
		script = fmt.Sprintf(`
const app = Application('Reminders');
const list = app.lists.byName('%s');
const rems = list.reminders;
const ids = rems.id();
const names = rems.name();
const bodies = rems.body();
const completed = rems.completed();
const flagged = rems.flagged();
const priorities = rems.priority();
const dueDates = rems.dueDate();
const remindMeDates = rems.remindMeDate();
const completionDates = rems.completionDate();
const creationDates = rems.creationDate();
const modDates = rems.modificationDate();
const listName = '%s';
const n = ids.length;
const listNames = new Array(n).fill(listName);
const fmtDate = d => d ? d.toISOString() : null;
JSON.stringify({
	ids, names,
	bodies: bodies.map(b => b || null),
	listNames,
	completed, flagged, priorities,
	dueDates: dueDates.map(fmtDate),
	remindMeDates: remindMeDates.map(fmtDate),
	completionDates: completionDates.map(fmtDate),
	creationDates: creationDates.map(fmtDate),
	modDates: modDates.map(fmtDate),
});`, EscapeJXA(filter.ListName), EscapeJXA(filter.ListName))
	} else {
		// For all lists, iterate through each list and use bulk access per list
		script = `
const app = Application('Reminders');
const lists = app.lists();
let allIds=[], allNames=[], allBodies=[], allListNames=[];
let allCompleted=[], allFlagged=[], allPriorities=[];
let allDueDates=[], allRemindMeDates=[], allCompletionDates=[];
let allCreationDates=[], allModDates=[];
const fmtDate = d => d ? d.toISOString() : null;
for (const list of lists) {
	const ln = list.name();
	const rems = list.reminders;
	const n = rems.length;
	if (n === 0) continue;
	const ids = rems.id();
	const names = rems.name();
	const bodies = rems.body();
	const completed = rems.completed();
	const flagged = rems.flagged();
	const priorities = rems.priority();
	const dueDates = rems.dueDate();
	const remindMeDates = rems.remindMeDate();
	const completionDates = rems.completionDate();
	const creationDates = rems.creationDate();
	const modDates = rems.modificationDate();
	for (let i = 0; i < n; i++) {
		allIds.push(ids[i]);
		allNames.push(names[i]);
		allBodies.push(bodies[i] || null);
		allListNames.push(ln);
		allCompleted.push(completed[i]);
		allFlagged.push(flagged[i]);
		allPriorities.push(priorities[i]);
		allDueDates.push(fmtDate(dueDates[i]));
		allRemindMeDates.push(fmtDate(remindMeDates[i]));
		allCompletionDates.push(fmtDate(completionDates[i]));
		allCreationDates.push(fmtDate(creationDates[i]));
		allModDates.push(fmtDate(modDates[i]));
	}
}
JSON.stringify({
	ids:allIds, names:allNames, bodies:allBodies, listNames:allListNames,
	completed:allCompleted, flagged:allFlagged, priorities:allPriorities,
	dueDates:allDueDates, remindMeDates:allRemindMeDates,
	completionDates:allCompletionDates, creationDates:allCreationDates,
	modDates:allModDates,
});`
	}

	output, err := s.exec.RunJXA(script)
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}

	if output == "" {
		return []*reminder.Reminder{}, nil
	}

	var bulk jxaBulkResult
	if err := json.Unmarshal([]byte(output), &bulk); err != nil {
		return nil, fmt.Errorf("failed to parse reminders: %w", err)
	}

	var reminders []*reminder.Reminder
	for i := range bulk.IDs {
		r := &reminder.Reminder{
			ID:        bulk.IDs[i],
			Name:      bulk.Names[i],
			ListName:  bulk.ListNames[i],
			Completed: bulk.Completed[i],
			Flagged:   bulk.Flagged[i],
			Priority:  reminder.Priority(bulk.Priorities[i]),
		}
		if i < len(bulk.Bodies) && bulk.Bodies[i] != nil {
			r.Body = *bulk.Bodies[i]
		}
		if i < len(bulk.DueDates) && bulk.DueDates[i] != nil {
			r.DueDate = parseISODate(*bulk.DueDates[i])
		}
		if i < len(bulk.RemindMeDates) && bulk.RemindMeDates[i] != nil {
			r.RemindMeDate = parseISODate(*bulk.RemindMeDates[i])
		}
		if i < len(bulk.CompletionDates) && bulk.CompletionDates[i] != nil {
			r.CompletionDate = parseISODate(*bulk.CompletionDates[i])
		}
		if i < len(bulk.CreationDates) && bulk.CreationDates[i] != nil {
			r.CreationDate = parseISODate(*bulk.CreationDates[i])
		}
		if i < len(bulk.ModDates) && bulk.ModDates[i] != nil {
			r.ModificationDate = parseISODate(*bulk.ModDates[i])
		}
		r.URL = extractURL(r.Body)

		// Apply filters
		if filter != nil {
			if filter.Completed != nil {
				if *filter.Completed != r.Completed {
					continue
				}
			}
			if filter.Flagged != nil && *filter.Flagged && !r.Flagged {
				continue
			}
			if filter.DueBefore != nil {
				if r.DueDate == nil || r.DueDate.After(*filter.DueBefore) {
					continue
				}
			}
			if filter.DueAfter != nil {
				if r.DueDate == nil || r.DueDate.Before(*filter.DueAfter) {
					continue
				}
			}
			if filter.SearchQuery != "" {
				query := strings.ToLower(filter.SearchQuery)
				nameMatch := strings.Contains(strings.ToLower(r.Name), query)
				bodyMatch := strings.Contains(strings.ToLower(r.Body), query)
				if !nameMatch && !bodyMatch {
					continue
				}
			}
		}

		reminders = append(reminders, r)
	}

	return reminders, nil
}

// UpdateReminder updates properties of an existing reminder.
func (s *ReminderService) UpdateReminder(id string, updates map[string]any) error {
	var dateSetup strings.Builder
	var setStatements []string

	for key, value := range updates {
		switch key {
		case "name":
			setStatements = append(setStatements, fmt.Sprintf(`set name of r to "%s"`, EscapeString(value.(string))))
		case "body":
			setStatements = append(setStatements, fmt.Sprintf(`set body of r to "%s"`, EscapeString(value.(string))))
		case "due_date":
			if value == nil {
				setStatements = append(setStatements, `set due date of r to missing value`)
			} else {
				t := value.(time.Time)
				dateSetup.WriteString(FormatDateInline("newDueDate", t) + "\n")
				setStatements = append(setStatements, `set due date of r to newDueDate`)
			}
		case "remind_me_date":
			if value == nil {
				setStatements = append(setStatements, `set remind me date of r to missing value`)
			} else {
				t := value.(time.Time)
				dateSetup.WriteString(FormatDateInline("newRemindDate", t) + "\n")
				setStatements = append(setStatements, `set remind me date of r to newRemindDate`)
			}
		case "priority":
			setStatements = append(setStatements, fmt.Sprintf(`set priority of r to %d`, value.(reminder.Priority)))
		case "flagged":
			if value.(bool) {
				setStatements = append(setStatements, `set flagged of r to true`)
			} else {
				setStatements = append(setStatements, `set flagged of r to false`)
			}
		case "completed":
			if value.(bool) {
				setStatements = append(setStatements, `set completed of r to true`)
			} else {
				setStatements = append(setStatements, `set completed of r to false`)
			}
		}
	}

	if len(setStatements) == 0 {
		return nil
	}

	script := fmt.Sprintf(`%s
tell application "Reminders"
	set r to first reminder whose id is "%s"
	%s
end tell`, dateSetup.String(), EscapeString(id), strings.Join(setStatements, "\n\t"))

	_, err := s.exec.Run(script)
	if err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}

	return nil
}

// DeleteReminder deletes a reminder by ID.
func (s *ReminderService) DeleteReminder(id string) error {
	script := fmt.Sprintf(`tell application "Reminders"
	delete (first reminder whose id is "%s")
end tell`, EscapeString(id))

	_, err := s.exec.Run(script)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}

	return nil
}

// CompleteReminder marks a reminder as completed.
func (s *ReminderService) CompleteReminder(id string) error {
	return s.UpdateReminder(id, map[string]any{"completed": true})
}

// UncompleteReminder marks a reminder as incomplete.
func (s *ReminderService) UncompleteReminder(id string) error {
	return s.UpdateReminder(id, map[string]any{"completed": false})
}

// FlagReminder flags a reminder.
func (s *ReminderService) FlagReminder(id string) error {
	return s.UpdateReminder(id, map[string]any{"flagged": true})
}

// UnflagReminder removes the flag from a reminder.
func (s *ReminderService) UnflagReminder(id string) error {
	return s.UpdateReminder(id, map[string]any{"flagged": false})
}
