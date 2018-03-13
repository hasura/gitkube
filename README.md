# Gitkube

Gitkube is a tool for building and deploying docker images on Kubernetes using `git push`. 

After a simple initial setup, users can simply keep `git push`-ing their repos to build and deploy to Kubernetes automatically.

![Gitkube](https://raw.githubusercontent.com/hasura/gitkube/master/artifacts/gitkube.gif)


## When should I use gitkube?
1. Ideal for development where you can push your WIP branch to the cluster to test.
2. Reference implementation for writing git-based automation on your server. Fork this repo and create your own CRD + controller + git remote hook that can do things on the Kubernetes cluster.

## Getting started

Install gitkube

```sh
$ kubectl create -f https://raw.githubusercontent.com/hasura/gitkube/master/manifests/gitkube-setup.yaml

$ #expose gitkubed service
$ kubectl --namespace kube-system expose deployment gitkubed --type=LoadBalancer --name=gitkubed
```

Follow this [example](https://github.com/hasura/gitkube-example) repo for a typical workflow of gitkube.


## How it works

Gitkube has three components:

1. Remote: Custom resource defined by a K8s CRD
2. gitkube-controller: Controller that manages Remote objects and propogates changes to gitkubed 
3. gitkubed: Git host that builds docker image from repo and rolls out deployment

### High level architecture

![Architecture](https://raw.githubusercontent.com/hasura/gitkube/master/artifacts/gitkube-v0.1.png)

### Workflow
- Local dev: User creates a base git repo for the application with Dockerfile and K8s deployment
- Setting Remote: User defined the spec for Remote containaing the rules for `git push` 
- Deploying application: Once a Remote is setup, application can be deployed to K8s using `git push <remote> master`

#### Local dev
User should have a git repo with source code and a Dockerfile. User should also create a base K8s deployment for the application.

#### Setting Remote
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

#### Deploying application

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

## Roadmap

Gitkube is open to evolution. Some of the features to be added in future include:  

- Allowing all apps (daemonset, statefulset) to be deployed using `git push`. Current support is limited to deployments. [#19](https://github.com/hasura/gitkube/issues/19)
- Allowing different git hooks to be integrated. [#20](https://github.com/hasura/gitkube/issues/20)

## Contributing

Gitkube is an open source project licensed under Apache License 2.0

Contributions are welcome. 

## Maintainers

This project has come out of the work at [hasura.io](https://hasura.io). 
Current maintainers @tirumaraiselvan @shahidhk. 
