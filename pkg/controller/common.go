package controller

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	v1alpha1 "github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"
	listers "github.com/hasura/gitkube/pkg/client/listers/gitkube.sh/v1alpha1"
)

// RestartDeployment takes a deployment and annotates the pod spec with current timestamp
// This causes a fresh rollout of the deployment
func RestartDeployment(kubeclientset *kubernetes.Clientset, deployment *v1beta1.Deployment) error {

	timeannotation := fmt.Sprintf("%v", time.Now().Unix())

	if len(deployment.Spec.Template.ObjectMeta.Annotations) == 0 {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations["gitkube/lasteventtimestamp"] = timeannotation

	_, err := kubeclientset.AppsV1beta1().Deployments(deployment.Namespace).Update(deployment)
	if err != nil {
		return err
	}

	return nil
}

// CreateGitkubeConf takes a list of remotes, reshapes it and marshals it into a string
func CreateGitkubeConf(kubeclientset *kubernetes.Clientset, remotelister listers.RemoteLister) string {
	remotes, err := remotelister.List(labels.Everything())
	if err != nil {
		//handle error
	}

	remotesMap := make(map[string]interface{})
	for _, remote := range remotes {
		qualifiedRemoteName := fmt.Sprintf("%s-%s", remote.Namespace, remote.Name)
		remotesMap[qualifiedRemoteName] = CreateRemoteJson(kubeclientset, remote)
	}

	bytes, err := json.Marshal(remotesMap)
	if err != nil {
		return ""
	}

	return string(bytes)

}

// CreateRemoteJson takes a remote and reshapes it
func CreateRemoteJson(kubeclientset *kubernetes.Clientset, remote *v1alpha1.Remote) interface{} {
	remoteMap := make(map[string]interface{})
	deploymentsMap := make(map[string]interface{})

	for _, deployment := range remote.Spec.Deployments {
		deploymentTag := fmt.Sprintf("%s.%s", remote.Namespace, deployment.Name)
		containersMap := make(map[string]interface{})
		for _, container := range deployment.Containers {
			containersMap[container.Name] = map[string]interface{}{
				"path":       container.Path,
				"dockerfile": container.Dockerfile,
				"buildArgs":  container.BuildArgs,
			}
		}
		deploymentsMap[deploymentTag] = containersMap
	}

	remoteMap["authorized-keys"] = strings.Join(remote.Spec.AuthorizedKeys, "\n")
	remoteMap["manifests"] = remote.Spec.Manifests
	remoteMap["registry"] = createRegistryJson(kubeclientset, remote)
	remoteMap["deployments"] = deploymentsMap

	return remoteMap
}

// createRegistryJson takes a remote and returns a reshaped map of its registry
func createRegistryJson(kubeclientset *kubernetes.Clientset, remote *v1alpha1.Remote) interface{} {
	registry := remote.Spec.Registry
	registryMap := make(map[string]interface{})

	if registry == (v1alpha1.RegistrySpec{}) {
		return nil
	}

	registryMap["prefix"] = registry.Url

	var err error
	var secret *v1.Secret

	if registry.Credentials.SecretRef != "" {
		secret, err = kubeclientset.CoreV1().Secrets(remote.Namespace).Get(
			registry.Credentials.SecretRef, metav1.GetOptions{})
	} else {
		secret, err = kubeclientset.CoreV1().Secrets(remote.Namespace).Get(
			registry.Credentials.SecretKeyRef.Name, metav1.GetOptions{})
	}

	if err != nil {
		return registryMap
	}

	secretType := secret.Type
	switch secretType {
	case "kubernetes.io/dockercfg":
		registryMap["dockercfg"] = string(secret.Data[".dockercfg"])
	case "kubernetes.io/dockerconfigjson":
		registryMap["dockerconfigjson"] = string(secret.Data[".dockerconfigjson"])
	default:
		return registryMap
	}

	return registryMap
}
