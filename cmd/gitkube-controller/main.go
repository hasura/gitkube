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

package main

import (
	"flag"
	"log"
	"os"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/hasura/gitkube/pkg/client/clientset/versioned"
	informers "github.com/hasura/gitkube/pkg/client/informers/externalversions"
	controller "github.com/hasura/gitkube/pkg/controller"
	"github.com/hasura/gitkube/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Parse()

	namespace := os.Getenv("GITKUBE_NAMESPACE")
	controller.SetGitkubeNamespace(namespace)

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	gkClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building example clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	gkInformerFactory := informers.NewSharedInformerFactory(gkClient, time.Second*30)

	controller := controller.NewController(kubeClient, gkClient, kubeInformerFactory, gkInformerFactory)

	go kubeInformerFactory.Start(stopCh)
	go gkInformerFactory.Start(stopCh)

	if err = controller.Run(stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
