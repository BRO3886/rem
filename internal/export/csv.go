package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/BRO3886/rem/internal/reminder"
)

var csvHeaders = []string{
	"id", "name", "body", "list_name", "due_date", "remind_me_date",
	"priority", "priority_label", "flagged", "completed", "url",
}

// ExportCSV writes reminders as CSV to the writer.
func ExportCSV(w io.Writer, reminders []*reminder.Reminder) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write(csvHeaders); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, r := range reminders {
		dueDate := ""
		if r.DueDate != nil {
			dueDate = r.DueDate.Format(timeFormat)
		}
		remindDate := ""
		if r.RemindMeDate != nil {
			remindDate = r.RemindMeDate.Format(timeFormat)
		}

		record := []string{
			r.ID,
			r.Name,
			r.Body,
			r.ListName,
			dueDate,
			remindDate,
			strconv.Itoa(int(r.Priority)),
			r.Priority.String(),
			strconv.FormatBool(r.Flagged),
			strconv.FormatBool(r.Completed),
			r.URL,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// ImportCSV reads reminders from a CSV reader.
func ImportCSV(r io.Reader) ([]*reminder.Reminder, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Build column index map
	colMap := make(map[string]int)
	for i, h := range header {
		colMap[strings.TrimSpace(strings.ToLower(h))] = i
	}

	var reminders []*reminder.Reminder
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		rem := &reminder.Reminder{}

		if idx, ok := colMap["name"]; ok && idx < len(record) {
			rem.Name = record[idx]
		}
		if idx, ok := colMap["body"]; ok && idx < len(record) {
			rem.Body = record[idx]
		}
		if idx, ok := colMap["list_name"]; ok && idx < len(record) {
			rem.ListName = record[idx]
		}
		if idx, ok := colMap["url"]; ok && idx < len(record) {
			rem.URL = record[idx]
		}
		if idx, ok := colMap["due_date"]; ok && idx < len(record) && record[idx] != "" {
			t, err := time.ParseInLocation(timeFormat, record[idx], time.Now().Location())
			if err == nil {
				rem.DueDate = &t
			}
		}
		if idx, ok := colMap["remind_me_date"]; ok && idx < len(record) && record[idx] != "" {
			t, err := time.ParseInLocation(timeFormat, record[idx], time.Now().Location())
			if err == nil {
				rem.RemindMeDate = &t
			}
		}
		if idx, ok := colMap["priority"]; ok && idx < len(record) {
			p, err := strconv.Atoi(record[idx])
			if err == nil {
				rem.Priority = reminder.Priority(p)
			}
		}
		if idx, ok := colMap["flagged"]; ok && idx < len(record) {
			rem.Flagged = strings.ToLower(record[idx]) == "true"
		}
		if idx, ok := colMap["completed"]; ok && idx < len(record) {
			rem.Completed = strings.ToLower(record[idx]) == "true"
		}

		reminders = append(reminders, rem)
	}

	return reminders, nil
}
