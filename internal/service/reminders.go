//go:build darwin

package service

import (
	"fmt"
	"sort"
	"time"

	eventkit "github.com/BRO3886/go-eventkit"
	"github.com/BRO3886/go-eventkit/reminders"
	"github.com/BRO3886/rem/internal/reminder"
)

// ReminderService provides operations for reminders using go-eventkit for
// all reads and writes, including flagged operations via the private
// ReminderKit bridge.
type ReminderService struct {
	client *reminders.Client
}

// NewReminderService creates a new ReminderService.
func NewReminderService(client *reminders.Client) *ReminderService {
	return &ReminderService{client: client}
}

// CreateReminder creates a new reminder and returns its ID.
func (s *ReminderService) CreateReminder(r *reminder.Reminder) (string, error) {
	if r.Name == "" {
		return "", fmt.Errorf("reminder name is required")
	}

	input := reminders.CreateReminderInput{
		Title:    r.Name,
		Notes:    r.Body,
		ListName: r.ListName,
		DueDate:  r.DueDate,
		Priority: reminders.Priority(r.Priority),
		Tags:     r.Tags,
	}

	if r.RemindMeDate != nil {
		input.RemindMeDate = r.RemindMeDate
	}

	if r.URL != "" {
		input.URL = r.URL
	}

	if r.Flagged {
		input.Flagged = true
	}

	for _, a := range r.Alarms {
		input.Alarms = append(input.Alarms, reminders.Alarm{
			AbsoluteDate:   a.AbsoluteDate,
			RelativeOffset: a.RelativeOffset,
		})
	}

	for _, rr := range r.RecurrenceRules {
		input.RecurrenceRules = append(input.RecurrenceRules, toEventKitRecurrenceRule(rr))
	}

	created, err := s.client.CreateReminder(input)
	if err != nil {
		return "", fmt.Errorf("failed to create reminder: %w", err)
	}

	return created.ID, nil
}

// GetReminder retrieves a single reminder by ID or ID prefix.
func (s *ReminderService) GetReminder(id string) (*reminder.Reminder, error) {
	r, err := s.client.Reminder(id)
	if err != nil {
		return nil, fmt.Errorf("reminder not found: %s", id)
	}
	return fromEventKitReminder(r), nil
}

// ListReminders returns reminders matching the given filter.
func (s *ReminderService) ListReminders(filter *reminder.ListFilter) ([]*reminder.Reminder, error) {
	var opts []reminders.ListOption

	if filter != nil {
		if filter.ListName != "" {
			opts = append(opts, reminders.WithList(filter.ListName))
		}
		if filter.Completed != nil {
			opts = append(opts, reminders.WithCompleted(*filter.Completed))
		}
		if filter.SearchQuery != "" {
			opts = append(opts, reminders.WithSearch(filter.SearchQuery))
		}
		if filter.DueBefore != nil {
			opts = append(opts, reminders.WithDueBefore(*filter.DueBefore))
		}
		if filter.DueAfter != nil {
			opts = append(opts, reminders.WithDueAfter(*filter.DueAfter))
		}
		if len(filter.Tags) > 0 {
			opts = append(opts, reminders.WithTags(filter.Tags...))
		}
	}

	ekReminders, err := s.client.Reminders(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list reminders: %w", err)
	}

	// Apply flagged filter — go-eventkit reads the flagged property via the
	// private ReminderKit bridge, so it's already populated on each Reminder.
	needsFlagged := filter != nil && filter.Flagged != nil && *filter.Flagged

	result := make([]*reminder.Reminder, 0, len(ekReminders))
	for i := range ekReminders {
		r := fromEventKitReminder(&ekReminders[i])

		if needsFlagged && !r.Flagged {
			continue
		}

		result = append(result, r)
	}

	sortReminders(result)

	return result, nil
}

// sortReminders sorts by due date ascending (nil last), then priority (higher first, none last).
func sortReminders(result []*reminder.Reminder) {
	sort.SliceStable(result, func(i, j int) bool {
		ri, rj := result[i], result[j]

		switch {
		case ri.DueDate == nil && rj.DueDate == nil:
			// fall through to priority
		case ri.DueDate == nil:
			return false
		case rj.DueDate == nil:
			return true
		default:
			if !ri.DueDate.Equal(*rj.DueDate) {
				return ri.DueDate.Before(*rj.DueDate)
			}
		}

		// Priority 0 (none) sorts last; otherwise lower value = higher priority
		if ri.Priority == reminder.PriorityNone {
			return false
		}
		if rj.Priority == reminder.PriorityNone {
			return true
		}
		return ri.Priority < rj.Priority
	})
}

