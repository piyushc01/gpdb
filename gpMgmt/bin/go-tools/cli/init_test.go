package cli_test

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/greenplum-db/gpdb/gp/idl/mock_idl"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/spf13/cobra"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/cli"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
)

func TestRunInitClusterCmd(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	t.Run("returns error when length of args less than 1", func(t *testing.T) {
		testStr := "please provide config file for cluster initialization"
		cmd := cobra.Command{}
		err := cli.RunInitClusterCmd(&cmd, nil)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got:%v, expected:%s", err, testStr)
		}
	})
	t.Run("returns error when length of args greater than 1", func(t *testing.T) {
		testStr := "more arguments than expected"
		cmd := cobra.Command{}
		args := []string{"/tmp/1", "/tmp/2"}
		err := cli.RunInitClusterCmd(&cmd, args)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got:%v, expected:%s", err, testStr)
		}
	})
	t.Run("returns error cluster creation fails", func(t *testing.T) {
		testStr := "test-error"
		cmd := cobra.Command{}
		args := []string{"/tmp/1"}
		cli.InitClusterService = func(inputConfigFile string, force, verbose bool) error {
			return fmt.Errorf(testStr)
		}
		defer resetCLIVars()
		err := cli.RunInitClusterCmd(&cmd, args)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got:%v, expected:%s", err, testStr)
		}
	})
	t.Run("returns error cluster is created successfully", func(t *testing.T) {

		cmd := cobra.Command{}
		args := []string{"/tmp/1"}
		cli.InitClusterService = func(inputConfigFile string, force, verbose bool) error {
			return nil
		}
		defer resetCLIVars()
		err := cli.RunInitClusterCmd(&cmd, args)
		if err != nil {
			t.Fatalf("got:%v, expected no error", err)
		}
	})
}
func TestInitClusterService(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("fails if input config file does not exist", func(t *testing.T) {
		defer resetCLIVars()
		err := cli.InitClusterService("/tmp/invalid_file", false, false)
		if err == nil {
			t.Fatalf("error was expected")
		}
	})

	t.Run("fails if LoadInputConfigToIdl returns error", func(t *testing.T) {
		testStr := "got an error"
		defer resetCLIVars()
		defer utils.ResetSystemFunctions()
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, nil
		}
		cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, fmt.Errorf(testStr)
		}
		err := cli.InitClusterService("/tmp/invalid_file", false, false)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got %v, want %s", err, testStr)
		}
	})
	t.Run("fails if ValidateInputConfigAndSetDefaults returns error", func(t *testing.T) {
		testStr := "got an error"
		defer resetCLIVars()
		defer utils.ResetSystemFunctions()
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, nil
		}
		cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return fmt.Errorf(testStr)
		}
		err := cli.InitClusterService("/tmp/invalid_file", false, false)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got %v, want %s", err, testStr)
		}
	})
	t.Run("fails if ValidateInputConfigAndSetDefaults returns error", func(t *testing.T) {
		testStr := "got an error"
		defer resetCLIVars()
		defer utils.ResetSystemFunctions()
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, nil
		}
		cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return fmt.Errorf(testStr)
		}
		err := cli.InitClusterService("/tmp/invalid_file", false, false)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got %v, want %s", err, testStr)
		}
	})
	t.Run("returns error if connect to hub fails", func(t *testing.T) {
		testStr := "test-error"
		defer resetCLIVars()
		defer utils.ResetSystemFunctions()
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, nil
		}
		cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return nil
		}
		cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, fmt.Errorf(testStr)
		}

		err := cli.InitClusterService("/tmp/invalid_file", false, false)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got %v, want %v", err, testStr)
		}
	})
	t.Run("returns error if RPC returns error", func(t *testing.T) {
		testStr := "test-error"
		defer resetCLIVars()
		defer utils.ResetSystemFunctions()
		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, nil
		}
		cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return nil
		}
		cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().MakeCluster(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf(testStr))
			return hubClient, nil
		}

		err := cli.InitClusterService("/tmp/invalid_file", false, false)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got %v, want %v", err, testStr)
		}
	})
	t.Run("returns error if stream receiver returns error", func(t *testing.T) {
		testStr := "test-error"
		defer resetCLIVars()
		defer utils.ResetSystemFunctions()

		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, nil
		}
		cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return nil
		}
		cli.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().MakeCluster(gomock.Any(), gomock.Any()).Return(nil, nil)
			return hubClient, nil
		}
		cli.ParseStreamResponse = func(stream cli.StreamReceiver) error {
			return fmt.Errorf(testStr)
		}
		err := cli.InitClusterService("/tmp/invalid_file", false, false)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got %v, want %v", err, testStr)
		}
	})
}

