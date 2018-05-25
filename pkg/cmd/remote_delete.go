package cmd

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newRemoteDeleteCmd(c *Context) *cobra.Command {
	var namespace string
	remoteDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Remote from the cluster",
		Long:  "Delete a Gitkube remote from the cluster",
		Example: `  # Delete remote called 'example':
  gitkube remote delete example`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.GitkubeClientSet.Gitkube().Remotes(namespace).Delete(args[0], &metav1.DeleteOptions{})
			if err != nil {
				return errors.Wrap(err, "deleting remote failed")
			}
			logrus.Infof("remote '%s' deleted", args[0])
			return nil
		},
	}

	f := remoteDeleteCmd.Flags()
	f.StringVarP(&namespace, "namespace", "n", "default", "namespace of the remote")

	return remoteDeleteCmd
}
