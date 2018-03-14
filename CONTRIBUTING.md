# Contributing to gitkube

Thank you for your interest in developing gitkube. Refer to this guide to get started.

## Pre-requisites

- Go 1.8+ (not required if building with docker)
- Docker to build and push images
- Kubernetes cluster to test your build

## Fork the repo

Fork the repo and clone it locally to your GOPATH

```
$ mkdir -p $GOPATH/src/github.com/hasura
$ cd $GOPATH/src/github.com/hasura
$ git clone git@github.com:<username>/gitkube.git
$ cd gitkube
```

For pulling changes from upstream, add this remote:

```sh
$ git remote add upstream https://github.com/hasura/gitkube.git
```

## Make changes, build and push

- Make changes to the code
- Set `IMAGE_REGISTRY` env var to your docker hub or any other image registry (e.g. `quay.io/<username>` or just `<docker-hub-username>`)

```sh
$ export IMAGE_REGISTRY=<your-image-registry>
$ make build-all
$ make push-all
```

This will edit the image name in `manifest/gitkube-setup.yaml` with the built docker image.

## Test

Run an end-to-end test if everything is working well. If your change adds a new feature, try to add it in the tests too.

```sh
$ cd e2e
$ ./test.bash

# teardown after the tests
$ ./teardown.bash
```

## Send PR

If all looks well, send a PR!

