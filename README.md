# gitkube


## Setup

* Create 'remotes' CRD 
```sh
$ kubectl create -f manifests/crd.yaml
```
* Create gitkube resources
```sh
$ kubectl create -f manifests/gitkube-setup.yaml
```
By default gitube is created in kube-system namespace, check that its running

```sh
$ kubectl get pods -n kube-system | grep gitkube
```

## Usage

* Create a Remote
```sh
$ kubectl create -f manifests/remote.yaml
```

* Fetch the remote Url from the 'status' spec of the Remote
```sh
$ kubectl get remote hasura -o yaml
```

* Put the Url in the .git/config file
* Finally, do a git push hasura mater



