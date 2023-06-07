PWD=$(shell pwd)
ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort )
TOOLS_MOD_DIR := ./internal/tools
ADDLICENSE=addlicense
ALL_SRC := $(shell find . -name '*.go' -o -name '*.sh' -o -name 'Dockerfile' -type f | sort)
JSON_EXCLUDED_FILES = model/parameter_test.go \
					  model/parameter.go 

JSON_LINT_FILES=$(shell find . -name '*.go' $(foreach file, $(JSON_EXCLUDED_FILES), -not -path "./$(file)"))

# Use 8 characters in order to match ArgoCD's behavior
# when using the template variable `head_short_sha`.
# https://github.com/argoproj/argo-cd/issues/11976#issue-1532285712
GIT_SHA=$(shell git rev-parse --short=8 HEAD)

NAMESPACE=bindplane-dev
OUTDIR=./build

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ifeq ($(GOARCH), amd64)
GOARCH_FULL=amd64_v1
else
GOARCH_FULL=$(GOARCH)
endif

.PHONY: help
help:
	@echo "TARGET\tDESCRIPTION" | expand -t 24
	@grep '^.PHONY: .* #' Makefile | sed 's/\.PHONY: \(.*\) # \(.*\)/\1\t\2/' | sort | expand -t 24

.PHONY: gomoddownload
gomoddownload:
	go mod download



.PHONY: install-tools # installs build tools
install-tools: 
	cd $(TOOLS_MOD_DIR) && go install github.com/securego/gosec/v2/cmd/gosec
	cd $(TOOLS_MOD_DIR) && go install github.com/google/addlicense
	cd $(TOOLS_MOD_DIR) && go install github.com/swaggo/swag/cmd/swag
	cd $(TOOLS_MOD_DIR) && go install github.com/99designs/gqlgen
	cd $(TOOLS_MOD_DIR) && go install github.com/mgechev/revive
	cd $(TOOLS_MOD_DIR) && go install github.com/uw-labs/lichen
	cd $(TOOLS_MOD_DIR) && go install honnef.co/go/tools/cmd/staticcheck
	cd $(TOOLS_MOD_DIR) && go install github.com/client9/misspell/cmd/misspell
	cd $(TOOLS_MOD_DIR) && go install github.com/ory/go-acc
	cd $(TOOLS_MOD_DIR) && go install github.com/vektra/mockery/v2
	cd $(TOOLS_MOD_DIR) && go install github.com/goreleaser/goreleaser

.PHONY: install-ui # [ui] npm install
install-ui:
	cd ui && npm install

.PHONY: install # install-tools && install-ui
install: install-tools install-ui

.PHONY: ci # [ui] npm ci
ci:
	cd ui && npm ci

ui/node_modules: ui/package.json ui/package-lock.json
	$(MAKE) ci

.PHONY: dev # runs go serve, ui proxy server, and ui graphql generator [primary development target]
dev: ui/node_modules prep
	./ui/node_modules/.bin/concurrently -c blue,magenta,cyan -n sv,ui,gq "go run ./cmd/bindplane/main.go serve --force-console-color --env development" "cd ui && npm start" "cd ui && npm run generate:watch"

.PHONY: test
test: prep
	go test ./... -race -timeout 60s

# Only runs integration tests and requires `make release-test`
# for container image output.
# This target runs in CI after the build stage due to the dependence
# on the container image. Regular tests run separately, and do not
# need to run as part of this target.
.PHONY: test-integration
test-integration:
	BINDPLANE_TEST_IMAGE="ghcr.io/observiq/bindplane-amd64:$(GIT_SHA)" go test ./client -tags integration

# Same as `test` but with codecov. Does not run integration tests.
.PHONY: test-with-cover
test-with-cover: prep
	go-acc --output=coverage.out --ignore=generated --ignore=mocks ./...

show-coverage: test-with-cover
	# Show coverage as HTML in the default browser.
	go tool cover -html=coverage.out

