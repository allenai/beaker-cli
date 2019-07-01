# Explicitly use bash and enable pipefail
SHELL=/bin/bash -o pipefail

# Project variables
BASE    = $(realpath .)
GOPATH  = $(shell go env GOPATH)
DIST    = $(BASE)/dist

VERSION = $(shell git describe --tags --abbrev=0 --match=v* 2> /dev/null || echo v0.0.0)
COMMIT  = $(shell git rev-parse HEAD)

# List of non-vendor packages. Can be overridden with a list of packages or the PKG variable.
PKGS = $(or $(PKG), $(BASE)/...)
DIRS = $(shell go list -f '{{ .Dir }}' $(PKGS))

# Go tools
GOIMPORTS = $(GOPATH)/bin/goimports

# Go
GO_SRC_FILES = $(shell find $(BASE) -type f -name '*.go')

# Goimports is an improvement over the gofmt tool which groups imports removes unused imports.
$(GOIMPORTS):
	go get golang.org/x/tools/cmd/goimports@1c3d964395ce8f04f3b03b30aaed0b096c08c3c6

#
# Meta targets
#

# Artificial targets declared in the order they appear below.
.PHONY: clean git-hooks dev release test check-format format
.DEFAULT_GOAL := dev

# TODO: Split clean into clean-all to also clean the dependencies
clean:
	@if [ -f "$(BASE)/beaker" ]; then rm "$(BASE)/beaker"; fi
	@if [ -f "$(BASE)/Gopkg.lock" ]; then rm "$(BASE)/Gopkg.lock"; fi
	@if [ -d "$(BASE)/vendor" ]; then rm -rf "$(BASE)/vendor"; fi
	@if [ -d "$(DIST)" ]; then rm -rf "$(DIST)"; fi

git-hooks: | $(GOIMPORTS)
	cp -f "$(BASE)/scripts/pre-commit" "$(BASE)/.git/hooks/pre-commit"

#
# Primary targets
#

# Build a dev binary against the current platform and place it.
dev: beaker
beaker: $(GO_SRC_FILES)
	@echo "Building for local development..."
	@go build -v --tags dev -o $@ -ldflags "\
		-X github.com/allenai/beaker/client.version=$(VERSION) \
		-X main.commit=$(COMMIT)" ./cmd/beaker
		-X main.version=$(VERSION) \

# Build release binaries for Beaker.
# This requires a github token to be set.
release:
	$(eval TEMP := $(shell mktemp -d))
ifeq ($(shell uname -s),Darwin)
	$(eval ARCHIVE := goreleaser_Darwin_x86_64.tar.gz)
else
	$(eval ARCHIVE := goreleaser_Linux_x86_64.tar.gz)
endif

	curl -L https://github.com/goreleaser/goreleaser/releases/download/v0.89.0/$(ARCHIVE) | tar -xvz -C$(TEMP) goreleaser
	$(TEMP)/goreleaser release --rm-dist
	rm -rf $(TEMP)

#
# Validation targets
#

test: check-format
	@echo Testing...
	@go test $(ARGS) $(PKGS)
	@echo Running 'go vet'...
	@go vet $(PKGS)

# Validate or automatically correct formatting.
check-format: | $(GOIMPORTS)
	@echo Validating formatting...
	@VERIFY=1 "$(BASE)/scripts/format.sh"
format: | $(GOIMPORTS)
	@$(BASE)/scripts/format.sh
