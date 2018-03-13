
# set constants
IMAGE_REGISTRY ?= "hasura"
CONTROLLER_IMAGE ?= "$(IMAGE_REGISTRY)/gitkube-controller"
GITKUBED_IMAGE ?= "$(IMAGE_REGISTRY)/gitkubed"
GITKUBED_DIR ?= "build/gitkubed"
SETUP_MANIFEST_FILE ?= "manifests/gitkube-setup.yaml"

VERSION := $(shell hack/get-version.sh)
GITKUBED_VERSION := $(shell hack/get-component-version.sh $(GITKUBED_DIR))

build-controller:
	docker build -t $(CONTROLLER_IMAGE):$(VERSION) .
	hack/update-image-version.sh $(SETUP_MANIFEST_FILE) $(CONTROLLER_IMAGE) $(VERSION)

push-controller:
	docker push $(CONTROLLER_IMAGE):$(VERSION)

build-gitkubed:
	docker build -t $(GITKUBED_IMAGE):$(GITKUBED_VERSION) $(GITKUBED_DIR)
	hack/update-image-version.sh $(SETUP_MANIFEST_FILE) $(GITKUBED_IMAGE) $(GITKUBED_VERSION)

push-gitkubed:
	docker push $(GITKUBED_IMAGE):$(GITKUBED_VERSION)

build-all: build-controller build-gitkubed
push-all: push-controller push-gitkubed

all: build-all push-all
