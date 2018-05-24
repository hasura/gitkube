package cmd

import (
	"fmt"

	"github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newRemoteGenerateCmd(c *Context) *cobra.Command {
	var opts remoteGenerateOptions
	var remoteSpec v1alpha1.RemoteSpec
	opts.Context = c
	opts.RemoteSpec = &remoteSpec
	remoteGenerateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate remote.yaml spec",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := opts.Run()
			if err != nil {
				return errors.Wrap(err, "generating remote failed")
			}
			return nil
		},
	}

	f := remoteGenerateCmd.Flags()
	f.StringVarP(&opts.OutputFormat, "output", "o", "yaml", "file format to output, supports yaml and json")

	return remoteGenerateCmd
}

type initMethod string
type initMethodHelm initMethod
type initMethodKubectl initMethod
type initMethodNone initMethod

type remoteGenerateOptions struct {
	Context      *Context
	OutputFormat string

	SSHPublicKeyFile string

	InitMethod         initMethod
	ManifestsDirectory string

	DockerRegistry dockerRegistry

	RemoteSpec *v1alpha1.RemoteSpec
}

type dockerRegistry struct {
	ConfigFile string
	URL        string
	Username   string
	Password   string
}

func (o *remoteGenerateOptions) Run() error {

	p := promptui.Prompt{
		Label:   "Remote name",
		Default: "myremote",
	}
	result, err := p.Run()
	if err != nil {
		return errors.Wrap(err, "error reading name prompt")
	}

	fmt.Println(result)

	return nil
}
