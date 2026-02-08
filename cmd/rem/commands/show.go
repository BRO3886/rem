package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/rem/internal/reminder"
	"github.com/BRO3886/rem/internal/ui"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show [id]",
	Aliases: []string{"get"},
	Short:   "Show details of a specific reminder",
	Long:    `Display all properties of a specific reminder by its ID.`,
	Example: `  rem show abc12345
  rem get abc12345 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		// Try to find the reminder by matching the prefix
		r, err := findReminderByID(id)
		if err != nil {
			return err
		}

		format := ui.ParseOutputFormat(outputFormat)
		ui.PrintReminderDetail(os.Stdout, r, format)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}

// findReminderByID finds a reminder by full or partial ID.
// Accepts full IDs (x-apple-reminder://UUID), bare UUIDs, or short prefixes.
func findReminderByID(id string) (*reminder.Reminder, error) {
	r, err := reminderSvc.GetReminder(id)
	if err != nil {
		return nil, fmt.Errorf("no reminder found with ID: %s", id)
	}
	return r, nil
}
