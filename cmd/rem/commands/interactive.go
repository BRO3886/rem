package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/BRO3886/rem/internal/parser"
	"github.com/BRO3886/rem/internal/reminder"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i"},
	Short:   "Interactive reminder management",
	Long:    `Launch an interactive session for creating and managing reminders.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractiveMenu()
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

func runInteractiveMenu() error {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n--- rem interactive ---")
		fmt.Println("1) Create a reminder")
		fmt.Println("2) List reminders")
		fmt.Println("3) Complete a reminder")
		fmt.Println("4) Delete a reminder")
		fmt.Println("5) List all lists")
		fmt.Println("6) Create a list")
		fmt.Println("q) Quit")
		fmt.Print("\nChoice: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			if err := runAddInteractive(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		case "2":
			reminders, err := reminderSvc.ListReminders(&reminder.ListFilter{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
			for i, r := range reminders {
				status := "[ ]"
				if r.Completed {
					status = "[x]"
				}
				dueStr := ""
				if r.DueDate != nil {
					dueStr = " (due: " + r.DueDate.Format("Jan 02, 15:04") + ")"
				}
				fmt.Printf("  %d. %s %s%s [%s]\n", i+1, status, r.Name, dueStr, r.ListName)
			}
		case "3":
			fmt.Print("Reminder ID (or prefix): ")
			id, _ := reader.ReadString('\n')
			id = strings.TrimSpace(id)
			r, err := findReminderByID(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
			if err := reminderSvc.CompleteReminder(r.ID); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
			fmt.Printf("Completed: %s\n", r.Name)
		case "4":
			fmt.Print("Reminder ID (or prefix): ")
			id, _ := reader.ReadString('\n')
			id = strings.TrimSpace(id)
			r, err := findReminderByID(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
			fmt.Printf("Delete '%s'? (y/N): ", r.Name)
			confirm, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(confirm)) == "y" {
				if err := reminderSvc.DeleteReminder(r.ID); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					continue
				}
				fmt.Printf("Deleted: %s\n", r.Name)
			}
		case "5":
			lists, err := listSvc.GetLists()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
			for _, l := range lists {
				fmt.Printf("  - %s (%d reminders)\n", l.Name, l.Count)
			}
		case "6":
			fmt.Print("List name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			if name == "" {
				fmt.Println("Name cannot be empty.")
				continue
			}
			list, err := listSvc.CreateList(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
			fmt.Printf("Created list: %s\n", list.Name)
		case "q", "Q", "quit", "exit":
			fmt.Println("Bye!")
			return nil
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

func runAddInteractive() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Title: ")
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("title is required")
	}

	// Get available lists
	lists, err := listSvc.GetLists()
	if err != nil {
		return err
	}

	fmt.Println("Available lists:")
	for i, l := range lists {
		fmt.Printf("  %d. %s\n", i+1, l.Name)
	}
	fmt.Print("Choose list (number or name, Enter for default): ")
	listChoice, _ := reader.ReadString('\n')
	listChoice = strings.TrimSpace(listChoice)

	listName := ""
	if listChoice != "" {
		// Try as number first
		for i, l := range lists {
			if fmt.Sprintf("%d", i+1) == listChoice {
				listName = l.Name
				break
			}
		}
		if listName == "" {
			listName = listChoice
		}
	}

	fmt.Print("Notes (optional): ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	fmt.Print("Due date (e.g., 'tomorrow', 'next friday at 2pm', Enter to skip): ")
	dueStr, _ := reader.ReadString('\n')
	dueStr = strings.TrimSpace(dueStr)

	fmt.Print("Priority (high/medium/low/none, Enter for none): ")
	priorityStr, _ := reader.ReadString('\n')
	priorityStr = strings.TrimSpace(priorityStr)

	fmt.Print("URL (optional): ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	fmt.Print("Flagged? (y/N): ")
	flaggedStr, _ := reader.ReadString('\n')
	flagged := strings.TrimSpace(strings.ToLower(flaggedStr)) == "y"

	r := &reminder.Reminder{
		Name:     title,
		Body:     notes,
		ListName: listName,
		URL:      url,
		Flagged:  flagged,
		Priority: reminder.ParsePriority(priorityStr),
	}

	if dueStr != "" {
		dueDate, err := parser.ParseDate(dueStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not parse due date '%s': %v\n", dueStr, err)
		} else {
			r.DueDate = &dueDate
		}
	}

	id, err := reminderSvc.CreateReminder(r)
	if err != nil {
		return err
	}

	fmt.Printf("Created reminder: %s (ID: %s)\n", r.Name, shortIDStr(id))
	return nil
}
