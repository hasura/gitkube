package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

func (c *GitController) deploymentEnqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.deploymentworkqueue.AddRateLimited(key)
}

func (c *GitController) runDeploymentWorker() {
	for c.processNextDeploymentWorkItem() {
	}
}

func (c *GitController) processNextDeploymentWorkItem() bool {
	obj, shutdown := c.deploymentworkqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.deploymentworkqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.deploymentworkqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in deploymentworkqueue but got %#v", obj))
			return nil
		}

		if err := c.syncDeploymentHandler(key); err != nil {
			return fmt.Errorf("error syncing %s: %s", key, err.Error())
		}

		c.deploymentworkqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *GitController) syncDeploymentHandler(key string) error {
	return nil
}
