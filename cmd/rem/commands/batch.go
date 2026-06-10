package commands

import (
	"fmt"
)

type batchTarget struct {
	id   string
	name string
}

// runBatchAction resolves each arg to a reminder, applies fn to each, and
// prints "<verb>: <name>" per success. All IDs are resolved before any
// mutation so a typo'd ID fails the whole command instead of half-applying.
func runBatchAction(args []string, verb string, fn func(id string) error) error {
	var targets []batchTarget
	for _, arg := range args {
		r, err := findReminderByID(arg)
		if err != nil {
			return err
		}
		targets = append(targets, batchTarget{id: r.ID, name: r.Name})
	}
	return applyBatch(targets, verb, fn)
}

func applyBatch(targets []batchTarget, verb string, fn func(id string) error) error {
	if len(targets) == 1 {
		if err := fn(targets[0].id); err != nil {
			return err
		}
		fmt.Printf("%s: %s\n", verb, targets[0].name)
		return nil
	}

	var failed int
	for _, t := range targets {
		if err := fn(t.id); err != nil {
			failed++
			fmt.Printf("Error: %s: %v\n", t.name, err)
			continue
		}
		fmt.Printf("%s: %s\n", verb, t.name)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d failed", failed, len(targets))
	}
	return nil
}