.PHONY: bench
bench:
	go test -benchmem -run=^$$ -bench ^* ./...

.PHONY: tidy # runs go mod tidy
tidy:
	$(MAKE) for-all CMD="go mod tidy"

.PHONY: lint # runs revive linter and npm run lint
lint:
	revive -config revive/config.toml -formatter=stylish -exclude "graphql/schema.resolvers.go"  -set_exit_status ./...
	cd ui && npm run lint && cd ..
	@for file in $(JSON_LINT_FILES); do \
        if grep -q "\"encoding/json\"" $$file; then \
            echo "Error: Standard JSON library used in $$file"; exit 1; \
        fi \
    done


.PHONY: vet # runs go vet
vet:
	GOOS=darwin go vet ./...
	GOOS=linux go vet ./...
	GOOS=windows go vet ./...

.PHONY: secure # runs gosec to identify security issues
secure: prep
	gosec -exclude-generated -exclude-dir internal/tools ./...

.PHONY: generate # runs go generate to generate graphql resolver and runs add-license
generate:
	go generate ./...
	@$(MAKE) add-license

.PHONY: swagger # generates the REST API documentation using swagger
swagger:
	swag init --parseDependency --parseInternal -g model/rest.go -o docs/swagger/
	@$(MAKE) add-license

.PHONY: init-server # runs bindplane init server to setup the server
init-server: prep
	go run cmd/bindplane/main.go init server

.PHONY: for-all
for-all:
	@set -e; for dir in $(ALL_MODULES); do \
	  (cd "$${dir}" && $${CMD} ); \
	done

# TODO(jsirianni): Add secure: https://github.com/observIQ/bindplane/issues/478
.PHONY: ci-check
ci-check: vet test lint check-license scan-licenses

.PHONY: check-license # checks for missing license header in source files
check-license:
	@ADDLICENSEOUT=`$(ADDLICENSE) -check $(ALL_SRC) 2>&1`; \
		if [ "$$ADDLICENSEOUT" ]; then \
			echo "$(ADDLICENSE) FAILED => add License errors:\n"; \
			echo "$$ADDLICENSEOUT\n"; \
			echo "Use 'make add-license' to fix this."; \
			exit 1; \
		else \
			echo "Check License finished successfully"; \
		fi

.PHONY: add-license # adds license header to source files
add-license:
	@ADDLICENSEOUT=`$(ADDLICENSE) -y "" -c "observIQ, Inc." $(ALL_SRC) 2>&1`; \
		if [ "$$ADDLICENSEOUT" ]; then \
			echo "$(ADDLICENSE) FAILED => add License errors:\n"; \
			echo "$$ADDLICENSEOUT\n"; \
			exit 1; \
		else \
			echo "Add License finished successfully"; \
		fi

.PHONY: scan-licenses # checks dependencies for permitted licenses
scan-licenses:
	lichen --config=./license.yaml $$(find build/bindplane* | xargs)

# TLS will run the tls generation script only when the
# tls directory is missing
tls:
	mkdir tls
	docker run \
		-v ${PWD}/scripts/generate-dev-certificates.sh:/generate-dev-certificates.sh \
		-v ${PWD}/tls:/tls \
		--entrypoint=/bin/sh \
		alpine/openssl /generate-dev-certificates.sh

.PHONY: docker-http
docker-http:
	docker run -d -p 3010:3001 \
		--name "bindplane-server-${GIT_SHA}-http" \
		-e BINDPLANE_CONFIG_SESSIONS_SECRET=403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		-e BINDPLANE_CONFIG_LOG_OUTPUT=stdout \
		-e BINDPLANE_CONFIG_SECRET_KEY=403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		"ghcr.io/observiq/bindplane-$(GOARCH):${GIT_SHA}" \
		--host 0.0.0.0 \
		--port "3001" \
		--remote-url http://localhost:3010 \
		--secret-key 403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		--store-type bbolt \
		--store-bbolt-path /data/storage
	docker logs "bindplane-server-${GIT_SHA}-http"

