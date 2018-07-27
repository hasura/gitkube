package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/homedir"
)

func newRemoteGenerateCmd(c *Context) *cobra.Command {
	var opts remoteGenerateOptions
	remote := v1alpha1.Remote{}
	remote.TypeMeta = metav1.TypeMeta{
		APIVersion: v1alpha1.SchemeGroupVersion.String(),
		Kind:       "Remote",
	}
	opts.Context = c
	opts.Remote = &remote
	remoteGenerateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a Remote spec in an interactive manner",
		Long:  "An interactive prompt-driven approach to generate Remote spec, rather than writing yaml by hand",
		Example: `  # Generate a remote and save it as 'example-remote.yaml':
  gitkube remote generate -f example-remote.yaml

  # Shows interactive prompts to type-in/select options.
  # Contacts the cluster to create docker registry secret (if provided)`,
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
	f.StringVarP(&opts.OutputFile, "output-file", "f", "", "write generated spec to this file")

	return remoteGenerateCmd
}

type remoteGenerateOptions struct {
	Context      *Context
	OutputFormat string
	OutputFile   string

	Remote *v1alpha1.Remote
}

type dockerRegistry struct {
	Server      string
	URL         string
	Username    string
	Password    string
	AuthEncoded string
}

type dockerConfigJson struct {
	Auths auths `json:"auths"`
}
type auths struct {
	IndexDockerIO indexDockerIO `json:"https://index.docker.io/v1/"`
}
type indexDockerIO struct {
	Auth string `json:"auth"`
}

const dockerIOServer = "https://index.docker.io/v1/"

type DockerConfigEntryWithAuth struct {
	// +optional
	Username string `json:"username,omitempty"`
	// +optional
	Password string `json:"password,omitempty"`
	// +optional
	Email string `json:"email,omitempty"`
	// +optional
	Auth string `json:"auth,omitempty"`
}

func handlePromptError(err error) error {
	if err == promptui.ErrInterrupt {
		return errors.New("cancelled by user")
	}
	return errors.Wrap(err, "prompt failed")
}

