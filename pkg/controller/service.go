package controller

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

func (c *GitController) serviceEnqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.serviceworkqueue.AddRateLimited(key)
}

func (c *GitController) runServiceWorker() {
	for c.processNextServiceWorkItem() {
	}
}

func (c *GitController) processNextServiceWorkItem() bool {
	obj, shutdown := c.serviceworkqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.serviceworkqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.serviceworkqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in serviceworkqueue but got %#v", obj))
			return nil
		}

		if err := c.syncServiceHandler(key); err != nil {
			return fmt.Errorf("error syncing %s: %s", key, err.Error())
		}

		c.serviceworkqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *GitController) syncServiceHandler(key string) error {
	return nil
}
