# generate version string for the build
VERSION := $(shell hack/get-version.sh)

# set constants
IMAGE_REGISTRY ?= "hasura"
CONTROLLER_IMAGE ?= "$(IMAGE_REGISTRY)/gitkube-controller"
GITKUBED_IMAGE ?= "$(IMAGE_REGISTRY)/gitkubed"
GITKUBED_DIR ?= "build/gitkubed"
SETUP_MANIFEST_FILE ?= "manifests/gitkube-setup.yaml"

build-controller:
	docker build -t $(CONTROLLER_IMAGE):$(VERSION) .
	docker push $(CONTROLLER_IMAGE):$(VERSION)
	hack/update-image-version.sh $(SETUP_MANIFEST_FILE) $(CONTROLLER_IMAGE) $(VERSION)

build-gitkubed:
	docker build -t $(GITKUBED_IMAGE):$(VERSION) $(GITKUBED_DIR)
	docker push $(GITKUBED_IMAGE):$(VERSION)
	hack/update-image-version.sh $(SETUP_MANIFEST_FILE) $(GITKUBED_IMAGE) $(VERSION)

build-all: build-controller build-gitkubed
