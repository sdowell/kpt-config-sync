#!/usr/bin/make
#
# Copyright 2018 CSP Config Management Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

######################################################################

##### CONFIG #####

REPO := kpt.dev/configsync

# List of dirs containing go code owned by Nomos
NOMOS_CODE_DIRS := pkg cmd e2e
NOMOS_GO_PKG := $(foreach dir,$(NOMOS_CODE_DIRS),./$(dir)/...)

# Directory containing all build artifacts.  We need abspath here since
# mounting a directory as a docker volume requires an abspath.  This should
# eventually be pushed down into where docker commands are run rather than
# requiring all our paths to be absolute.
OUTPUT_DIR := $(abspath .output)

# Self-contained GOPATH dir.
GO_DIR := $(OUTPUT_DIR)/go

# Directory containing installed go binaries.
BIN_DIR := $(GO_DIR)/bin
KUSTOMIZE_VERSION := v5.0.1
HELM_VERSION := v3.11.3

# Directory used for staging Docker contexts.
STAGING_DIR := $(OUTPUT_DIR)/staging

# Directory used for staging the manifest primitives.
NOMOS_MANIFEST_STAGING_DIR := $(STAGING_DIR)/operator
OSS_MANIFEST_STAGING_DIR := $(STAGING_DIR)/oss

# Directory that gets mounted into /tmp for build and test containers.
TEMP_OUTPUT_DIR := $(OUTPUT_DIR)/tmp

# Directory containing Deployment yamls with image paths spliced in
GEN_DEPLOYMENT_DIR := $(OUTPUT_DIR)/deployment

# Directory containing generated test yamls
TEST_GEN_YAML_DIR := $(OUTPUT_DIR)/test/yaml

# Use git tags to set version string.
ifeq ($(origin VERSION), undefined)
VERSION := $(shell git describe --tags --always --dirty --long)
endif

# UID of current user
UID := $(shell id -u)

# GID of current user
GID := $(shell id -g)

# GCP project that owns container registry for local dev build.
ifeq ($(origin GCP_PROJECT), undefined)
GCP_PROJECT := $(shell gcloud config get-value project)
endif

# Use BUILD_ID as a unique identifier for prow jobs, otherwise default to USER
BUILD_ID ?= $(USER)
GCR_PREFIX ?= $(GCP_PROJECT)/$(BUILD_ID)

# Allow arbitrary registry name, but default to GCR with prefix if not provided
REGISTRY ?= gcr.io/$(GCR_PREFIX)

# Registry used for retagging previously built images
OLD_REGISTRY ?= $(REGISTRY)

# Docker image used for build and test. This image does not support CGO.
BUILDENV_IMAGE ?= gcr.io/stolos-dev/buildenv:v0.2.12

# Nomos docker images containing all binaries.
RECONCILER_IMAGE := reconciler
RECONCILER_MANAGER_IMAGE := reconciler-manager
ADMISSION_WEBHOOK_IMAGE := admission-webhook
HYDRATION_CONTROLLER_IMAGE := hydration-controller
HYDRATION_CONTROLLER_WITH_SHELL_IMAGE := $(HYDRATION_CONTROLLER_IMAGE)-with-shell
OCI_SYNC_IMAGE := oci-sync
HELM_SYNC_IMAGE := helm-sync
NOMOS_IMAGE := nomos
NOTIFICATION_IMAGE := notification

# nomos binary for local run.
NOMOS_LOCAL := $(BIN_DIR)/linux_amd64/nomos

# Allows an interactive docker build or test session to be interrupted
# by Ctrl-C.  This must be turned off in case of non-interactive runs,
# like in CI/CD.
DOCKER_INTERACTIVE := --interactive
HAVE_TTY := $(shell [ -t 0 ] && echo 1 || echo 0)
ifeq ($(HAVE_TTY), 1)
	DOCKER_INTERACTIVE += --tty
endif

# Suppresses docker output on build.
#
# To turn off, for example:
#   make DOCKER_BUILD_QUIET="" deploy
DOCKER_BUILD_QUIET ?= --quiet

# Suppresses gcloud output.
GCLOUD_QUIET := --quiet

