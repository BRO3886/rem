package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/go-eventkit/reminders"
	"github.com/BRO3886/rem/internal/service"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	noColor      bool

	exec        *service.Executor
	reminderSvc *service.ReminderService
	listSvc     *service.ListService
)

func init() {
	client, err := reminders.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize Reminders access: %v\n", err)
		os.Exit(1)
	}
	exec = service.NewExecutor()
	reminderSvc = service.NewReminderService(client, exec)
	listSvc = service.NewListService(client, exec)
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
