package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/BRO3886/rem/cmd/rem/commands"
)

// Set by ldflags at build time.
var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	if runtime.GOOS != "darwin" {
		fmt.Fprintln(os.Stderr, "Error: rem requires macOS (uses AppleScript to interact with the Reminders app)")
		os.Exit(2)
	}

	commands.Version = version
	commands.BuildTime = buildTime

	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
