package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/agent"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
	"google.golang.org/grpc"
)

func TestParseConfig(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("ParseConfig parses the configuration correctly without any error", func(t *testing.T) {
		origReadFile := ReadFile
		defer func() { ReadFile = origReadFile }()
		ReadFile = func(name string) ([]byte, error) {
			return []byte(""), nil
		}

		origUnmarshal := Unmarshal
		defer func() { Unmarshal = origUnmarshal }()
		Unmarshal = func(data []byte, v any) error {
			return nil

		}
		_, err := ParseConfig("testfile.test")
		if err != nil {
			t.Fatalf("Got the error when no error expected in the config file parsing")
		}
	})

	t.Run("ParseConfig fails when reading the config file", func(t *testing.T) {
		expectedErr := "Error opening file. File does not exists."
		origReadFile := ReadFile
		defer func() { ReadFile = origReadFile }()
		ReadFile = func(name string) ([]byte, error) {
			return []byte(""), errors.New(expectedErr)
		}

		origUnmarshal := Unmarshal
		defer func() { Unmarshal = origUnmarshal }()
		Unmarshal = func(data []byte, v any) error {
			return nil

		}
		_, err := ParseConfig("testfile.test")
		if err == nil {
			t.Fatalf("Got the error when no error expected in the config file parsing")
		}
	})
	t.Run("ParseConfig fails when json parsing fails to parse the config file", func(t *testing.T) {
		expectedErr := "Error parsing json file."
		origReadFile := ReadFile
		defer func() { ReadFile = origReadFile }()
		ReadFile = func(name string) ([]byte, error) {
			return []byte(""), nil
		}

		origUnmarshal := Unmarshal
		defer func() { Unmarshal = origUnmarshal }()
		Unmarshal = func(data []byte, v any) error {
			return errors.New(expectedErr)

		}
		_, err := ParseConfig("testfile.test")
		if err == nil {
			t.Fatalf("Got the error when no error expected in the config file parsing")
		}
	})
	t.Run("ParseConfig fails when actual json parsing ", func(t *testing.T) {
		origReadFile := ReadFile
		defer func() { ReadFile = origReadFile }()
		ReadFile = func(name string) ([]byte, error) {
			return []byte(" TEST DATA"), nil
		}

		_, err := ParseConfig("testfile.test")
		if err == nil {
			t.Fatalf("Got the error when no error expected in the config file parsing")
		}
	})
	t.Run("ParseConfig properly populates the configuration : hub port, agent port, log dir, service name, gphome", func(t *testing.T) {
		confData := []byte("{\n\t\"hubPort\": 5555,\n\t\"agentPort\": 8888,\n\t\"hubLogDir\": \"/testlogdir\",\n\t\"serviceName\": \"testServiceName\",\n\t\"gphome\": \"/test/testgphome\"\n}")
		origReadFile := ReadFile
		defer func() { ReadFile = origReadFile }()
		ReadFile = func(name string) ([]byte, error) {
			return []byte(confData), nil
		}

		conf, err := ParseConfig("testfile.test")
		if err != nil {
			t.Fatalf("Got the error when no error expected in the config file parsing. error:%s", err.Error())
		}
		if conf.Port != 5555 {
			t.Fatalf("Config file hub port 5555 is not matching with port from the parsed config:%d", conf.Port)
		}
		if conf.AgentPort != 8888 {
			t.Fatalf("Config file agent port 8888 is not matching with port from parsed config: %d", conf.AgentPort)
		}
		if conf.LogDir != "/testlogdir" {
			t.Fatalf("Config file log directory is not matching with log directory from config file:%s", conf.LogDir)
		}
		if conf.ServiceName != "testServiceName" {
			t.Fatalf("Config file service name not matching with servicename from the config:%s", conf.ServiceName)
		}
		if conf.GpHome != "/test/testgphome" {
			t.Fatalf("Config file gphome not matching with gphome from the config:%s", conf.GpHome)
		}
	})
	t.Run("ParseConfig properly populates the configuration hostlist, credentials", func(t *testing.T) {
		confData := []byte("{\n  \"hostnames\": [\n    \"sdw1\",\n    \"sdw2\",\n    \"sdw3\"\n  ],\n  \"Credentials\": {\n    \"caCert\": \"/test/ca-cert.pem\",\n    \"caKey\": \"/test/ca-key.pem\",\n    \"serverCert\": \"/test/server-cert.pem\",\n    \"serverKey\": \"/test/server-key.pem\"\n  }\n}")
		origReadFile := ReadFile
		defer func() { ReadFile = origReadFile }()
		ReadFile = func(name string) ([]byte, error) {
			return []byte(confData), nil
		}

		conf, err := ParseConfig("testfile.test")
		if err != nil {
			t.Fatalf("Got the error when no error expected in the config file parsing. error:%s", err.Error())
		}
		if len(conf.Hostnames) != 3 {
			t.Fatalf("Hostnames list length expected 3, got length %d", len(conf.Hostnames))
		}
		if conf.Hostnames[0] != "sdw1" && conf.Hostnames[1] != "sdw2" && conf.Hostnames[2] != "sdw3" {
			t.Fatalf("Hostnames from the config file not matching with the actual config. Got 1:%s, 2:%s 3:%s", conf.Hostnames[0], conf.Hostnames[1], conf.Hostnames[2])
		}
		// Check credentials has proper value
		CACertPath, CAKeyPath, ServerCertPath, ServerKeyPath := conf.Credentials.GetClientServerCredsPath()
		if CACertPath != "/test/ca-cert.pem" {
			t.Fatalf("CACertPath not matching with the path from the conf:%s", CACertPath)
		}
		if CAKeyPath != "/test/ca-key.pem" {
			t.Fatalf("CACertPath not matching with the path from the conf:%s", CAKeyPath)
		}
		if ServerCertPath != "/test/server-cert.pem" {
			t.Fatalf("ServerCertPath not matching with the path from the conf:%s", ServerCertPath)
		}
		if ServerKeyPath != "/test/server-key.pem" {
			t.Fatalf("ServerCertPath not matching with the path from the conf:%s", ServerKeyPath)
		}

	})
}

