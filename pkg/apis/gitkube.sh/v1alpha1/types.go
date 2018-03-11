package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//Remote
type Remote struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemoteSpec   `json:"spec"`
	Status RemoteStatus `json:"status"`
}

type RemoteSpec struct {
	AuthorizedKeys []string         `json:"authorizedKeys"`
	Registry       RegistrySpec     `json:"registry"`
	Deployments    []DeploymentSpec `json:"deployments"`
}

type RemoteStatus struct {
	RemoteUrl     string `json:"remoteUrl"`
	RemoteUrlDesc string `json:"remoteUrlDesc"`
}

type RegistrySpec struct {
	Url         string          `json:"url,omitempty"`
	Credentials CredentialsSpec `json:"credentials,omitempty"`
}

type CredentialsSpec struct {
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type DeploymentSpec struct {
	Name       string          `json:"name"`
	Containers []ContainerSpec `json:"containers"`
}

type ContainerSpec struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Dockerfile string `json:"dockerfile"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//RemoteList
type RemoteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Remote `json:"items"`
}
