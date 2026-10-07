package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/BRO3886/rem/internal/reminder"
)

func TestResolveParentID(t *testing.T) {
	child := &reminder.Reminder{ID: "CHILD-1111", Name: "child"}
	lookup := func(id string) (*reminder.Reminder, error) {
		switch id {
		case "PAR":
			return &reminder.Reminder{ID: "PARENT-2222", Name: "parent"}, nil
		case "CHI":
			return child, nil
		}
		return nil, errors.New("reminder not found: " + id)
	}

	tests := []struct {
		value   string
		want    string
		wantErr string
	}{
		{value: "none", want: ""},
		{value: "", want: ""},
		{value: "PAR", want: "PARENT-2222"},
		{value: "CHI", wantErr: "own parent"},
		{value: "NOPE", wantErr: "not found"},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := resolveParentID(tt.value, child, lookup)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
