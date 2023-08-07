package cli

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/idl/mock_idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
	"testing"
)

func TestWaitAndRetryHubConnect(t *testing.T) {
	testhelper.SetupTestLogger()

	t.Run("WaitAndRetryHubConnect returns success on success", func(t *testing.T) {
		origConnectHub := connectToHub
		defer func() { connectToHub = origConnectHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, nil
		}
		err := WaitAndRetryHubConnect()
		if err != nil {
			t.Fatalf("Excted no error, received error:%s", err.Error())
		}
	})
	t.Run("WaitAndRetryHubConnect returns failure upon failure to connect", func(t *testing.T) {
		origConnectHub := connectToHub
		expectedErr := "Hub service started but failed to connect. Bailing out."
		defer func() { connectToHub = origConnectHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New(expectedErr)
		}
		err := WaitAndRetryHubConnect()
		if err == nil {
			t.Fatalf("Excted an error, received no error")
		}
	})
}

func TestStartAgentsAll(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("starts all agents without any error", func(t *testing.T) {
		origConnectHub := connectToHub
		defer func() { connectToHub = origConnectHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StartAgents(gomock.Any(), gomock.Any())
			return hubClient, nil
		}
		_, err := startAgentsAll(conf)
		if err != nil {
			t.Fatalf("Expected no error, got error:%s", err.Error())
		}

	})
	t.Run("start all agents fails on error connecting hub", func(t *testing.T) {
		origConnectHub := connectToHub
		defer func() { connectToHub = origConnectHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, errors.New("Error connecting hub")
		}
		_, err := startAgentsAll(conf)
		if err == nil {
			t.Fatalf("Expected error, got no error:")
		}
	})
	t.Run("start all agent fails when error starting agents", func(t *testing.T) {
		origConnectHub := connectToHub
		defer func() { connectToHub = origConnectHub }()
		connectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().StartAgents(gomock.Any(), gomock.Any()).Return(nil, errors.New("TEST: Agent Start ERROR"))
			return hubClient, nil
		}
		_, err := startAgentsAll(conf)
		if err == nil {
			t.Fatalf("Expected error, got no error:")
		}
	})
}

func TestRunStartService(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()

	t.Run("Run services when there's no error", func(t *testing.T) {
		origStartHubService := startHubService
		defer func() { startHubService = origStartHubService }()
		startHubService = func(serviceName string) error {
			return nil
		}
		origWaitAndRetryHubConnect := WaitAndRetryHubConnect
		defer func() { WaitAndRetryHubConnect = origWaitAndRetryHubConnect }()
		WaitAndRetryHubConnect = func() error {
			return nil
		}
		origStartAgentsAll := startAgentsAll
		defer func() { startAgentsAll = origStartAgentsAll }()
		startAgentsAll = func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, nil
		}
		err := RunStartService(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, got error:%s", err.Error())
		}
	})
	t.Run("Run services when there's error starting hub", func(t *testing.T) {
		origStartHubService := startHubService
		defer func() { startHubService = origStartHubService }()
		startHubService = func(serviceName string) error {
			return errors.New("TEST ERROR Starting Hub service")
		}
		err := RunStartService(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error, got no error")
		}
	})
	t.Run("Run services when there's error connecting Hub", func(t *testing.T) {
		origStartHubService := startHubService
		defer func() { startHubService = origStartHubService }()
		startHubService = func(serviceName string) error {
			return nil
		}
		origWaitAndRetryHubConnect := WaitAndRetryHubConnect
		defer func() { WaitAndRetryHubConnect = origWaitAndRetryHubConnect }()
		WaitAndRetryHubConnect = func() error {
			return errors.New("TEST ERROR while connecting Hub service")
		}
		err := RunStartService(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error, got no error")
		}
	})
	t.Run("Run services when there's error starting agents", func(t *testing.T) {
		origStartHubService := startHubService
		defer func() { startHubService = origStartHubService }()
		startHubService = func(serviceName string) error {
			return nil
		}
		origWaitAndRetryHubConnect := WaitAndRetryHubConnect
		defer func() { WaitAndRetryHubConnect = origWaitAndRetryHubConnect }()
		WaitAndRetryHubConnect = func() error {
			return nil
		}
		origStartAgentsAll := startAgentsAll
		defer func() { startAgentsAll = origStartAgentsAll }()
		startAgentsAll = func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, errors.New("TEST ERROR while starting agents")
		}
		err := RunStartService(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error, got no error")
		}
	})
}

func TestRunStartAgent(t *testing.T) {
	testhelper.SetupTestLogger()
	t.Run("Run start agent starts agents when no failure", func(t *testing.T) {
		origStartAgentsAll := startAgentsAll
		defer func() { startAgentsAll = origStartAgentsAll }()
		startAgentsAll = func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, nil
		}
		err := RunStartAgent(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error. Got error:%s", err.Error())
		}
	})
	t.Run("Run start agent starts agents when starting agents fails", func(t *testing.T) {
		origStartAgentsAll := startAgentsAll
		defer func() { startAgentsAll = origStartAgentsAll }()
		startAgentsAll = func(hubConfig *hub.Config) (idl.HubClient, error) {
			return nil, errors.New("TEST Error when starting agents")
		}
		err := RunStartAgent(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error. Got no error")
		}
	})
}

func TestRunStartHub(t *testing.T) {
	testhelper.SetupTestLogger()
	conf = testutils.InitializeTestEnv()

	t.Run("Run Start Hub throws no error when there none", func(t *testing.T) {
		origStartHubService := startHubService
		defer func() { startHubService = origStartHubService }()
		startHubService = func(serviceName string) error {
			return nil
		}
		origShowHubStatus := ShowHubStatus
		defer func() { ShowHubStatus = origShowHubStatus }()
		ShowHubStatus = func(conf *hub.Config, skipHeader bool) error {
			return nil
		}
		err := RunStartHub(nil, nil)
		if err != nil {
			t.Fatalf("Expected no error, received error:%s", err.Error())
		}
	})
	t.Run("Run Start Hub throws error when start hub service fails", func(t *testing.T) {
		origStartHubService := startHubService
		defer func() { startHubService = origStartHubService }()
		startHubService = func(serviceName string) error {
			return errors.New("TEST Error: Failed to start Hub service")
		}
		err := RunStartHub(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error, received no error:")
		}
	})
	t.Run("Run Start Hub throws no error when there none", func(t *testing.T) {
		origStartHubService := startHubService
		defer func() { startHubService = origStartHubService }()
		startHubService = func(serviceName string) error {
			return nil
		}
		origShowHubStatus := ShowHubStatus
		defer func() { ShowHubStatus = origShowHubStatus }()
		verbose = true
		ShowHubStatus = func(conf *hub.Config, skipHeader bool) error {
			return errors.New("Test Error in ShowHubStatus ")
		}
		err := RunStartHub(nil, nil)
		if err == nil {
			t.Fatalf("Expected an error, received no error")
		}
	})
}
