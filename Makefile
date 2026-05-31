.PHONY: configure-image build-github docker-build-push-github docker-build docker-publish docker-push ensure-image-tag local run tail-watch tail-prod

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
	@true

ensure-image-tag: configure-image
	@test -n "$(strip $(SHORT_SHA))" || (echo "Unable to determine git short SHA. Ensure this is a git repository with at least one commit." >&2; exit 1)

run:
	air

local:
	make -j 2 tail-watch run

build: tail-prod docker-build docker-push

docker-build: ensure-image-tag
	docker build -f Docker/Dockerfile \
		--build-arg LOG_LEVEL=$(LOG_LEVEL) \
		--build-arg VERSION=$(VERSION) \
		--build-arg REVISION=$(REVISION) \
		--build-arg SOURCE_URL=$(SOURCE_URL) \
		$(OCI_LABEL_ARGS) \
		-t $(IMAGE) \
		.

docker-push: ensure-image-tag
	docker push $(IMAGE)

docker-publish:
	make docker-build docker-push

tail-watch:
	tailwindcss -i ./tailwind/styles.css -o ./static/styles.css --watch

tail-prod:
	tailwindcss -i ./tailwind/styles.css -o ./static/styles.css --minify

build-github: tail-prod configure-image docker-build-push-github

docker-build-push-github:
	echo ">> Building and pushing $(IMAGE)"
	-docker buildx inspect >/dev/null 2>&1 || docker buildx create --use
	docker buildx build -f Docker/Dockerfile . \
		--platform=linux/arm64/v8 \
		--build-arg LOG_LEVEL=$(LOG_LEVEL) \
		--build-arg VERSION=$(VERSION) \
		--build-arg REVISION=$(REVISION) \
		--build-arg SOURCE_URL=$(SOURCE_URL) \
		$(OCI_LABEL_ARGS) \
		-t $(IMAGE) \
		--push
