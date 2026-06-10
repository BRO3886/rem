package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	completeInteractive bool
	completeList        string
	completeFlagged     bool

	uncompleteInteractive bool
	uncompleteList        string
)

var completeCmd = &cobra.Command{
	Use:     "complete [id...]",
	Aliases: []string{"done"},
	Short:   "Mark one or more reminders as complete",
	Example: `  rem complete abc12345
  rem complete abc12345 def67890
  rem done abc12345
  rem complete -i
  rem complete -i --list Work --flagged`,
	Args: func(cmd *cobra.Command, args []string) error {
		if completeInteractive {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if completeInteractive {
			return runCompleteInteractive(false)
		}
		return runBatchAction(args, "Completed", reminderSvc.CompleteReminder)
	},
}

var uncompleteCmd = &cobra.Command{
	Use:   "uncomplete [id...]",
	Short: "Mark one or more reminders as incomplete",
	Example: `  rem uncomplete abc12345
  rem uncomplete abc12345 def67890
  rem uncomplete -i
  rem uncomplete -i --list Work`,
	Args: func(cmd *cobra.Command, args []string) error {
		if uncompleteInteractive {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if uncompleteInteractive {
			return runCompleteInteractive(true)
		}
		return runBatchAction(args, "Marked incomplete", reminderSvc.UncompleteReminder)
	},
}

func init() {
	completeCmd.Flags().BoolVarP(&completeInteractive, "interactive", "i", false, "Select reminders interactively")
	completeCmd.Flags().StringVarP(&completeList, "list", "l", "", "Filter by list name")
	completeCmd.Flags().BoolVar(&completeFlagged, "flagged", false, "Filter to flagged reminders only")
	rootCmd.AddCommand(completeCmd)

	uncompleteCmd.Flags().BoolVarP(&uncompleteInteractive, "interactive", "i", false, "Select reminders interactively")
	uncompleteCmd.Flags().StringVarP(&uncompleteList, "list", "l", "", "Filter by list name")
	rootCmd.AddCommand(uncompleteCmd)
}

// runCompleteInteractive runs the interactive multi-select flow for completing or uncompleting reminders.
// If uncomplete is true, shows completed reminders and marks them incomplete.
func runCompleteInteractive(uncomplete bool) error {
	if err := requireInteractive(); err != nil {
		return err
	}

	listName := completeList
	if uncomplete {
		listName = uncompleteList
	}

	reminders, err := reminderSvc.ListReminders(completeFilter(uncomplete, listName, completeFlagged))
	if err != nil {
		return err
	}

	title := "Select reminders to complete"
	if uncomplete {
		title = "Select reminders to mark incomplete"
	}

	selected, err := reminderMultiSelect(title, reminders)
	if err != nil {
		return err
	}
	if selected == nil {
		return nil // cancelled
	}

	for _, id := range selected {
		if uncomplete {
			if err := reminderSvc.UncompleteReminder(id); err != nil {
				fmt.Printf("Error uncompleting %s: %v\n", shortIDStr(id), err)
				continue
			}
		} else {
			if err := reminderSvc.CompleteReminder(id); err != nil {
				fmt.Printf("Error completing %s: %v\n", shortIDStr(id), err)
				continue
			}
		}
	}

	action := "Completed"
	if uncomplete {
		action = "Marked incomplete"
	}
	fmt.Printf("%s %d reminder(s)\n", action, len(selected))
	return nil
}
