all: cert build

.DEFAULT_GOAL := all
MODULE_NAME=gp

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

build:
	$(BUILD_ENV) go build -o gp $(BUILD_FLAGS) github.com/greenplum-db/gpdb/gp

proto:
	go generate ./idl

cert:
	./generate_test_tls_certificates.sh `hostname`

BUILD_FLAGS = -gcflags="all=-N -l"

install:
	GOBIN=$(GPHOME)/bin go install $(BUILD_FLAGS) github.com/greenplum-db/gpdb/gp

clean:
	rm -f gp
	rm -f ./idl/*.pb.go
	rm -rf ./certificates
	rm -rf /tmp/go-build*
	rm -rf /tmp/gexec_artifacts*
	rm -rf $(GPHOME)/bin/gp
