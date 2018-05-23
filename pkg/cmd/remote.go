package cmd

import "github.com/spf13/cobra"

func newRemoteCmd(c *Context) *cobra.Command {
	remoteCmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage gitkube remotes on a cluster",
	}

	remoteCmd.AddCommand(
		newRemoteGenerateCmd(c),
	)

	return remoteCmd
}
