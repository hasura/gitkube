package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version of the CLI/API, set during build time
	version = "v0.0.0-unset"
)

// GetVersion returns the current version string
func GetVersion() string {
	return version
}

var versionCmd = &cobra.Command{
	Use:               "version",
	Short:             "Output the cli version",
	Args:              cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) { return nil },
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(GetVersion())
		return nil
	},
	Example: `  # Print the version string:
  gitkube version`,
}
