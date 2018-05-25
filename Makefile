# set constants
IMAGE_REGISTRY ?= "hasura"
CONTROLLER_IMAGE_NAME ?= "gitkube-controller"
CONTROLLER_IMAGE ?= "$(IMAGE_REGISTRY)/$(CONTROLLER_IMAGE_NAME)"
GITKUBED_IMAGE_NAME ?= "gitkubed"
GITKUBED_IMAGE ?= "$(IMAGE_REGISTRY)/$(GITKUBED_IMAGE_NAME)"
GITKUBED_DIR ?= "build/gitkubed"
SETUP_MANIFEST_FILE ?= "manifests/gitkube-setup.yaml"

PWD := $(shell pwd)
VERSION := $(shell hack/get-version.sh)

build-controller:
	docker build -t $(CONTROLLER_IMAGE):$(VERSION) .
	$(shell sed -i -E "s@image: .+\/$(CONTROLLER_IMAGE_NAME):.+@image: $(CONTROLLER_IMAGE):$(VERSION)@" $(SETUP_MANIFEST_FILE))

push-controller:
	docker push $(CONTROLLER_IMAGE):$(VERSION)

build-gitkubed:
	docker build -t $(GITKUBED_IMAGE):$(VERSION) $(GITKUBED_DIR)
	$(shell sed -i -E "s@image: .+\/$(GITKUBED_IMAGE_NAME):.+@image: $(GITKUBED_IMAGE):$(VERSION)@" $(SETUP_MANIFEST_FILE))

push-gitkubed:
	docker push $(GITKUBED_IMAGE):$(VERSION)

build-all: build-controller build-gitkubed
push-all: push-controller push-gitkubed

controller: build-controller push-controller
gitkubed: build-gitkubed push-gitkubed

# build cli locally, for all given platform/arch
build-cli:
	go get github.com/mitchellh/gox
	gox -ldflags "-X github.com/hasura/gitkube/pkg/cmd.version=$(VERSION)" \
	-os="linux darwin windows" \
	-arch="amd64" \
	-output="_output/$(VERSION)/gitkube_{{.OS}}_{{.Arch}}" \
	./cmd/gitkube-cli/

# build cli inside a docker container
build-cli-in-docker:
	docker build -t gitkube-cli-builder -f build/cli-builder.dockerfile build
	docker run --rm -it \
	-v $(PWD):/go/src/github.com/hasura/gitkube \
	gitkube-cli-builder \
	make build-cli

all: build-all push-all build-cli-in-docker
