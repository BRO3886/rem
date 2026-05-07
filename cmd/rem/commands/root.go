package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/go-eventkit/reminders"
	"github.com/BRO3886/rem/internal/service"
	"github.com/BRO3886/rem/internal/skills"
	"github.com/BRO3886/rem/internal/update"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	noColor      bool

	exec        *service.Executor
	reminderSvc *service.ReminderService
	listSvc     *service.ListService
)

// updateResultCh receives the background update check result (if any).
var updateResultCh = make(chan *update.Result, 1)

func init() {
	client, err := reminders.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize Reminders access: %v\n", err)
		os.Exit(1)
	}
	exec = service.NewExecutor()
	reminderSvc = service.NewReminderService(client)
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor || os.Getenv("NO_COLOR") != "" {
			color.NoColor = true
		}

		// Start background update check
		if shouldCheckForUpdate(cmd) {
			go func() {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					updateResultCh <- nil
					return
				}
				updateResultCh <- update.Check(homeDir, Version)
			}()
		} else {
			updateResultCh <- nil
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		printUpdateNotice(cmd)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, plain")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// shouldCheckForUpdate returns false for commands/contexts where the check should be skipped.
func shouldCheckForUpdate(cmd *cobra.Command) bool {
	// Skip if env var set
	if os.Getenv("REM_NO_UPDATE_CHECK") != "" {
		return false
	}

	// Skip for dev builds
	if Version == "" || Version == "dev" {
		return false
	}

	// Skip for meta commands
	name := cmd.Name()
	if name == "version" || name == "completion" || name == "skills" {
		return false
	}

	// Skip if --output json (scripting context)
	if outputFormat == "json" {
		return false
	}

	// Skip if stdout is not a TTY (piped output)
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	if fi.Mode()&os.ModeCharDevice == 0 {
		return false
	}

	return true
}

// printUpdateNotice prints update and skills staleness notices to stderr.
func printUpdateNotice(_ *cobra.Command) {
	// Collect update result (non-blocking — if goroutine isn't done, skip)
	var result *update.Result
	select {
	case result = <-updateResultCh:
	default:
		// Goroutine still running, don't wait
		result = nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	yellow := color.New(color.FgYellow)

	if result != nil && result.HasUpdate {
		fmt.Fprintln(os.Stderr)
		yellow.Fprintf(os.Stderr, "A new version of rem is available: %s → %s\n", Version, result.Latest)
		fmt.Fprintf(os.Stderr, "Update: curl -fsSL https://rem.sidv.dev/install | bash\n")
	}

	// Check skills staleness (local only, no HTTP)
	printSkillsStalenessNotice(homeDir)
}

// printSkillsStalenessNotice checks if installed skills are outdated.
func printSkillsStalenessNotice(homeDir string) {
	if Version == "" || Version == "dev" {
		return
	}

	targets := skills.InstalledTargets(skills.DefaultTargets(homeDir))
	for _, t := range targets {
		installed := skills.InstalledVersion(t)
		if installed != "" && installed != Version {
			yellow := color.New(color.FgYellow)
			fmt.Fprintln(os.Stderr)
			yellow.Fprintf(os.Stderr, "Installed skills are outdated (%s). Run: rem skills install\n", installed)
			return // Only show once
		}
	}
}
