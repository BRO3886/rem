//go:build darwin

package service

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Executor runs AppleScript commands via osascript.
// Used only for the default list name query.
type Executor struct {
	timeout time.Duration
}

// NewExecutor creates a new Executor.
func NewExecutor() *Executor {
	return &Executor{
		timeout: 30 * time.Second,
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

