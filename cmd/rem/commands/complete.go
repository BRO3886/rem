package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:     "complete [id]",
	Aliases: []string{"done"},
	Short:   "Mark a reminder as complete",
	Example: `  rem complete abc12345
  rem done abc12345`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := findReminderByID(args[0])
		if err != nil {
			return err
		}

		if err := reminderSvc.CompleteReminder(r.ID); err != nil {
			return err
		}

		fmt.Printf("Completed: %s\n", r.Name)
		return nil
	},
}

var uncompleteCmd = &cobra.Command{
	Use:   "uncomplete [id]",
	Short: "Mark a reminder as incomplete",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := findReminderByID(args[0])
		if err != nil {
			return err
		}

		if err := reminderSvc.UncompleteReminder(r.ID); err != nil {
			return err
		}

		fmt.Printf("Marked incomplete: %s\n", r.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(uncompleteCmd)
}
