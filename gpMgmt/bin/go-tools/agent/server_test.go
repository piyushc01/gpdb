package agent_test

import (
	"errors"
	"fmt"
	"github.com/greenplum-db/gpdb/gp/cli"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/utils"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	agent "github.com/greenplum-db/gpdb/gp/agent"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
)

func TestStartServer(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("successfully starts the server", func(t *testing.T) {
		credentials := &testutils.MockCredentials{}
		agentServer := agent.New(agent.Config{
			Port:        cli.DefaultAgentPort,
			ServiceName: cli.DefaultServiceName,
			Credentials: credentials,
		})
		errChan := make(chan error, 1)
		go func() {
			errChan <- agentServer.Start()
		}()
		defer agentServer.Shutdown()
		select {
		case err := <-errChan:
			if err != nil {
				t.Fatalf("unexpected error: %#v", err)
			}
		case <-time.After(1 * time.Second):
			t.Log("server started listening")
		}
	})
	t.Run("failed to start if the load credential fail", func(t *testing.T) {
		credentials := &testutils.MockCredentials{}
		credentials.SetCredsError("Test credential error")
		agentServer := agent.New(agent.Config{
			Port:        cli.DefaultAgentPort,
			ServiceName: cli.DefaultServiceName,
			Credentials: credentials,
		})
		errChan := make(chan error, 1)
		go func() {
			errChan <- agentServer.Start()
		}()
		defer agentServer.Shutdown()
		select {
		case err := <-errChan:
			expected := "Could not load credentials"
			if !strings.Contains(err.Error(), expected) {
				t.Fatalf("Expected %q to contain %q", err.Error(), expected)
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("Expected to fail, but the server started")
		}
	})
	t.Run("Listen fails when starting the server", func(t *testing.T) {
		credentials := &testutils.MockCredentials{}
		listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cli.DefaultAgentPort))
		if err != nil {
			t.Fatalf("Port: %d already in use. Error: %q", cli.DefaultAgentPort, err.Error())
		}
		defer listener.Close()
		agentServer := agent.New(agent.Config{
			Port:        cli.DefaultAgentPort,
			ServiceName: cli.DefaultServiceName,
			Credentials: credentials,
		})
		errChan := make(chan error, 1)
		go func() {
			errChan <- agentServer.Start()
		}()
		defer agentServer.Shutdown()
		select {
		case err := <-errChan:
			expected := fmt.Sprintf("Could not listen on port %d:", cli.DefaultAgentPort)
			if !strings.Contains(err.Error(), expected) {
				t.Fatalf("Expected %q to contain:%q", err.Error(), expected)
			}
		case <-time.After(1 * time.Second):
			t.Log("Failed to raise error if listener fail")
		}
	})
}

func TestGetStatus(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("service status returns unknown when no agent is running", func(t *testing.T) {
		credentials := &testutils.MockCredentials{}
		agentServer := agent.New(agent.Config{
			Port:        cli.DefaultAgentPort,
			ServiceName: cli.DefaultServiceName,
			Credentials: credentials,
		})
		actualStatus, err := agentServer.GetStatus()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		expected := &idl.ServiceStatus{
			Status: "Unknown",
			Pid:    0,
			Uptime: "Unknown",
		}
		if !reflect.DeepEqual(actualStatus, expected) {
			t.Fatalf("expected: %v got: %v", expected, actualStatus)
		}
	})
	t.Run("service status returns running and uptime when hub and agent is running", func(t *testing.T) {
		credentials := &testutils.MockCredentials{}
		agentServer := agent.New(agent.Config{
			Port:        cli.DefaultAgentPort,
			ServiceName: cli.DefaultServiceName,
			Credentials: credentials,
		})
		platform := &testutils.MockPlatform{}
		expected := &idl.ServiceStatus{
			Status: "Running",
			Uptime: "10ms",
			Pid:    uint32(1234),
		}
		platform.RetStatus = idl.ServiceStatus{
			Status: "Running",
			Uptime: "10ms",
			Pid:    uint32(1234),
		}
		platform.Err = nil
		agent.SetPlatform(platform)
		defer agent.ResetPlatform()
		/*start the hub and make sure it connects*/
		actualStatus, err := agentServer.GetStatus()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		if !reflect.DeepEqual(actualStatus, expected) {
			t.Fatalf("expected: %v got: %v", expected, actualStatus)
		}
	})
	t.Run("get service status when raised error", func(t *testing.T) {
		credentials := &testutils.MockCredentials{}
		agentServer := agent.New(agent.Config{
			Port:        cli.DefaultHubPort,
			ServiceName: cli.DefaultServiceName,
			Credentials: credentials,
		})
		platform := &testutils.MockPlatform{}
		platform.Err = errors.New("TEST Error")
		agent.SetPlatform(platform)
		defer agent.ResetPlatform()
		/*start the hub and make sure it connects*/
		_, err := agentServer.GetStatus()
		if err == nil {
			t.Fatalf("Expected error but found success : %#v", err)
		}
	})
}

