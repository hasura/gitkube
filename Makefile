# set constants
IMAGE_REGISTRY ?= "hasura"
CONTROLLER_IMAGE_NAME ?= "gitkube-controller"
CONTROLLER_IMAGE ?= "$(IMAGE_REGISTRY)/$(CONTROLLER_IMAGE_NAME)"
GITKUBED_IMAGE_NAME ?= "gitkubed"
GITKUBED_IMAGE ?= "$(IMAGE_REGISTRY)/$(GITKUBED_IMAGE_NAME)"
GITKUBED_DIR ?= "build/gitkubed"
SETUP_MANIFEST_FILE ?= "manifests/gitkube-setup.yaml"

CONTROLLER_VERSION := $(shell hack/get-version.sh)
GITKUBED_VERSION := $(shell hack/get-component-version.sh $(GITKUBED_DIR))

build-controller:
	docker build -t $(CONTROLLER_IMAGE):$(CONTROLLER_VERSION) .
	$(shell sed -i -E "s@image: .+\/$(CONTROLLER_IMAGE_NAME):.+@image: $(CONTROLLER_IMAGE):$(CONTROLLER_VERSION)@" $(SETUP_MANIFEST_FILE))

push-controller:
	docker push $(CONTROLLER_IMAGE):$(CONTROLLER_VERSION)

build-gitkubed:
	docker build -t $(GITKUBED_IMAGE):$(GITKUBED_VERSION) $(GITKUBED_DIR)
	$(shell sed -i -E "s@image: .+\/$(GITKUBED_IMAGE_NAME):.+@image: $(GITKUBED_IMAGE):$(GITKUBED_VERSION)@" $(SETUP_MANIFEST_FILE))

push-gitkubed:
	docker push $(GITKUBED_IMAGE):$(GITKUBED_VERSION)

build-all: build-controller build-gitkubed
push-all: push-controller push-gitkubed

controller: build-controller push-controller
gitkubed: build-gitkubed push-gitkubed

all: build-all push-all
