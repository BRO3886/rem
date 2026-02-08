package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a reminder",
	Example: `  rem delete abc12345
  rem rm abc12345 --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := findReminderByID(args[0])
		if err != nil {
			return err
		}

		if !deleteForce {
			fmt.Printf("Delete reminder '%s'? (y/N): ", r.Name)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := reminderSvc.DeleteReminder(r.ID); err != nil {
			return err
		}

		fmt.Printf("Deleted: %s\n", r.Name)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt")
	rootCmd.AddCommand(deleteCmd)
}
