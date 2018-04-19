# Minikube

## Setup

#### Install Minikube

Follow the instructions here to install Minikube: https://kubernetes.io/docs/tasks/tools/install-minikube/

Post-installation, you will have kubeconfig set to your minikube cluster

#### Install Gitkube

Create gitkube resources:

```sh
$ kubectl create -f https://storage.googleapis.com/gitkube/gitkube-setup-stable.yaml
```

Expose gitkubed service as Nodeport:

```sh
$ kubectl --namespace kube-system expose deployment gitkubed --type=NodePort --name=gitkubed
```

## Example workflow

Follow the example here for a typical workflow: https://github.com/hasura/gitkube-example

After finishing the Remote resource creation step as per the example repo, please read the following notes:

Since the gitkubed service is exposed as NodePort, the Remote resource will not have a `remoteUrl` set automatically.
The key `remoteUrlDesc` will have instructions to construct the `remoteUrl` manually.

```sh
$ kubectl get remotes example -o yaml

...
status:
  remoteUrl: ""
  remoteUrlDesc: 'manually configure remote for gitkube service of type NodePort.
    E.g.: ssh://<namespace>-<remote-name>@<any-node-ip>:<node-port>/~/git/<namespace>-<remote-name>'
```

For e.g.

Get the `node-ip` and `node-port` for the minikube cluster

```sh
$ minikube service -n kube-system gitkubed --url
http://192.168.99.100:30731
```

If the Remote is called `example` in the `default` namespace, the `remoteUrl` should be `ssh://default-example@192.168.99.100:30731/~/git/default-example`

Add the `remoteUrl` as a git remote:
```sh
git remote add example ssh://default-example@192.168.99.100:30731/~/git/default-example
```

Continue with the rest of the workflow as per the example repo.

