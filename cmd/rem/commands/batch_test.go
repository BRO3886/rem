package commands

import (
	"errors"
	"testing"
)

func TestApplyBatchSingleSuccess(t *testing.T) {
	var called []string
	err := applyBatch([]batchTarget{{id: "a", name: "A"}}, "Completed", func(id string) error {
		called = append(called, id)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(called) != 1 || called[0] != "a" {
		t.Fatalf("expected fn called with [a], got %v", called)
	}
}

func TestApplyBatchSingleFailureReturnsRawError(t *testing.T) {
	sentinel := errors.New("boom")
	err := applyBatch([]batchTarget{{id: "a", name: "A"}}, "Completed", func(id string) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected raw sentinel error for single target, got %v", err)
	}
}

func TestApplyBatchMultiSuccess(t *testing.T) {
	var called []string
	targets := []batchTarget{{id: "a", name: "A"}, {id: "b", name: "B"}, {id: "c", name: "C"}}
	err := applyBatch(targets, "Flagged", func(id string) error {
		called = append(called, id)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(called) != 3 || called[0] != "a" || called[1] != "b" || called[2] != "c" {
		t.Fatalf("expected fn called with [a b c] in order, got %v", called)
	}
}

func TestApplyBatchMultiPartialFailureContinues(t *testing.T) {
	var called []string
	targets := []batchTarget{{id: "a", name: "A"}, {id: "b", name: "B"}, {id: "c", name: "C"}}
	err := applyBatch(targets, "Completed", func(id string) error {
		called = append(called, id)
		if id == "b" {
			return errors.New("boom")
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected error for partial failure")
	}
	if got, want := err.Error(), "1 of 3 failed"; got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if len(called) != 3 {
		t.Fatalf("expected all 3 targets attempted despite failure, got %v", called)
	}
}

func TestApplyBatchEmptyTargets(t *testing.T) {
	err := applyBatch(nil, "Completed", func(id string) error {
		t.Fatal("fn should not be called")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