// UpdateReminder updates properties of an existing reminder.
func (s *ReminderService) UpdateReminder(id string, updates map[string]any) error {
	input := reminders.UpdateReminderInput{}

	for key, value := range updates {
		switch key {
		case "name":
			v := value.(string)
			input.Title = &v
		case "body":
			v := value.(string)
			input.Notes = &v
		case "due_date":
			if value == nil {
				input.ClearDueDate = true
			} else {
				t := value.(time.Time)
				input.DueDate = &t
			}
		case "remind_me_date":
			if value == nil {
				// Clear remind me date by setting to zero time
				t := time.Time{}
				input.RemindMeDate = &t
			} else {
				t := value.(time.Time)
				input.RemindMeDate = &t
			}
		case "priority":
			p := reminders.Priority(value.(reminder.Priority))
			input.Priority = &p
		case "flagged":
			v := value.(bool)
			input.Flagged = &v
		case "completed":
			v := value.(bool)
			input.Completed = &v
		case "url":
			v := value.(string)
			input.URL = &v
		case "tags":
			v := value.([]string)
			input.Tags = &v
		case "list":
			v := value.(string)
			input.ListName = &v
		case "alarms":
			if value == nil {
				empty := []reminders.Alarm{}
				input.Alarms = &empty
			} else {
				alarms := value.([]reminder.Alarm)
				ekAlarms := make([]reminders.Alarm, len(alarms))
				for i, a := range alarms {
					ekAlarms[i] = reminders.Alarm{
						AbsoluteDate:   a.AbsoluteDate,
						RelativeOffset: a.RelativeOffset,
					}
				}
				input.Alarms = &ekAlarms
			}
		case "recurrence":
			if value == nil {
				empty := []eventkit.RecurrenceRule{}
				input.RecurrenceRules = &empty
			} else {
				rules := value.([]reminder.RecurrenceRule)
				ekRules := make([]eventkit.RecurrenceRule, len(rules))
				for i, rr := range rules {
					ekRules[i] = toEventKitRecurrenceRule(rr)
				}
				input.RecurrenceRules = &ekRules
			}
		}
	}

	if _, err := s.client.UpdateReminder(id, input); err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}

	return nil
}

// DeleteReminder deletes a reminder by ID.
func (s *ReminderService) DeleteReminder(id string) error {
	if err := s.client.DeleteReminder(id); err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}
	return nil
}

// DeleteReminders deletes multiple reminders in a single batch call.
// Returns a map of reminder ID to error for any that failed.
func (s *ReminderService) DeleteReminders(ids []string) map[string]error {
	return s.client.DeleteReminders(ids)
}

// CompleteReminder marks a reminder as completed.
func (s *ReminderService) CompleteReminder(id string) error {
	if _, err := s.client.CompleteReminder(id); err != nil {
		return fmt.Errorf("failed to complete reminder: %w", err)
	}
	return nil
}

// UncompleteReminder marks a reminder as incomplete.
func (s *ReminderService) UncompleteReminder(id string) error {
	if _, err := s.client.UncompleteReminder(id); err != nil {
		return fmt.Errorf("failed to uncomplete reminder: %w", err)
	}
	return nil
}

// FlagReminder flags a reminder via the private ReminderKit bridge.
func (s *ReminderService) FlagReminder(id string) error {
	flagged := true
	if _, err := s.client.UpdateReminder(id, reminders.UpdateReminderInput{Flagged: &flagged}); err != nil {
		return fmt.Errorf("failed to flag reminder: %w", err)
	}
	return nil
}

// UnflagReminder removes the flag from a reminder via the private ReminderKit bridge.
func (s *ReminderService) UnflagReminder(id string) error {
	flagged := false
	if _, err := s.client.UpdateReminder(id, reminders.UpdateReminderInput{Flagged: &flagged}); err != nil {
		return fmt.Errorf("failed to unflag reminder: %w", err)
	}
	return nil
}

// toEventKitRecurrenceRule converts a domain RecurrenceRule to an eventkit RecurrenceRule.
func toEventKitRecurrenceRule(rr reminder.RecurrenceRule) eventkit.RecurrenceRule {
	switch rr.Frequency {
	case reminder.FrequencyDaily:
		return eventkit.Daily(rr.Interval)
	case reminder.FrequencyWeekly:
		var days []eventkit.Weekday
		for _, d := range rr.DaysOfWeekNums {
			days = append(days, eventkit.Weekday(d))
		}
		return eventkit.Weekly(rr.Interval, days...)
	case reminder.FrequencyMonthly:
		return eventkit.Monthly(rr.Interval, rr.DaysOfMonth...)
	case reminder.FrequencyYearly:
		return eventkit.Yearly(rr.Interval)
	default:
		return eventkit.Daily(1)
	}
}

// fromEventKitReminder converts a go-eventkit Reminder to an internal Reminder.
func fromEventKitReminder(r *reminders.Reminder) *reminder.Reminder {
	result := &reminder.Reminder{
		ID:               r.ID,
		Name:             r.Title,
		Body:             r.Notes,
		ListName:         r.List,
		DueDate:          r.DueDate,
		RemindMeDate:     r.RemindMeDate,
		CompletionDate:   r.CompletionDate,
		CreationDate:     r.CreatedAt,
		ModificationDate: r.ModifiedAt,
		Priority:         reminder.Priority(r.Priority),
		Completed:        r.Completed,
		Flagged:          r.Flagged,
		URL:              r.URL,
		Tags:             append([]string(nil), r.Tags...),
		Recurring:        r.Recurring,
		HasAlarms:        r.HasAlarms,
	}

	// Map recurrence rules
	for _, rule := range r.RecurrenceRules {
		rr := reminder.RecurrenceRule{
			Frequency:   reminder.RecurrenceFrequency(rule.Frequency),
			Interval:    rule.Interval,
			DaysOfMonth: rule.DaysOfTheMonth,
		}
		for _, dow := range rule.DaysOfTheWeek {
			rr.DaysOfWeek = append(rr.DaysOfWeek, dow.DayOfTheWeek.String())
			rr.DaysOfWeekNums = append(rr.DaysOfWeekNums, int(dow.DayOfTheWeek))
		}
		result.RecurrenceRules = append(result.RecurrenceRules, rr)
	}

	// Map alarms
	for _, a := range r.Alarms {
		result.Alarms = append(result.Alarms, reminder.Alarm{
			AbsoluteDate:   a.AbsoluteDate,
			RelativeOffset: a.RelativeOffset,
		})
	}

	// For backwards compatibility: if URL is empty but notes contain a URL, extract it
	if result.URL == "" && result.Body != "" {
		result.URL = extractURL(result.Body)
	}

	return result
}
