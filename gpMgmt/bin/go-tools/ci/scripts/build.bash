#!/bin/bash

set -eux -o pipefail

ccp_src/scripts/setup_ssh_to_cluster.sh

scp cluster_env_files/hostfile_all cdw:/tmp

# Install patchelf. We need to SSH as root, hence the use of
# cluster_env_files.
ssh -t ccp-$(cat cluster_env_files/terraform/name)-0 "sudo bash -c '
    source /home/gpadmin/gpdb_src/concourse/scripts/common.bash
'"

ssh -n cdw "
    set -eux -o pipefail

    source /usr/local/greenplum-db-devel/greenplum_path.sh
    export PATH=/usr/local/go/bin:$PATH
    cd /home/gpadmin/gpdb_src/gpMgmt/bin/go-tools
    make depend-dev
    make build
    make install
    source /usr/local/greenplum-db-devel/greenplum_path.sh
    gpsync -f /tmp/hostfile_all /usr/local/greenplum-db-devel/bin/gp =:/usr/local/greenplum-db-devel/bin/gp
    ./ci/scripts/generate_ssl_cert_multi_host.bash

    gp install --hostfile /tmp/hostfile_all --server-certificate /tmp/certificates/server-cert.pem --server-key /tmp/certificates/server-key.pem --ca-certificate /tmp/certificates/ca-cert.pem --ca-key /tmp/certificates/ca-key.pem
    gp start hub
    gp status hub
    gp start agents
    gp status agents
    gp stop agents
    gp stop hub
"
