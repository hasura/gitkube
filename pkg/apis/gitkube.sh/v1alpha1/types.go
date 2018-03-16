package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Remote is the definition of a remote
type Remote struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemoteSpec   `json:"spec"`
	Status RemoteStatus `json:"status"`
}

type RemoteSpec struct {
	// SSH public keys for git push authorization
	AuthorizedKeys []string `json:"authorizedKeys"`

	// Registry details for pushing and pulling from external registry
	// +optional
	Registry RegistrySpec `json:"registry"`

	// List of deployment spec.
	// Deployment spec defines which deployments are under gitkube management
	Deployments []DeploymentSpec `json:"deployments"`
}

type RemoteStatus struct {
	// Url of the git remote where the repo is pushed
	RemoteUrl string `json:"remoteUrl"`

	// Description of RemoteUrl
	// Contains error description if RemoteUrl is not available
	RemoteUrlDesc string `json:"remoteUrlDesc"`
}

type RegistrySpec struct {
	// Url of the external registry where built images should be pushed
	// E.g. registry.harbor.io/library
	Url string `json:"url,omitempty"`

	// Credentials for registry
	Credentials CredentialsSpec `json:"credentials,omitempty"`
}

type CredentialsSpec struct {
	// Secret should point to a secret of type docker-registry
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type DeploymentSpec struct {
	// Name of the deployment
	Name string `json:"name"`

	// List of container spec which are part of the deployment
	Containers []ContainerSpec `json:"containers"`
}

type ContainerSpec struct {
	// Name of container
	Name string `json:"name"`

	// Location of source code in the git repo for the container
	Path string `json:"path"`

	// Location of dockerfile for the container
	Dockerfile string `json:"dockerfile"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RemoteList is a list of Remotes
type RemoteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Remote `json:"items"`
}