func resetConfHostnames() {
	cli.Conf.Hostnames = []string{"cdw", "sdw1", "sdw2"}
}

func TestValidateInputConfigAndSetDefaults(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	var request = &idl.MakeClusterRequest{}

	_, _, logfile := testhelper.SetupTestLogger()
	initializeRequest := func() {
		coordinator := &idl.Segment{
			HostAddress:   "cdw",
			HostName:      "cdw",
			Port:          700,
			DataDirectory: "/tmp/coordinator/",
		}
		gparray := idl.GpArray{
			Coordinator: coordinator,
			Primaries: []*idl.Segment{
				{
					HostAddress:   "sdw1",
					HostName:      "sdw1",
					Port:          7002,
					DataDirectory: "/tmp/demo/1",
				},
				{
					HostAddress:   "sdw1",
					HostName:      "sdw1",
					Port:          7003,
					DataDirectory: "/tmp/demo/2",
				},
				{
					HostAddress:   "sdw2",
					HostName:      "sdw2",
					Port:          7004,
					DataDirectory: "/tmp/demo/3",
				},
				{
					HostAddress:   "sdw2",
					HostName:      "sdw2",
					Port:          7005,
					DataDirectory: "/tmp/demo/4",
				},
			},
		}
		clusterparamas := idl.ClusterParams{
			CoordinatorConfig: map[string]string{
				"max_connections": "50",
			},
			SegmentConfig: map[string]string{
				"max_connections":    "150",
				"debug_pretty_print": "off",
				"log_min_messages":   "warning",
			},
			CommonConfig: map[string]string{
				"shared_buffers": "128000kB",
			},
			Locale: &idl.Locale{
				LcAll:      "en_US.UTF-8",
				LcCtype:    "en_US.UTF-8",
				LcTime:     "en_US.UTF-8",
				LcNumeric:  "en_US.UTF-8",
				LcMonetory: "en_US.UTF-8",
				LcMessages: "en_US.UTF-8",
				LcCollate:  "en_US.UTF-8",
			},
			HbaHostnames:  false,
			Encoding:      "Unicode",
			SuPassword:    "gp",
			DbName:        "gpadmin",
			DataChecksums: false,
		}
		request = &idl.MakeClusterRequest{
			GpArray:       &gparray,
			ClusterParams: &clusterparamas,
			ForceFlag:     false,
			Verbose:       false,
		}
	}

	cliHandler := viper.New()
	cliHandler.Set("coordinator", idl.Segment{})
	cliHandler.Set("primary-segments-array", []idl.Segment{})
	cliHandler.Set("locale", idl.Locale{})

	initializeRequest()
	resetConfHostnames()
	t.Run("fails if coordinator segment is not provided in input config", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()

		expectedError := "no coordinator segments are provided in input config file"
		request.GpArray.Coordinator = nil
		v := viper.New()
		err := cli.ValidateInputConfigAndSetDefaults(request, v)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("fails if 0 primary segments are provided in input config file", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		expectedError := "no primary segments are provided in input config file"
		request.GpArray.Primaries = []*idl.Segment{}
		v := viper.New()
		v.Set("coordinator", idl.Segment{})
		err := cli.ValidateInputConfigAndSetDefaults(request, v)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})

	t.Run("fails if some of hosts do not have gp services configured", func(t *testing.T) {
		defer resetCLIVars()
		defer resetConfHostnames()
		defer initializeRequest()
		cli.Conf.Hostnames = []string{"cdw", "sdw1"}
		expectedError := "following hostnames [sdw2 sdw2] do not have gp services configured. Please configure the services"

		err := cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got: %v, want: %v", err, expectedError)
		}
	})
	t.Run("succeeds with info if encoding is not provided", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		defer resetConfHostnames()
		cli.CheckForDuplicatPortAndDataDirectory = func(primaries []*idl.Segment) error {
			return nil
		}
		request.ClusterParams.Encoding = ""

		err := cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err != nil {
			t.Fatalf("got an unexpected error %v", err)
		}
		expectedLogMsg := `Could not find encoding in cluster config, defaulting to UTF-8`
		testutils.AssertLogMessage(t, logfile, expectedLogMsg)
	})
	t.Run("fails if provided encoding is SQL_ASCII", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		defer resetConfHostnames()
		request.ClusterParams.Encoding = "SQL_ASCII"
		expectedError := "SQL_ASCII is no longer supported as a server encoding"

		err := cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("succeeds with info if coordinator max_connection is not provided", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		defer resetConfHostnames()
		delete(request.ClusterParams.CoordinatorConfig, "max_connections")
		delete(request.ClusterParams.CommonConfig, "max_connections")

		err := cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err != nil {
			t.Fatalf("got an unexpected error %v", err)
		}
		expectedLogMsg := `COORDINATOR max_connections not set, will set to default value 150`
		testutils.AssertLogMessage(t, logfile, expectedLogMsg)
	})
	t.Run("fails if provided coordinator max_connection is less than 1", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		defer resetConfHostnames()
		request.ClusterParams.CoordinatorConfig["max_connections"] = "-1"
		expectedError := "COORDINATOR max_connections value -1 is too small. Should be more than 1"

		err := cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("fails if provided coordinator max_connection is less than 1 in common-config", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		defer resetConfHostnames()
		delete(request.ClusterParams.CoordinatorConfig, "max_connections")
		request.ClusterParams.CommonConfig["max_connections"] = "-1"
		expectedError := "COORDINATOR max_connections value -1 is too small. Should be more than 1"

		err := cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("max_connections picks value from common config if not provided in coordinator-config", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		defer resetConfHostnames()
		delete(request.ClusterParams.CoordinatorConfig, "max_connections")
		request.ClusterParams.CommonConfig["max_connections"] = "300"
		expectedLogMsg := "COORDINATOR max_connections set to value: 300"

		err := cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err != nil {
			t.Fatalf("got %v, want no error", err)
		}
		testutils.AssertLogMessage(t, logfile, expectedLogMsg)
	})
	t.Run("succeeds with info if shared_buffers are not provided", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		defer resetConfHostnames()
		delete(request.ClusterParams.CommonConfig, "shared_buffers")

		err := cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err != nil {
			t.Fatalf("got an unexpected error %v", err)
		}
		expectedLogMsg := `shared_buffers is not set, will set to default value 128000kB`
		testutils.AssertLogMessage(t, logfile, expectedLogMsg)
	})
	t.Run("fails if port/directory duplicate check fails returns error", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		expectedError := "Got an error"
		cli.CheckForDuplicatPortAndDataDirectory = func(primaries []*idl.Segment) error {
			return errors.New(expectedError)
		}

		err := cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
}

