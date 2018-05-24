package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	// enable gcp auth provider
	gitkubeCS "github.com/hasura/gitkube/pkg/client/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var rootCmd = &cobra.Command{
	Use:           "gitkube",
	Short:         "Manage Gitkube installation on a Kubernetes cluster",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		// create kubernetes client
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{
			ClusterDefaults: clientcmd.ClusterDefaults,
		}
		if currentContext.KubeContext != "" {
			configOverrides.CurrentContext = currentContext.KubeContext
		}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
		kconfig, err := kubeConfig.ClientConfig()
		if err != nil {
			return errors.Wrap(err, "unable to build kubeconfig")
		}
		// create the clientset
		clientset, err := kubernetes.NewForConfig(kconfig)
		if err != nil {
			return errors.Wrap(err, "unable to build clientset")
		}
		currentContext.KubeClientSet = clientset

		gitkubeclientset, err := gitkubeCS.NewForConfig(kconfig)
		if err != nil {
			return errors.Wrap(err, "unable to build gitkube clientset")
		}
		currentContext.GitkubeClientSet = gitkubeclientset

		apiextensionsclientset, err := apiextensionsclient.NewForConfig(kconfig)
		if err != nil {
			return errors.Wrap(err, "unable to build apiextensionsclientset")
		}
		currentContext.APIExtensionsClientSet = apiextensionsclientset

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

// Context holds the contextual information for each execution
type Context struct {
	KubeContext            string
	Namespace              string
	KubeClientSet          *kubernetes.Clientset
	GitkubeClientSet       *gitkubeCS.Clientset
	APIExtensionsClientSet *apiextensionsclient.Clientset
}

var currentContext Context

func init() {
	// global flags
	// TODO: read defaults from env vars
	rootCmd.PersistentFlags().StringVar(&currentContext.KubeContext, "kube-context", "", "kubecontext to connect")

	// sub-commands
	rootCmd.AddCommand(
		docsCmd,
		versionCmd,
		NewInstallCmd(&currentContext),
		newRemoteCmd(&currentContext),
	)

}
