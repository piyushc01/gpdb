package gpservice_cli_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/greenplum-db/gpdb/gpservice/gpservice-cli"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gpdb/gpservice/hub"
	"github.com/greenplum-db/gpdb/gpservice/idl"
	"github.com/greenplum-db/gpdb/gpservice/idl/mock_idl"
)

func TestStopAgentService(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("StopAgentService stops the agent service when theres no error", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StopAgents(gomock.Any(), gomock.Any())
			return hubClient, nil
		}

		err := gpservice_cli.StopAgentService()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("StopAgentService returns an error when theres error connecting hub", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error connecting Hub"
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}

		err := gpservice_cli.StopAgentService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
	t.Run("StopAgentService returns error when theres error stopping agents", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error stopping agents"
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StopAgents(gomock.Any(), gomock.Any()).Return(nil, errors.New(expectedStr))
			return hubClient, nil
		}
		err := gpservice_cli.StopAgentService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestStopHubService(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("Stops hub when theres no error", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().Stop(gomock.Any(), gomock.Any())
			return hubClient, nil
		}

		err := gpservice_cli.StopHubService()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("Stops hub when theres error connecting hub service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error connecting Hub"
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}

		err := gpservice_cli.StopHubService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})

	t.Run("Stop returns error when there's error stopping Hub", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error stopping Hub"
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(nil, errors.New(expectedStr))
			return hubClient, nil
		}

		err := gpservice_cli.StopHubService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestRunStopServices(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("Returns no error when there is none", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.StopAgentService = funcNilError()
		gpservice_cli.StopHubService = funcNilError()

		err := gpservice_cli.RunStopServices(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("Returns  error when error stopping Agent service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error stopping Agent service"
		gpservice_cli.StopAgentService = funcErrorMessage(expectedStr)

		err := gpservice_cli.RunStopServices(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})

	t.Run("Returns error when error stopping hub service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "Error stopping Hub service"
		gpservice_cli.StopAgentService = funcNilError()
		gpservice_cli.StopHubService = funcErrorMessage(expectedStr)

		err := gpservice_cli.RunStopServices(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestRunStopAgents(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("returns no error where there none no verbose", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.StopAgentService = funcNilError()
		gpservice_cli.Verbose = false

		err := gpservice_cli.RunStopAgents(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

	})
	t.Run("returns no error where there none in verbose", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.StopAgentService = funcNilError()
		gpservice_cli.ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return nil
		}
		gpservice_cli.Verbose = true

		err := gpservice_cli.RunStopAgents(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("returns error where there is error stopping agent service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error stopping agents"
		gpservice_cli.StopAgentService = funcErrorMessage(expectedStr)

		err := gpservice_cli.RunStopAgents(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestRunStopHub(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("return no error when there is none non verbose mode", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.StopHubService = funcNilError()
		gpservice_cli.Verbose = false

		err := gpservice_cli.RunStopHub(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("return no error when there is none verbose mode", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.StopHubService = funcNilError()
		gpservice_cli.ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		gpservice_cli.Verbose = true

		err := gpservice_cli.RunStopHub(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("return error when there is error stoppig Hub service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error while stopping Hub Service"
		gpservice_cli.StopHubService = funcErrorMessage(expectedStr)
		gpservice_cli.Verbose = false

		err := gpservice_cli.RunStopHub(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}
