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
)

func TestStopAgentService(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("StopAgentService stops the agent service when theres no error", func(t *testing.T) {
		origConnectToHub := connectToHub
		defer func() { connectToHub = origConnectToHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StopAgents(gomock.Any(), gomock.Any())
			return hubClient, nil
		}
		err := StopAgentService()
		if err != nil {
			t.Fatalf("No error expected. Got and error:%s", err.Error())
		}
	})
	t.Run("StopAgentService returns an error when theres error connecting hub", func(t *testing.T) {
		origConnectToHub := connectToHub
		defer func() { connectToHub = origConnectToHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New("TEST Error connecting Hub")
		}
		err := StopAgentService()
		if err == nil {
			t.Fatalf("Expected an error. Got no error.")
		}
	})
	t.Run("StopAgentService returns error when theres error stopping agents", func(t *testing.T) {
		origConnectToHub := connectToHub
		defer func() { connectToHub = origConnectToHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StopAgents(gomock.Any(), gomock.Any()).Return(nil, errors.New("TEST Error stopping agents"))
			return hubClient, nil
		}
		err := StopAgentService()
		if err == nil {
			t.Fatalf("Expected an error. Got no error.")
		}
	})
}

func TestStopHubService(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Stops hub when theres no error", func(t *testing.T) {
		origConnectToHub := connectToHub
		defer func() { connectToHub = origConnectToHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().Stop(gomock.Any(), gomock.Any())
			return hubClient, nil
		}
		err := StopHubService()
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})

	t.Run("Stops hub when theres error connecting hub service", func(t *testing.T) {
		origConnectToHub := connectToHub
		defer func() { connectToHub = origConnectToHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New("TEST Error connecting Hub")
		}
		err := StopHubService()
		if err == nil {
			t.Fatalf("Expected an error. Got no error:")
		}
	})

	t.Run("Stop returns error when there's error stopping Hub", func(t *testing.T) {
		origConnectToHub := connectToHub
		defer func() { connectToHub = origConnectToHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().Stop(gomock.Any(), gomock.Any()).Return(nil, errors.New("TEST Error stopping Hub"))
			return hubClient, nil
		}
		err := StopHubService()
		if err == nil {
			t.Fatalf("Expected an error. Got no error:")
		}
	})
}

func TestRunStopServices(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("Returns no error when there is none", func(t *testing.T) {
		origStopAgentService := StopAgentService
		defer func() { StopAgentService = origStopAgentService }()
		StopAgentService = func() error {
			return nil
		}
		origStopHubService := StopHubService
		defer func() { StopHubService = origStopHubService }()
		StopHubService = func() error {
			return nil
		}

		err := RunStopServices(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})

	t.Run("Returns  error whenerror stopping Agent service", func(t *testing.T) {
		origStopAgentService := StopAgentService
		defer func() { StopAgentService = origStopAgentService }()
		StopAgentService = func() error {
			return errors.New("TEST Error stopping Agent service")
		}

		err := RunStopServices(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})

	t.Run("Returns error when error stopping hub service", func(t *testing.T) {
		origStopAgentService := StopAgentService
		defer func() { StopAgentService = origStopAgentService }()
		StopAgentService = func() error {
			return nil
		}
		origStopHubService := StopHubService
		defer func() { StopHubService = origStopHubService }()
		StopHubService = func() error {
			return errors.New("Error stopping Hub service")
		}

		err := RunStopServices(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})

}

func TestRunStopAgents(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("returns no error where there none no verbose", func(t *testing.T) {
		origStopAgentService := StopAgentService
		defer func() { StopAgentService = origStopAgentService }()
		StopAgentService = func() error {
			return nil
		}
		verbose = false

		err := RunStopAgents(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}

	})
	t.Run("returns no error where there none in verbose", func(t *testing.T) {
		origStopAgentService := StopAgentService
		defer func() { StopAgentService = origStopAgentService }()
		StopAgentService = func() error {
			return nil
		}
		origShowAgentsStatus := ShowAgentsStatus
		defer func() { ShowAgentsStatus = origShowAgentsStatus }()
		ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
			return nil
		}
		verbose = true

		err := RunStopAgents(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error:%s", err.Error())
		}
	})
	t.Run("returns error where there is error stopping agent service", func(t *testing.T) {
		origStopAgentService := StopAgentService
		defer func() { StopAgentService = origStopAgentService }()
		StopAgentService = func() error {
			return errors.New("TEST Error stopping agents")
		}

		err := RunStopAgents(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}

	})

}

func TestRunStopHub(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("return no error when there is none non verbose mode", func(t *testing.T) {
		origStopHubService := StopHubService
		defer func() { StopHubService = origStopHubService }()
		StopHubService = func() error {
			return nil
		}
		verbose = false

		err := RunStopHub(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error: %s", err.Error())
		}
	})
	t.Run("return no error when there is none verbose mode", func(t *testing.T) {
		origStopHubService := StopHubService
		defer func() { StopHubService = origStopHubService }()
		StopHubService = func() error {
			return nil
		}
		origShowHubStatus := ShowHubStatus
		defer func() { ShowHubStatus = origShowHubStatus }()
		ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
			return true, nil
		}
		verbose = true

		err := RunStopHub(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got an error: %s", err.Error())
		}
	})
	t.Run("return error when there is error stoppig Hub service", func(t *testing.T) {
		origStopHubService := StopHubService
		defer func() { StopHubService = origStopHubService }()
		StopHubService = func() error {
			return errors.New("TEST Error while stopping Hub Service")
		}
		verbose = false

		err := RunStopHub(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})
}
