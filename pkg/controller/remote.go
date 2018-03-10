package controller

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	l "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	util "github.com/hasura/gitkube/pkg/controller/util"

	v1alpha1 "github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"
	listers "github.com/hasura/gitkube/pkg/client/listers/gitkube/v1alpha1"
)

func (c *GitController) remoteEnqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.remoteworkqueue.AddRateLimited(key)
}

func (c *GitController) runRemoteWorker() {
	for c.processNextRemoteWorkItem() {
	}
}

func (c *GitController) processNextRemoteWorkItem() bool {
	obj, shutdown := c.remoteworkqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.remoteworkqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.remoteworkqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in remoteworkqueue but got %#v", obj))
			return nil
		}

		if err := c.syncRemoteHandler(key); err != nil {
			return fmt.Errorf("error syncing %s: %s", key, err.Error())
		}

		c.remoteworkqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *GitController) syncRemoteHandler(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	l.Infof("syncing remote: %s.%s", ns, name)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	remote, err := c.remotesLister.Remotes(ns).Get(name)

	if err != nil {
		return err
	}

	ciconf, err := c.kubeclientset.CoreV1().ConfigMaps(gitkubeNamespace).Get(gitkubeConfigMapName, metav1.GetOptions{})
	if err != nil {
		//create config map
		return err
	}

	ciconf.Data = make(map[string]string)
	ciconf.Data["remotes.json"] = CreateGitkubeConf(c.remotesLister)

	_, err = c.kubeclientset.CoreV1().ConfigMaps(gitkubeNamespace).Update(ciconf)

	gitkubedeployment, err := c.kubeclientset.AppsV1beta1().Deployments(gitkubeNamespace).Get(gitkubeDeploymentName, metav1.GetOptions{})

	timeannotation := fmt.Sprintf("%v", time.Now().Unix())

	if len(gitkubedeployment.Spec.Template.ObjectMeta.Annotations) == 0 {
		gitkubedeployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}

	gitkubedeployment.Spec.Template.ObjectMeta.Annotations["gitkube/lasteventtimestamp"] = timeannotation
	_, err = c.kubeclientset.AppsV1beta1().Deployments(gitkubeNamespace).Update(gitkubedeployment)

	if err != nil {
		return err
	}

	gitkubeservice, err := c.kubeclientset.CoreV1().Services(gitkubeNamespace).Get(gitkubeServiceName, metav1.GetOptions{})

	extserviceIP, err := util.GetExternalIP(c.kubeclientset, gitkubeservice)

	remoteCopy := remote.DeepCopy()

	if err != nil {
		remoteCopy.Status.RemoteUrl = ""
		remoteCopy.Status.RemoteUrlDesc = err.Error()
	} else {
		qualifiedRemoteName := fmt.Sprintf("%s-%s", ns, name)
		remoteCopy.Status.RemoteUrl = fmt.Sprintf("ssh://%s@%s/~/git/%s", qualifiedRemoteName, extserviceIP, qualifiedRemoteName)
		remoteCopy.Status.RemoteUrlDesc = ""
	}

	_, err = c.remotesGetter.Remotes(ns).Update(remoteCopy)

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
		l.Error(err.Error())
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
