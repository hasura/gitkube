package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func newInstallCmd(c *Context) *cobra.Command {
	var opts installOptions
	opts.Context = c

	// installCmd defines the install command
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install Gitkube on a Kubernetes cluster",
		Long:  "Install all Gitkube components on the cluster and expose the gitkubed deployment",
		Example: `  # Install Gitkube in 'kube-system' namespace:
  gitkube install

  # Install in another namespace:
  gitkube install --namespace <your-namespace>

  # The command prompts for a ServiceType to expose gitkubed deployment.
  # Use '--expose' flag to set a ServiceType and skip the prompt
  # Say, 'LoadBalancer':
  gitkube install --expose LoadBalancer`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := opts.installManifests()
			if err != nil {
				return errors.Wrap(err, "installing gitkube failed")
			}
			return nil
		},
	}

	f := installCmd.Flags()

	f.StringVarP(&opts.Expose, "expose", "e", "", "k8s service type to expose the gitkubed deployment")
	f.StringVarP(&opts.Namespace, "namespace", "n", "kube-system", "namespace to create install gitkube resources in")

	return installCmd
}

type installOptions struct {
	Context   *Context
	Expose    string
	Namespace string
}

// InstallManifests installs all gitkube related manifests on the cluster
func (o *installOptions) installManifests() error {
	// create CRD
	crd := newCRD()
	_, err := o.Context.APIExtensionsClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Create(&crd)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			logrus.Warnf("CustomResourceDefinition %s already exists, nothing to do", CRDName)
		} else {
			return errors.Wrap(err, "error creating CustomResourceDefinition")
		}
	} else {
		logrus.Infof("CustomResourceDefinition %s created", crd.GetName())
	}

	// create SA
	sa := newSA(o.Namespace)
	_, err = o.Context.KubeClientSet.Core().ServiceAccounts(o.Namespace).Create(&sa)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			logrus.Warnf("ServiceAccount %s already exists, nothing to do", SAName)
		} else {
			return errors.Wrap(err, "error creating ServiceAccount")
		}
	} else {
		logrus.Infof("ServiceAccount %s created", sa.GetName())
	}

	// create CRB
	crb := newCRB(o.Namespace)
	_, err = o.Context.KubeClientSet.Rbac().ClusterRoleBindings().Create(&crb)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			logrus.Warnf("ClusterRoleBinding %s already exists, nothing to do", CRBName)
		} else {
			return errors.Wrap(err, "error creating ClusterRoleBinding")
		}
	} else {
		logrus.Infof("ClusterRoleBinding %s created", crb.GetName())
	}

	// create CM
	cm := newCM(o.Namespace)
	_, err = o.Context.KubeClientSet.Core().ConfigMaps(o.Namespace).Create(&cm)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			logrus.Warnf("ConfigMap %s already exists, nothing to do", CMName)
		} else {
			return errors.Wrap(err, "error creating ConfigMap")
		}
	} else {
		logrus.Infof("ConfigMap %s created", cm.GetName())
	}

	// create gitkubed
	gitkubed := newGitkubed(o.Namespace)
	_, err = o.Context.KubeClientSet.ExtensionsV1beta1().Deployments(o.Namespace).Create(&gitkubed)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			logrus.Warnf("Deployment %s already exists, nothing to do", GitkubedDeploymentName)
		} else {
			return errors.Wrapf(err, "error creating %s Deployment", gitkubed.GetName())
		}
	} else {
		logrus.Infof("Deployment %s created", gitkubed.GetName())
	}

	// create gitkube-controller
	gitkubeController := newGitkubeController(o.Namespace)
	_, err = o.Context.KubeClientSet.ExtensionsV1beta1().Deployments(o.Namespace).Create(&gitkubeController)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			logrus.Warnf("Deployment %s already exists, nothing to do", GitkubeControllerDeploymentName)
		} else {
			return errors.Wrapf(err, "error creating %s Deployment", gitkubeController.GetName())
		}
	} else {
		logrus.Infof("Deployment %s created", gitkubeController.GetName())
	}

	// expose gitkubed deployment
	if o.Expose == "" {
		p := promptui.Select{
			Label: "Choose the k8s service type to expose gitkubed",
			Items: []string{
				string(corev1.ServiceTypeClusterIP),
				string(corev1.ServiceTypeExternalName),
				string(corev1.ServiceTypeLoadBalancer),
				string(corev1.ServiceTypeNodePort),
			},
		}
		_, o.Expose, err = p.Run()
		if err != nil {
			return errors.Wrap(err, "unable to read prompt")
		}
	}
	svcType := corev1.ServiceType(o.Expose)
	svc := newSVC(o.Namespace, svcType)
	_, err = o.Context.KubeClientSet.Core().Services(o.Namespace).Create(&svc)
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			logrus.Warnf("Service %s already exists, nothing to do", SVCName)
		} else {
			return errors.Wrapf(err, "error creating Service")
		}
	} else {
		logrus.Infof("Service %s created", svc.GetName())
	}

	logrus.Infof("gitkube installed in '%s' namespace", o.Namespace)

	return nil
}