func TestConfigWrite(t *testing.T) {
	caCertPath := "/tmp/test"
	caKeyPath := "/tmp/test"
	serverCertPath := "/tmp/test"
	serverKeyPath := "/tmp/test"
	hubPort := cli.DefaultHubPort
	agentPort := cli.DefaultAgentPort
	hostnames := []string{"localhost"}
	hubLogDir := "/tmp/test"
	serviceName := cli.DefaultServiceName
	cli.ConfigFilePath = "/tmp/gp.test_config"
	gphome := "/tmp/gphome"
	testhelper.SetupTestLogger()
	t.Run("config write followed by read returns same data", func(t *testing.T) {
		credentials := &utils.GpCredentials{
			CACertPath:     caCertPath,
			CAKeyPath:      caKeyPath,
			ServerCertPath: serverCertPath,
			ServerKeyPath:  serverKeyPath,
		}
		origCopyConfigFileToAgents := hub.CopyConfigFileToAgents
		defer func() { hub.CopyConfigFileToAgents = origCopyConfigFileToAgents }()
		hub.CopyConfigFileToAgents = func(conf *hub.Config, ConfigFilePath string) error {
			return nil
		}
		conf := &hub.Config{
			Port:        hubPort,
			AgentPort:   agentPort,
			Hostnames:   hostnames,
			LogDir:      hubLogDir,
			ServiceName: serviceName,
			GpHome:      gphome,
			Credentials: credentials,
		}
		err := conf.Write(cli.ConfigFilePath)
		if err != nil {
			t.Fatalf("Expected no error, got an error while creating config file:%v", err)
		}
		conf2 := &hub.Config{}
		err = conf2.Load(cli.ConfigFilePath)
		if err != nil {
			t.Fatalf("Expected no error, got an error while reading config file:%v", err)
		}
		if reflect.DeepEqual(conf, conf2) != true {
			t.Fatalf("Expected config:%v not same as Read Config:%v", conf, conf2)
		}
	})
	t.Run("config write returns error when copying to other hosts fails ", func(t *testing.T) {
		credentials := &utils.GpCredentials{
			CACertPath:     caCertPath,
			CAKeyPath:      caKeyPath,
			ServerCertPath: serverCertPath,
			ServerKeyPath:  serverKeyPath,
		}
		origCopyConfigFileToAgents := hub.CopyConfigFileToAgents
		defer func() { hub.CopyConfigFileToAgents = origCopyConfigFileToAgents }()
		hub.CopyConfigFileToAgents = func(conf *hub.Config, ConfigFilePath string) error {
			return fmt.Errorf("TEST Error copying files")
		}
		conf := &hub.Config{
			Port:        hubPort,
			AgentPort:   agentPort,
			Hostnames:   hostnames,
			LogDir:      hubLogDir,
			ServiceName: serviceName,
			GpHome:      gphome,
			Credentials: credentials,
		}
		err := conf.Write(cli.ConfigFilePath)
		if err == nil {
			t.Fatalf("Expected file copy error, got no error")
		}
	})
	t.Run("config write returns error when json marshalling fails ", func(t *testing.T) {
		credentials := &utils.GpCredentials{
			CACertPath:     caCertPath,
			CAKeyPath:      caKeyPath,
			ServerCertPath: serverCertPath,
			ServerKeyPath:  serverKeyPath,
		}
		origMasrshalIndent := hub.MasrshalIndent
		defer func() { hub.MasrshalIndent = origMasrshalIndent }()
		hub.MasrshalIndent = func(v any, prefix, indent string) ([]byte, error) {
			return nil, fmt.Errorf("TEST Error jason marshalling")
		}
		conf := &hub.Config{
			Port:        hubPort,
			AgentPort:   agentPort,
			Hostnames:   hostnames,
			LogDir:      hubLogDir,
			ServiceName: serviceName,
			GpHome:      gphome,
			Credentials: credentials,
		}
		err := conf.Write(cli.ConfigFilePath)
		if err == nil {
			t.Fatalf("Expected json marshalling error, got no error")
		}
	})
}
