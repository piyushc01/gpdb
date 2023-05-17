all: build

.DEFAULT_GOAL := all
MODULE_NAME=gp


LINUX_ENV := env GOOS=linux GOARCH=amd64
MAC_ENV := env GOOS=darwin GOARCH=amd64


.PHONY: unit test
TEST_PACKAGES := ./...
unit: TEST_PACKAGES := $(shell go list ./...)
unit:
	go test -v -count=1 $(TEST_PACKAGES) --cover

test: unit

.PHONY: depend-dev
depend-dev: export GOBIN := $(CURDIR)/dev-bin
depend-dev: export GOFLAGS := -mod=readonly # do not update dependencies during installation
depend-dev:
	mkdir -p $(GOBIN)
	go install github.com/golang/protobuf/protoc-gen-go@v1.3.2
	go install github.com/golang/mock/mockgen@v1.6.0
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.52.2

.PHONY: lint
lint:
	golangci-lint run -e ^copylocks

BUILD_ENV = $($(OS)_ENV)

.PHONY: build build_linux build_mac
build:
	$(BUILD_ENV) go build -o gp $(BUILD_FLAGS) github.com/greenplum-db/gpdb/gp

build_linux: OS := LINUX
build_mac: OS := MAC
build_linux build_mac: build

proto:
	go generate ./idl

cert:
	./generate_test_tls_certificates.sh `hostname`

BUILD_FLAGS = -gcflags="all=-N -l"

install:
	GOBIN=$(GPHOME)/bin go install $(BUILD_FLAGS) github.com/greenplum-db/gpdb/gp

# goimports supports a superset of the formatting rules gofmt supports, but
# gofmt allows custom formatting rules so we include both for now
format:
	goimports -l -w agent/ cli/ hub/ testutils/ utils/
	gofmt -l -w agent/ cli/ hub/ testutils/ utils/

# You can override these from the command line.
GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
GIT_REMOTE ?= $(shell git ls-remote --get-url)
PIPELINE_NAME ?= gpdb-dev-$(shell git rev-parse --abbrev-ref HEAD | tr '/' '-')
pipeline:
	fly -t dev set-pipeline -p $(PIPELINE_NAME) \
		-c ci/pipelines/pipeline.yml \
		-l ../../../concourse/vars/common_prod.yml \
		-l ../../../concourse/vars/common_dev.yml \
		-v gpdb-git-remote=$(GIT_REMOTE) \
		-v gpdb-git-branch=$(GIT_BRANCH) \
		-v pipeline-name=$(PIPELINE_NAME)

clean:
	rm -f gp
	rm -f ./idl/*.pb.go
	rm -rf ./certificates
	rm -rf /tmp/go-build*
	rm -rf /tmp/gexec_artifacts*
	rm -rf $(GPHOME)/bin/gp