func TestConnectToHub(t *testing.T) {
	testhelper.SetupTestLogger()
	creds := &testutils.MockCredentials{}
	os := &testutils.MockPlatform{}
	os.RetStatus = idl.ServiceStatus{Status: "", Uptime: "", Pid: uint32(0)}
	os.Err = nil
	agent.SetPlatform(os)
	defer agent.ResetPlatform()
	hostlist := []string{"sdw1", "sdw2", "sdw3"}
	config := hub.Config{Port: 4444, AgentPort: 5555, Hostnames: hostlist, LogDir: "/tmp", ServiceName: "gp", GpHome: "/usr/local/gpdb/", Credentials: creds}

	t.Run("Connect to hub succeeds and no error thrown", func(t *testing.T) {
		origDialContext := DialContext
		defer func() { DialContext = origDialContext }()
		DialContext = func(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
			return &grpc.ClientConn{}, nil
		}
		_, err := connectToHub(&config)
		if err != nil {
			t.Fatalf("Unexpected error happened when connecting to hub:%s", err.Error())
		}
	})
	t.Run("Connect to hub throws error when Dial context fails", func(t *testing.T) {
		origDialContext := DialContext
		defer func() { DialContext = origDialContext }()
		DialContext = func(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
			return nil, errors.New("TEST ERROR while dialing context")
		}

		_, err := connectToHub(&config)
		if err == nil {
			t.Fatalf("Expected dial error error, but did not get any error")
		}
	})
	t.Run("Connect to hub throws error when load client credentials fail", func(t *testing.T) {
		origDialContext := DialContext
		defer func() { DialContext = origDialContext }()
		DialContext = func(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
			return &grpc.ClientConn{}, nil
		}

		creds.SetCredsError("Load credentials error")
		defer creds.ResetCredsError()
		_, err := connectToHub(&config)
		if err == nil {
			t.Fatalf("Expected load credential error error, but did not get any error")
		}
	})
}
