package gpservice_cli_test

import (
	"errors"
	cli2 "github.com/greenplum-db/gpdb/gp/gpservice/gpservice-cli"
	"github.com/greenplum-db/gpdb/gp/gpservice/hub"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/idl/mock_idl"
)

func TestStopAgentService(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("StopAgentService stops the agent service when theres no error", func(t *testing.T) {
		defer resetCLIVars()
		cli2.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StopAgents(gomock.Any(), gomock.Any())
			return hubClient, nil
		}

		err := cli2.StopAgentService()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("StopAgentService returns an error when theres error connecting hub", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error connecting Hub"
		cli2.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}

		err := cli2.StopAgentService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
	t.Run("StopAgentService returns error when theres error stopping agents", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error stopping agents"
		cli2.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StopAgents(gomock.Any(), gomock.Any()).Return(nil, errors.New(expectedStr))
			return hubClient, nil
		}
		err := cli2.StopAgentService()
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
		cli2.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().Stop(gomock.Any(), gomock.Any())
			return hubClient, nil
		}

		err := cli2.StopHubService()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("Stops hub when theres error connecting hub service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error connecting Hub"
		cli2.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}

		err := cli2.StopHubService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})

	t.Run("Stop returns error when there's error stopping Hub", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error stopping Hub"
		cli2.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(nil, errors.New(expectedStr))
			return hubClient, nil
		}

		err := cli2.StopHubService()
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
		cli2.StopAgentService = funcNilError()
		cli2.StopHubService = funcNilError()

		err := cli2.RunStopServices(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("Returns  error when error stopping Agent service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error stopping Agent service"
		cli2.StopAgentService = funcErrorMessage(expectedStr)

		err := cli2.RunStopServices(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})

	t.Run("Returns error when error stopping hub service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "Error stopping Hub service"
		cli2.StopAgentService = funcNilError()
		cli2.StopHubService = funcErrorMessage(expectedStr)

		err := cli2.RunStopServices(nil, nil)
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
		cli2.StopAgentService = funcNilError()
		cli2.Verbose = false

		err := cli2.RunStopAgents(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}

	})
	t.Run("returns no error where there none in verbose", func(t *testing.T) {
		defer resetCLIVars()
		cli2.StopAgentService = funcNilError()
		cli2.ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return nil
		}
		cli2.Verbose = true

		err := cli2.RunStopAgents(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("returns error where there is error stopping agent service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error stopping agents"
		cli2.StopAgentService = funcErrorMessage(expectedStr)

		err := cli2.RunStopAgents(nil, nil)
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
		cli2.StopHubService = funcNilError()
		cli2.Verbose = false

		err := cli2.RunStopHub(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("return no error when there is none verbose mode", func(t *testing.T) {
		defer resetCLIVars()
		cli2.StopHubService = funcNilError()
		cli2.ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		cli2.Verbose = true

		err := cli2.RunStopHub(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("return error when there is error stoppig Hub service", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error while stopping Hub Service"
		cli2.StopHubService = funcErrorMessage(expectedStr)
		cli2.Verbose = false

		err := cli2.RunStopHub(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}
