package commands

import (
	"fmt"

	"github.com/BRO3886/rem/internal/parser"
	"github.com/BRO3886/rem/internal/reminder"
	"github.com/spf13/cobra"
)

var (
	updateName     string
	updateNotes    string
	updateDue      string
	updatePriority string
	updateURL      string
	updateFlagged  string
	updateInteractive bool
)

var updateCmd = &cobra.Command{
	Use:     "update [id]",
	Aliases: []string{"edit"},
	Short:   "Update an existing reminder",
	Long:    `Update properties of an existing reminder by its ID.`,
	Example: `  rem update abc12345 --due "next monday"
  rem update abc12345 --notes "Updated notes" --priority medium
  rem edit abc12345 --name "New title"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

		r, err := findReminderByID(id)
		if err != nil {
			return err
		}

		updates := make(map[string]any)

		if cmd.Flags().Changed("name") {
			updates["name"] = updateName
		}
		if cmd.Flags().Changed("notes") {
			body := updateNotes
			if updateURL != "" {
				body = body + "\n\nURL: " + updateURL
			}
			updates["body"] = body
		} else if cmd.Flags().Changed("url") {
			body := r.Body
			if body != "" {
				body = body + "\n\nURL: " + updateURL
			} else {
				body = "URL: " + updateURL
			}
			updates["body"] = body
		}
		if cmd.Flags().Changed("due") {
			if updateDue == "" || updateDue == "none" {
				updates["due_date"] = nil
			} else {
				t, err := parser.ParseDate(updateDue)
				if err != nil {
					return fmt.Errorf("invalid due date: %w", err)
				}
				updates["due_date"] = t
			}
		}
		if cmd.Flags().Changed("priority") {
			updates["priority"] = reminder.ParsePriority(updatePriority)
		}
		if cmd.Flags().Changed("flagged") {
			updates["flagged"] = updateFlagged == "true" || updateFlagged == "yes"
		}

		if len(updates) == 0 {
			return fmt.Errorf("no updates specified")
		}

		err = reminderSvc.UpdateReminder(r.ID, updates)
		if err != nil {
			return err
		}

		fmt.Printf("Updated reminder: %s\n", r.Name)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "New name/title")
	updateCmd.Flags().StringVarP(&updateNotes, "notes", "n", "", "New notes/body")
	updateCmd.Flags().StringVarP(&updateDue, "due", "d", "", "New due date (use 'none' to clear)")
	updateCmd.Flags().StringVarP(&updatePriority, "priority", "p", "", "New priority: high, medium, low, none")
	updateCmd.Flags().StringVarP(&updateURL, "url", "u", "", "New URL")
	updateCmd.Flags().StringVar(&updateFlagged, "flagged", "", "Set flagged status: true/false")
	updateCmd.Flags().BoolVarP(&updateInteractive, "interactive", "i", false, "Update interactively")

	rootCmd.AddCommand(updateCmd)
}