func (o *remoteGenerateOptions) Run() error {

	p := promptui.Prompt{
		Label:   "Remote name",
		Default: "example",
	}
	r, err := p.Run()
	if err != nil {
		return handlePromptError(err)
	}
	o.Remote.SetName(r)

	p = promptui.Prompt{
		Label:   "Namespace",
		Default: "default",
	}
	r, err = p.Run()
	if err != nil {
		return handlePromptError(err)
	}
	o.Remote.SetNamespace(r)

	keyFile := filepath.Join(homedir.HomeDir(), ".ssh", "id_rsa.pub")
	p = promptui.Prompt{
		Label:     "SSH public key file",
		Default:   keyFile,
		IsVimMode: true,
	}
	r, err = p.Run()
	if err != nil {
		return handlePromptError(err)
	}
	sshKey, err := ioutil.ReadFile(r)
	if err != nil {
		// TODO: (show command to) generate SSH key?
		return errors.Wrap(err, "cannot read ssh key file")
	}
	o.Remote.Spec.AuthorizedKeys = []string{string(sshKey)}

	INIT_YAML := "K8s YAML Manifests"
	INIT_HELM := "Helm Chart"
	INIT_NONE := "None"
	ps := promptui.Select{
		Label: "Initialisation",
		Items: []string{INIT_YAML, INIT_HELM, INIT_NONE},
	}
	_, r, err = ps.Run()
	if err != nil {
		return handlePromptError(err)
	}
	if r == INIT_HELM || r == INIT_YAML {
		p = promptui.Prompt{
			Label: "Manifests/Chart directory",
		}
		r, err := p.Run()
		if err != nil {
			return handlePromptError(err)
		}
		o.Remote.Spec.Manifests = v1alpha1.ManifestSpec{
			Path: r,
		}
		// TODO: ask for HELM Specific stuff
	}

	// Docker registry
	var server, registry, username, password, email string

	optionOne := "docker.io login info (not found)"
	dr, err := getDockerRegistryDetails()
	if err == nil {
		// getting docker registry succeeded, dr has details
		optionOne = fmt.Sprintf("docker.io/%s", dr.Username)
	}
dockerRegistryPrompt:
	ps = promptui.Select{
		Label: "Choose docker registry",
		Items: []string{
			optionOne,
			"Specify a different registry",
			"Skip for now",
		},
	}
	i, r, err := ps.Run()
	if err != nil {
		return handlePromptError(err)
	}

	var skipRegistry bool
	switch i {
	case 0:
		if dr == nil {
			logrus.Error("couldn't find docker.io login on this system, choose another option")
			goto dockerRegistryPrompt
		} else {
			// use existing username and password
			registry = dr.URL
			server = dr.Server
			username = dr.Username
			password = dr.Password
			email = fmt.Sprintf("%s@%s", dr.Username, dr.URL)
		}
	case 1:
		// prompt for registry, username, password
		p = promptui.Prompt{
			Label:   "Docker registry server",
			Default: dockerIOServer,
		}
		server, err = p.Run()
		if err != nil {
			return handlePromptError(err)
		}
		p = promptui.Prompt{
			Label:   "Registry URL",
			Default: "docker.io",
		}
		registry, err = p.Run()
		if err != nil {
			return handlePromptError(err)
		}
		p = promptui.Prompt{
			Label: "Username",
		}
		username, err = p.Run()
		if err != nil {
			return handlePromptError(err)
		}
		p = promptui.Prompt{
			Label: "Password",
			Mask:  '*',
		}
		password, err = p.Run()
		if err != nil {
			return handlePromptError(err)
		}
		p = promptui.Prompt{
			Label:   "Email",
			Default: fmt.Sprintf("%s@%s", username, registry),
		}
		email, err = p.Run()
		if err != nil {
			return handlePromptError(err)
		}
	case 2:
		// do nothing
		skipRegistry = true
	}
	if !skipRegistry {
		s := corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      o.Remote.GetName() + "-regsecret",
				Namespace: o.Remote.GetNamespace(),
			},
			Type: corev1.SecretTypeDockerConfigJson,
		}
		dcewa := DockerConfigEntryWithAuth{
			Username: username,
			Password: password,
			Email:    email,
			Auth:     getDockerAuthString(username, password),
		}
		sd := map[string]map[string]DockerConfigEntryWithAuth{
			"auths": {
				server: dcewa,
			},
		}
		data, err := json.Marshal(sd)
		if err != nil {
			return errors.Wrap(err, "error converting dockerconfig to json")
		}
		s.StringData = map[string]string{
			".dockerconfigjson": string(data),
		}

		client := o.Context.KubeClientSet
		_, err = client.CoreV1().Secrets(o.Remote.GetNamespace()).Create(&s)
		if err != nil {
			return errors.Wrap(err, "error creating docker-registry secret")
		}
		logrus.Info("Created docker-registry secret")

		o.Remote.Spec.Registry = v1alpha1.RegistrySpec{
			Url: fmt.Sprintf("%s/%s", registry, username),
			Credentials: v1alpha1.CredentialsSpec{
				SecretRef: o.Remote.GetName() + "-regsecret",
			},
		}
	}
	// Deployments
	o.Remote.Spec.Deployments = []v1alpha1.DeploymentSpec{}
takeDeployment:
	d := v1alpha1.DeploymentSpec{}
	d.Containers = []v1alpha1.ContainerSpec{}
	p = promptui.Prompt{
		Label:   "Deployment name",
		Default: "app",
	}
	r, err = p.Run()
	if err != nil {
		return handlePromptError(err)
	}
	d.Name = r
