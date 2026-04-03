# Copyright (c) The Kowabunga Project
# Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
# SPDX-License-Identifier: Apache-2.0

PKG_NAME=github.com/kowabunga-cloud/kahuna/internal/kahuna
VERSION=0.64.1
DIST=noble
CODENAME=NoFuture

SRC_DIR = internal
SDK_GENERATOR = go-server
SDK_PACKAGE_NAME = sdk
SDK_VERSION = "tags/v0.53.2"
#SDK_VERSION = "heads/main"
SDK_OPENAPI_SPEC = "https://raw.githubusercontent.com/kowabunga-cloud/openapi/refs/$(SDK_VERSION)/openapi.yaml"

#export GOOS=linux
#export GOARCH=amd64

# Make sure GOPATH is NOT set to this folder or we'll get an error "$GOPATH/go.mod exists but should not"
#export GOPATH = ""
export GO111MODULE = on
BINDIR = bin

NODE_DIR = ./node_modules
YARN = $(NODE_DIR)/.bin/yarn
OPENAPI_GENERATOR = $(NODE_DIR)/.bin/openapi-generator-cli

GOLINT = $(BINDIR)/golangci-lint
GOVULNCHECK = $(BINDIR)/govulncheck
GOSEC = $(BINDIR)/gosec

PKGS = $(shell go list ./... | grep -v /$(SDK_PACKAGE_NAME))
PKGS_SHORT = $(shell go list ./... | grep -v /$(SDK_PACKAGE_NAME) | sed 's%github.com/kowabunga-cloud/kahuna/%%')

V = 0
Q = $(if $(filter 1,$V),,@)
PROD = 0
ifeq ($(PROD),1)
DEBUG = -w -s
endif
M = $(shell printf "\033[34;1m▶\033[0m")

ifeq ($(V), 1)
  OUT = ""
else
  OUT = ">/dev/null"
endif

# This is our default target
# it does not build/run the tests
.PHONY: all
all: mod fmt vet lint build ; @ ## Do all
	$Q echo "done"

.PHONY: get-yarn
get-yarn: bin ;$(info $(M) [NPM] installing yarn…) @
	$Q test -x $(YARN) || npm install --silent yarn

.PHONY: get-openapi-generator
get-openapi-generator: get-yarn ;$(info $(M) [Yarn] installing openapi-generator-cli…) @
	$Q test -x $(OPENAPI_GENERATOR) || $(YARN) add --silent @openapitools/openapi-generator-cli 2> /dev/null
	$Q chmod a+x $(OPENAPI_GENERATOR)

# Generates server-side SDK from OpenAPI specification
.PHONY: sdk
sdk: get-openapi-generator ; $(info $(M) generate server-side SDK from OpenAPI specifications…) @
	$Q git rm -qrf $(SRC_DIR)/$(SDK_PACKAGE_NAME) || true
	$Q $(OPENAPI_GENERATOR) generate \
	  -g $(SDK_GENERATOR) \
	  --package-name $(SDK_PACKAGE_NAME) \
	  --openapi-normalizer KEEP_ONLY_FIRST_TAG_IN_OPERATION=true \
	  -p outputAsLibrary=true \
	  -p sourceFolder=$(SDK_PACKAGE_NAME) \
	  -i "$(SDK_OPENAPI_SPEC)" \
	  -o $(SRC_DIR) \
	  $(OUT)
	$Q rm -f $(SRC_DIR)/README.md
	$Q rm -f $(SRC_DIR)/.openapi-generator-ignore
	$Q rm -rf $(SRC_DIR)/.openapi-generator
	$Q rm -rf $(SRC_DIR)/api
	$Q git add "$(SRC_DIR)/$(SDK_PACKAGE_NAME)"

# This target grabs the necessary go modules
.PHONY: mod
mod: ; $(info $(M) collecting modules…) @
	$Q go mod download
	$Q go mod tidy

# Updates all go modules
update: ; $(info $(M) updating modules…) @
	$Q go get -u ./...
	$Q go mod tidy

# Makes sure bin directory is created
.PHONY: bin
bin: ; $(info $(M) create local bin directory) @
	$Q mkdir -p $(BINDIR)

.PHONY: build
build: bin ; $(info $(M) building Kahuna orchestrator…) @
	$Q go build \
		-gcflags="internal/...=-e" \
		-ldflags='$(DEBUG) -X $(PKG_NAME).version=$(VERSION) -X $(PKG_NAME).codename=$(CODENAME)' \
		-o $(BINDIR) ./cmd/kahuna

.PHONY: tests
tests: ; $(info $(M) testing Kowabunga suite…) @
	$Q go test ./... -count=1 -coverprofile=coverage.txt

.PHONY: deb
deb: ; $(info $(M) building Debian package…) @
	$Q VERSION=$(VERSION) DIST=$(DIST) ./debian.sh

.PHONY: apk
apk: ; $(info $(M) building Alpine package…) @
	$Q VERSION=$(VERSION) DIST=$(DIST) ./alpine.sh

.PHONY: get-lint
get-lint: ; $(info $(M) downloading go-lint…) @
	$Q test -x $(GOLINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s

.PHONY: lint
lint: get-lint ; $(info $(M) running go-lint…) @
	$Q $(GOLINT) run ./... ; exit 0

.PHONY: get-govulncheck
get-govulncheck: ; $(info $(M) downloading govulncheck…) @
	$Q test -x $(GOVULNCHECK) || GOBIN="$(PWD)/$(BINDIR)/" go install golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: vuln
vuln: get-govulncheck ; $(info $(M) running govulncheck…) @ ## Check for known vulnerabilities
	$Q $(GOVULNCHECK) ./... ; exit 0

.PHONY: get-gosec
get-gosec: ; $(info $(M) downloading gosec…) @
	$Q test -x $(GOSEC) || GOBIN="$(PWD)/$(BINDIR)/" go install github.com/securego/gosec/v2/cmd/gosec@latest

.PHONY: sec
sec: get-gosec ; $(info $(M) running gosec…) @ ## AST / SSA code checks
	$Q $(GOSEC) -terse -exclude=G101,G115 ./... ; exit 0

.PHONY: vet
vet: ; $(info $(M) running go vet…) @
	$Q go vet $(PKGS) ; exit 0

.PHONY: fmt
fmt: ; $(info $(M) running go fmt…) @
	$Q gofmt -w -s $(PKGS_SHORT)

.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	$Q rm -rf $(BINDIR)
	$Q rm -rf $(NODE_DIR)
	$Q rm -f package-lock.json
	$Q rm -f package.json
	$Q rm -f yarn.lock
	$Q rm -f openapitools.json

# This target parse this makefile and extract special comments to build a help
.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# This target count all the lines of .go files (no matter if empty lines or comments)
.PHONY: lc
lc: ; @
	@find . -name "*.go" -exec cat {} \; | wc -l | awk '{print $$1}'

# This target count the lines of go code only (ignore empty lines, comments, etc.)
# it requires gosloc
.PHONY: sloc
sloc: ; @
	@find . -name "*.go" -exec cat {} \; | gosloc