# Developer focused docker image tag.
ifeq ($(origin IMAGE_TAG), undefined)
IMAGE_TAG := $(shell git describe --tags --always --dirty --long)
endif
# Tag used for retagging previously built images
OLD_IMAGE_TAG ?= $(IMAGE_TAG)

# Base image names as given on gcr.io
RECONCILER_GCR := $(REGISTRY)/$(RECONCILER_IMAGE)
RECONCILER_MANAGER_GCR := $(REGISTRY)/$(RECONCILER_MANAGER_IMAGE)
ADMISSION_WEBHOOK_GCR := $(REGISTRY)/$(ADMISSION_WEBHOOK_IMAGE)
HYDRATION_CONTROLLER_GCR := $(REGISTRY)/$(HYDRATION_CONTROLLER_IMAGE)
HYDRATION_CONTROLLER_WITH_SHELL_GCR := $(REGISTRY)/$(HYDRATION_CONTROLLER_WITH_SHELL_IMAGE)
OCI_SYNC_GCR := $(REGISTRY)/$(OCI_SYNC_IMAGE)
HELM_SYNC_GCR := $(REGISTRY)/$(HELM_SYNC_IMAGE)
NOMOS_GCR := $(REGISTRY)/$(NOMOS_IMAGE)
NOTIFICATION_GCR := $(REGISTRY)/$(NOTIFICATION_IMAGE)
# Full image tags as given on gcr.io
RECONCILER_TAG := $(RECONCILER_GCR):$(IMAGE_TAG)
RECONCILER_MANAGER_TAG := $(RECONCILER_MANAGER_GCR):$(IMAGE_TAG)
ADMISSION_WEBHOOK_TAG := $(ADMISSION_WEBHOOK_GCR):$(IMAGE_TAG)
HYDRATION_CONTROLLER_TAG := $(HYDRATION_CONTROLLER_GCR):$(IMAGE_TAG)
HYDRATION_CONTROLLER_WITH_SHELL_TAG := $(HYDRATION_CONTROLLER_WITH_SHELL_GCR):$(IMAGE_TAG)
OCI_SYNC_TAG := $(OCI_SYNC_GCR):$(IMAGE_TAG)
HELM_SYNC_TAG := $(HELM_SYNC_GCR):$(IMAGE_TAG)
NOMOS_TAG := $(NOMOS_GCR):$(IMAGE_TAG)
NOTIFICATION_TAG := $(NOTIFICATION_GCR):$(IMAGE_TAG)

DOCKER_RUN_ARGS = \
	$(DOCKER_INTERACTIVE)                                              \
	-u $(UID):$(GID)                                                   \
	-v $(GO_DIR):/go                                                   \
	-v $$(pwd):/go/src/$(REPO)                                         \
	-v $(GO_DIR)/std/linux_amd64_static:/usr/local/go/pkg/linux_amd64_static    \
	-v $(GO_DIR)/std/linux_arm64_static:/usr/local/go/pkg/linux_arm64_static    \
	-v $(GO_DIR)/std/darwin_amd64_static:/usr/local/go/pkg/darwin_amd64_static   \
	-v $(GO_DIR)/std/darwin_arm64_static:/usr/local/go/pkg/darwin_arm64_static   \
	-v $(GO_DIR)/std/windows_amd64_static:/usr/local/go/pkg/windows_amd64_static   \
	-v $(TEMP_OUTPUT_DIR):/tmp                                         \
	-w /go/src/$(REPO)                                                 \
	--rm                                                               \
	$(BUILDENV_IMAGE)                                                  \

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

##### SETUP #####

.DEFAULT_GOAL := all

.PHONY: $(OUTPUT_DIR)
$(OUTPUT_DIR):
	@echo "+++ Creating the local build output directory: $(OUTPUT_DIR)"
	@mkdir -p \
		$(STAGING_DIR) \
		$(TEMP_OUTPUT_DIR) \
		$(TEST_GEN_YAML_DIR) \
		$(NOMOS_MANIFEST_STAGING_DIR) \
		$(OSS_MANIFEST_STAGING_DIR) \
		$(GEN_DEPLOYMENT_DIR)

