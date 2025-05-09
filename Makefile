.DEFAULT_GOAL = help

SRC = $(shell find . -name "*.go" | grep -v "_test\." )

.PHONY: help
help: ## list Makefile targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: download ginkgo ## run all build, static analysis, and test steps

build: download $(SRC) ## build the provider
	goreleaser build --rm-dist --snapshot

.PHONY: clean
clean: ## clean up build artifacts
	- rm -rf dist
	- rm -rf /tmp/tpcsbmysql-non-fake.txt
	- rm -rf /tmp/tpcsbmysql-pkgs.txt
	- rm -rf /tmp/tpcsbmysql-coverage.out

download: ## download dependencies
	go mod download

.PHONY: ginkgo
ginkgo: ## run the tests with Ginkgo
	go tool ginkgo -r

.PHONY: ginkgo-coverage
ginkgo-coverage: ## ginkgo coverage score
	go list ./... | grep -v fake > /tmp/tpcsbmysql-non-fake.txt
	paste -sd "," /tmp/tpcsbmysql-non-fake.txt > /tmp/tpcsbmysql-pkgs.txt
	go test -coverpkg=`cat /tmp/tpcsbmysql-pkgs.txt` -coverprofile=/tmp/tpcsbmysql-coverage.out ./...
	go tool cover -func /tmp/tpcsbmysql-coverage.out | grep total
