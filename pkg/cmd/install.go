package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gobuffalo/packr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	_ "k8s.io/api/extensions/install"
	k8sv1beta1 "k8s.io/api/extensions/v1beta1"
	_ "k8s.io/api/install"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	_ "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/install"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	_ "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
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
	// c := o.Context.KubeClientSet

	box := packr.NewBox("../../manifests")

	setupYaml, err := box.MustBytes("gitkube-setup.yaml")
	if err != nil {
		return errors.Wrap(err, "error reading embedded manifest files")
	}

	d := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(setupYaml), 4096)

	for {
		o := map[string]interface{}{}
		if err := d.Decode(&o); err != nil {
			if err == io.EOF {
				break
			}
		}
		j, err := json.Marshal(o)
		if err != nil {
			return errors.Wrap(err, "error converting manifest to json")
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode

		obj, gvk, err := decode(j, nil, nil)
		if err != nil {
			return errors.Wrap(err, "error decoding manifest to k8s type")
		}
		switch r := obj.(type) {
		case *apiextensionsv1beta1.CustomResourceDefinition:
			fmt.Println("CRD")
		case *rbacv1beta1.ClusterRoleBinding:
			fmt.Println("CRB")
			fmt.Println(r)
		case *v1.ServiceAccount:
			fmt.Println("SA")
		case *v1.ConfigMap:
			fmt.Println("CM")
		case *k8sv1beta1.Deployment:
			fmt.Println("DPL")
		default:
			fmt.Println(gvk)
			//o is unknown for us
		}
		fmt.Println(o)
	}

	return nil
}
