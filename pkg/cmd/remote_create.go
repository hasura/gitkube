package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"
	gitkubescheme "github.com/hasura/gitkube/pkg/client/clientset/versioned/scheme"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func newRemoteCreateCmd(c *Context) *cobra.Command {
	var opts remoteCreateOptions
	opts.Context = c
	var remote v1alpha1.Remote
	opts.Remote = &remote
	remoteCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create gitkube remote on a cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := opts.run()
			if err != nil {
				return errors.Wrap(err, "creating remote failed")
			}
			return nil
		},
	}

	f := remoteCreateCmd.Flags()
	f.StringVarP(&opts.SpecFile, "file", "f", "", "spec file to read")

	return remoteCreateCmd
}

type remoteCreateOptions struct {
	SpecFile string
	RawData  []byte

	Remote *v1alpha1.Remote

	Context *Context
}

func (o *remoteCreateOptions) run() error {
	data, err := ioutil.ReadFile(o.SpecFile)
	if err != nil {
		return errors.Wrap(err, "error reading file")
	}
	o.RawData = data
	gclient := o.Context.GitkubeClientSet

	d := gitkubescheme.Codecs.UniversalDeserializer()
	obj, _, err := d.Decode(o.RawData, nil, nil)
	if err != nil {
		return errors.Wrap(err, "parsing yaml as a valid remote failed")
	}
	o.Remote = obj.(*v1alpha1.Remote)

	_, err = gclient.Gitkube().Remotes(o.Remote.GetNamespace()).Create(o.Remote)
	if err != nil {
		return errors.Wrap(err, "k8s api error")
	}
	logrus.Infof("remote %s created", o.Remote.GetName())

	logrus.Info("waiting for remote url")

	w, err := gclient.Gitkube().Remotes(o.Remote.GetNamespace()).Watch(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "watching remote failed")
	}
	for ev := range w.ResultChan() {
		if ev.Type == watch.Modified || ev.Type == watch.Added {
			r := ev.Object.(*v1alpha1.Remote)
			if r.GetName() == o.Remote.GetName() {
				status := r.Status
				if status.RemoteUrl != "" {
					fmt.Println(status.RemoteUrl)
					// TODO: print git remote add instructions
					break
				}
				if status.RemoteUrlDesc != "" {
					// TODO: desc appear only on errors?
					logrus.Errorln(status.RemoteUrlDesc)
				}
			}
		}
	}
	return nil
}
