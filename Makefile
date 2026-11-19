.PHONY: configure-image ensure-image-tag local build docker-build docker-push build-github vet lint test check vuln js-check

# Keep in sync with .github/workflows/pr-checks.yml (golangci/golangci-lint-action version).
GOLANGCI_LINT_VERSION ?= v2.15.0
# Pinned to match the sibling repos (dined, noted); never @latest.
GOVULNCHECK_VERSION ?= v1.7.0

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
	docker buildx build $(DOCKER_BUILD_ARGS) --platform=linux/arm64/v8 $(if $(METADATA_FILE),--metadata-file "$(METADATA_FILE)") --push .

vet:
	go vet ./...

# Linters and the gofmt/goimports formatters are configured in .golangci.yml;
# `golangci-lint fmt` applies the formatting fixes.
lint:
	golangci-lint run

test:
	go test -race ./...

# Fails on reachable vulnerabilities. Needs network access for the vulnerability database.
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

# Syntax-checks every script under static/ (needs Node on PATH). `node --check`
# catches syntax errors, and scripts/check-classic-script.js compiles each file as a
# classic browser script (no top-level `return`, `import`, or `export`). `node --check`
# accepts a top-level `return` as CommonJS, and `import`/`export` when it parses the
# file as an ES module. Neither executes the script.
# The static JavaScript ships as-is, so this is its only build-time gate.
# A separate PR check step, like `make vuln`; not part of `make check`.
js-check:
	@set -e; for f in static/*.js; do \
		echo "node --check $$f"; \
		node --check "$$f"; \
	done
	node scripts/check-classic-script.js static/*.js

# Mirrors the PR checks job (vet, golangci-lint with gofmt/goimports, race tests, build).
# `make vuln` is a separate PR check step and is not part of this target.
check: vet lint test
	go build ./cmd/bob
