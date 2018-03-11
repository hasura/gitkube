package controller

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/api/apps/v1beta1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	v1alpha1 "github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"
	listers "github.com/hasura/gitkube/pkg/client/listers/gitkube/v1alpha1"
)

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

func CreateGitkubeConf(remotelister listers.RemoteLister) string {
	remotes, err := remotelister.List(labels.Everything())
	if err != nil {
		//handle error
	}

	remotesMap := make(map[string]interface{})
	for _, remote := range remotes {
		qualifiedRemoteName := fmt.Sprintf("%s-%s", remote.Namespace, remote.Name)
		remotesMap[qualifiedRemoteName] = CreateRemoteJson(remote)
	}

	bytes, err := json.Marshal(remotesMap)
	if err != nil {
		return ""
	}

	return string(bytes)

}

func CreateRemoteJson(remote *v1alpha1.Remote) interface{} {
	remoteMap := make(map[string]interface{})
	deploymentsMap := make(map[string]interface{})

	for _, deployment := range remote.Spec.Deployments {
		deploymentTag := fmt.Sprintf("%s.%s", remote.Namespace, deployment.Name)
		containersMap := make(map[string]interface{})
		for _, container := range deployment.Containers {
			containersMap[container.Name] = map[string]interface{}{
				"path":       container.Path,
				"dockerfile": container.Dockerfile,
			}
		}
		deploymentsMap[deploymentTag] = containersMap
	}

	remoteMap["authorized-keys"] = strings.Join(remote.Spec.AuthorizedKeys, "\n")
	remoteMap["registry"] = ""
	remoteMap["deployments"] = deploymentsMap

	return remoteMap

}
