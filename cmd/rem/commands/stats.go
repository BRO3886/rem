package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/BRO3886/rem/internal/reminder"
	"github.com/BRO3886/rem/internal/ui"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show reminder statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		format := ui.ParseOutputFormat(outputFormat)

		allReminders, err := reminderSvc.ListReminders(nil)
		if err != nil {
			return err
		}

		lists, err := listSvc.GetLists()
		if err != nil {
			return err
		}

		total := len(allReminders)
		completed := 0
		flagged := 0
		overdue := 0
		now := time.Now()

		for _, r := range allReminders {
			if r.Completed {
				completed++
			}
			if r.Flagged {
				flagged++
			}
			if r.IsOverdue(now) {
				overdue++
			}
		}

		incomplete := total - completed
		completionRate := 0.0
		if total > 0 {
			completionRate = float64(completed) / float64(total) * 100
		}

		if format == ui.FormatJSON {
			fmt.Fprintf(os.Stdout, `{
  "total": %d,
  "completed": %d,
  "incomplete": %d,
  "flagged": %d,
  "overdue": %d,
  "completion_rate": %.1f,
  "lists": %d
}
`, total, completed, incomplete, flagged, overdue, completionRate, len(lists))
			return nil
		}

		fmt.Println("Reminder Statistics")
		fmt.Println("===================")
		fmt.Printf("Total:           %d\n", total)
		fmt.Printf("Completed:       %d\n", completed)
		fmt.Printf("Incomplete:      %d\n", incomplete)
		fmt.Printf("Flagged:         %d\n", flagged)
		fmt.Printf("Overdue:         %d\n", overdue)
		fmt.Printf("Completion Rate: %.1f%%\n", completionRate)
		fmt.Printf("Lists:           %d\n", len(lists))

		if len(lists) > 0 {
			fmt.Println("\nPer List:")
			ui.PrintLists(os.Stdout, lists, ui.FormatTable, true)
		}

		return nil
	},
}

var overdueCmd = &cobra.Command{
	Use:   "overdue",
	Short: "Show overdue reminders",
	RunE: func(cmd *cobra.Command, args []string) error {
		incomplete := false
		reminders, err := reminderSvc.ListReminders(&reminder.ListFilter{
			ListName:  overdueList,
			Completed: &incomplete,
		})
		if err != nil {
			return err
		}

		now := time.Now()
		var overdue []*reminder.Reminder
		for _, r := range reminders {
			if r.IsOverdue(now) {
				overdue = append(overdue, r)
			}
		}

		format := ui.ParseOutputFormat(outputFormat)
		if len(overdue) == 0 {
			if format != ui.FormatJSON {
				fmt.Println("No overdue reminders!")
			} else {
				fmt.Println("[]")
			}
			return nil
		}

		ui.PrintReminders(os.Stdout, overdue, format)
		return nil
	},
}

var overdueList string
var upcomingDays int
var upcomingList string

var upcomingCmd = &cobra.Command{
	Use:   "upcoming",
	Short: "Show upcoming reminders",
	RunE: func(cmd *cobra.Command, args []string) error {
		incomplete := false
		now := time.Now()
		cutoff := now.AddDate(0, 0, upcomingDays)
		reminders, err := reminderSvc.ListReminders(&reminder.ListFilter{
			ListName:  upcomingList,
			Completed: &incomplete,
			DueBefore: &cutoff,
			DueAfter:  &now,
		})
		if err != nil {
			return err
		}

		format := ui.ParseOutputFormat(outputFormat)
		if len(reminders) == 0 {
			if format != ui.FormatJSON {
				fmt.Printf("No reminders due in the next %d days.\n", upcomingDays)
			} else {
				fmt.Println("[]")
			}
			return nil
		}

		ui.PrintReminders(os.Stdout, reminders, format)
		return nil
	},
}

func init() {
	overdueCmd.Flags().StringVarP(&overdueList, "list", "l", "", "Filter by list name")
	upcomingCmd.Flags().IntVar(&upcomingDays, "days", 7, "Number of days to look ahead")
	upcomingCmd.Flags().StringVarP(&upcomingList, "list", "l", "", "Filter by list name")
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(overdueCmd)
	rootCmd.AddCommand(upcomingCmd)
}
