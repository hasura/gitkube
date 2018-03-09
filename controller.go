/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	l "github.com/sirupsen/logrus"
	// corev1 "k8s.io/api/core/v1"
	// kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/apimachinery/pkg/runtime/schema"
	// appsv1beta2 "k8s.io/api/apps/v1beta2"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	util "github.com/hasura/gitkube/util"

	v1alpha1 "github.com/hasura/gitkube/pkg/apis/gitkube.sh/v1alpha1"

	clientset "github.com/hasura/gitkube/pkg/client/clientset/versioned"

	typed "github.com/hasura/gitkube/pkg/client/clientset/versioned/typed/gitkube/v1alpha1"
	informers "github.com/hasura/gitkube/pkg/client/informers/externalversions"
	listers "github.com/hasura/gitkube/pkg/client/listers/gitkube/v1alpha1"
)

const (
	gitkubeDeploymentName = "gitkubed"
	gitkubeServiceName    = "gitkubed"
	gitkubeConfigMapName  = "gitkube-ci-conf"
	gitkubeNamespace      = "kube-system"
)

type GitController struct {
	kubeclientset *kubernetes.Clientset

	remotesGetter typed.RemotesGetter
	remoteLister  listers.RemoteLister
	remotesSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface
}

func NewController(
	kubeclientset *kubernetes.Clientset,
	clientset *clientset.Clientset,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	informerFactory informers.SharedInformerFactory) *GitController {

	// deploymentInformer := kubeInformerFactory.Apps().V1beta2().Deployments()
	remoteInformer := informerFactory.Gitkube().V1alpha1().Remotes()

	controller := &GitController{
		kubeclientset: kubeclientset,
		remotesGetter: clientset.GitkubeV1alpha1(),
		remoteLister:  remoteInformer.Lister(),
		remotesSynced: remoteInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "git-kube"),
	}

	l.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	remoteInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.enqueue(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			oldRemote := old.(*v1alpha1.Remote)
			newRemote := new.(*v1alpha1.Remote)
			if oldRemote.ResourceVersion == newRemote.ResourceVersion {
				return
			}

			controller.enqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			controller.enqueue(obj)
		},
	})
	return controller
}

func (c *GitController) enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *GitController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	l.Info("Starting Danava controller")

	l.Info("waiting for cache sync")
	if !cache.WaitForCacheSync(stopCh, c.remotesSynced) {
		return fmt.Errorf("timed out waiting for cache sync")
	}
	l.Info("caches are synced")

	l.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	// wait until we're told to stop
	l.Info("waiting for stop signal")
	<-stopCh
	l.Info("received stop signal")

	return nil
}

func (c *GitController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *GitController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing %s: %s", key, err.Error())
		}

		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *GitController) syncHandler(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	l.Infof("syncing remote: %s.%s", ns, name)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	remote, err := c.remoteLister.Remotes(ns).Get(name)

	if err != nil {
		return err
	}

	ciconf, err := c.kubeclientset.CoreV1().ConfigMaps(gitkubeNamespace).Get(gitkubeConfigMapName, metav1.GetOptions{})
	if err != nil {
		//create config map
		return err
	}

	ciconf.Data = make(map[string]string)
	ciconf.Data["remotes.json"] = CreateGitkubeConf(c.remoteLister)

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
