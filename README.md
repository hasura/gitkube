# Gitkube

Gitkube is a tool for building and deploying docker images on Kubernetes using `git push`. 

## Workflow

Typical workflow consists of two parts:
- Setting Remote: Remote is a Kubernetes custom resource which contains rules for `git push` 
- Deploying application: Once a Remote is setup, applications can be deployed using `git push <remote> master`

### Setting Remote
A Remote resource consists of 3 parts:
1. authorizedKeys: List of ssh-keys for authorizing `git push`.
2. registry: Details of docker registry where images are pushed post-build.
3. deployments: Spec for building docker image and updating corresponding K8s deployment.

Here is a typical spec for a Remote:
```yaml
apiVersion: gitkube.sh/v1alpha1
kind: Remote
metadata:
  name: sampleremote
  namespace: default
spec:

# Insert ssh-keys for allowing users to git push
  authorizedKeys:
  - "ssh-rsa your-ssh-public-key"

# Provide registry details for pushing and pulling image from/into the cluster 
  registry:
    url: "registry.io/user"
    credentials:
    # dockercfg secret
      secretKeyRef:
        name: regsecret
        key: .dockercfg

# Define deployment rules
  deployments:
  - name: www                             # Name of K8s deployment which is updated on git push
    containers: 
    - name: www                           # Name of container in the deployment which is built during git push
      path: example/www                   # Location of source code in the git repo
      dockerfile: example/www/Dockerfile  # Location of Dockerfile for the source code
```

### Deploying application

Once a Remote is created, it gets a git remote url which you can find in its `status` spec

```sh
$ kubectl get remote sampleremote -o yaml
...
status:
  remoteUrl: ssh://default-sampleremote@35.225.226.96/~/git/default-sampleremote
  remoteUrlDesc: ""
```

Add the generated `remoteUrl` in git

```sh
$ git add remote sampleremote ssh://default-sampleremote@35.225.226.96/~/git/default-sampleremote
```

And finally, `git push`

```sh
$ git push sampleremote master
```

## Install

```sh
$ wget https://raw.githubusercontent.com/hasura/gitkube/readme/manifests/gitkube-setup.yaml
$ kubectl create -f gitkube-setup.yaml
```

## Example

A full example is available at https://github.com/hasura/gitkube-example/

## Roadmap

Gitkube is open to evolution. Some of the features to be added in future include:  

- Allowing all apps (daemonset, statefulset) to be deployed using `git push`. Current support is limited to deployments.
- Allowing different git hooks to be integrated

## Contributing

Gitkube is an open source project licensed under Apache License 2.0

Contributions are welcome. Please follow the [contributing guide](https://github.com/hasura/gitkube/blob/master/CONTRIBUTING) to get started.

