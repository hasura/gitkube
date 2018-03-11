# gitkube examplmple

The following example assumes `gitkube` has been setup on the cluster.

### 1) Create Remote
First, create a Remote object. Specify ssh-keys, registry and the deployment values in the Remote.
```sh
$ kubectl create -f remote.yaml
```
### 2) Setup git remote
Get the git URL of your Remote created in Step 1.
```sh
$ kubectl get remote sampleremote -o yaml
```
Copy `remoteUrl` from `status` spec of the output above. Next, add this as a git remote.

```sh
$ git remote add myremote <remoteUrl>
```

### 3) Prepare application
Create all k8s resources required by your application. It is mandatory to have a `Deployment` object for your application. You can provide any proxy image in the `Deployment` spec for the first time.
```sh
$ kubectl create -f www/k8s.yaml
```

### 4) Deploy application
Commit and push the code.
```sh
$ git add . && git commit -m "init"
$ git push myremote master
```

