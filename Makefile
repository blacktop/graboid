REPO=blacktop
NAME=graboid
CUR_VERSION=$(shell svu current)
NEXT_VERSION=$(shell svu patch)


.PHONY: setup
setup: ## Install all the build and lint dependencies
	@echo "===> Installing deps"
	go get -u github.com/alecthomas/gometalinter
	go install github.com/goreleaser/goreleaser
	go get -u github.com/pierrre/gotestcover
	go get -u github.com/spf13/cobra/cobra
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/caarlos0/svu
	gometalinter --install

.PHONY: update_mod
update_mod:
	rm go.sum
	go mod download
	go mod tidy

.PHONY: dry_release
dry_release:
	@echo " > Creating Pre-release Build ${NEXT_VERSION}"
	@goreleaser build --rm-dist --skip-validate

.PHONY: release
release: ## Create a new release from the NEXT_VERSION
	@echo " > Creating Release ${NEXT_VERSION}"
	@.hack/make/release ${NEXT_VERSION}
	@goreleaser --rm-dist

.PHONY: release-minor
release-minor: ## Create a new minor semver release
	@echo " > Creating Release $(shell svu minor)"
	@.hack/make/release $(shell svu minor)
	@goreleaser --rm-dist

destroy: ## Remove release from the CUR_VERSION
	@echo " > Deleting Release ${CUR_VERSION}"
	rm -rf dist
	git tag -d ${CUR_VERSION}
	git push origin :refs/tags/${CUR_VERSION}

build: ## Build a beta version of malice
	@echo "===> Building Binaries"
	go build -mod=vendor

clean: ## Clean up artifacts
	rm *.tar
	rm -rf dist

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