# These directories get mounted by DOCKER_RUN_ARGS, so we have to create them
# before invoking docker.
.PHONY: buildenv-dirs
buildenv-dirs:
	@mkdir -p \
		$(BIN_DIR) \
		$(GO_DIR)/src/$(REPO) \
			$(GO_DIR)/std/linux_amd64_static \
			$(GO_DIR)/std/linux_arm64_static \
			$(GO_DIR)/std/darwin_amd64_static \
			$(GO_DIR)/std/darwin_arm64_static \
			$(GO_DIR)/std/windows_amd64_static \
		$(TEMP_OUTPUT_DIR)

##### TARGETS #####

include Makefile.build
-include Makefile.docs
include Makefile.e2e
-include Makefile.prow
include Makefile.gen
include Makefile.oss
include Makefile.oss.prow
include Makefile.reconcilermanager
-include Makefile.release

.PHONY: all
all: test deps configsync-crds

# Cleans all artifacts.
.PHONY: clean
clean:
	@echo "+++ Cleaning $(OUTPUT_DIR)"
	@rm -rf $(OUTPUT_DIR)

.PHONY: test-unit
test-unit: pull-buildenv buildenv-dirs "$(BIN_DIR)/kustomize"
	@echo "+++ Running unit tests in a docker container"
	@docker run $(DOCKER_RUN_ARGS) ./scripts/test-unit.sh $(NOMOS_GO_PKG)

# Runs unit tests and linter.
.PHONY: test
test: test-unit lint

# The presubmits have made side-effects in the past, which makes their
# validation suspect (as the repository they are validating is different
# than what is checked in).  Run `test` but verify that repository is clean.
.PHONY: test-presubmit
test-presubmit: all
	@./scripts/fail-if-dirty-repo.sh

# Runs all tests.
# This only runs on local dev environment not CI environment.
.PHONY: test-all-local
test-all-local: test test-e2e

# Runs gofmt and goimports.
# Even though goimports runs gofmt, it runs it without the -s (simplify) flag
# and offers no option to turn it on. So we run them in sequence.
.PHONY: fmt-go
fmt-go: pull-buildenv buildenv-dirs
	@docker run $(DOCKER_RUN_ARGS) gofmt -s -w $(NOMOS_CODE_DIRS)
	@docker run $(DOCKER_RUN_ARGS) goimports -w $(NOMOS_CODE_DIRS)

# Runs shfmt.
.PHONY: fmt-sh
fmt-sh:
	@./scripts/fmt-bash.sh

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: deps
deps: tidy vendor

.PHONY: lint
lint: lint-go lint-bash lint-yaml lint-license lint-license-headers

.PHONY: lint-go
lint-go: pull-buildenv buildenv-dirs
	@docker run $(DOCKER_RUN_ARGS) ./scripts/lint-go.sh $(NOMOS_GO_PKG)

.PHONY: lint-bash
lint-bash:
	@./scripts/lint-bash.sh

.PHONY: lint-license
lint-license: pull-buildenv buildenv-dirs
	@docker run $(DOCKER_RUN_ARGS) ./scripts/lint-license.sh

"$(GOBIN)/addlicense":
	go install github.com/google/addlicense@v1.0.0

"$(GOBIN)/kustomize":
	CGO_ENABLED=0 go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

# install kustomize binary for containerized testing
"$(BIN_DIR)/kustomize": "$(GOBIN)/kustomize"
	cp $(GOBIN)/kustomize $(BIN_DIR)/kustomize

.PHONY: license-headers
license-headers: "$(GOBIN)/addlicense"
	GOBIN=$(GOBIN) ./scripts/license-headers.sh add

.PHONY: lint-license-headers
lint-license-headers: "$(GOBIN)/addlicense"
	GOBIN=$(GOBIN) ./scripts/license-headers.sh lint

.PHONY: lint-yaml
lint-yaml:
	@./scripts/lint-yaml.sh

# Print the value of a variable
print-%:
	@echo $($*)

####################################################################################################
# MANUAL TESTING COMMANDS

# Reapply the resources, including a new image.  Use this to update with your code changes.
.PHONY: manual-test-refresh
manual-test-refresh: config-sync-manifest
	kubectl apply -f ${NOMOS_MANIFEST_STAGING_DIR}/config-sync-manifest.yaml

####################################################################################################