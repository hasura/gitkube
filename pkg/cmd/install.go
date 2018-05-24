package cmd

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
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

	// create CRB
	crb := newCRB(o.Namespace)
	_, err = o.Context.KubeClientSet.Rbac().ClusterRoleBindings().Create(&crb)
	if err != nil {
		return errors.Wrap(err, "error creating ClusterRoleBinding")
	}
	logrus.Infof("ClusterRoleBinding %s created", crb.GetName())

	// create CM
	cm := newCM(o.Namespace)
	_, err = o.Context.KubeClientSet.Core().ConfigMaps(o.Namespace).Create(&cm)
	if err != nil {
		return errors.Wrap(err, "error creating ConfigMap")
	}
	logrus.Infof("ConfigMap %s created", cm.GetName())

	// create gitkubed
	gitkubed := newGitkubed(o.Namespace)
	_, err = o.Context.KubeClientSet.ExtensionsV1beta1().Deployments(o.Namespace).Create(&gitkubed)
	if err != nil {
		return errors.Wrapf(err, "error creating %s Deployment", gitkubed.GetName())
	}
	logrus.Infof("Deployment %s created", gitkubed.GetName())

	// create gitkube-controller
	gitkubeController := newGitkubeController(o.Namespace)
	_, err = o.Context.KubeClientSet.ExtensionsV1beta1().Deployments(o.Namespace).Create(&gitkubeController)
	if err != nil {
		return errors.Wrapf(err, "error creating %s Deployment", gitkubeController.GetName())
	}
	logrus.Infof("Deployment %s created", gitkubeController.GetName())

	// expose gitkubed deployment
	svcType := corev1.ServiceType(o.Expose)
	svc := newSVC(o.Namespace, svcType)
	_, err = o.Context.KubeClientSet.Core().Services(o.Namespace).Create(&svc)
	if err != nil {
		return errors.Wrapf(err, "error creating Service")
	}
	logrus.Infof("Service %s created", svc.GetName())

	logrus.Infof("gitkube installed in '%s' namespace", o.Namespace)

	return nil
}
