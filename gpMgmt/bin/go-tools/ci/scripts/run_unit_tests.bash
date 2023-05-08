#!/bin/bash

set -eux -o pipefail

# Install GPDB
mkdir -p /usr/local/greenplum-db-devel
tar -xzf gpdb_binary/bin_gpdb.tar.gz -C /usr/local/greenplum-db-devel

# Run tests
export PATH=/usr/local/go/bin:$PATH
source /usr/local/greenplum-db-devel/greenplum_path.sh
cd gpdb_src/gpMgmt/bin/go-tools
make test
