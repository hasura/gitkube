package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docsCmd = &cobra.Command{
	Use:               "docs",
	Short:             "Generate markdown docs for all the commands",
	Hidden:            true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) { return nil },
	RunE: func(cmd *cobra.Command, args []string) error {
		err := doc.GenMarkdownTree(rootCmd, "./")
		if err != nil {
			return errors.Wrap(err, "generating docs failed")
		}
		return nil
	},
}
