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

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	corev1 "k8s.io/api/core/v1"

	listerappsv1beta1 "k8s.io/client-go/listers/apps/v1beta1"
	listercorev1 "k8s.io/client-go/listers/core/v1"

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

	servicesLister listercorev1.ServiceLister
	servicesSynced cache.InformerSynced

	deploymentsLister listerappsv1beta1.DeploymentLister
	deploymentsSynced cache.InformerSynced

	remotesGetter typed.RemotesGetter
	remotesLister listers.RemoteLister
	remotesSynced cache.InformerSynced

	remoteworkqueue     workqueue.RateLimitingInterface
	deploymentworkqueue workqueue.RateLimitingInterface
	serviceworkqueue    workqueue.RateLimitingInterface
}

func NewController(
	kubeclientset *kubernetes.Clientset,
	clientset *clientset.Clientset,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	informerFactory informers.SharedInformerFactory) *GitController {

	deploymentInformer := kubeInformerFactory.Apps().V1beta1().Deployments()
	serviceInformer := kubeInformerFactory.Core().V1().Services()
	remoteInformer := informerFactory.Gitkube().V1alpha1().Remotes()

	controller := &GitController{
		kubeclientset: kubeclientset,

		servicesLister: serviceInformer.Lister(),
		servicesSynced: serviceInformer.Informer().HasSynced,

		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,

		remotesGetter: clientset.GitkubeV1alpha1(),
		remotesLister: remoteInformer.Lister(),
		remotesSynced: remoteInformer.Informer().HasSynced,

		remoteworkqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "remote"),
		deploymentworkqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "gitkubed-deploy"),
		serviceworkqueue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "gitkubed-svc"),
	}

	l.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			service := obj.(*corev1.Service)

			if service.Name != gitkubeServiceName {
				return
			}

			controller.serviceEnqueue(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			oldService := old.(*corev1.Service)
			newService := new.(*corev1.Service)

			if oldService.Name != gitkubeServiceName {
				return
			}

			if oldService.ResourceVersion == newService.ResourceVersion {
				return
			}

			controller.serviceEnqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			service := obj.(*corev1.Service)

			if service.Name != gitkubeServiceName {
				return
			}

			controller.serviceEnqueue(obj)
		},
	})

	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment := obj.(*appsv1beta1.Deployment)

			if deployment.Name != gitkubeDeploymentName {
				return
			}

			controller.deploymentEnqueue(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			oldDeployment := old.(*appsv1beta1.Deployment)
			newDeployment := new.(*appsv1beta1.Deployment)

			if oldDeployment.Name != gitkubeDeploymentName {
				return
			}

			if oldDeployment.ResourceVersion == newDeployment.ResourceVersion {
				return
			}

			controller.deploymentEnqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*appsv1beta1.Deployment)

			if deployment.Name != gitkubeDeploymentName {
				return
			}

			controller.deploymentEnqueue(obj)
		},
	})

	remoteInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			controller.remoteEnqueue(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			oldRemote := old.(*v1alpha1.Remote)
			newRemote := new.(*v1alpha1.Remote)
			if oldRemote.ResourceVersion == newRemote.ResourceVersion {
				return
			}

			controller.remoteEnqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			controller.remoteEnqueue(obj)
		},
	})
	return controller
}

func (c *GitController) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.remoteworkqueue.ShutDown()
	defer c.serviceworkqueue.ShutDown()
	defer c.deploymentworkqueue.ShutDown()

	l.Info("Initialising gitkube")

	l.Info("Waiting for cache sync")
	if !cache.WaitForCacheSync(stopCh, c.remotesSynced, c.deploymentsSynced, c.servicesSynced) {
		return fmt.Errorf("timed out waiting for cache sync")
	}
	l.Info("Caches are synced")

	l.Info("Starting remote worker")
	go wait.Until(c.runRemoteWorker, time.Second, stopCh)

	l.Info("Starting deployment worker")
	go wait.Until(c.runDeploymentWorker, time.Second, stopCh)

	l.Info("Starting service worker")
	go wait.Until(c.runServiceWorker, time.Second, stopCh)

	l.Info("Waiting for stop signal")
	<-stopCh
	l.Info("Received stop signal")

	return nil
}
