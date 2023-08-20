package hub_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/constants"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/testutils"
	"google.golang.org/grpc"
)

func TestStartServer(t *testing.T) {

	testhelper.SetupTestLogger()
	host, _ := os.Hostname()
	gpHome := os.Getenv("GPHOME")

	t.Run("successfully starts the hub server", func(t *testing.T) {

		credCmd := &testutils.MockCredentials{}

		conf := &hub.Config{
			1234,
			8080,
			[]string{host},
			"/tmp/logDir",
			"gp",
			gpHome,
			credCmd,
		}

		hubServer := hub.New(conf, grpc.DialContext)

		errChan := make(chan error, 1)
		go func() {
			errChan <- hubServer.Start()
		}()

		defer hubServer.Shutdown()

		select {
		case err := <-errChan:
			if err != nil {
				t.Fatalf("unexpected error: %#v", err)
			}
		case <-time.After(1 * time.Second):
			t.Log("hub server started listening")
		}

	})

	t.Run("failed to start if the load credential fail", func(t *testing.T) {

		credCmd := &testutils.MockCredentials{}
		credCmd.SetCredsError("Test error in loading creds")

		conf := &hub.Config{
			1235,
			8080,
			[]string{host},
			"/tmp/logDir",
			"gp",
			gpHome,
			credCmd,
		}
		hubServer := hub.New(conf, grpc.DialContext)

		errChan := make(chan error, 1)
		go func() {
			errChan <- hubServer.Start()
		}()
		defer hubServer.Shutdown()

		select {
		case err := <-errChan:
			if !strings.Contains(err.Error(), "Could not load credentials") {
				t.Fatalf("want \"Could not load credentials\" but get: %q", err.Error())
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("Failed to raise error if load credential fail")
		}
	})
}

func TestStartAgents(t *testing.T) {
	host, _ := os.Hostname()
	gpHome := os.Getenv("GPHOME")

	testCases := []struct {
		name        string
		conf        *hub.Config
		expectedErr string
	}{
		{
			name: "successfully starts the agents from hub",
			conf: &hub.Config{
				constants.DefaultHubPort,
				constants.DefaultAgentPort,
				[]string{host},
				"/tmp/logDir",
				constants.DefaultServiceName,
				gpHome,
				&testutils.MockCredentials{},
			},
			expectedErr: "",
		},
		{
			name: "failed to start if the host is not reachable",
			conf: &hub.Config{
				constants.DefaultHubPort,
				constants.DefaultAgentPort,
				[]string{"test"},
				"/tmp/logDir",
				constants.DefaultServiceName,
				gpHome,
				&testutils.MockCredentials{},
			},
			expectedErr: "unable to login",
		},
		{
			name: "failed to start if the gphome is not set",
			conf: &hub.Config{
				constants.DefaultHubPort,
				constants.DefaultAgentPort,
				[]string{host},
				"/tmp/logDir",
				constants.DefaultServiceName,
				"gphome",
				&testutils.MockCredentials{},
			},
			expectedErr: "No such file or directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testhelper.SetupTestLogger()
			hubServer := hub.New(tc.conf, grpc.DialContext)
			defer hubServer.Shutdown()

			err := hubServer.StartAllAgents()

			if tc.expectedErr != "" {
				if !strings.Contains(err.Error(), tc.expectedErr) {
					t.Fatalf("expected %s, but got: %#v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %#v", err)
				}
			}
		})
	}
}
