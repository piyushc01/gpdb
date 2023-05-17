#!/bin/bash

set -eux -o pipefail

pushd gpdb_src/gpMgmt/bin/go-tools
make build
popd

tar -czf gp_binary/gp.tgz -C gpdb_src/gpMgmt/bin/go-tools gp
