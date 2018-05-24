package cmd

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

	f := installCmd.Flags()

	f.StringVar(&opts.Expose, "expose", "LoadBalancer", "k8s service type to expose the gitkubed deployment")
	f.StringVarP(&opts.Namespace, "namespace", "n", "kube-system", "namespace to create gitkube resources")

	return installCmd
}

type InstallOptions struct {
	Context   *Context
	Expose    string
	Namespace string
}

// InstallManifests installs all gitkube related manifests on the cluster
func (o *InstallOptions) InstallManifests() error {
	// create CRD
	crd := newCRD()
	_, err := o.Context.APIExtensionsClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&crd)
	if err != nil {
		return errors.Wrap(err, "error creating CustomResourceDefinition")
	}
	logrus.Infof("CustomResourceDefinition %s created", crd.GetName())

	// create SA
	sa := newSA(o.Namespace)
	_, err = o.Context.KubeClientSet.Core().ServiceAccounts(o.Namespace).Create(&sa)
	if err != nil {
		return errors.Wrap(err, "error creating ServiceAccount")
	}
	logrus.Infof("ServiceAccount %s created", sa.GetName())

	return nil
}
