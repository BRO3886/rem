package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version info set by ldflags at build time.
var (
	Version   = "dev"
	BuildTime = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("rem %s (built %s)\n", Version, BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
