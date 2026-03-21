package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/rem/internal/ui"
	"github.com/charmbracelet/huh"
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

var (
	listCreateInteractive bool
	listRenameInteractive bool
	listDeleteInteractive bool
	listDeleteForce       bool
)

var listCreateCmd = &cobra.Command{
	Use:     "create [name]",
	Aliases: []string{"new"},
	Short:   "Create a new reminder list",
	Example: `  rem lm create "Shopping"
  rem lm create -i`,
	Args: func(cmd *cobra.Command, args []string) error {
		if listCreateInteractive {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if listCreateInteractive {
			return runListCreateInteractive()
		}

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
	Example: `  rem lm rename "Shopping" "Groceries"
  rem lm rename -i`,
	Args: func(cmd *cobra.Command, args []string) error {
		if listRenameInteractive {
			return cobra.RangeArgs(0, 2)(cmd, args)
		}
		return cobra.ExactArgs(2)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if listRenameInteractive {
			return runListRenameInteractive()
		}

		if err := listSvc.RenameList(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Renamed list '%s' to '%s'\n", args[0], args[1])
		return nil
	},
}

var listDeleteCmd = &cobra.Command{
	Use:     "delete [name]",
	Aliases: []string{"rm"},
	Short:   "Delete a reminder list",
	Example: `  rem lm delete "Shopping"
  rem lm delete "Shopping" --force
  rem lm delete -i`,
	Args: func(cmd *cobra.Command, args []string) error {
		if listDeleteInteractive {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if listDeleteInteractive {
			return runListDeleteInteractive("")
		}

		name := args[0]

		if !listDeleteForce {
			if isTTY() {
				confirmed, err := huhConfirm(fmt.Sprintf("Delete list '%s' and all its reminders?", name))
				if err != nil {
					return err
				}
				if !confirmed {
					return nil
				}
			} else {
				return fmt.Errorf("use --force/-y to delete non-interactively, or run in a terminal")
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
	Use:     "list-mgmt",
	Aliases: []string{"lm"},
	Short:   "Manage reminder lists (create, rename, delete)",
	Long: `Manage reminder lists. Use subcommands to create, rename, or delete lists.

Note: Use 'rem lists' (plural) to view all lists.
Use 'rem list-mgmt' or 'rem lm' for list management operations.`,
}

func init() {
	listsCmd.Flags().BoolVarP(&listsShowCount, "count", "c", false, "Show reminder count per list")
	rootCmd.AddCommand(listsCmd)

	listCreateCmd.Flags().BoolVarP(&listCreateInteractive, "interactive", "i", false, "Create list interactively")
	listRenameCmd.Flags().BoolVarP(&listRenameInteractive, "interactive", "i", false, "Rename list interactively")
	listDeleteCmd.Flags().BoolVarP(&listDeleteInteractive, "interactive", "i", false, "Delete list interactively")
	listDeleteCmd.Flags().BoolVarP(&listDeleteForce, "force", "y", false, "Skip confirmation prompt")
	listDeleteCmd.Flags().BoolVar(&listDeleteForce, "yes", false, "Skip confirmation prompt (alias for --force)")

	listMgmtCmd.AddCommand(listCreateCmd)
	listMgmtCmd.AddCommand(listRenameCmd)
	listMgmtCmd.AddCommand(listDeleteCmd)
	rootCmd.AddCommand(listMgmtCmd)
}

// runListCreateInteractive runs the interactive list creation flow.
func runListCreateInteractive() error {
	if err := requireInteractive(); err != nil {
		return err
	}

	var name string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("List name").
				Value(&name).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				}),
		),
	).WithTheme(huhTheme())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return err
	}

	list, err := listSvc.CreateList(name)
	if err != nil {
		return err
	}

	fmt.Printf("Created list: %s\n", list.Name)
	return nil
}

// runListRenameInteractive runs the interactive list rename flow.
func runListRenameInteractive() error {
	if err := requireInteractive(); err != nil {
		return err
	}

	selected, err := listSelect("Select list to rename")
	if err != nil {
		return err
	}
	if selected == "" {
		return nil // cancelled
	}

	var newName string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("New name").
				Description(fmt.Sprintf("Renaming '%s'", selected)).
				Value(&newName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				}),
		),
	).WithTheme(huhTheme())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return err
	}

	if err := listSvc.RenameList(selected, newName); err != nil {
		return err
	}

	fmt.Printf("Renamed list '%s' to '%s'\n", selected, newName)
	return nil
}

// runListDeleteInteractive runs the interactive list deletion flow.
func runListDeleteInteractive(preselected string) error {
	if err := requireInteractive(); err != nil {
		return err
	}

	selected := preselected
	if selected == "" {
		var err error
		selected, err = listSelect("Select list to delete")
		if err != nil {
			return err
		}
		if selected == "" {
			return nil // cancelled
		}
	}

	confirmed, err := huhConfirm(fmt.Sprintf("Delete list '%s' and all its reminders?", selected))
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	if err := listSvc.DeleteList(selected); err != nil {
		return err
	}

	fmt.Printf("Deleted list: %s\n", selected)
	return nil
}
