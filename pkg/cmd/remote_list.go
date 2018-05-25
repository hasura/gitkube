package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newRemoteListCmd(c *Context) *cobra.Command {
	var namespace string
	remoteListCmd := &cobra.Command{
		Use:   "list",
		Short: "List Remotes",
		Long:  "List Gitkube Remotes on a cluster",
		Example: `  # List all remotes:
  gitkube remote list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rl, err := c.GitkubeClientSet.Gitkube().Remotes(namespace).List(metav1.ListOptions{})
			if err != nil {
				return errors.Wrap(err, "listing remote failed")
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME \t URL \t ERROR")
			for _, r := range rl.Items {
				fmt.Fprintf(w, "%s \t %s \t %s\r\n", r.GetName(), r.Status.RemoteUrl, r.Status.RemoteUrlDesc)
			}
			w.Flush()
			return nil
		},
	}

	f := remoteListCmd.Flags()
	f.StringVarP(&namespace, "namespace", "n", "default", "namespace for listing")

	return remoteListCmd
}
