package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/rem/internal/export"
	"github.com/BRO3886/rem/internal/reminder"
	"github.com/spf13/cobra"
)

var (
	exportList       string
	exportFormat     string
	exportOutputFile string
	exportIncomplete bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export reminders to JSON or CSV",
	Example: `  rem export --list Work --format json > work.json
  rem export --format csv --output reminders.csv
  rem export --incomplete --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := &reminder.ListFilter{
			ListName: exportList,
		}
		if exportIncomplete {
			v := false
			filter.Completed = &v
		}

		reminders, err := reminderSvc.ListReminders(filter)
		if err != nil {
			return err
		}

		var w *os.File
		if exportOutputFile != "" {
			f, err := os.Create(exportOutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer f.Close()
			w = f
		} else {
			w = os.Stdout
		}

		switch exportFormat {
		case "csv":
			return export.ExportCSV(w, reminders)
		default:
			return export.ExportJSON(w, reminders)
		}
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportList, "list", "l", "", "Export reminders from a specific list")
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "Export format: json, csv")
	exportCmd.Flags().StringVar(&exportOutputFile, "output-file", "", "Output file path (default: stdout)")
	exportCmd.Flags().BoolVar(&exportIncomplete, "incomplete", false, "Export only incomplete reminders")
	rootCmd.AddCommand(exportCmd)
}
