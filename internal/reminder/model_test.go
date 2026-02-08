package reminder

import "testing"

func TestPriorityString(t *testing.T) {
	tests := []struct {
		p    Priority
		want string
	}{
		{PriorityNone, "none"},
		{PriorityHigh, "high"},
		{Priority(2), "high"},
		{Priority(3), "high"},
		{Priority(4), "high"},
		{PriorityMedium, "medium"},
		{PriorityLow, "low"},
		{Priority(6), "low"},
		{Priority(7), "low"},
		{Priority(8), "low"},
		{Priority(-1), "none"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.p.String()
			if got != tt.want {
				t.Errorf("Priority(%d).String() = %q, want %q", tt.p, got, tt.want)
			}
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input string
		want  Priority
	}{
		{"high", PriorityHigh},
		{"h", PriorityHigh},
		{"1", PriorityHigh},
		{"medium", PriorityMedium},
		{"med", PriorityMedium},
		{"m", PriorityMedium},
		{"5", PriorityMedium},
		{"low", PriorityLow},
		{"l", PriorityLow},
		{"9", PriorityLow},
		{"none", PriorityNone},
		{"", PriorityNone},
		{"invalid", PriorityNone},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParsePriority(tt.input)
			if got != tt.want {
				t.Errorf("ParsePriority(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
