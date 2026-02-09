package applescript

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Executor runs AppleScript and JXA commands via osascript.
// Reads are handled by the internal/eventkit package (cgo + EventKit).
type Executor struct {
	timeout time.Duration
}

// NewExecutor creates a new Executor.
func NewExecutor() *Executor {
	return &Executor{
		timeout: 120 * time.Second,
	}
}

// Run executes an AppleScript and returns the output.
func (e *Executor) Run(script string) (string, error) {
	return e.RunContext(context.Background(), script)
}

// RunContext executes an AppleScript with a context for cancellation.
func (e *Executor) RunContext(ctx context.Context, script string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("applescript error: %s", errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunJXA executes a JavaScript for Automation script and returns the output.
func (e *Executor) RunJXA(script string) (string, error) {
	return e.RunJXAContext(context.Background(), script)
}

// RunJXAContext executes a JXA script with a context for cancellation.
func (e *Executor) RunJXAContext(ctx context.Context, script string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "osascript", "-l", "JavaScript", "-e", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("jxa error: %s", errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// EscapeString escapes a string for safe use in AppleScript.
func EscapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// EscapeJXA escapes a string for safe use in JXA JavaScript strings.
func EscapeJXA(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	return s
}

// FormatDate formats a Go time.Time into an AppleScript date construction string.
func FormatDate(t time.Time) string {
	return fmt.Sprintf(`(current date) + 0
set year of result to %d
set month of result to %d
set day of result to %d
set hours of result to %d
set minutes of result to %d
set seconds of result to %d
set theDate to result`,
		t.Year(), int(t.Month()), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// FormatDateInline creates a self-contained AppleScript date expression
// that can be used inline in property lists.
func FormatDateInline(varName string, t time.Time) string {
	return fmt.Sprintf(`set %s to current date
set year of %s to %d
set month of %s to %d
set day of %s to %d
set hours of %s to %d
set minutes of %s to %d
set seconds of %s to %d`,
		varName,
		varName, t.Year(),
		varName, int(t.Month()),
		varName, t.Day(),
		varName, t.Hour(),
		varName, t.Minute(),
		varName, t.Second())
}
