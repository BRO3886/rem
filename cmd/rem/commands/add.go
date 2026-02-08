package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/BRO3886/rem/internal/parser"
	"github.com/BRO3886/rem/internal/reminder"
	"github.com/BRO3886/rem/internal/ui"
	"github.com/spf13/cobra"
)

var (
	addList     string
	addDue      string
	addPriority string
	addNotes    string
	addURL      string
	addFlagged  bool
	addInteractive bool
)

var addCmd = &cobra.Command{
	Use:     "add [title]",
	Aliases: []string{"create", "new"},
	Short:   "Create a new reminder",
	Long:    `Create a new reminder with optional properties like due date, priority, notes, and URL.`,
	Example: `  rem add "Buy groceries" --list Personal --due tomorrow --priority high
  rem add "Review PR" --due "next friday at 2pm" --url https://github.com/org/repo/pull/123
  rem add "Call dentist" --due "in 2 days" --notes "Ask about cleaning"
  rem add -i  # Interactive mode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if addInteractive {
			return runAddInteractive()
		}

		if len(args) == 0 {
			return fmt.Errorf("reminder title is required (or use -i for interactive mode)")
		}

		r := &reminder.Reminder{
			Name:     args[0],
			Body:     addNotes,
			ListName: addList,
			URL:      addURL,
			Flagged:  addFlagged,
			Priority: reminder.ParsePriority(addPriority),
		}

		if addDue != "" {
			dueDate, err := parser.ParseDate(addDue)
			if err != nil {
				return fmt.Errorf("invalid due date: %w", err)
			}
			r.DueDate = &dueDate
		}

		id, err := reminderSvc.CreateReminder(r)
		if err != nil {
			return err
		}

		format := ui.ParseOutputFormat(outputFormat)
		if format == ui.FormatJSON {
			fmt.Fprintf(os.Stdout, `{"id": "%s", "name": "%s"}`+"\n", id, r.Name)
		} else {
			fmt.Fprintf(os.Stdout, "Created reminder: %s (ID: %s)\n", r.Name, shortIDStr(id))
		}

		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addList, "list", "l", "", "Reminder list name (default: system default list)")
	addCmd.Flags().StringVarP(&addDue, "due", "d", "", "Due date (e.g., 'tomorrow', 'next friday at 2pm', '2026-02-15')")
	addCmd.Flags().StringVarP(&addPriority, "priority", "p", "", "Priority: high, medium, low, or none")
	addCmd.Flags().StringVarP(&addNotes, "notes", "n", "", "Notes/body for the reminder")
	addCmd.Flags().StringVarP(&addURL, "url", "u", "", "URL to attach to the reminder")
	addCmd.Flags().BoolVarP(&addFlagged, "flagged", "f", false, "Flag the reminder")
	addCmd.Flags().BoolVarP(&addInteractive, "interactive", "i", false, "Create reminder interactively")

	rootCmd.AddCommand(addCmd)
}

func shortIDStr(id string) string {
	s := strings.TrimPrefix(id, "x-apple-reminder://")
	if len(s) > 8 {
		return s[:8]
	}
	return s
}
