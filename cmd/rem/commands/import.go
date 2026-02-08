package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BRO3886/rem/internal/export"
	"github.com/spf13/cobra"
)

var (
	importList   string
	importDryRun bool
)

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import reminders from JSON or CSV file",
	Example: `  rem import work.json
  rem import reminders.csv --list "Imported"
  rem import --dry-run data.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer f.Close()

		ext := strings.ToLower(filepath.Ext(filePath))

		var importFunc func() error

		switch ext {
		case ".csv":
			reminders, err := export.ImportCSV(f)
			if err != nil {
				return err
			}
			importFunc = func() error {
				for _, r := range reminders {
					if importList != "" {
						r.ListName = importList
					}
					if importDryRun {
						dueStr := ""
						if r.DueDate != nil {
							dueStr = " (due: " + r.DueDate.Format("2006-01-02 15:04") + ")"
						}
						fmt.Printf("[dry-run] Would create: %s%s [%s]\n", r.Name, dueStr, r.ListName)
						continue
					}
					id, err := reminderSvc.CreateReminder(r)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to create '%s': %v\n", r.Name, err)
						continue
					}
					fmt.Printf("Created: %s (ID: %s)\n", r.Name, shortIDStr(id))
				}
				return nil
			}
		case ".json":
			reminders, err := export.ImportJSON(f)
			if err != nil {
				return err
			}
			importFunc = func() error {
				for _, r := range reminders {
					if importList != "" {
						r.ListName = importList
					}
					if importDryRun {
						dueStr := ""
						if r.DueDate != nil {
							dueStr = " (due: " + r.DueDate.Format("2006-01-02 15:04") + ")"
						}
						fmt.Printf("[dry-run] Would create: %s%s [%s]\n", r.Name, dueStr, r.ListName)
						continue
					}
					id, err := reminderSvc.CreateReminder(r)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to create '%s': %v\n", r.Name, err)
						continue
					}
					fmt.Printf("Created: %s (ID: %s)\n", r.Name, shortIDStr(id))
				}
				return nil
			}
		default:
			return fmt.Errorf("unsupported file format: %s (use .json or .csv)", ext)
		}

		return importFunc()
	},
}

func init() {
	importCmd.Flags().StringVarP(&importList, "list", "l", "", "Import all reminders into this list")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Preview import without creating reminders")
	rootCmd.AddCommand(importCmd)
}
