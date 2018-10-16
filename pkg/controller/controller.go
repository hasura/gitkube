/*
Copyright 2018 The Gitkube Authors.

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

package controller

import (
	"fmt"
	"time"

	l "github.com/sirupsen/logrus"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	corev1 "k8s.io/api/core/v1"
	listercorev1 "k8s.io/client-go/listers/core/v1"

	clientset "github.com/hasura/gitkube/pkg/client/clientset/versioned"
	typed "github.com/hasura/gitkube/pkg/client/clientset/versioned/typed/gitkube.sh/v1alpha1"
	informers "github.com/hasura/gitkube/pkg/client/informers/externalversions"
	listers "github.com/hasura/gitkube/pkg/client/listers/gitkube.sh/v1alpha1"
)

const (
	gitkubeDeploymentName = "gitkubed"
	gitkubeServiceName    = "gitkubed"
	gitkubeConfigMapName  = "gitkube-ci-conf"
)

var (
	gitkubeNamespace string
)

func SetGitkubeNamespace(ns string) {
	if ns == "" {
		gitkubeNamespace = "kube-system"

	} else {
		gitkubeNamespace = ns
	}
}

type GitController struct {
	kubeclientset *kubernetes.Clientset

	configmapsLister listercorev1.ConfigMapLister
	configmapsSynced cache.InformerSynced

	remotesGetter typed.RemotesGetter
	remotesLister listers.RemoteLister
	remotesSynced cache.InformerSynced

	remoteworkqueue    workqueue.RateLimitingInterface
	configmapworkqueue workqueue.RateLimitingInterface
}

// NewController returns a GitController
func NewController(
	kubeclientset *kubernetes.Clientset,
	clientset *clientset.Clientset,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	informerFactory informers.SharedInformerFactory) *GitController {

	configmapInformer := kubeInformerFactory.Core().V1().ConfigMaps()
	remoteInformer := informerFactory.Gitkube().V1alpha1().Remotes()

	controller := &GitController{
		kubeclientset: kubeclientset,

		configmapsLister: configmapInformer.Lister(),
		configmapsSynced: configmapInformer.Informer().HasSynced,

		remotesGetter: clientset.GitkubeV1alpha1(),
		remotesLister: remoteInformer.Lister(),
		remotesSynced: remoteInformer.Informer().HasSynced,

		remoteworkqueue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "remote"),
		configmapworkqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "gitkube-ci-conf"),
	}

	l.Info("Setting up event handlers")
	configmapInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cm := obj.(*corev1.ConfigMap)

			if cm.Name != gitkubeConfigMapName {
				return
			}

			controller.configmapEnqueue(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			oldCm := old.(*corev1.ConfigMap)
			newCm := new.(*corev1.ConfigMap)

			if oldCm.Name != gitkubeConfigMapName {
				return
			}

			if oldCm.ResourceVersion == newCm.ResourceVersion {
				return
			}

			controller.configmapEnqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			cm := obj.(*corev1.ConfigMap)

			if cm.Name != gitkubeConfigMapName {
				return
			}

			controller.configmapEnqueue(obj)
		},
	})

	remoteInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.remoteEnqueue(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			controller.remoteEnqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			controller.remoteEnqueue(obj)
		},
	})
	return controller
}

// Run starts the worker threads for remote and configmap work queues
func (c *GitController) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.remoteworkqueue.ShutDown()
	defer c.configmapworkqueue.ShutDown()

	l.Info("Initialising gitkube")

	l.Info("Waiting for cache sync")
	if !cache.WaitForCacheSync(stopCh, c.remotesSynced, c.configmapsSynced) {
		return fmt.Errorf("timed out waiting for cache sync")
	}
	l.Info("Caches are synced")

	l.Info("Starting remote worker")
	go wait.Until(c.runRemoteWorker, time.Second, stopCh)

	l.Info("Starting configmap worker")
	go wait.Until(c.runConfigMapWorker, time.Second, stopCh)

	l.Info("Waiting for stop signal")
	<-stopCh
	l.Info("Received stop signal")

	return nil
}