func TestCheckForDuplicatePortAndDataDirectoryFn(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	var primaries = []*idl.Segment{}
	initializeData := func() {
		var primary0 = idl.Segment{
			HostAddress:   "sdw1",
			HostName:      "sdw1",
			Port:          7002,
			DataDirectory: "/tmp/demo/1",
		}
		var primary1 = idl.Segment{
			HostAddress:   "sdw1",
			HostName:      "sdw1",
			Port:          7003,
			DataDirectory: "/tmp/demo/2",
		}
		var primary2 = idl.Segment{
			HostAddress:   "sdw2",
			HostName:      "sdw2",
			Port:          7004,
			DataDirectory: "/tmp/demo/3",
		}
		var primary3 = idl.Segment{
			HostAddress:   "sdw2",
			HostName:      "sdw2",
			Port:          7005,
			DataDirectory: "/tmp/demo/4",
		}
		primaries = []*idl.Segment{
			&primary0, &primary1, &primary2, &primary3,
		}
	}

	initializeData()
	t.Run("fails if duplicate data-directory entry is found for a host", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeData()
		expectedError := "duplicate data directory entry /tmp/demo/1 found for host sdw1"
		primaries[1].DataDirectory = "/tmp/demo/1"
		err := cli.CheckForDuplicatPortAndDataDirectory(primaries)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}

	})
	t.Run("fails if duplicate port entry is found for a host", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeData()
		expectedError := "duplicate port entry 7002 found for host sdw1"
		primaries[1].Port = 7002
		err := cli.CheckForDuplicatPortAndDataDirectory(primaries)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("succeeds if no duplicate port/datadir entry is found for any of the hosts", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeData()
		err := cli.CheckForDuplicatPortAndDataDirectory(primaries)
		if err != nil {
			t.Fatalf("got an unexpected error")
		}
	})
}
