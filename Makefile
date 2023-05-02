.DEFAULT_GOAL = help

GO-VERSION = 1.20.4
GO-VER = go$(GO-VERSION)

SRC = $(shell find . -name "*.go" | grep -v "_test\." )

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: version download ginkgo ## run all build, static analysis, and test steps

build: version download $(SRC) ## build the provider
	goreleaser build --rm-dist --snapshot

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf dist

download: ## download dependencies
	go mod download

.PHONY: ginkgo
ginkgo: ## run the tests with Ginkgo
	go run github.com/onsi/ginkgo/v2/ginkgo -r

.PHONY: version
version:
	@@go version
	@@if [ "$$(go version | awk '{print $$3}')" != "${GO-VER}" ]; then \
		echo "Go version does not match: expected: ${GO-VER}, got $$(go version | awk '{print $$3}')"; \
		exit 1; \
	fi
