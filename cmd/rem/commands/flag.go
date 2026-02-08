package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var flagCmd = &cobra.Command{
	Use:   "flag [id]",
	Short: "Flag a reminder",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := findReminderByID(args[0])
		if err != nil {
			return err
		}

		if err := reminderSvc.FlagReminder(r.ID); err != nil {
			return err
		}

		fmt.Printf("Flagged: %s\n", r.Name)
		return nil
	},
}

var unflagCmd = &cobra.Command{
	Use:   "unflag [id]",
	Short: "Remove flag from a reminder",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := findReminderByID(args[0])
		if err != nil {
			return err
		}

		if err := reminderSvc.UnflagReminder(r.ID); err != nil {
			return err
		}

		fmt.Printf("Unflagged: %s\n", r.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(flagCmd)
	rootCmd.AddCommand(unflagCmd)
}
