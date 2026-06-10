package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	flagInteractive bool
	flagList        string

	unflagInteractive bool
	unflagList        string
)

var flagCmd = &cobra.Command{
	Use:   "flag [id...]",
	Short: "Flag one or more reminders",
	Example: `  rem flag abc12345
  rem flag abc12345 def67890
  rem flag -i
  rem flag -i --list Work`,
	Args: func(cmd *cobra.Command, args []string) error {
		if flagInteractive {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagInteractive {
			return runFlagInteractive(flagList)
		}
		return runBatchAction(args, "Flagged", reminderSvc.FlagReminder)
	},
}

var unflagCmd = &cobra.Command{
	Use:   "unflag [id...]",
	Short: "Remove flag from one or more reminders",
	Example: `  rem unflag abc12345
  rem unflag abc12345 def67890
  rem unflag -i
  rem unflag -i --list Work`,
	Args: func(cmd *cobra.Command, args []string) error {
		if unflagInteractive {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if unflagInteractive {
			return runUnflagInteractive(unflagList)
		}
		return runBatchAction(args, "Unflagged", reminderSvc.UnflagReminder)
	},
}

func init() {
	flagCmd.Flags().BoolVarP(&flagInteractive, "interactive", "i", false, "Select reminders interactively")
	flagCmd.Flags().StringVarP(&flagList, "list", "l", "", "Filter by list name")
	rootCmd.AddCommand(flagCmd)

	unflagCmd.Flags().BoolVarP(&unflagInteractive, "interactive", "i", false, "Select reminders interactively")
	unflagCmd.Flags().StringVarP(&unflagList, "list", "l", "", "Filter by list name")
	rootCmd.AddCommand(unflagCmd)
}

// runFlagInteractive runs the interactive multi-select flow for flagging reminders.
func runFlagInteractive(listName string) error {
	if err := requireInteractive(); err != nil {
		return err
	}

	reminders, err := reminderSvc.ListReminders(flagFilter(listName))
	if err != nil {
		return err
	}

	selected, err := reminderMultiSelect("Select reminders to flag", reminders)
	if err != nil {
		return err
	}
	if selected == nil {
		return nil // cancelled
	}

	for _, id := range selected {
		if err := reminderSvc.FlagReminder(id); err != nil {
			fmt.Printf("Error flagging %s: %v\n", shortIDStr(id), err)
			continue
		}
	}

	fmt.Printf("Flagged %d reminder(s)\n", len(selected))
	return nil
}

// runUnflagInteractive runs the interactive multi-select flow for unflagging reminders.
func runUnflagInteractive(listName string) error {
	if err := requireInteractive(); err != nil {
		return err
	}

	reminders, err := reminderSvc.ListReminders(unflagFilter(listName))
	if err != nil {
		return err
	}

	selected, err := reminderMultiSelect("Select reminders to unflag", reminders)
	if err != nil {
		return err
	}
	if selected == nil {
		return nil // cancelled
	}

	for _, id := range selected {
		if err := reminderSvc.UnflagReminder(id); err != nil {
			fmt.Printf("Error unflagging %s: %v\n", shortIDStr(id), err)
			continue
		}
	}

	fmt.Printf("Unflagged %d reminder(s)\n", len(selected))
	return nil
}
