package controller

import (
	"fmt"

	l "github.com/sirupsen/logrus"

	"github.com/hasura/gitkube/pkg/controller/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

func (c *GitController) configmapEnqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.configmapworkqueue.AddRateLimited(key)
}

func (c *GitController) runConfigMapWorker() {
	for c.processNextConfigMapWorkItem() {
	}
}

func (c *GitController) processNextConfigMapWorkItem() bool {
	obj, shutdown := c.configmapworkqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.configmapworkqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.configmapworkqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in configmapworkqueue but got %#v", obj))
			return nil
		}

		if err := c.syncConfigMapHandler(key); err != nil {
			return fmt.Errorf("error syncing %s: %s", key, err.Error())
		}

		c.configmapworkqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *GitController) syncConfigMapHandler(key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	l.Infof("syncing configmap: %s.%s", ns, name)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	ciconf, err := c.configmapsLister.ConfigMaps(gitkubeNamespace).Get(gitkubeConfigMapName)
	if err != nil {
		//create config map
		return err
	}

	var oldHash, newHash string

	if len(ciconf.Data) > 0 {
		oldHash = util.GetMD5Hash(ciconf.Data["remotes.json"])
	}

	ciconfCopy := ciconf.DeepCopy()

	ciconfCopy.Data = make(map[string]string)
	ciconfCopy.Data["remotes.json"] = CreateGitkubeConf(c.kubeclientset, c.remotesLister)

	newHash = util.GetMD5Hash(ciconfCopy.Data["remotes.json"])

	if newHash == oldHash {
		return nil
	}

	_, err = c.kubeclientset.CoreV1().ConfigMaps(gitkubeNamespace).Update(ciconfCopy)
	if err != nil {
		return err
	}

	gitkubedeployment, err := c.kubeclientset.AppsV1beta1().Deployments(gitkubeNamespace).Get(gitkubeDeploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	err = RestartDeployment(c.kubeclientset, gitkubedeployment)
	if err != nil {
		return nil
	}

	return nil
}