.PHONY: docker-http-profile
docker-http-profile:
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile set docker-http \
		--remote-url http://localhost:3010
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile use docker-http

.PHONY: docker-ubi8-http
docker-ubi8-http:
	docker run -d -p 3011:3001 \
		--name "bindplane-server-${GIT_SHA}-ubi8-http" \
		-e BINDPLANE_CONFIG_SESSIONS_SECRET=403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		-e BINDPLANE_CONFIG_LOG_OUTPUT=stdout \
		-e BINDPLANE_CONFIG_SECRET_KEY=403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		"ghcr.io/observiq/bindplane-$(GOARCH):${GIT_SHA}-ubi8" \
		--host 0.0.0.0 \
		--port "3001" \
		--remote-url http://localhost:3011 \
		--secret-key 403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		--store-type bbolt \
		--store-bbolt-path /data/storage
	docker logs "bindplane-server-${GIT_SHA}-ubi8-http"

.PHONY: docker-ubi8-http-profile
docker-ubi8-http-profile:
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile set docker-http \
		--server-url http://localhost:3010 --remote-url ws://localhost:3010
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile use docker-http

.PHONY: docker-https
docker-https: tls
	docker run -d \
		-p 3013:3001 \
		--name "bindplane-server-${GIT_SHA}-https" \
		-e BINDPLANE_SESSION_SECRET=403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		-e BINDPLANE_LOGGING_OUTPUT=stdout \
		-v "${PWD}/tls:/tls" \
		"ghcr.io/observiq/bindplane-$(GOARCH):latest" \
			--tls-cert /tls/bindplane.crt --tls-key /tls/bindplane.key \
			--host 0.0.0.0 \
			--port "3001" \
			--remote-url https://localhost:3013 \
			--secret-key 403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
			--store-type bbolt \
			--store-bbolt-path /data/storage
	docker logs "bindplane-server-${GIT_SHA}-https"

.PHONY: docker-https-profile
docker-https-profile: tls
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile set docker-https \
		--remote-url https://localhost:3013 \
		--tls-ca tls/bindplane-ca.crt
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile use docker-https

.PHONY: docker-https-mtls
docker-https-mtls: tls
	docker run -d \
		-p 3012:3001 \
		--name "bindplane-server-${GIT_SHA}-https-mtls" \
		-e BINDPLANE_SESSION_SECRET=403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
		-e BINDPLANE_LOGGING_OUTPUT=stdout \
		-v "${PWD}/tls:/tls" \
		"ghcr.io/observiq/bindplane-$(GOARCH):latest" \
			--tls-cert /tls/bindplane.crt --tls-key /tls/bindplane.key --tls-ca /tls/bindplane-ca.crt --tls-ca /tls/test-ca.crt \
			--host 0.0.0.0 \
			--port "3001" \
			--remote-url https://localhost:3012 \
			--secret-key 403dd8ff-72a9-4401-9a66-e54b37d6e0ce \
			--store-type bbolt \
			--store-bbolt-path /data/storage
	docker logs  "bindplane-server-${GIT_SHA}-https-mtls"

.PHONY: docker-https-mtls-profile
docker-https-mtls-profile: tls
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile set docker-https-mtls \
		--remote-url https://localhost:3012 \
		--tls-cert tls/bindplane-client.crt --tls-key ./tls/bindplane-client.key --tls-ca tls/bindplane-ca.crt
	dist/bindplane_$(GOOS)_$(GOARCH_FULL)/bindplane profile use docker-https-mtls

.PHONY: docker-all
docker-all: docker-clean docker-http docker-https docker-https-mtls

.PHONY: docker-clean
docker-clean:
	docker ps -a | grep bindplane-server | awk '{print $$1}' | xargs -I{} docker rm --force {}

