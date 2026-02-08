package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BRO3886/rem/internal/export"
	"github.com/BRO3886/rem/internal/reminder"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// OutputFormat represents the output format type.
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatPlain OutputFormat = "plain"
)

// ParseOutputFormat parses an output format string.
func ParseOutputFormat(s string) OutputFormat {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "plain", "text":
		return FormatPlain
	default:
		return FormatTable
	}
}

// ColorsEnabled returns true if color output should be used.
func ColorsEnabled() bool {
	return os.Getenv("NO_COLOR") == ""
}

// PrintReminders outputs reminders in the specified format.
func PrintReminders(w io.Writer, reminders []*reminder.Reminder, format OutputFormat) {
	switch format {
	case FormatJSON:
		printRemindersJSON(w, reminders)
	case FormatPlain:
		printRemindersPlain(w, reminders)
	default:
		printRemindersTable(w, reminders)
	}
}

// PrintReminderDetail prints a single reminder with all details.
func PrintReminderDetail(w io.Writer, r *reminder.Reminder, format OutputFormat) {
	switch format {
	case FormatJSON:
		jr := export.ToJSON(r)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(jr)
	case FormatPlain:
		printReminderPlainDetail(w, r)
	default:
		printReminderRichDetail(w, r)
	}
}

// PrintLists outputs reminder lists in the specified format.
func PrintLists(w io.Writer, lists []*reminder.List, format OutputFormat, showCount bool) {
	switch format {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(lists)
	case FormatPlain:
		for _, l := range lists {
			if showCount {
				fmt.Fprintf(w, "%s (%d)\n", l.Name, l.Count)
			} else {
				fmt.Fprintln(w, l.Name)
			}
		}
	default:
		printListsTable(w, lists, showCount)
	}
}

func newTable(w io.Writer) *tablewriter.Table {
	return tablewriter.NewTable(w,
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithRendition(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenColumns: tw.On,
				},
			},
		}),
	)
}

func printRemindersTable(w io.Writer, reminders []*reminder.Reminder) {
	if len(reminders) == 0 {
		fmt.Fprintln(w, "No reminders found.")
		return
	}

	table := newTable(w)
	table.Header("ID", "Name", "List", "Due", "Priority", "Status")

	for _, r := range reminders {
		id := shortID(r.ID)
		dueStr := ""
		if r.DueDate != nil {
			dueStr = r.DueDate.Format("Jan 02, 15:04")
		}
		priority := r.Priority.String()
		status := statusString(r)

		table.Append([]string{id, r.Name, r.ListName, dueStr, priority, status})
	}

	table.Render()
}

func printRemindersPlain(w io.Writer, reminders []*reminder.Reminder) {
	for _, r := range reminders {
		dueStr := ""
		if r.DueDate != nil {
			dueStr = " (due: " + r.DueDate.Format("2006-01-02 15:04") + ")"
		}
		statusMark := "[ ]"
		if r.Completed {
			statusMark = "[x]"
		}
		fmt.Fprintf(w, "%s %s %s%s [%s]\n", statusMark, shortID(r.ID), r.Name, dueStr, r.ListName)
	}
}

func printRemindersJSON(w io.Writer, reminders []*reminder.Reminder) {
	export.ExportJSON(w, reminders)
}

func printReminderRichDetail(w io.Writer, r *reminder.Reminder) {
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Fprintf(w, "%s %s\n", bold("Name:"), r.Name)
	fmt.Fprintf(w, "%s %s\n", bold("ID:"), r.ID)
	fmt.Fprintf(w, "%s %s\n", bold("List:"), cyan(r.ListName))

	if r.Body != "" {
		fmt.Fprintf(w, "%s %s\n", bold("Notes:"), r.Body)
	}
	if r.URL != "" {
		fmt.Fprintf(w, "%s %s\n", bold("URL:"), cyan(r.URL))
	}
	if r.DueDate != nil {
		fmt.Fprintf(w, "%s %s\n", bold("Due:"), r.DueDate.Format("Mon Jan 02, 2006 at 3:04 PM"))
	}
	if r.RemindMeDate != nil {
		fmt.Fprintf(w, "%s %s\n", bold("Remind:"), r.RemindMeDate.Format("Mon Jan 02, 2006 at 3:04 PM"))
	}

	priorityStr := r.Priority.String()
	switch {
	case r.Priority >= 1 && r.Priority <= 4:
		priorityStr = red(priorityStr)
	case r.Priority == 5:
		priorityStr = yellow(priorityStr)
	case r.Priority >= 6:
		priorityStr = green(priorityStr)
	}
	fmt.Fprintf(w, "%s %s\n", bold("Priority:"), priorityStr)

	if r.Completed {
		fmt.Fprintf(w, "%s %s\n", bold("Status:"), green("completed"))
	} else {
		fmt.Fprintf(w, "%s %s\n", bold("Status:"), "incomplete")
	}

	if r.Flagged {
		fmt.Fprintf(w, "%s %s\n", bold("Flagged:"), yellow("yes"))
	}

	if r.CreationDate != nil {
		fmt.Fprintf(w, "%s %s\n", bold("Created:"), r.CreationDate.Format("Mon Jan 02, 2006 at 3:04 PM"))
	}
	if r.ModificationDate != nil {
		fmt.Fprintf(w, "%s %s\n", bold("Modified:"), r.ModificationDate.Format("Mon Jan 02, 2006 at 3:04 PM"))
	}
}

func printReminderPlainDetail(w io.Writer, r *reminder.Reminder) {
	fmt.Fprintf(w, "Name: %s\n", r.Name)
	fmt.Fprintf(w, "ID: %s\n", r.ID)
	fmt.Fprintf(w, "List: %s\n", r.ListName)
	if r.Body != "" {
		fmt.Fprintf(w, "Notes: %s\n", r.Body)
	}
	if r.URL != "" {
		fmt.Fprintf(w, "URL: %s\n", r.URL)
	}
	if r.DueDate != nil {
		fmt.Fprintf(w, "Due: %s\n", r.DueDate.Format("2006-01-02 15:04"))
	}
	fmt.Fprintf(w, "Priority: %s\n", r.Priority.String())
	if r.Completed {
		fmt.Fprintf(w, "Status: completed\n")
	} else {
		fmt.Fprintf(w, "Status: incomplete\n")
	}
	if r.Flagged {
		fmt.Fprintf(w, "Flagged: yes\n")
	}
}

func printListsTable(w io.Writer, lists []*reminder.List, showCount bool) {
	if len(lists) == 0 {
		fmt.Fprintln(w, "No lists found.")
		return
	}

	table := newTable(w)
	if showCount {
		table.Header("Name", "Reminders")
	} else {
		table.Header("Name")
	}

	for _, l := range lists {
		if showCount {
			table.Append([]string{l.Name, fmt.Sprintf("%d", l.Count)})
		} else {
			table.Append([]string{l.Name})
		}
	}

	table.Render()
}

func shortID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func statusString(r *reminder.Reminder) string {
	parts := []string{}
	if r.Completed {
		parts = append(parts, "done")
	}
	if r.Flagged {
		parts = append(parts, "flagged")
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}
