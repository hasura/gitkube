package cmd

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/api/extensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newUninstallCmd(c *Context) *cobra.Command {
	var opts uninstallOptions
	opts.Context = c

	// installCmd defines the install command
	var uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Gitkube components from a cluster",
		Long:  "Remove Gitkube and all dependencies from the cluster",
		Example: `  # Uninstall gitkube:
  gitkube uninstall`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := opts.run()
			if err != nil {
				return errors.Wrap(err, "uninstalling gitkube failed")
			}
			return nil
		},
	}

	f := uninstallCmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "kube-system", "namespace where gitkube is installed")

	return uninstallCmd
}

type uninstallOptions struct {
	Namespace string

	Context *Context
}

func (o *uninstallOptions) run() error {
	// delete CRD
	err := o.Context.APIExtensionsClientSet.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(CRDName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Warnf("CustomResourceDefinition %s does not exist, nothing to do", CRDName)
		} else {
			return errors.Wrapf(err, "deleting CustomResourceDefinition %s failed", CRDName)
		}
	} else {
		logrus.Infof("CustomResourceDefinition %s deleted", CRDName)
	}

	// delete SA
	err = o.Context.KubeClientSet.Core().ServiceAccounts(o.Namespace).Delete(SAName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Warnf("ServiceAccount %s does not exist, nothing to do", SAName)
		} else {
			return errors.Wrapf(err, "deleting ServiceAccount %s failed", SAName)
		}
	} else {
		logrus.Infof("ServiceAccount %s deleted", SAName)
	}

	// delete CRB
	err = o.Context.KubeClientSet.RbacV1().ClusterRoleBindings().Delete(CRBName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Warnf("ClusterRoleBinding %s does not exist, nothing to do", CRBName)
		} else {
			return errors.Wrapf(err, "deleting ClusterRoleBinding %s failed", CRBName)
		}
	} else {
		logrus.Infof("ClusterRoleBinding %s deleted", CRBName)
	}

	// delete CM
	err = o.Context.KubeClientSet.Core().ConfigMaps(o.Namespace).Delete(CMName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Warnf("ConfigMap %s does not exist, nothing to do", CMName)
		} else {
			return errors.Wrapf(err, "deleting ConfigMap %s failed", CMName)
		}
	} else {
		logrus.Infof("ConfigMap %s deleted", CMName)
	}

	// delete gitkubed deployment
	_, err = o.Context.KubeClientSet.ExtensionsV1beta1().Deployments(o.Namespace).UpdateScale(GitkubedDeploymentName, &v1beta1.Scale{ObjectMeta: metav1.ObjectMeta{Name: GitkubedDeploymentName, Namespace: o.Namespace}, Spec: v1beta1.ScaleSpec{Replicas: 0}})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Warnf("Deployment %s does not exist, nothing to do", GitkubedDeploymentName)
		} else {
			return errors.Wrapf(err, "scaling Deployment %s to 0 failed", GitkubedDeploymentName)
		}
	} else {
		logrus.Infof("Deployment %s scaled to zero", GitkubedDeploymentName)
	}
	err = o.Context.KubeClientSet.ExtensionsV1beta1().Deployments(o.Namespace).Delete(GitkubedDeploymentName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Warnf("Deployment %s does not exist, nothing to do", GitkubedDeploymentName)
		} else {
			return errors.Wrapf(err, "deleting Deployment %s failed", GitkubedDeploymentName)
		}
	} else {
		logrus.Infof("Deployment %s deleted", GitkubedDeploymentName)
	}

	// delete gitkube-controller deployment
	_, err = o.Context.KubeClientSet.ExtensionsV1beta1().Deployments(o.Namespace).UpdateScale(GitkubeControllerDeploymentName, &v1beta1.Scale{ObjectMeta: metav1.ObjectMeta{Name: GitkubeControllerDeploymentName, Namespace: o.Namespace}, Spec: v1beta1.ScaleSpec{Replicas: 0}})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Warnf("Deployment %s does not exist, nothing to do", GitkubeControllerDeploymentName)
		} else {
			return errors.Wrapf(err, "scaling Deployment %s to 0 failed", GitkubeControllerDeploymentName)
		}
	} else {
		logrus.Infof("Deployment %s scaled to zero", GitkubeControllerDeploymentName)
	}
	err = o.Context.KubeClientSet.ExtensionsV1beta1().Deployments(o.Namespace).Delete(GitkubeControllerDeploymentName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Warnf("Deployment %s does not exist, nothing to do", GitkubeControllerDeploymentName)
		} else {
			return errors.Wrapf(err, "deleting Deployment %s failed", GitkubeControllerDeploymentName)
		}
	} else {
		logrus.Infof("Deployment %s deleted", GitkubeControllerDeploymentName)
	}

	// delete SVC
	err = o.Context.KubeClientSet.Core().Services(o.Namespace).Delete(SVCName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logrus.Warnf("Service %s does not exist, nothing to do", SVCName)
		} else {
			return errors.Wrapf(err, "deleting Service %s failed", SVCName)
		}
	} else {
		logrus.Infof("Service %s deleted", SVCName)
	}

	logrus.Infof("all gitkube components removed from '%s' namespace", o.Namespace)

	return nil
}
