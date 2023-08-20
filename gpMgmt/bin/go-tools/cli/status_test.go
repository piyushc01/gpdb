package cli

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/idl/mock_idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
	"github.com/greenplum-db/gpdb/gp/utils"
)

func TestPrintServicesStatus(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()

	t.Run("returns no error when there's none", func(t *testing.T) {
		origShowHubStatus := ShowHubStatus
		defer func() { ShowHubStatus = origShowHubStatus }()
		ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		origShowAgentsStatus := ShowAgentsStatus
		defer func() { ShowAgentsStatus = origShowAgentsStatus }()
		ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return nil
		}
		err := PrintServicesStatus()
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns an error when error printing Hub status", func(t *testing.T) {
		origShowHubStatus := ShowHubStatus
		defer func() { ShowHubStatus = origShowHubStatus }()
		ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return false, errors.New("TEST Error printing Hub status")
		}
		err := PrintServicesStatus()
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})
	t.Run("returns an error when error printing Agent status", func(t *testing.T) {
		origShowHubStatus := ShowHubStatus
		defer func() { ShowHubStatus = origShowHubStatus }()
		ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		origShowAgentsStatus := ShowAgentsStatus
		defer func() { ShowAgentsStatus = origShowAgentsStatus }()
		ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return errors.New("TEST Error printing Agent status")
		}
		err := PrintServicesStatus()
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})
}

func TestRunServiceStatus(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()

	t.Run("returns no error when there is none", func(t *testing.T) {
		origPrintServicesStatus := PrintServicesStatus
		defer func() { PrintServicesStatus = origPrintServicesStatus }()
		PrintServicesStatus = func() error {
			return nil
		}
		err := RunServiceStatus(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns error when print service status fails", func(t *testing.T) {
		origPrintServicesStatus := PrintServicesStatus
		defer func() { PrintServicesStatus = origPrintServicesStatus }()
		PrintServicesStatus = func() error {
			return errors.New("TEST Error printing service status")
		}
		err := RunServiceStatus(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})
}

func TestShowAgentStatys(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("returns no error when there's none with  header", func(t *testing.T) {
		origConnectHub := connectToHub
		defer func() { connectToHub = origConnectHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StatusAgents(gomock.Any(), gomock.Any()).Return(&idl.StatusAgentsReply{}, nil)
			return hubClient, nil
		}

		err := ShowAgentsStatus(conf, false)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns no error when there's none with no header", func(t *testing.T) {
		origConnectHub := connectToHub
		defer func() { connectToHub = origConnectHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StatusAgents(gomock.Any(), gomock.Any()).Return(&idl.StatusAgentsReply{}, nil)
			return hubClient, nil
		}

		err := ShowAgentsStatus(conf, true)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns error when there error connecting Hub", func(t *testing.T) {
		origConnectHub := connectToHub
		defer func() { connectToHub = origConnectHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New("TEST Error connecting Hub")
		}

		err := ShowAgentsStatus(conf, true)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})
}

func TestShowHubStatus(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()

	t.Run("returns no error when there is none", func(t *testing.T) {
		os := &testutils.MockPlatform{}
		os.RetStatus = idl.ServiceStatus{Status: "Running", Uptime: "10ms", Pid: uint32(1234)}
		os.Err = nil
		platform = os
		defer func() { platform = utils.GetPlatform() }()

		_, err := ShowHubStatus(conf, true)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns error when error getting service status", func(t *testing.T) {
		os := &testutils.MockPlatform{}
		os.RetStatus = idl.ServiceStatus{Status: "Running", Uptime: "10ms", Pid: uint32(1234)}
		os.Err = errors.New("TEST Error getting service status")
		os.ServiceStatusMessage = ""
		platform = os
		defer func() { platform = utils.GetPlatform() }()

		_, err := ShowHubStatus(conf, true)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})

}

func TestRunStatusAgent(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()

	t.Run("return no error when there is none", func(t *testing.T) {
		origShowAgentsStatus := ShowAgentsStatus
		defer func() { ShowAgentsStatus = origShowAgentsStatus }()
		ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return nil
		}
		err := RunStatusAgent(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("return error when there is error getting agent status", func(t *testing.T) {
		origShowAgentsStatus := ShowAgentsStatus
		defer func() { ShowAgentsStatus = origShowAgentsStatus }()
		ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return errors.New("TEST Error getting agent status")
		}
		err := RunStatusAgent(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})
}

func TestRunStatusHub(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()

	t.Run("return no error when there is none", func(t *testing.T) {
		origShowHubStatus := ShowHubStatus
		defer func() { ShowHubStatus = origShowHubStatus }()
		ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		err := RunStatusHub(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("return error when there is error getting agent status", func(t *testing.T) {
		origShowHubStatus := ShowHubStatus
		defer func() { ShowHubStatus = origShowHubStatus }()
		ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return false, errors.New("TEST Error getting agent status")
		}
		err := RunStatusHub(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})
}
