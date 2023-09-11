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
)

func TestWaitAndRetryHubConnect(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("WaitAndRetryHubConnect returns success on success", func(t *testing.T) {
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			return nil, nil
		}
		setConnectToHub(mockConnectToHub)
		err := cli.WaitAndRetryHubConnect()
		if err != nil {
			t.Fatalf("Excted no error, received error:%s", err.Error())
		}
	})
	t.Run("WaitAndRetryHubConnect returns failure upon failure to connect", func(t *testing.T) {
		expectedErr := "Hub service started but failed to connect"
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedErr)
		}
		setConnectToHub(mockConnectToHub)

		err := cli.WaitAndRetryHubConnect()
		if !strings.Contains(err.Error(), expectedErr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedErr)
		}
	})
}

func TestStartAgentsAll(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("starts all agents without any error", func(t *testing.T) {
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StartAgents(gomock.Any(), gomock.Any())
			return hubClient, nil
		}
		setConnectToHub(mockConnectToHub)
		_, err := cli.StartAgentsAll(cli.Conf)
		if err != nil {
			t.Fatalf("Expected no error, got error:%s", err.Error())
		}
	})
	t.Run("start all agents fails on error connecting hub", func(t *testing.T) {
		defer resetConnectToHub()
		expectedStr := "error connecting hub"
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}
		setConnectToHub(mockConnectToHub)

		_, err := cli.StartAgentsAll(cli.Conf)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
	t.Run("start all agent fails when error starting agents", func(t *testing.T) {
		expectedStr := "TEST: Agent Start ERROR"
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StartAgents(gomock.Any(), gomock.Any()).Return(nil, errors.New(expectedStr))
			return hubClient, nil
		}
		setConnectToHub(mockConnectToHub)

		_, err := cli.StartAgentsAll(cli.Conf)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
}

func TestRunStartService(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()

	t.Run("Run services when there's no error", func(t *testing.T) {
		defer resetStartHubService()
		mockStartHubService := func(serviceName string) error {
			return nil
		}
		setStartHubService(mockStartHubService)

		defer resetWaitAndRetryHubConnect()
		mockWaitAndRetryHubConnect := func() error {
			return nil
		}
		setWaitAndRetryHubConnect(mockWaitAndRetryHubConnect)

		defer resetStartAgentsAll()
		mockStartAgentsAll := func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, nil
		}
		setStartAgentsAll(mockStartAgentsAll)

		err := cli.RunStartService(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got error:%s", err.Error())
		}
	})
	t.Run("Run services when there's error starting hub", func(t *testing.T) {
		expectedStr := "TEST ERROR Starting Hub service"
		defer resetStartHubService()
		mockStartHubService := func(serviceName string) error {
			return errors.New(expectedStr)
		}
		setStartHubService(mockStartHubService)
		err := cli.RunStartService(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Want:\"%s\" But got: %q", expectedStr, err.Error())
		}
	})
	t.Run("Run services when there's error connecting Hub", func(t *testing.T) {
		expectedStr := "TEST ERROR while connecting Hub service"
		defer resetStartHubService()
		mockStartHubService := func(serviceName string) error {
			return nil
		}
		setStartHubService(mockStartHubService)

		defer resetWaitAndRetryHubConnect()
		mockWaitAndRetryHubConnect := func() error {
			return errors.New(expectedStr)
		}
		setWaitAndRetryHubConnect(mockWaitAndRetryHubConnect)
		err := cli.RunStartService(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Want:\"%s\" But got: %q", expectedStr, err.Error())
		}
	})
	t.Run("Run services when there's error starting agents", func(t *testing.T) {
		expectedStr := "TEST ERROR while starting agents"
		defer resetStartHubService()
		mockStartHubService := func(serviceName string) error {
			return nil
		}
		setStartHubService(mockStartHubService)

		defer resetWaitAndRetryHubConnect()
		mockWaitAndRetryHubConnect := func() error {
			return nil
		}
		setWaitAndRetryHubConnect(mockWaitAndRetryHubConnect)

		defer resetStartAgentsAll()
		mockStartAgentsAll := func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}
		setStartAgentsAll(mockStartAgentsAll)

		err := cli.RunStartService(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Want:\"%s\" But got: %q", expectedStr, err.Error())
		}
	})
}

