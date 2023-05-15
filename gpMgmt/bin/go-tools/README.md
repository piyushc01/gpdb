### Setup environment
### Install go
#### Linux:
```
yum install go
```
#### For Mac:
```
brew install go
```

### Build and Install
```
mkdir ~/workspace & cd ~/workspace
git clone git@github.com:greenplum-db/gpdb.git
cd ~/workspace/gpdb/gpMgmt/bin/go-tools
```

#### Get dependency and set env
```
make depend-dev    # set gobin path and fetch the protoc and mock dependency
```

#### Build Certificate
```
make cert    # generate certificates for given host
```
#### Build protobuf
```
make proto   # compile protobuf files to generate grpc code for hub and agents
```

#### Build & Install
```
make               # execute build target to compile and build gp binary
make install       # installs gp binary in GPHOME/bin
```

#### Cross-compile with:
```
make build_linux   # build gp binary for Linux platform
make build_mac     # build gp binary for Mac platform
```

#### Lint
```
make lint       # run lint
```

### Running gp utility

#### gp install to generate config file

```
gp install       # to generate config file with given conf setting
gp install --help  # to view the config options

example:
gp install --host <host> --server-certificate <path/to/server-cert.pem> --server-key < path/to/server-key.pem> --ca-certificate <path/to/ca-cert.pem> --ca-key <path/to/ca-key.pem>
```

#### Start/Stop hub
```
gp start hub
gp stop hub
```

#### Start/Stop agents
```
gp start agents
gp stop agents
```

#### Status of agent/hub
```
gp status hub
gp status agents
```

### Running Tests

#### Unit tests
```
make test     # run unit test in verbose mode
```

#### End-to-End tests
Creates a Concourse pipeline that includes various multi-host unit/functional tests.
```
make pipeline
```
To update the pipeline edit the yaml files in the `ci` directory and run
`make pipeline`.


#### Log Locations
logs are located in the path provided in configuration file. by default it will be generated in `/tmp` directory.
Logs are located on **_all hosts_**.