# Call 'release-test' first.
.PHONY: inspec-continer-image
inspec-continer-image: prep docker-http docker-ubi8-http
	docker exec -u root bindplane-server-${GIT_SHA}-http apt-get update -qq
	docker exec -u root bindplane-server-${GIT_SHA}-http apt-get install -qq -y procps net-tools
	cinc-auditor exec test/inspec/docker/integration.rb -t "docker://bindplane-server-${GIT_SHA}-http"

	docker exec -u root bindplane-server-${GIT_SHA}-ubi8-http dnf install -y procps net-tools
	cinc-auditor exec test/inspec/docker/integration.rb -t "docker://bindplane-server-${GIT_SHA}-ubi8-http"

.PHONY: run
run: docker-http

# Called by commands such as 'vet', useful when the ui has not
# been built before (in ci)
prep: ui/build
ui/build:
	mkdir ui/build
	touch ui/build/index.html

.PHONY: ui-test # [ui] runs ui tests in watch mode
ui-test:
	cd ui && CI=true npm run test --watchAll

.PHONY: ui-test-with-cover # [ui] runs ui tests with coverage
ui-test-with-cover:
	cd ui && CI=true npm test -- --coverage

# ui-build builds the static site to be embeded into the Go binary.
# make install should be called before, if you are not up to date.
.PHONY: ui-build # [ui] builds the static site to be embeded into the Go binary
ui-build:
	cd ui && npm run build

# goreleaser will call ui-build to ensure the static site
# is up to date. goreleaser will not call `make install`.
.PHONY: build # builds bindplane and bindplanectl using goreleaser
build:
	goreleaser build --clean --skip-validate --single-target --snapshot

.PHONY: clean # removes the dist folder
clean:
	rm -rf $(OUTDIR)

# Build all binaries, packages, and container images. Add current git hash
# tags for use with "make inspec-continer-image".
.PHONY: release-test
release-test:
	goreleaser release --clean --skip-publish --skip-validate --snapshot
	@docker tag ghcr.io/observiq/bindplane-arm64:latest ghcr.io/observiq/bindplane-arm64:${GIT_SHA}
	@docker tag ghcr.io/observiq/bindplane-amd64:latest ghcr.io/observiq/bindplane-amd64:${GIT_SHA}
	@docker tag ghcr.io/observiq/bindplane-arm64:latest-ubi8 ghcr.io/observiq/bindplane-arm64:${GIT_SHA}-ubi8
	@docker tag ghcr.io/observiq/bindplane-amd64:latest-ubi8 ghcr.io/observiq/bindplane-amd64:${GIT_SHA}-ubi8

# Kitchen prep will build a release and ensure the required
# gems are installed for using Kitchen with GCE
.PHONY: kitchen-prep
kitchen-prep: release-test
	sudo cinc gem install --no-user-install kitchen-google
	sudo cinc gem install --no-user-install kitchen-sync
	mkdir -p dist/kitchen
	cp dist/bindplane_*amd64.deb dist/kitchen
	cp dist/bindplane_*amd64.rpm dist/kitchen
	cp scripts/install-linux.sh dist/kitchen

# Assumes you have a ssh key pair at ~/.ssh/id_rsa && ~/.ssh/id_rsa.pub
# Assumes you are authenticated to GCP with Gcloud SDK
#
# Run all tests:
#   make kitchen
# Run tests against specific OS:
#   make kitchen ARGS=sles
.PHONY: kitchen
kitchen:
	kitchen test -c 10 $(ARGS)

.PHONY: kitchen-clean
kitchen-clean:
	kitchen destroy -c 10

ALLDOC=$(shell find . \( -name "*.md" -o -name "*.yaml" \) | grep -v ui/node_modules)

.PHONY: misspell # checks for spelling errors in .md and .yaml files
misspell:
	misspell -error $(ALLDOC)

.PHONY: misspell-fix # fixes spelling errors in .md and .yaml files
misspell-fix:
	misspell -w $(ALLDOC)
