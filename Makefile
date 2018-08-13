# Explicitly use bash and enable pipefail
SHELL=/bin/bash -o pipefail

# Project variables
PACKAGE = github.com/allenai/beaker
BASE    = $(GOPATH)/src/$(PACKAGE)
DIST    = $(BASE)/dist
VERSION = $(shell git describe --tags --abbrev=0 --match=v* 2> /dev/null || echo v0.0.0)
COMMIT  = $(shell git rev-parse HEAD)

# Standard paths
GOPATH = $(shell go env GOPATH)
GOBIN = $(GOPATH)/bin

# List of non-vendor packages. Can be overridden with a list of packages or the PKG variable.
PKGS = $(or $(PKG), $(PACKAGE)/...)
DIRS = $(shell go list -f '{{ .Dir }}' $(PKGS))

# Go tools
DEP = $(GOBIN)/dep
GOIMPORTS = $(GOBIN)/goimports
GOMETALINTER = $(GOBIN)/gometalinter

# Go
GO_SRC_FILES = $(shell find $(BASE) -type f -name '*.go')

# Dep is the soon-to-be-standard dependency management tool.
$(DEP):
	go get -u github.com/golang/dep/cmd/dep

# Goimports is an improvement over the gofmt tool which groups imports removes unused imports.
$(GOIMPORTS):
	go get -u golang.org/x/tools/cmd/goimports

# gometalinter combines multiple go linters.
$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

#
# Meta targets
#

# Artificial targets declared in the order they appear below.
.PHONY: clean git-hooks dev release dep test vet check-format format
.DEFAULT_GOAL := dev

# TODO: Split clean into clean-all to also clean the dependencies
clean:
	@if [ -f "$(BASE)/beaker" ]; then rm "$(BASE)/beaker"; fi
	@if [ -f "$(BASE)/Gopkg.lock" ]; then rm "$(BASE)/Gopkg.lock"; fi
	@if [ -d "$(BASE)/vendor" ]; then rm -rf "$(BASE)/vendor"; fi
	@if [ -d "$(DIST)" ]; then rm -rf "$(DIST)"; fi
	@if [ -f "$(GOBIN)/beaker" ]; then rm "$(GOBIN)/beaker"; fi

git-hooks: | $(GOIMPORTS)
	cp -f "$(BASE)/scripts/pre-commit" "$(BASE)/.git/hooks/pre-commit"

#
# Primary targets
#

# Build a dev binary against the current platform and place it.
dev: dep $(GOBIN)/beaker
$(GOBIN)/beaker: $(GO_SRC_FILES)
	@echo "Building for local development..."
	@go build -v --tags dev -o $@ -ldflags "\
		-X main.version=$(VERSION) \
		-X main.commit=$(COMMIT)" ./cmd/beaker

# Build release binaries for Beaker.
# This requires a github token to be set.
release: dep
	$(eval TEMP := $(shell mktemp -d))
ifeq ($(shell uname -s),Darwin)
	$(eval ARCHIVE := goreleaser_Darwin_x86_64.tar.gz)
else
	$(eval ARCHIVE := goreleaser_Linux_x86_64.tar.gz)
endif

	curl -L https://github.com/goreleaser/goreleaser/releases/download/v0.80.1/$(ARCHIVE) | tar -xvz -C$(TEMP) goreleaser
	$(TEMP)/goreleaser release --rm-dist
	rm -rf $(TEMP)

# Ensure dependencies
dep: $(BASE)/Gopkg.lock
$(BASE)/Gopkg.lock: Gopkg.toml | $(DEP)
	@$(DEP) ensure -v
	@touch -m $@

#
# Validation targets
#

test: dep | vet
	@echo Testing...
	@go test $(ARGS) $(PKGS)

# Run static analysis tools.
vet: check-format dep | $(GOMETALINTER)
	@echo Running 'go vet'...
	@go vet $(PKGS)

	@echo Running 'gometalinter'...
	@$(GOMETALINTER) --config=gometalinter-config.json $(DIRS)

# Validate or automatically correct formatting.
check-format: | $(GOIMPORTS)
	@echo Validating formatting...
	@VERIFY=1 "$(BASE)/scripts/format.sh"
format: | $(GOIMPORTS)
	@$(BASE)/scripts/format.sh
