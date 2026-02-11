//go:build !darwin

package client

import "fmt"

var errUnsupported = fmt.Errorf("rem client: only supported on macOS (darwin)")

// Client provides methods for interacting with macOS Reminders.
// On non-darwin platforms, all methods return an unsupported error.
type Client struct{}

// New creates a new Reminders client.
// Returns an error on non-darwin platforms.
func New() (*Client, error) { return nil, errUnsupported }
