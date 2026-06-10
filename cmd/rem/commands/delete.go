package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	deleteForce       bool
	deleteInteractive bool
	deleteList        string
	deleteFlagged     bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete [id...]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete one or more reminders",
	Example: `  rem delete abc12345
  rem delete abc12345 def67890 --force
  rem rm abc12345 --force
  rem delete -i
  rem delete -i --list Work --flagged`,
	Args: func(cmd *cobra.Command, args []string) error {
		if deleteInteractive {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if deleteInteractive {
			return runDeleteInteractive()
		}

		// Resolve all IDs first
		type resolved struct {
			id   string
			name string
		}
		var toDelete []resolved
		for _, arg := range args {
			r, err := findReminderByID(arg)
			if err != nil {
				return err
			}
			toDelete = append(toDelete, resolved{id: r.ID, name: r.Name})
		}

		if !deleteForce {
			if isTTY() {
				msg := fmt.Sprintf("Delete reminder '%s'?", toDelete[0].name)
				if len(toDelete) > 1 {
					msg = fmt.Sprintf("Delete %d reminders?", len(toDelete))
				}
				confirmed, err := huhConfirm(msg)
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			} else {
				return fmt.Errorf("use --force/-f to delete non-interactively, or run in a terminal")
			}
		}

		if len(toDelete) == 1 {
			if err := reminderSvc.DeleteReminder(toDelete[0].id); err != nil {
				return err
			}
			fmt.Printf("Deleted: %s\n", toDelete[0].name)
		} else {
			ids := make([]string, len(toDelete))
			for i, r := range toDelete {
				ids[i] = r.id
			}
			errs := reminderSvc.DeleteReminders(ids)
			for id, err := range errs {
				fmt.Printf("Error deleting %s: %v\n", shortIDStr(id), err)
			}
			fmt.Printf("Deleted %d reminder(s)\n", len(toDelete)-len(errs))
		}

		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
	deleteCmd.Flags().BoolVarP(&deleteForce, "yes", "y", false, "Skip confirmation prompt (alias for --force)")
	deleteCmd.Flags().BoolVarP(&deleteInteractive, "interactive", "i", false, "Select reminders interactively")
	deleteCmd.Flags().StringVarP(&deleteList, "list", "l", "", "Filter by list name")
	deleteCmd.Flags().BoolVar(&deleteFlagged, "flagged", false, "Filter to flagged reminders only")
	rootCmd.AddCommand(deleteCmd)
}

// runDeleteInteractive runs the interactive multi-select flow for deleting reminders.
func runDeleteInteractive() error {
	if err := requireInteractive(); err != nil {
		return err
	}

	reminders, err := reminderSvc.ListReminders(deleteFilter(deleteList, deleteFlagged))
	if err != nil {
		return err
	}

	selected, err := reminderMultiSelect("Select reminders to delete", reminders)
	if err != nil {
		return err
	}
	if selected == nil {
		return nil // cancelled
	}

	confirmed, err := huhConfirm(fmt.Sprintf("Delete %d reminder(s)?", len(selected)))
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	errs := reminderSvc.DeleteReminders(selected)
	for id, err := range errs {
		fmt.Printf("Error deleting %s: %v\n", shortIDStr(id), err)
	}

	fmt.Printf("Deleted %d reminder(s)\n", len(selected)-len(errs))
	return nil
}
