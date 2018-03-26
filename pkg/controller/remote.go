package controller

import (
	"fmt"

	l "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	util "github.com/hasura/gitkube/pkg/controller/util"
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

	ciconf, err := c.configmapsLister.ConfigMaps(gitkubeNamespace).Get(gitkubeConfigMapName)
	if err != nil {
		//create config map?
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

	//if remotes changed, then update config map and restart deployment
	if newHash != oldHash {

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
			return err
		}

	}

	gitkubeservice, err := c.kubeclientset.CoreV1().Services(gitkubeNamespace).Get(gitkubeServiceName, metav1.GetOptions{})

	extserviceIP, err := util.GetExternalIP(c.kubeclientset, gitkubeservice)

	var remoteUrl, remoteUrlDesc string

	if err != nil {
		remoteUrl = ""
		remoteUrlDesc = err.Error()
	} else {
		qualifiedRemoteName := fmt.Sprintf("%s-%s", ns, name)
		remoteUrl = fmt.Sprintf("ssh://%s@%s/~/git/%s", qualifiedRemoteName, extserviceIP, qualifiedRemoteName)
		remoteUrlDesc = ""
	}

	//if remote url changed or empty, then update remote
	if (remote.Status.RemoteUrl != remoteUrl) || (remoteUrl == "") {
		remoteCopy := remote.DeepCopy()
		remoteCopy.Status.RemoteUrl = remoteUrl
		remoteCopy.Status.RemoteUrlDesc = remoteUrlDesc

		_, err = c.remotesGetter.Remotes(ns).Update(remoteCopy)
		if err != nil {
			return err
		}
	}

	return nil
}
