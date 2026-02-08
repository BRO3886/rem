package commands

import (
	"github.com/BRO3886/rem/internal/applescript"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	noColor      bool

	exec        *applescript.Executor
	reminderSvc *applescript.ReminderService
	listSvc     *applescript.ListService
)

func init() {
	exec = applescript.NewExecutor()
	reminderSvc = applescript.NewReminderService(exec)
	listSvc = applescript.NewListService(exec)
}

var rootCmd = &cobra.Command{
	Use:   "rem",
	Short: "A powerful CLI for macOS Reminders",
	Long: `rem is a command-line interface for interacting with the macOS Reminders app.
It provides full CRUD operations for reminders and lists, natural language date parsing,
import/export capabilities, and a clean terminal UI.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, plain")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
