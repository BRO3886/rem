package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/BRO3886/rem/internal/ui"
	"github.com/spf13/cobra"
)

var listsShowCount bool

var listsCmd = &cobra.Command{
	Use:   "lists",
	Short: "List all reminder lists",
	Example: `  rem lists
  rem lists --count
  rem lists --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lists, err := listSvc.GetLists()
		if err != nil {
			return err
		}

		format := ui.ParseOutputFormat(outputFormat)
		ui.PrintLists(os.Stdout, lists, format, listsShowCount)
		return nil
	},
}

var listCreateCmd = &cobra.Command{
	Use:     "create [name]",
	Aliases: []string{"new"},
	Short:   "Create a new reminder list",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := listSvc.CreateList(args[0])
		if err != nil {
			return err
		}

		format := ui.ParseOutputFormat(outputFormat)
		if format == ui.FormatJSON {
			fmt.Fprintf(os.Stdout, `{"id": "%s", "name": "%s"}`+"\n", list.ID, list.Name)
		} else {
			fmt.Printf("Created list: %s\n", list.Name)
		}
		return nil
	},
}

var listRenameCmd = &cobra.Command{
	Use:   "rename [old-name] [new-name]",
	Short: "Rename a reminder list",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := listSvc.RenameList(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Renamed list '%s' to '%s'\n", args[0], args[1])
		return nil
	},
}

var listDeleteForce bool

var listDeleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Aliases: []string{"rm"},
	Short:   "Delete a reminder list",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if !listDeleteForce {
			fmt.Printf("Delete list '%s' and all its reminders? (y/N): ", name)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := listSvc.DeleteList(name); err != nil {
			return err
		}
		fmt.Printf("Deleted list: %s\n", name)
		return nil
	},
}

// listMgmtCmd is the parent command for list management operations.
var listMgmtCmd = &cobra.Command{
	Use:   "list-mgmt",
	Aliases: []string{"lm"},
	Short: "Manage reminder lists (create, rename, delete)",
	Long: `Manage reminder lists. Use subcommands to create, rename, or delete lists.

Note: Use 'rem lists' (plural) to view all lists.
Use 'rem list-mgmt' or 'rem lm' for list management operations.`,
}

func init() {
	listsCmd.Flags().BoolVarP(&listsShowCount, "count", "c", false, "Show reminder count per list")
	rootCmd.AddCommand(listsCmd)

	listDeleteCmd.Flags().BoolVar(&listDeleteForce, "force", false, "Skip confirmation prompt")

	listMgmtCmd.AddCommand(listCreateCmd)
	listMgmtCmd.AddCommand(listRenameCmd)
	listMgmtCmd.AddCommand(listDeleteCmd)
	rootCmd.AddCommand(listMgmtCmd)
}
