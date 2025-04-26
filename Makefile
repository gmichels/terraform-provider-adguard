TEST?=$$(go list ./adguard | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=gmichels
NAME=adguard
BINARY=terraform-provider-${NAME}
VERSION=0.0.1
GOOS=darwin
GOARCH=amd64
OS_ARCH=${GOOS}_${GOARCH}
TERRAFORM_PLUGINS=~/.terraform.d/plugins

all, help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nMakefile help:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

default: install

build: ## Build the provider
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${BINARY}

release: ### Build and release binaries
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build ### Build and install the provider to the terraform plugins directory
	mkdir -p ${TERRAFORM_PLUGINS}/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ${TERRAFORM_PLUGINS}/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/${BINARY}_v${VERSION}

test: ## Run tests
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: ## Run acceptance tests using the docker-compose file
	docker compose -f ./docker/docker-compose.yaml up -d
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 10m
	docker compose -f ./docker/docker-compose.yaml down
	git checkout HEAD -- ./docker/conf/AdGuardHome.yaml

cleanup: ## Cleanup leftovers after failed tests
	docker compose -f ./docker/docker-compose.yaml down
	rm -fr ./docker
	git checkout HEAD -- ./docker
