package cli_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/greenplum-db/gpdb/gp/cli"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/idl/mock_idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
)

func TestStopAgentService(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("StopAgentService stops the agent service when theres no error", func(t *testing.T) {

		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StopAgents(gomock.Any(), gomock.Any())
			return hubClient, nil
		}
		setConnectToHub(mockConnectToHub)

		err := cli.StopAgentService()
		if err != nil {
			t.Fatalf("No error expected. Got and error:%s", err.Error())
		}
	})
	t.Run("StopAgentService returns an error when theres error connecting hub", func(t *testing.T) {
		expectedStr := "TEST Error connecting Hub"
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}
		setConnectToHub(mockConnectToHub)

		err := cli.StopAgentService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Got: %q Want:\"%s\" ", err.Error(), expectedStr)
		}
	})
	t.Run("StopAgentService returns error when theres error stopping agents", func(t *testing.T) {
		expectedStr := "TEST Error stopping agents"
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StopAgents(gomock.Any(), gomock.Any()).Return(nil, errors.New(expectedStr))
			return hubClient, nil
		}
		setConnectToHub(mockConnectToHub)
		err := cli.StopAgentService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Got: %q Want:\"%s\" ", err.Error(), expectedStr)
		}
	})
}

func TestStopHubService(t *testing.T) {
	testhelper.SetupTestLogger()
	cli.Conf = testutils.InitializeTestEnv()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Stops hub when theres no error", func(t *testing.T) {
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().Stop(gomock.Any(), gomock.Any())
			return hubClient, nil
		}
		setConnectToHub(mockConnectToHub)

		err := cli.StopHubService()
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})

	t.Run("Stops hub when theres error connecting hub service", func(t *testing.T) {
		expectedStr := "TEST Error connecting Hub"
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedStr)
		}
		setConnectToHub(mockConnectToHub)

		err := cli.StopHubService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Got: %q Want:\"%s\" ", err.Error(), expectedStr)
		}
	})

	t.Run("Stop returns error when there's error stopping Hub", func(t *testing.T) {
		expectedStr := "TEST Error stopping Hub"
		defer resetConnectToHub()
		mockConnectToHub := func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(nil, errors.New(expectedStr))
			return hubClient, nil
		}
		setConnectToHub(mockConnectToHub)

		err := cli.StopHubService()
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Got: %q Want:\"%s\" ", err.Error(), expectedStr)
		}
	})
}

func TestRunStopServices(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("Returns no error when there is none", func(t *testing.T) {
		defer resetStopAgentService()
		mockStopAgentService := func() error {
			return nil
		}
		setStopAgentService(mockStopAgentService)

		defer func() { resetStopHubService() }()
		mockStopHubService := func() error {
			return nil
		}
		setStopHubService(mockStopHubService)

		err := cli.RunStopServices(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})

	t.Run("Returns  error when error stopping Agent service", func(t *testing.T) {
		expectedStr := "TEST Error stopping Agent service"
		defer resetStopAgentService()
		mockStopAgentService := func() error {
			return errors.New(expectedStr)
		}
		setStopAgentService(mockStopAgentService)

		err := cli.RunStopServices(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Got: %q Want:\"%s\" ", err.Error(), expectedStr)
		}
	})

	t.Run("Returns error when error stopping hub service", func(t *testing.T) {
		expectedStr := "Error stopping Hub service"
		defer resetStopAgentService()
		mockStopAgentService := func() error {
			return nil
		}
		setStopAgentService(mockStopAgentService)

		defer resetStopHubService()
		mockStopHubService := func() error {
			return errors.New(expectedStr)
		}
		setStopHubService(mockStopHubService)

		err := cli.RunStopServices(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Got: %q Want:\"%s\" ", err.Error(), expectedStr)
		}
	})
}

func TestRunStopAgents(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("returns no error where there none no verbose", func(t *testing.T) {

		defer resetStopAgentService()
		mockStopAgentService := func() error {
			return nil
		}
		setStopAgentService(mockStopAgentService)
		cli.Verbose = false

		err := cli.RunStopAgents(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}

	})
	t.Run("returns no error where there none in verbose", func(t *testing.T) {
		defer resetStopAgentService()
		mockStopAgentService := func() error {
			return nil
		}
		setStopAgentService(mockStopAgentService)

		defer resetShowAgentsStatus()
		mockShowAgentsStatus := func(conf *hub.Config, skipHeader bool) error {
			return nil
		}
		setShowAgentsStatus(mockShowAgentsStatus)
		cli.Verbose = true

		err := cli.RunStopAgents(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns error where there is error stopping agent service", func(t *testing.T) {
		expectedStr := "TEST Error stopping agents"
		defer resetStopAgentService()
		mockStopAgentService := func() error {
			return errors.New(expectedStr)
		}
		setStopAgentService(mockStopAgentService)

		err := cli.RunStopAgents(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Got: %q Want:\"%s\" ", err.Error(), expectedStr)
		}
	})
}

func TestRunStopHub(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("return no error when there is none non verbose mode", func(t *testing.T) {
		defer resetStopHubService()
		mockStopHubService := func() error {
			return nil
		}
		setStopHubService(mockStopHubService)
		cli.Verbose = false

		err := cli.RunStopHub(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error: %s", err.Error())
		}
	})
	t.Run("return no error when there is none verbose mode", func(t *testing.T) {

		defer resetStopHubService()
		mockStopHubService := func() error {
			return nil
		}
		setStopHubService(mockStopHubService)

		defer resetShowHubStatus()
		mockShowHubStatus := func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		setShowHubStatus(mockShowHubStatus)
		cli.Verbose = true

		err := cli.RunStopHub(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error: %s", err.Error())
		}
	})
	t.Run("return error when there is error stoppig Hub service", func(t *testing.T) {
		expectedStr := "TEST Error while stopping Hub Service"
		defer resetStopHubService()
		mockStopHubService := func() error {
			return errors.New(expectedStr)
		}
		setStopHubService(mockStopHubService)
		cli.Verbose = false

		err := cli.RunStopHub(nil, nil)
		if !strings.Contains(err.Error(), expectedStr) {
			t.Fatalf("Got: %q Want:\"%s\" ", err.Error(), expectedStr)
		}
	})
}

func setStopAgentService(customFn func() error) {
	cli.StopAgentService = customFn
}
func resetStopAgentService() {
	cli.StopAgentService = cli.StopAgentServiceFn
}

func setStopHubService(customFn func() error) {
	cli.StopHubService = customFn
}
func resetStopHubService() {
	cli.StopHubService = cli.StopHubServiceFn
}
