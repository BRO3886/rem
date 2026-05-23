package commands

import (
	"fmt"
	"strings"

	"github.com/BRO3886/rem/internal/reminder"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	updateName        string
	updateNotes       string
	updateDue         string
	updatePriority    string
	updateURL         string
	updateFlagged     string
	updateAddTagsBulk string
	updateRemoveBulk  string
	updateList        string
	updateRemindMe    string
	updateRepeat      string
	updateInteractive bool
)

var updateCmd = &cobra.Command{
	Use:     "update [id]",
	Aliases: []string{"edit"},
	Short:   "Update an existing reminder",
	Long:    `Update properties of an existing reminder by its ID.`,
	Example: `  rem update abc12345 --due "next monday"
  rem update abc12345 --notes "Updated notes" --priority medium
  rem edit abc12345 --name "New title"
  rem update abc12345 --list "Work"
  rem update abc12345 --remind-me 15m
  rem update abc12345 --repeat "weekly on mon,fri"
  rem update abc12345 --repeat none
  rem update -i
  rem update -i abc12345`,
	Args: func(cmd *cobra.Command, args []string) error {
		if updateInteractive {
			return cobra.RangeArgs(0, 1)(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateInteractive {
			var id string
			if len(args) > 0 {
				id = args[0]
			}
			return runUpdateInteractive(id)
		}

		id := args[0]

		r, err := findReminderByID(id)
		if err != nil {
			return err
		}

		updates := make(map[string]any)

		if cmd.Flags().Changed("name") {
			updates["name"] = updateName
		}
		if cmd.Flags().Changed("notes") {
			updates["body"] = updateNotes
		}
		if cmd.Flags().Changed("url") {
			updates["url"] = updateURL
		}
		if cmd.Flags().Changed("due") {
			if updateDue == "" || updateDue == "none" {
				updates["due_date"] = nil
			} else {
				t, err := parseDate(updateDue)
				if err != nil {
					return fmt.Errorf("invalid due date: %w", err)
				}
				updates["due_date"] = t
			}
		}
		if cmd.Flags().Changed("priority") {
			updates["priority"] = reminder.ParsePriority(updatePriority)
		}
		if cmd.Flags().Changed("flagged") {
			updates["flagged"] = updateFlagged == "true" || updateFlagged == "yes"
		}
		titleTags := []string(nil)
		if cmd.Flags().Changed("name") {
			titleTags = tagsFromTitle(updateName)
		}
		if len(titleTags) > 0 || updateAddTagsBulk != "" || updateRemoveBulk != "" {
			updates["tags"] = mergeTagUpdates(r.Tags, titleTags, updateAddTagsBulk, updateRemoveBulk)
		}
		if cmd.Flags().Changed("list") {
			updates["list"] = updateList
		}
		if cmd.Flags().Changed("remind-me") {
			if updateRemindMe == "" || updateRemindMe == "none" {
				updates["alarms"] = nil
			} else {
				alarm, err := parseAlarm(updateRemindMe)
				if err != nil {
					return err
				}
				updates["alarms"] = []reminder.Alarm{alarm}
			}
		}
		if cmd.Flags().Changed("repeat") {
			if updateRepeat == "" || updateRepeat == "none" {
				updates["recurrence"] = nil
			} else {
				rule, err := parseRecurrence(updateRepeat)
				if err != nil {
					return err
				}
				updates["recurrence"] = []reminder.RecurrenceRule{rule}
			}
		}

		if len(updates) == 0 {
			return fmt.Errorf("no updates specified")
		}

		err = reminderSvc.UpdateReminder(r.ID, updates)
		if err != nil {
			return err
		}

		fmt.Printf("Updated reminder: %s\n", r.Name)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateName, "name", "", "New name/title")
	updateCmd.Flags().StringVarP(&updateNotes, "notes", "n", "", "New notes/body")
	updateCmd.Flags().StringVarP(&updateDue, "due", "d", "", "New due date (use 'none' to clear)")
	updateCmd.Flags().StringVarP(&updatePriority, "priority", "p", "", "New priority: high, medium, low, none")
	updateCmd.Flags().StringVarP(&updateURL, "url", "u", "", "New URL")
	updateCmd.Flags().StringVar(&updateFlagged, "flagged", "", "Set flagged status: true/false")
	updateCmd.Flags().StringVar(&updateAddTagsBulk, "add-tags", "", "Add comma-separated native tags")
	updateCmd.Flags().StringVar(&updateRemoveBulk, "remove-tags", "", "Remove comma-separated native tags")
	updateCmd.Flags().StringVarP(&updateList, "list", "l", "", "Move reminder to a different list")
	updateCmd.Flags().StringVar(&updateRemindMe, "remind-me", "", "Set alarm: duration before due (15m, 1h, 2d), 'none' to clear")
	updateCmd.Flags().StringVar(&updateRepeat, "repeat", "", "Set recurrence: daily, weekly, 'weekly on mon,wed,fri', 'none' to clear")
	updateCmd.Flags().BoolVarP(&updateInteractive, "interactive", "i", false, "Update interactively")

	rootCmd.AddCommand(updateCmd)
}

// runUpdateInteractive runs the interactive update form.
func runUpdateInteractive(idArg string) error {
	if err := requireInteractive(); err != nil {
		return err
	}

	var r *reminder.Reminder

	if idArg != "" {
		var err error
		r, err = findReminderByID(idArg)
		if err != nil {
			return err
		}
	} else {
		// Pick a reminder interactively
		incomplete := false
		reminders, err := reminderSvc.ListReminders(&reminder.ListFilter{
			Completed: &incomplete,
		})
		if err != nil {
			return err
		}

		selectedID, err := reminderSelect("Select reminder to update", reminders)
		if err != nil {
			return err
		}
		if selectedID == "" {
			return nil // cancelled
		}

		r, err = reminderSvc.GetReminder(selectedID)
		if err != nil {
			return err
		}
	}

	// Pre-populate form values from current reminder
	name := r.Name
	notes := r.Body
	url := r.URL
	tagsStr := strings.Join(r.Tags, ", ")
	dueStr := ""
	if r.DueDate != nil {
		dueStr = r.DueDate.Local().Format("Jan 02, 2006 3:04 PM")
	}
	priorityStr := r.Priority.String()
	flaggedStr := "no"
	if r.Flagged {
		flaggedStr = "yes"
	}

	// Get lists for the select
	lists, err := listSvc.GetLists()
	if err != nil {
		return err
	}

	listOptions := make([]huh.Option[string], len(lists))
	for i, l := range lists {
		listOptions[i] = huh.NewOption(l.Name, l.Name)
	}
	listName := r.ListName

	priorityOptions := []huh.Option[string]{
		huh.NewOption("None", "none"),
		huh.NewOption("Low", "low"),
		huh.NewOption("Medium", "medium"),
		huh.NewOption("High", "high"),
	}

	flaggedOptions := []huh.Option[string]{
		huh.NewOption("No", "no"),
		huh.NewOption("Yes", "yes"),
	}

	dueDescription := "e.g., 'tomorrow', 'next friday at 2pm', 'none' to clear"
	if r.DueDate != nil {
		dueDescription = fmt.Sprintf("Current: %s — enter new date, 'none' to clear, or leave as-is", r.DueDate.Local().Format("Jan 02, 2006 3:04 PM"))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Value(&name).
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
				Value(&notes),
			huh.NewInput().
				Title("Due date").
				Description(dueDescription).
				Value(&dueStr),
			huh.NewSelect[string]().
				Title("Priority").
				Options(priorityOptions...).
				Value(&priorityStr),
			huh.NewInput().
				Title("URL").
				Value(&url),
			huh.NewInput().
				Title("Tags").
				Description("Comma-separated; clear to remove all tags").
				Value(&tagsStr),
			huh.NewSelect[string]().
				Title("Flagged").
				Options(flaggedOptions...).
				Value(&flaggedStr),
		),
	).WithTheme(huhTheme())

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil
		}
		return err
	}

	// Build updates map by comparing with original values
	updates := make(map[string]any)

	if name != r.Name {
		updates["name"] = name
	}
	if notes != r.Body {
		updates["body"] = notes
	}
	if url != r.URL {
		updates["url"] = url
	}
	if listName != r.ListName {
		updates["list"] = listName
	}
	if priorityStr != r.Priority.String() {
		updates["priority"] = reminder.ParsePriority(priorityStr)
	}
	if tagsStr != strings.Join(r.Tags, ", ") {
		updates["tags"] = mergeTags(nil, append(tagsFromTitle(name), parseTagList(tagsStr)...), nil)
	} else if name != r.Name {
		titleTags := tagsFromTitle(name)
		if len(titleTags) > 0 {
			updates["tags"] = mergeTags(r.Tags, titleTags, nil)
		}
	}

	// Handle due date changes
	newFlagged := flaggedStr == "yes"
	origDueStr := ""
	if r.DueDate != nil {
		origDueStr = r.DueDate.Local().Format("Jan 02, 2006 3:04 PM")
	}
	if dueStr != origDueStr {
		if dueStr == "" || dueStr == "none" {
			updates["due_date"] = nil
		} else {
			t, err := parseDate(dueStr)
			if err != nil {
				return fmt.Errorf("invalid due date: %w", err)
			}
			updates["due_date"] = t
		}
	}

	if newFlagged != r.Flagged {
		updates["flagged"] = newFlagged
	}

	if len(updates) == 0 {
		fmt.Println("No changes made.")
		return nil
	}

	if err := reminderSvc.UpdateReminder(r.ID, updates); err != nil {
		return err
	}

	fmt.Printf("Updated reminder: %s\n", name)
	return nil
}
