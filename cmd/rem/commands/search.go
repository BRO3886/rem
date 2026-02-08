package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/rem/internal/reminder"
	"github.com/BRO3886/rem/internal/ui"
	"github.com/spf13/cobra"
)

var (
	searchList       string
	searchIncomplete bool
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search reminders by title and notes",
	Example: `  rem search "groceries"
  rem search "meeting" --list Work --incomplete`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := &reminder.ListFilter{
			ListName:    searchList,
			SearchQuery: args[0],
		}

		if searchIncomplete {
			v := false
			filter.Completed = &v
		}

		reminders, err := reminderSvc.ListReminders(filter)
		if err != nil {
			return err
		}

		if len(reminders) == 0 {
			fmt.Fprintf(os.Stderr, "No reminders matching '%s'\n", args[0])
			return nil
		}

		format := ui.ParseOutputFormat(outputFormat)
		ui.PrintReminders(os.Stdout, reminders, format)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchList, "list", "l", "", "Search within a specific list")
	searchCmd.Flags().BoolVar(&searchIncomplete, "incomplete", false, "Search only incomplete reminders")
	rootCmd.AddCommand(searchCmd)
}
