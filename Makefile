.PHONY: configure-image ensure-image-tag local build docker-build docker-push build-github

configure-image:
	$(eval REGISTRY ?= registry.tail209cfc.ts.net)
	$(eval IMAGE_NAME ?= $(REGISTRY)/bob)
	$(eval SHORT_SHA := $(shell git rev-parse --short HEAD))
	$(eval REVISION ?= $(shell git rev-parse HEAD 2>/dev/null))
	$(eval IMAGE_TAG ?= $(SHORT_SHA))
	$(eval VERSION ?= $(IMAGE_TAG))
	$(eval SOURCE_URL ?= https://github.com/bitofbytes-io/bitofbytes)
	$(eval IMAGE := $(IMAGE_NAME):$(IMAGE_TAG))
	$(eval LOG_LEVEL ?= warn)
	$(eval OCI_LABEL_ARGS := --label "org.opencontainers.image.source=$(SOURCE_URL)" --label "org.opencontainers.image.revision=$(REVISION)" --label "org.opencontainers.image.version=$(VERSION)" --label "org.opencontainers.image.title=bitofbytes" --label "org.opencontainers.image.description=BitOfBytes web application")
	$(eval DOCKER_BUILD_ARGS := -f Docker/Dockerfile --build-arg LOG_LEVEL=$(LOG_LEVEL) --build-arg VERSION=$(VERSION) --build-arg REVISION=$(REVISION) --build-arg SOURCE_URL=$(SOURCE_URL) $(OCI_LABEL_ARGS) -t $(IMAGE))
	@true

ensure-image-tag: configure-image
	@test -n "$(strip $(SHORT_SHA))" || (echo "Unable to determine git short SHA. Ensure this is a git repository with at least one commit." >&2; exit 1)

local:
	air

build: docker-build docker-push

docker-build: ensure-image-tag
	docker build $(DOCKER_BUILD_ARGS) .

docker-push: ensure-image-tag
	docker push $(IMAGE)

# CI: build the arm64 image and push it in one step. METADATA_FILE, when set,
# receives buildx's build metadata; CI reads the pushed digest from it.
build-github: configure-image
	echo ">> Building and pushing $(IMAGE)"
	-docker buildx inspect >/dev/null 2>&1 || docker buildx create --use
	docker buildx build $(DOCKER_BUILD_ARGS) --platform=linux/arm64/v8 $(if $(METADATA_FILE),--metadata-file $(METADATA_FILE)) --push .
