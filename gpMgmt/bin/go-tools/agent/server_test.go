package agent_test

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	agent "github.com/greenplum-db/gpdb/gp/agent"
	"github.com/greenplum-db/gpdb/gp/constants"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
)

func TestStartServer(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("successfully starts the server", func(t *testing.T) {
		credentials := &testutils.MockCredentials{}
		agentServer := agent.New(agent.Config{
			Port:        constants.DefaultAgentPort,
			ServiceName: constants.DefaultServiceName,
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
			Port:        constants.DefaultAgentPort,
			ServiceName: constants.DefaultServiceName,
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
		listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", constants.DefaultAgentPort))
		if err != nil {
			t.Fatalf("Port: %d already in use. Error: %q", constants.DefaultAgentPort, err.Error())
		}
		defer listener.Close()
		agentServer := agent.New(agent.Config{
			Port:        constants.DefaultAgentPort,
			ServiceName: constants.DefaultServiceName,
			Credentials: credentials,
		})
		errChan := make(chan error, 1)
		go func() {
			errChan <- agentServer.Start()
		}()
		defer agentServer.Shutdown()
		select {
		case err := <-errChan:
			expected := fmt.Sprintf("Could not listen on port %d:", constants.DefaultAgentPort)
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

	t.Run("returns appropriate status when no agent is running", func(t *testing.T) {
		credentials := &testutils.MockCredentials{}
		agentServer := agent.New(agent.Config{
			Port:        constants.DefaultAgentPort,
			ServiceName: constants.DefaultServiceName,
			Credentials: credentials,
		})
		actualStatus, err := agentServer.GetStatus()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
		expected := &idl.ServiceStatus{
			Status: "not running",
			Pid:    0,
			Uptime: "",
		}
		if !reflect.DeepEqual(actualStatus, expected) {
			t.Fatalf("expected: %v got: %v", expected, actualStatus)
		}
	})
	t.Run("service status returns running and uptime when hub and agent is running", func(t *testing.T) {
		credentials := &testutils.MockCredentials{}
		agentServer := agent.New(agent.Config{
			Port:        constants.DefaultAgentPort,
			ServiceName: constants.DefaultServiceName,
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
			Port:        constants.DefaultHubPort,
			ServiceName: constants.DefaultServiceName,
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
