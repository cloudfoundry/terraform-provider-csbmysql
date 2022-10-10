default: install

generate:
	go generate ./...

install:
	go install .

test:
	ginkgo ./...

testacc:
	TF_ACC=1 ginkgo -timeout 10m -v ./...
