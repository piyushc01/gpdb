package gpservice_cli_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gpdb/gpservice/gpservice-cli"
	"github.com/greenplum-db/gpdb/gpservice/hub"
	"github.com/greenplum-db/gpdb/gpservice/idl"
	"github.com/greenplum-db/gpdb/gpservice/idl/mock_idl"
)

func TestWaitAndRetryHubConnect(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("WaitAndRetryHubConnect returns success on success", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, nil
		}
		err := gpservice_cli.WaitAndRetryHubConnect()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("WaitAndRetryHubConnect returns failure upon failure to connect", func(t *testing.T) {
		defer resetCLIVars()
		expectedErr := "failed to connect to hub service. Check hub service log for details."
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedErr)
		}

		err := gpservice_cli.WaitAndRetryHubConnect()
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("got %v, want %v", err, expectedErr)
		}
	})
}

func TestStartAgentsAll(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("starts all agents without any error", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StartAgents(gomock.Any(), gomock.Any())
			return hubClient, nil
		}
		_, err := gpservice_cli.StartAgentsAll(gpservice_cli.Conf)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("start all agents fails on error connecting hub", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "error connecting hub"
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}

		_, err := gpservice_cli.StartAgentsAll(gpservice_cli.Conf)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
	t.Run("start all agent fails when error starting agents", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST: Agent Start ERROR"
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StartAgents(gomock.Any(), gomock.Any()).Return(nil, errors.New(expectedStr))
			return hubClient, nil
		}

		_, err := gpservice_cli.StartAgentsAll(gpservice_cli.Conf)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestRunStartService(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("Run services when there's no error", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.StartHubService = func(serviceName string) error {
			return nil
		}

		gpservice_cli.WaitAndRetryHubConnect = funcNilError()

		gpservice_cli.StartAgentsAll = func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, nil
		}

		err := gpservice_cli.RunStartService(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("Run services when there's error starting hub", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST ERROR Starting Hub service"
		gpservice_cli.StartHubService = func(serviceName string) error {
			return errors.New(expectedStr)
		}

		err := gpservice_cli.RunStartService(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
	t.Run("Run services when there's error connecting Hub", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST ERROR while connecting Hub service"
		gpservice_cli.StartHubService = func(serviceName string) error {
			return nil
		}

		gpservice_cli.WaitAndRetryHubConnect = funcErrorMessage(expectedStr)
		err := gpservice_cli.RunStartService(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
	t.Run("Run services when there's error starting agents", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST ERROR while starting agents"
		gpservice_cli.StartHubService = func(serviceName string) error {
			return nil
		}

		gpservice_cli.WaitAndRetryHubConnect = funcNilError()

		gpservice_cli.StartAgentsAll = func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}

		err := gpservice_cli.RunStartService(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestRunStartAgent(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("Run start agent starts agents when no failure", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.StartAgentsAll = func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, nil
		}

		err := gpservice_cli.RunStartAgent(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("Run start agent starts agents when starting agents fails", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error when starting agents"
		gpservice_cli.StartAgentsAll = func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}

		err := gpservice_cli.RunStartAgent(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestRunStartHub(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("Run Start Hub throws no error when there none", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.StartHubService = func(serviceName string) error {
			return nil
		}

		gpservice_cli.ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}

		gpservice_cli.WaitAndRetryHubConnect = funcNilError()

		err := gpservice_cli.RunStartHub(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("Run Start Hub throws error when start hub service fails", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error: Failed to start Hub service"
		gpservice_cli.StartHubService = func(serviceName string) error {
			return errors.New(expectedStr)
		}

		err := gpservice_cli.RunStartHub(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
	t.Run("Run Start Hub throws no error when there none in verbose mode", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "Test Error in ShowHubStatus"
		gpservice_cli.StartHubService = func(serviceName string) error {
			return nil
		}
		gpservice_cli.Verbose = true
		gpservice_cli.ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return false, errors.New(expectedStr)
		}
		gpservice_cli.WaitAndRetryHubConnect = funcNilError()

		err := gpservice_cli.RunStartHub(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}