takeContainer:
	c := v1alpha1.ContainerSpec{}
	p = promptui.Prompt{
		Label:   "Container name",
		Default: "app",
	}
	r, err = p.Run()
	if err != nil {
		return handlePromptError(err)
	}
	c.Name = r
	p = promptui.Prompt{
		Label:   "Dockerfile path",
		Default: "Dockerfile",
	}
	r, err = p.Run()
	if err != nil {
		return handlePromptError(err)
	}
	c.Dockerfile = r
	p = promptui.Prompt{
		Label:   "Build context path",
		Default: "./",
	}
	r, err = p.Run()
	if err != nil {
		return handlePromptError(err)
	}
	c.Path = r

	d.Containers = append(d.Containers, c)

	p = promptui.Prompt{
		Label:     "Add another container",
		IsConfirm: true,
	}
	r, err = p.Run()
	if err != nil && err != promptui.ErrAbort {
		return handlePromptError(err)
	}
	if r == "y" {
		goto takeContainer
	} // TODO: handle else?

	o.Remote.Spec.Deployments = append(o.Remote.Spec.Deployments, d)

	p = promptui.Prompt{
		Label:     "Add another deployment",
		IsConfirm: true,
	}
	r, err = p.Run()
	if err != nil && err != promptui.ErrAbort {
		return handlePromptError(err)
	}
	if r == "y" {
		goto takeDeployment
	} // TODO: handle else?

	spec, err := yaml.Marshal(o.Remote)
	if err != nil {
		errors.Wrap(err, "error marshalling remote spec")
	}

	var output []byte
	switch o.OutputFormat {
	case "json":
		// TODO: pretty print json
		output, err = yaml.YAMLToJSON(spec)
		if err != nil {
			return errors.Wrap(err, "error converting to json")
		}
	case "yaml":
		output = spec
	default:
		// TODO: do this validation earlier itself
		return errors.Wrap(err, "unknown output format")
	}
	fmt.Println("")
	fmt.Print(string(output))
	fmt.Println("")

	if o.OutputFile != "" {
		var overwrite bool
		if _, err := os.Stat(o.OutputFile); err == nil {
			// file exists
			p = promptui.Prompt{
				Label:     fmt.Sprintf("Overwrite %s", o.OutputFile),
				IsConfirm: true,
			}
			r, err = p.Run()
			if err != nil && err != promptui.ErrAbort {
				return handlePromptError(err)
			}
			if r == "y" {
				overwrite = true
			}
		} else {
			overwrite = true
		}
		if overwrite {
			err := ioutil.WriteFile(o.OutputFile, []byte(output), 0644)
			if err != nil {
				return errors.Wrap(err, "writing to file failed")
			}
			logrus.Infof("spec saved as '%s'\r\n", o.OutputFile)
		}
	}
	return nil
}

func getDockerAuthString(username, password string) string {
	data := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func getDockerRegistryDetails() (*dockerRegistry, error) {
	dockerconfigjsonFile := filepath.Join(homedir.HomeDir(), ".docker", "config.json")
	dockerconfigjsonData, err := ioutil.ReadFile(dockerconfigjsonFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading docker config failed")
	}
	var dcj dockerConfigJson
	err = json.Unmarshal(dockerconfigjsonData, &dcj)
	if err != nil {
		return nil, errors.Wrap(err, "parsing dockerconfigjson failed")
	}

	var username, password string

	if a := dcj.Auths.IndexDockerIO.Auth; a != "" {
		d, err := base64.StdEncoding.DecodeString(a)
		if err != nil {
			return nil, errors.Wrap(err, "decoding docker auth failed")
		}
		auth := strings.Split(string(d), ":")
		username = auth[0]
		password = auth[1]
	}

	return &dockerRegistry{
		URL:         "docker.io",
		Server:      dockerIOServer,
		Username:    username,
		Password:    password,
		AuthEncoded: dcj.Auths.IndexDockerIO.Auth,
	}, nil
}
