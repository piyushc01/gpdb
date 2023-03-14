all: build

.DEFAULT_GOAL := all
MODULE_NAME=gp

.PHONY: unit test

TEST_PACKAGES := ./...

unit: TEST_PACKAGES := $(shell go list ./...)
unit:
	go test -count=1 $(TEST_PACKAGES)

test: unit

build:
	$(BUILD_ENV) go build -o gp $(BUILD_FLAGS) github.com/greenplum-db/gpdb/gp

proto:
	go generate ./idl

cert:
	./generate_test_tls_certificates.sh

BUILD_FLAGS = -gcflags="all=-N -l"

install:
	go install $(BUILD_FLAGS) github.com/greenplum-db/gpdb/gp

clean:
	rm -f gp
	rm -f ./idl/*.pb.go
	rm -rf ./certificates
	rm -rf /tmp/go-build*
	rm -rf /tmp/gexec_artifacts*
