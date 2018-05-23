package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	_ "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	_ "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

func NewInstallCmd(c *Context) *cobra.Command {
	var opts InstallOptions
	opts.Context = c

	// installCmd defines the install command
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install Gitkube on a Kubernetes cluster",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := opts.InstallManifests()
			if err != nil {
				return errors.Wrap(err, "installing gitkube on cluster failed")
			}
			return nil
		},
	}

	installCmd.Flags().StringVar(&opts.Expose, "expose", "LoadBalancer", "k8s service type to expose the gitkubed deployment")

	return installCmd
}

type InstallOptions struct {
	Context *Context
	Expose  string
}

// InstallManifests installs all gitkube related manifests on the cluster
func (o *InstallOptions) InstallManifests() error {

	return nil
}
