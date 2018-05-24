package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newRemoteListCmd(c *Context) *cobra.Command {
	var namespace string
	remoteListCmd := &cobra.Command{
		Use:   "list",
		Short: "List Remotes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			rl, err := c.GitkubeClientSet.Gitkube().Remotes(namespace).List(metav1.ListOptions{})
			if err != nil {
				return errors.Wrap(err, "listing remote failed")
			}
			for _, r := range rl.Items {
				fmt.Println(r.GetName())
			}
			return nil
		},
	}

	f := remoteListCmd.Flags()
	f.StringVarP(&namespace, "namespace", "n", "default", "namespace for listing")

	return remoteListCmd
}
