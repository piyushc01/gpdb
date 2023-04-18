package agent_test

import (
	"errors"
	"github.com/greenplum-db/gpdb/gp/agent"
	"google.golang.org/grpc/credentials"
	"strings"
	"testing"
	"time"
)

type MockCredentials struct {
	TlsConnection credentials.TransportCredentials
	err error
}

func (s* MockCredentials)LoadServerCredentials() (credentials.TransportCredentials, error) {
	return s.TlsConnection,s.err
}

func (s* MockCredentials) LoadClientCredentials() (credentials.TransportCredentials, error){
	return s.TlsConnection,s.err
}

func TestGetStatus(t *testing.T){

	t.Run("successfully starts the server", func(t *testing.T) {

		credCmd := &MockCredentials{nil,nil}

		agentServer := agent.New(agent.Config{
			Port: 8000,
			ServiceName: "gp",
			CredentialsInterface: credCmd,
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

		credCmd := &MockCredentials{nil,errors.New("")}

		agentServer := agent.New(agent.Config{
			Port: 8000,
			ServiceName: "gp",
			CredentialsInterface: credCmd,
		})

		errChan := make(chan error, 1)
		go func() {
			errChan <- agentServer.Start()
		}()
		defer agentServer.Shutdown()

		select {
		case err := <-errChan:
			if err == nil || !strings.Contains(err.Error(),"Could not load credentials")  {
				t.Fatalf("want \"Could not load credentials\" but get: %q", err.Error())
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("Failed to raise error if load credential fail")
		}
	})
}
