# Contributing to gitkube

Thank you for your interest in developing gitkube. Refer to this guide to get started.

## Pre-requisites

- You must have Go 1.8+
- You need docker to build image
- You need a k8s cluster to test your build

## Fork the repo

Fork the repo and clone it locally at your GOPATH

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

- Make changes to the code.
- Edit `Makefile` and set `IMAGE_REGISTRY`

```sh
$ make build-all
$ make push-all
```

This will edit the registry in `manifest/gitkube-setup.yaml` with the built docker image.

## Test

```sh
$ cd e2e
$ ./test.bash

```

## Send PR

If all looks well, send a PR and if accepted, maintainers will release it.