func TestRunStartAgent(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("Run start agent starts agents when no failure", func(t *testing.T) {
		defer resetStartAgentsAll()
		mockStartAgentsAll := func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, nil
		}
		setStartAgentsAll(mockStartAgentsAll)

		err := cli.RunStartAgent(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got error:%s", err.Error())
		}
	})
	t.Run("Run start agent starts agents when starting agents fails", func(t *testing.T) {
		expectedStr := "TEST Error when starting agents"
		defer resetStartAgentsAll()
		mockStartAgentsAll := func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}
		setStartAgentsAll(mockStartAgentsAll)

		err := cli.RunStartAgent(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
}

func TestRunStartHub(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()

	t.Run("Run Start Hub throws no error when there none", func(t *testing.T) {
		defer resetStartHubService()
		mockStartHubService := func(serviceName string) error {
			return nil
		}
		setStartHubService(mockStartHubService)

		defer func() { resetShowHubStatus() }()
		mockShowHubStatus := func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		setShowHubStatus(mockShowHubStatus)

		err := cli.RunStartHub(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, received error:%s", err.Error())
		}
	})
	t.Run("Run Start Hub throws error when start hub service fails", func(t *testing.T) {
		expectedStr := "TEST Error: Failed to start Hub service"
		defer resetStartHubService()
		mockStartHubService := func(serviceName string) error {
			return errors.New(expectedStr)
		}
		setStartHubService(mockStartHubService)

		err := cli.RunStartHub(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
	t.Run("Run Start Hub throws no error when there none", func(t *testing.T) {
		expectedStr := "Test Error in ShowHubStatus"
		defer resetStartHubService()
		mockStartHubService := func(serviceName string) error {
			return nil
		}
		setStartHubService(mockStartHubService)

		defer resetShowHubStatus()
		cli.Verbose = true
		mockShowHubStatus := func(conf *hub.Config, skipHeader bool) (bool, error) {
			return false, errors.New(expectedStr)
		}
		setShowHubStatus(mockShowHubStatus)

		err := cli.RunStartHub(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("got: %q want:\"%s\"", err.Error(), expectedStr)
		}
	})
}

func setConnectToHub(customFunc func(conf *hub.Config) (idl.HubClient, error)) {
	cli.ConnectToHub = customFunc
}

func resetConnectToHub() {
	cli.ConnectToHub = cli.ConnectToHubFn
}

func setStartHubService(customFn func(serviceName string) error) {
	cli.StartHubService = customFn
}
func resetStartHubService() {
	cli.StartHubService = cli.StartHubServiceFn
}

func setWaitAndRetryHubConnect(customFn func() error) {
	cli.WaitAndRetryHubConnect = customFn
}

func resetWaitAndRetryHubConnect() {
	cli.WaitAndRetryHubConnect = cli.WaitAndRetryHubConnectFn
}

func setShowHubStatus(customFn func(conf *hub.Config, skipHeader bool) (bool, error)) {
	cli.ShowHubStatus = customFn
}
func resetShowHubStatus() {
	cli.ShowHubStatus = cli.ShowHubStatusFn
}

func setStartAgentsAll(customFn func(hubConfig *hub.Config) (idl.HubClient, error)) {
	cli.StartAgentsAll = customFn
}

func resetStartAgentsAll() {
	cli.StartAgentsAll = cli.StartAgentsAllFn
}

func setShowAgentsStatus(customFn func(conf *hub.Config, skipHeader bool) error) {
	cli.ShowAgentsStatus = customFn
}

func resetShowAgentsStatus() {
	cli.ShowAgentsStatus = cli.ShowAgentsStatusFn

}

func setPrintServicesStatus(customFn func() error) {
	cli.PrintServicesStatus = customFn
}

func resetPrintServicesStatus() {
	cli.PrintServicesStatus = cli.PrintServicesStatusFn

}
