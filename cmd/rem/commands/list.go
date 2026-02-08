package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/rem/internal/parser"
	"github.com/BRO3886/rem/internal/reminder"
	"github.com/BRO3886/rem/internal/ui"
	"github.com/spf13/cobra"
)

var (
	listListName   string
	listIncomplete bool
	listCompleted  bool
	listFlagged    bool
	listDueBefore  string
	listDueAfter   string
	listSearch     string
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List reminders",
	Long:    `List reminders with optional filtering by list, completion status, date range, and more.`,
	Example: `  rem list --list Work --incomplete
  rem list --due-before "2026-02-15" --output json
  rem list --flagged
  rem ls`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := &reminder.ListFilter{
			ListName:    listListName,
			SearchQuery: listSearch,
		}

		if listIncomplete {
			v := false
			filter.Completed = &v
		}
		if listCompleted {
			v := true
			filter.Completed = &v
		}
		if listFlagged {
			v := true
			filter.Flagged = &v
		}
		if listDueBefore != "" {
			t, err := parser.ParseDate(listDueBefore)
			if err != nil {
				return fmt.Errorf("invalid --due-before date: %w", err)
			}
			filter.DueBefore = &t
		}
		if listDueAfter != "" {
			t, err := parser.ParseDate(listDueAfter)
			if err != nil {
				return fmt.Errorf("invalid --due-after date: %w", err)
			}
			filter.DueAfter = &t
		}

		reminders, err := reminderSvc.ListReminders(filter)
		if err != nil {
			return err
		}

		format := ui.ParseOutputFormat(outputFormat)
		ui.PrintReminders(os.Stdout, reminders, format)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&listListName, "list", "l", "", "Filter by list name")
	listCmd.Flags().BoolVar(&listIncomplete, "incomplete", false, "Show only incomplete reminders")
	listCmd.Flags().BoolVar(&listCompleted, "completed", false, "Show only completed reminders")
	listCmd.Flags().BoolVar(&listFlagged, "flagged", false, "Show only flagged reminders")
	listCmd.Flags().StringVar(&listDueBefore, "due-before", "", "Show reminders due before this date")
	listCmd.Flags().StringVar(&listDueAfter, "due-after", "", "Show reminders due after this date")
	listCmd.Flags().StringVarP(&listSearch, "search", "s", "", "Search in title and notes")

	rootCmd.AddCommand(listCmd)
}
