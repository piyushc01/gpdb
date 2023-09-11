package cli_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/cli"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/idl/mock_idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
	"github.com/greenplum-db/gpdb/gp/utils"
)

func TestPrintServicesStatus(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()

	t.Run("returns no error when there's none", func(t *testing.T) {

		defer resetShowHubStatus()
		mockShowHubStatus := func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		setShowHubStatus(mockShowHubStatus)

		defer resetShowAgentsStatus()
		mockShowAgentsStatus := func(conf *hub.Config, skipHeader bool) error {
			return nil
		}
		setShowAgentsStatus(mockShowAgentsStatus)

		err := cli.PrintServicesStatus()
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns an error when error printing Hub status", func(t *testing.T) {
		expectedStr := "TEST Error printing Hub status"
		defer resetShowHubStatus()
		mockShowHubStatus := func(conf *hub.Config, skipHeader bool) (bool, error) {
			return false, errors.New(expectedStr)
		}
		setShowHubStatus(mockShowHubStatus)
		err := cli.PrintServicesStatus()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
	t.Run("returns an error when error printing Agent status", func(t *testing.T) {
		expectedStr := "TEST Error printing Agent status"
		defer resetShowHubStatus()
		mockShowHubStatus := func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		setShowHubStatus(mockShowHubStatus)

		defer resetShowAgentsStatus()
		mockShowAgentsStatus := func(conf *hub.Config, skipHeader bool) error {
			return errors.New(expectedStr)
		}
		setShowAgentsStatus(mockShowAgentsStatus)

		err := cli.PrintServicesStatus()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
}

func TestRunServiceStatus(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()

	t.Run("returns no error when there is none", func(t *testing.T) {
		defer resetPrintServicesStatus()
		mockPrintServicesStatus := func() error {
			return nil
		}
		setPrintServicesStatus(mockPrintServicesStatus)

		err := cli.RunServiceStatus(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns error when print service status fails", func(t *testing.T) {
		expectedStr := "TEST Error printing service status"
		defer resetPrintServicesStatus()
		mockPrintServicesStatus := func() error {
			return errors.New(expectedStr)
		}
		setPrintServicesStatus(mockPrintServicesStatus)

		err := cli.RunServiceStatus(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
}

func TestShowAgentStatys(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("returns no error when there's none with  header", func(t *testing.T) {
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StatusAgents(gomock.Any(), gomock.Any()).Return(&idl.StatusAgentsReply{}, nil)
			return hubClient, nil
		}
		setConnectToHub(mockConnectToHub)

		err := cli.ShowAgentsStatus(cli.Conf, false)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns no error when there's none with no header", func(t *testing.T) {
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StatusAgents(gomock.Any(), gomock.Any()).Return(&idl.StatusAgentsReply{}, nil)
			return hubClient, nil
		}
		setConnectToHub(mockConnectToHub)

		err := cli.ShowAgentsStatus(cli.Conf, true)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns error when there error connecting Hub", func(t *testing.T) {
		expectedStr := "TEST Error connecting Hub"
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}
		setConnectToHub(mockConnectToHub)

		err := cli.ShowAgentsStatus(cli.Conf, true)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
}

func TestShowHubStatus(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()

	t.Run("returns no error when there is none", func(t *testing.T) {
		mockPlatform := &testutils.MockPlatform{Err: nil}
		mockPlatform.RetStatus = &idl.ServiceStatus{Status: "Running", Uptime: "10ms", Pid: uint32(1234)}
		cli.Platform = mockPlatform
		defer func() { cli.Platform = utils.GetPlatform() }()

		_, err := cli.ShowHubStatus(cli.Conf, true)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns error when error getting service status", func(t *testing.T) {
		expectedStr := "TEST Error getting service status"
		mockPlatform := &testutils.MockPlatform{Err: errors.New(expectedStr), ServiceStatusMessage: ""}
		mockPlatform.RetStatus = &idl.ServiceStatus{Status: "Running", Uptime: "10ms", Pid: uint32(1234)}
		cli.Platform = mockPlatform
		defer func() { cli.Platform = utils.GetPlatform() }()

		_, err := cli.ShowHubStatus(cli.Conf, true)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
}

func TestRunStatusAgent(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()

	t.Run("return no error when there is none", func(t *testing.T) {
		defer resetShowAgentsStatus()
		mockShowAgentsStatus := func(conf *hub.Config, skipHeader bool) error {
			return nil
		}
		setShowAgentsStatus(mockShowAgentsStatus)

		err := cli.RunStatusAgent(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("return error when there is error getting agent status", func(t *testing.T) {
		expectedStr := "TEST Error getting agent status"
		defer resetShowAgentsStatus()
		mockShowAgentsStatus := func(conf *hub.Config, skipHeader bool) error {
			return errors.New(expectedStr)
		}
		setShowAgentsStatus(mockShowAgentsStatus)

		err := cli.RunStatusAgent(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
}

func TestRunStatusHub(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()

	t.Run("return no error when there is none", func(t *testing.T) {

		defer resetShowHubStatus()
		mockShowHubStatus := func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		setShowHubStatus(mockShowHubStatus)

		err := cli.RunStatusHub(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("return error when there is error getting agent status", func(t *testing.T) {
		expectedStr := "TEST Error getting agent status"
		defer resetShowHubStatus()
		mockShowHubStatus := func(conf *hub.Config, skipHeader bool) (bool, error) {
			return false, errors.New(expectedStr)
		}
		setShowHubStatus(mockShowHubStatus)

		err := cli.RunStatusHub(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
}
