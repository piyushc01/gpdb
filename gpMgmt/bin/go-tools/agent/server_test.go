package agent_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/agent"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
	"google.golang.org/grpc/credentials"
)

type MockCredentials struct {
	TlsConnection credentials.TransportCredentials
	err           error
}

func (s *MockCredentials) LoadServerCredentials() (credentials.TransportCredentials, error) {
	return s.TlsConnection, s.err
}

func (s *MockCredentials) LoadClientCredentials() (credentials.TransportCredentials, error) {
	return s.TlsConnection, s.err
}

func TestStartServer(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("successfully starts the server", func(t *testing.T) {

		credCmd := &MockCredentials{nil, nil}

		agentServer := agent.New(agent.Config{
			Port:        8000,
			ServiceName: "gp",
			Credentials: credCmd,
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

		credCmd := &MockCredentials{nil, errors.New("")}

		agentServer := agent.New(agent.Config{
			Port:        8001,
			ServiceName: "gp",
			Credentials: credCmd,
		})

		errChan := make(chan error, 1)
		go func() {
			errChan <- agentServer.Start()
		}()
		defer agentServer.Shutdown()

		select {
		case err := <-errChan:
			if err == nil || !strings.Contains(err.Error(), "Could not load credentials") {
				t.Fatalf("want \"Could not load credentials\" but get: %q", err.Error())
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("Failed to raise error if load credential fail")
		}
	})
}

func TestGetStatus(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("get service status when no agent is running", func(t *testing.T) {

		credCmd := &MockCredentials{nil, nil}

		agentServer := agent.New(agent.Config{
			Port:        8000,
			ServiceName: "gp",
			Credentials: credCmd,
		})

		msg, err := agentServer.GetStatus()

		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		if msg.Status != "Unknown" || msg.Pid != 0 || msg.Uptime != "Unknown" {
			t.Fatalf("expected unknown status not found")
		}

	})

	t.Run("get service status when hub and agent is running", func(t *testing.T) {

		credCmd := &MockCredentials{nil, nil}

		agentServer := agent.New(agent.Config{
			Port:        8000,
			ServiceName: "gp",
			Credentials: credCmd,
		})

		os := testutils.MockPlatform{}
		os.RetStatus = idl.ServiceStatus{Status: "running", Uptime: "10ms", Pid: uint32(1234)}
		os.Err = nil
		agent.SetPlatform(os)
		defer agent.ResetPlatform()

		/*start the hub and make sure it connects*/
		msg, err := agentServer.GetStatus()

		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

		if msg.Status == "Unknown" || msg.Pid == 0 || msg.Uptime == "Unknown" {
			t.Fatalf("expected unknown status not found")
		}
	})

	t.Run("get service status when raised error", func(t *testing.T) {

		credCmd := &MockCredentials{nil, nil}

		agentServer := agent.New(agent.Config{
			Port:        8000,
			ServiceName: "gp",
			Credentials: credCmd,
		})

		os := testutils.MockPlatform{}
		os.Err = errors.New("")
		agent.SetPlatform(os)
		defer agent.ResetPlatform()

		/*start the hub and make sure it connects*/
		_, err := agentServer.GetStatus()

		if err == nil {
			t.Fatalf("Expected error but found success : %#v", err)
		}
	})
}
