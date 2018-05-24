package cmd

import "github.com/spf13/cobra"

func newRemoteCmd(c *Context) *cobra.Command {
	remoteCmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage Gitkube Remotes on a cluster",
	}

	remoteCmd.AddCommand(
		newRemoteGenerateCmd(c),
		newRemoteCreateCmd(c),
		newRemoteDeleteCmd(c),
		newRemoteListCmd(c),
	)

	return remoteCmd
}
