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
	"github.com/greenplum-db/gpdb/gpservice/testutils"
	"github.com/greenplum-db/gpdb/gpservice/utils"
)

func TestPrintServicesStatus(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("returns no error when there's none", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		gpservice_cli.ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return nil
		}

		err := gpservice_cli.PrintServicesStatus()
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("returns an error when error printing Hub status", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error printing Hub status"
		gpservice_cli.ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return false, errors.New(expectedStr)
		}
		err := gpservice_cli.PrintServicesStatus()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
	t.Run("returns an error when error printing Agent status", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error printing Agent status"
		gpservice_cli.ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		gpservice_cli.ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return errors.New(expectedStr)
		}

		err := gpservice_cli.PrintServicesStatus()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestRunServiceStatus(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("returns no error when there is none", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.PrintServicesStatus = funcNilError()

		err := gpservice_cli.RunServiceStatus(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("returns error when print service status fails", func(t *testing.T) {
		expectedStr := "TEST Error printing service status"
		defer resetCLIVars()
		gpservice_cli.PrintServicesStatus = func() error {
			return errors.New(expectedStr)
		}

		err := gpservice_cli.RunServiceStatus(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestShowAgentStatus(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("returns no error when there's none with  header", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StatusAgents(gomock.Any(), gomock.Any()).Return(&idl.StatusAgentsReply{}, nil)
			return hubClient, nil
		}

		err := gpservice_cli.ShowAgentsStatus(gpservice_cli.Conf, false)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("returns no error when there's none with no header", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StatusAgents(gomock.Any(), gomock.Any()).Return(&idl.StatusAgentsReply{}, nil)
			return hubClient, nil
		}

		err := gpservice_cli.ShowAgentsStatus(gpservice_cli.Conf, true)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("returns error when there error connecting Hub", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error connecting Hub"
		gpservice_cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}

		err := gpservice_cli.ShowAgentsStatus(gpservice_cli.Conf, true)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestShowHubStatus(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("returns no error when there is none", func(t *testing.T) {
		mockPlatform := &testutils.MockPlatform{Err: nil}
		mockPlatform.RetStatus = &idl.ServiceStatus{Status: "Running", Uptime: "10ms", Pid: uint32(1234)}
		gpservice_cli.Platform = mockPlatform
		defer func() { gpservice_cli.Platform = utils.GetPlatform() }()

		_, err := gpservice_cli.ShowHubStatus(gpservice_cli.Conf, true)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("returns error when error getting service status", func(t *testing.T) {
		expectedStr := "TEST Error getting service status"
		mockPlatform := &testutils.MockPlatform{Err: errors.New(expectedStr), ServiceStatusMessage: ""}
		mockPlatform.RetStatus = &idl.ServiceStatus{Status: "Running", Uptime: "10ms", Pid: uint32(1234)}
		gpservice_cli.Platform = mockPlatform
		defer func() { gpservice_cli.Platform = utils.GetPlatform() }()

		_, err := gpservice_cli.ShowHubStatus(gpservice_cli.Conf, true)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestRunStatusAgent(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("return no error when there is none", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return nil
		}

		err := gpservice_cli.RunStatusAgent(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("return error when there is error getting agent status", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error getting agent status"
		gpservice_cli.ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return errors.New(expectedStr)
		}

		err := gpservice_cli.RunStatusAgent(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}

func TestRunStatusHub(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("return no error when there is none", func(t *testing.T) {
		defer resetCLIVars()
		gpservice_cli.ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}

		err := gpservice_cli.RunStatusHub(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
	t.Run("return error when there is error getting agent status", func(t *testing.T) {
		defer resetCLIVars()
		expectedStr := "TEST Error getting agent status"
		gpservice_cli.ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return false, errors.New(expectedStr)
		}

		err := gpservice_cli.RunStatusHub(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got %v, want %v", err, expectedStr)
		}
	})
}
