package commands

import (
	"fmt"
	"os"

	"github.com/BRO3886/rem/internal/reminder"
	"github.com/BRO3886/rem/internal/ui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i"},
	Short:   "Interactive reminder management",
	Long:    `Launch an interactive session for creating and managing reminders.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireInteractive(); err != nil {
			return err
		}
		return runInteractiveMenu()
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

func runInteractiveMenu() error {
	for {
		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("rem interactive").
					Options(
						huh.NewOption("Create a reminder", "create"),
						huh.NewOption("List reminders", "list"),
						huh.NewOption("Complete reminders", "complete"),
						huh.NewOption("Delete reminders", "delete"),
						huh.NewOption("Flag/unflag reminders", "flag"),
						huh.NewOption("Manage lists", "lists"),
						huh.NewOption("Quit", "quit"),
					).
					Value(&choice),
			),
		).WithTheme(huhTheme())

		if err := form.Run(); err != nil {
			if err == huh.ErrUserAborted {
				fmt.Println("Bye!")
				return nil
			}
			return err
		}

		var err error
		switch choice {
		case "create":
			err = runAddInteractive()
		case "list":
			err = runListInteractive()
		case "complete":
			err = runCompleteInteractive(false)
		case "delete":
			err = runDeleteInteractive()
		case "flag":
			err = runFlagMenuInteractive()
		case "lists":
			err = runListMgmtMenuInteractive()
		case "quit":
			fmt.Println("Bye!")
			return nil
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

func runListInteractive() error {
	reminders, err := reminderSvc.ListReminders(&reminder.ListFilter{})
	if err != nil {
		return err
	}
	format := ui.ParseOutputFormat(outputFormat)
	ui.PrintReminders(os.Stdout, reminders, format)
	return nil
}

func runAddInteractive() error {
	lists, err := listSvc.GetLists()
	if err != nil {
		return err
	}

	listOptions := make([]huh.Option[string], 0, len(lists)+1)
	listOptions = append(listOptions, huh.NewOption("Default", ""))
	for _, l := range lists {
		listOptions = append(listOptions, huh.NewOption(l.Name, l.Name))
	}

	priorityOptions := []huh.Option[string]{
		huh.NewOption("None", "none"),
		huh.NewOption("Low", "low"),
		huh.NewOption("Medium", "medium"),
		huh.NewOption("High", "high"),
	}

	var (
		title       string
		notes       string
		dueStr      string
		listName    string
		priorityStr string
		url         string
		tagsStr     string
		flagged     bool
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Value(&title).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("title is required")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("List").
				Options(listOptions...).
				Value(&listName),
			huh.NewInput().
				Title("Notes").
				Description("Optional").
				Value(&notes),
			huh.NewInput().
				Title("Due date").
				Description("e.g., 'tomorrow', 'next friday at 2pm' (optional)").
				Value(&dueStr),
			huh.NewSelect[string]().
				Title("Priority").
				Options(priorityOptions...).
				Value(&priorityStr),
			huh.NewInput().
				Title("URL").
				Description("Optional").
				Value(&url),
			huh.NewInput().
				Title("Tags").
				Description("Comma-separated, optional").
				Value(&tagsStr),
			huh.NewConfirm().
				Title("Flagged?").
				Affirmative("Yes").
				Negative("No").
				Value(&flagged),
		),
	).WithTheme(huhTheme())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return err
	}

	r := &reminder.Reminder{
		Name:     title,
		Body:     notes,
		ListName: listName,
		URL:      url,
		Flagged:  flagged,
		Tags:     mergeTags(nil, append(tagsFromTitle(title), parseTagList(tagsStr)...), nil),
		Priority: reminder.ParsePriority(priorityStr),
	}

	if dueStr != "" {
		dueDate, err := parseDate(dueStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not parse due date '%s': %v\n", dueStr, err)
		} else {
			r.DueDate = &dueDate
		}
	}
	// Interactive mode has no --silent/--remind-me equivalents yet, so the
	// default "alarm at due time" behavior always applies when a due date was
	// entered. Uses the same helper as the non-interactive path.
	alarms, err := buildAlarms(r.DueDate != nil, "", false)
	if err != nil {
		return err
	}
	r.Alarms = alarms

	id, err := reminderSvc.CreateReminder(r)
	if err != nil {
		return err
	}

	fmt.Printf("Created reminder: %s (ID: %s)\n", r.Name, shortIDStr(id))
	return nil
}

func runFlagMenuInteractive() error {
	var action string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Flag action").
				Options(
					huh.NewOption("Flag reminders", "flag"),
					huh.NewOption("Unflag reminders", "unflag"),
				).
				Value(&action),
		),
	).WithTheme(huhTheme())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return err
	}

	switch action {
	case "flag":
		return runFlagInteractive("")
	case "unflag":
		return runUnflagInteractive("")
	}
	return nil
}

func runListMgmtMenuInteractive() error {
	var action string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("List management").
				Options(
					huh.NewOption("Create a list", "create"),
					huh.NewOption("Rename a list", "rename"),
					huh.NewOption("Delete a list", "delete"),
				).
				Value(&action),
		),
	).WithTheme(huhTheme())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return err
	}

	switch action {
	case "create":
		return runListCreateInteractive()
	case "rename":
		return runListRenameInteractive()
	case "delete":
		return runListDeleteInteractive("")
	}
	return nil
}
