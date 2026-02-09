package applescript

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Executor runs AppleScript and JXA commands via osascript,
// and the compiled Swift helper for fast reads via EventKit.
type Executor struct {
	timeout    time.Duration
	helperPath string
}

// NewExecutor creates a new Executor.
func NewExecutor() *Executor {
	return &Executor{
		timeout:    120 * time.Second,
		helperPath: findHelperPath(),
	}
}

// findHelperPath locates the reminders-helper binary.
// It checks: next to the current executable, then in PATH.
func findHelperPath() string {
	// Check next to the current executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidate := filepath.Join(dir, "reminders-helper")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// Check PATH
	if p, err := exec.LookPath("reminders-helper"); err == nil {
		return p
	}
	return ""
}

// RunHelper executes the Swift EventKit helper with the given arguments.
func (e *Executor) RunHelper(args ...string) (string, error) {
	if e.helperPath == "" {
		return "", fmt.Errorf("reminders-helper not found; run 'make build' to compile it")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.helperPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("helper error: %s", errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
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
