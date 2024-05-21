package gpctl_cli_test

import (
	"errors"
	"fmt"
	"github.com/greenplum-db/gpdb/gp/gpctl/gpctl_cli"
	gpserviceCLI "github.com/greenplum-db/gpdb/gp/gpservice/gpservice-cli"
	"github.com/greenplum-db/gpdb/gp/gpservice/hub"
	"google.golang.org/grpc"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/greenplum-db/gpdb/gp/constants"

	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/idl/mock_idl"
	"github.com/greenplum-db/gpdb/gp/testutils"
	"github.com/greenplum-db/gpdb/gp/testutils/exectest"
	"github.com/greenplum-db/gpdb/gp/utils"
)

var (
	ctrl *gomock.Controller
)

func setupTest(t *testing.T) {
	testhelper.SetupTestLogger()
	gpserviceCLI.Conf = testutils.InitializeTestEnv()
	ctrl = gomock.NewController(t)
}

func teardownTest() {
	ctrl.Finish()
}

func init() {
	exectest.RegisterMains(
		CommandSuccess,
		CommandFailure,
	)
}

func resetCLIVars() {
	gpserviceCLI.DialContextFunc = grpc.DialContext
	gpserviceCLI.ConnectToHub = gpserviceCLI.ConnectToHubFunc
	gpserviceCLI.StartHubService = gpserviceCLI.StartHubServiceFunc
	gpserviceCLI.WaitAndRetryHubConnect = gpserviceCLI.WaitAndRetryHubConnectFunc
	gpserviceCLI.ShowHubStatus = gpserviceCLI.ShowHubStatusFunc
	gpserviceCLI.StartAgentsAll = gpserviceCLI.StartAgentsAllFunc
	gpserviceCLI.ShowAgentsStatus = gpserviceCLI.ShowAgentsStatusFunc
	gpserviceCLI.PrintServicesStatus = gpserviceCLI.PrintServicesStatusFunc
	gpserviceCLI.StopAgentService = gpserviceCLI.StopAgentServiceFunc
	gpserviceCLI.StopHubService = gpserviceCLI.StopHubServiceFunc
	gpctl_cli.InitClusterService = gpctl_cli.InitClusterServiceFn
	gpctl_cli.LoadInputConfigToIdl = gpctl_cli.LoadInputConfigToIdlFn
	gpctl_cli.ValidateInputConfigAndSetDefaults = gpctl_cli.ValidateInputConfigAndSetDefaultsFn
	gpctl_cli.CheckForDuplicatPortAndDataDirectory = gpctl_cli.CheckForDuplicatePortAndDataDirectoryFn
	gpctl_cli.GetSystemLocale = gpctl_cli.GetSystemLocaleFn
	gpctl_cli.SetDefaultLocale = gpctl_cli.SetDefaultLocaleFn
	gpctl_cli.ParseStreamResponse = gpserviceCLI.ParseStreamResponseFn
	gpctl_cli.IsGpServicesEnabled = gpctl_cli.IsGpServicesEnabledFn
}

func TestMain(m *testing.M) {
	os.Exit(exectest.Run(m))
}

func TestValidateStringArray(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	t.Run("returns true when a valid array is provided", func(t *testing.T) {
		input := []string{"test1", "test2"}

		result := gpctl_cli.ValidateStringArray(input)
		if !result {
			t.Fatalf("Got:%v, expected: true", result)
		}
	})

	t.Run("returns false when a invalid array is provided", func(t *testing.T) {
		input := []string{"test1", "test2", ""}

		result := gpctl_cli.ValidateStringArray(input)
		if result {
			t.Fatalf("Got:%v, expected: true", result)
		}
	})
}

func TestValidateMultiHomeConfig(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	t.Run("returns error when the hosts in the host list has different number of interfaces", func(t *testing.T) {
		testStr := "multi-home validation failed, all hosts should have same number of interfaces/aliase"
		config := gpctl_cli.InitConfig{}
		addressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2", "sdw1-3", "sdw1-4"},
			"sdw2": {"sdw2-1", "sdw2-2", "sdw2-3", "sdw2-4"},
			"sdw3": {"sdw3-1", "sdw3-2", "sdw3-3"},
			"sdw4": {"sdw4-1", "sdw4-2", "sdw4-3", "sdw4-4"},
		}

		_, err := gpctl_cli.ValidateMultiHomeConfig(config, addressMap)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Want: %s", err, testStr)
		}
	})

	t.Run("returns error when the directories per host are not in multiple of number of interfaces", func(t *testing.T) {
		testStr := "multi-host setup must have data-directories in multiple of number of addresses or more."
		config := gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/tmp", "/tmp", "/tmp"}}
		addressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2"},
			"sdw2": {"sdw2-1", "sdw2-2"},
			"sdw3": {"sdw3-1", "sdw3-2"},
			"sdw4": {"sdw4-1", "sdw4-2"},
		}

		_, err := gpctl_cli.ValidateMultiHomeConfig(config, addressMap)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Want: %s", err, testStr)
		}
	})

	t.Run("returns error when spread mirroring and number of unique hosts is equal", func(t *testing.T) {
		testStr := "to enable spread mirroring, number of hosts should be more than number of primary segments per host."
		config := gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/tmp", "/tmp", "/tmp", "/tmp"},
			MirrorDataDirectories: []string{"/tmp", "/tmp", "/tmp", "/tmp"},
			MirroringType:         constants.SpreadMirroring}
		addressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2", "sdw1-3", "sdw1-4"},
			"sdw2": {"sdw2-1", "sdw2-2", "sdw2-3", "sdw2-4"},
			"sdw3": {"sdw3-1", "sdw3-2", "sdw3-3", "sdw3-4"},
			"sdw4": {"sdw4-1", "sdw4-2", "sdw4-3", "sdw4-4"},
		}
		gpctl_cli.ContainsMirror = true

		_, err := gpctl_cli.ValidateMultiHomeConfig(config, addressMap)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Want: %s", err, testStr)
		}
	})

	t.Run("returns error when spread mirroring and number of unique hosts is lesser", func(t *testing.T) {
		testStr := "to enable spread mirroring, number of hosts should be more than number of primary segments per host."
		config := gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/tmp", "/tmp", "/tmp", "/tmp"},
			MirrorDataDirectories: []string{"/tmp", "/tmp", "/tmp", "/tmp"},
			MirroringType:         constants.SpreadMirroring}
		addressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2", "sdw1-3", "sdw1-4"},
			"sdw2": {"sdw2-1", "sdw2-2", "sdw2-3", "sdw2-4"},
			"sdw3": {"sdw3-1", "sdw3-2", "sdw3-3", "sdw3-4"},
		}
		gpctl_cli.ContainsMirror = true

		_, err := gpctl_cli.ValidateMultiHomeConfig(config, addressMap)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Want: %s", err, testStr)
		}
	})
}

func TestValidateExpansionConfigAndSetDefault(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("returns error when primary data directories not provided", func(t *testing.T) {
		testStr := "primary-data-directories not specified. Please specify primary-data-directories to continue"
		cliHandle := viper.New()
		config := &gpctl_cli.InitConfig{}

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns error when hostlist not provided", func(t *testing.T) {
		testStr := "hostlist not specified. Please specify hostlist to continue"
		cliHandle := viper.New()
		cliHandle.Set("primary-data-directories", []string{"/test"})
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}}

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns error when primary data directory contains empty string", func(t *testing.T) {
		testStr := "empty primary-data-directories entry provided, please provide valid directory"
		cliHandle := viper.New()
		cliHandle.Set("primary-data-directories", []string{""})
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{""}}

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns error when primary data directory contains empty and non-empty string", func(t *testing.T) {
		testStr := "empty primary-data-directories entry provided, please provide valid directory"
		cliHandle := viper.New()
		cliHandle.Set("primary-data-directories", []string{"/test1", "", "/test2"})
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test1", "", "/test2"}}

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns error when hostlist contains empty string", func(t *testing.T) {
		testStr := "empty hostlist entry detected, Please provide valid hostlist"
		cliHandle := viper.New()
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{""})
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{""}}

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns error when hostlist contains empty and non-empty string", func(t *testing.T) {
		testStr := "empty hostlist entry detected, Please provide valid hostlist"
		cliHandle := viper.New()
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"/tmp/1", "", "/tmp/2"})
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"/tmp/1", "", "/tmp/2"}}

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})
	t.Run("returns error when primary base port value is same as coordinator port", func(t *testing.T) {
		testStr := "coordinator port and primary-base-port value cannot be same"
		cliHandle := viper.New()
		basePort := 9000
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"swd1"})
		cliHandle.Set("primary-base-port", basePort)
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"swd1"}, Coordinator: gpctl_cli.Segment{Port: basePort}, PrimaryBasePort: basePort}

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("sets ContainsMirror to false and default value to primary base port", func(t *testing.T) {
		cliHandle := viper.New()
		basePort := 9000
		expectedPrimaryPort := basePort + 2
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"swd1"})
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"swd1"}, Coordinator: gpctl_cli.Segment{Port: basePort}}

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err != nil {
			t.Fatalf("Got:%v, Expected no error", err)
		}
		if config.PrimaryBasePort != expectedPrimaryPort {
			t.Fatalf("PrimaryBasePort Got:%d, expected;%d", config.PrimaryBasePort, expectedPrimaryPort)
		}
		if gpctl_cli.ContainsMirror {
			t.Fatalf("PrimaryBasePort got:%v, expected:false", gpctl_cli.ContainsMirror)
		}
	})

	t.Run("returns error when segment array is defined and expansion parameters also present", func(t *testing.T) {
		testStr := "cannot specify segments-array and primary-base-directories together"
		cliHandle := viper.New()
		basePort := 9000
		segArray := []gpctl_cli.SegmentPair{{Primary: &gpctl_cli.Segment{Hostname: "sdw1"}}}
		cliHandle.Set("segment-array", segArray)
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"swd1"}, Coordinator: gpctl_cli.Segment{Port: basePort},
			SegmentArray: segArray}

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns error when mirror data directories not provided", func(t *testing.T) {
		testStr := "number of primary-data-directories should be equal to number of mirror-data-directories"
		cliHandle := viper.New()
		basePort := 9000
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"swd1"}, Coordinator: gpctl_cli.Segment{Port: basePort},
			MirrorBasePort: 10000}
		cliHandle.Set("mirror-base-port", basePort)
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"swd1"})

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns error when mirror data directories count not equal to primary data directories", func(t *testing.T) {
		testStr := "number of primary-data-directories should be equal to number of mirror-data-directories"
		cliHandle := viper.New()
		basePort := 9000
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"swd1"}, Coordinator: gpctl_cli.Segment{Port: basePort},
			MirrorBasePort: 10000, MirrorDataDirectories: []string{"/test1", "/test2"}}
		cliHandle.Set("mirror-base-port", basePort)
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"swd1"})

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns error when unknown mirroring type provided", func(t *testing.T) {
		testStr := "invalid mirroring-Type: unknown. Valid options are 'group' and 'spread'"
		cliHandle := viper.New()
		basePort := 9000
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"swd1"}, Coordinator: gpctl_cli.Segment{Port: basePort},
			MirrorBasePort: 10000, MirrorDataDirectories: []string{"/test1"}, MirroringType: "UNKNOWN"}
		cliHandle.Set("mirror-base-port", basePort)
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"swd1"})
		cliHandle.Set("mirroring-type", "Unknown")

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("sets default mirror port value properly", func(t *testing.T) {
		cliHandle := viper.New()
		basePort := 9000
		expectedMirrorPort := basePort + 1002
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"swd1"}, Coordinator: gpctl_cli.Segment{Port: basePort}, MirrorDataDirectories: []string{"/test1"}, MirroringType: constants.GroupMirroring}
		cliHandle.Set("mirroring-type", constants.GroupMirroring)
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"swd1"})

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err != nil {
			t.Fatalf("Got:%v, Expected no error", err)
		}
		if config.MirrorBasePort != expectedMirrorPort {
			t.Fatalf("mirror-base-port Got:%d, Expected:%d", config.MirrorBasePort, expectedMirrorPort)
		}
	})

	t.Run("sets default mirror port value properly", func(t *testing.T) {
		testStr := "coordinator port and mirror-base-port value cannot be same. Please provide different values"
		cliHandle := viper.New()
		basePort := 9000
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"swd1"}, Coordinator: gpctl_cli.Segment{Port: basePort},
			MirrorDataDirectories: []string{"/test1"}, MirroringType: constants.GroupMirroring, MirrorBasePort: basePort}
		cliHandle.Set("mirroring-type", constants.GroupMirroring)
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"swd1"})
		cliHandle.Set("mirror-base-port", basePort)

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("sets default mirroring type value properly", func(t *testing.T) {
		cliHandle := viper.New()
		basePort := 9000
		expectedMirroringType := "group"
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test"}, HostList: []string{"swd1"}, Coordinator: gpctl_cli.Segment{Port: basePort}, MirrorDataDirectories: []string{"/test1"}, MirrorBasePort: 10000}
		cliHandle.Set("mirror-base-port", basePort)
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"swd1"})

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err != nil {
			t.Fatalf("Got:%v, Expected no error", err)
		}
		if config.MirroringType != constants.GroupMirroring {
			t.Fatalf("mirroring-typet Got:%s, Expected:%s", config.MirroringType, expectedMirroringType)
		}
	})

	t.Run("returns error when number of hosts are less in spread mirroring", func(t *testing.T) {
		testStr := "To enable spread mirroring, number of hosts should be more than number of primary segments per host."
		cliHandle := viper.New()
		basePort := 9000
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test1", "/test2", "/test3"}, HostList: []string{"swd1", "sdw2"}, Coordinator: gpctl_cli.Segment{Port: basePort}, MirrorDataDirectories: []string{"/mirror1", "/mirror2", "/mirror3"}, MirroringType: constants.SpreadMirroring}
		cliHandle.Set("mirroring-type", constants.SpreadMirroring)
		cliHandle.Set("primary-data-directories", []string{"/test"})
		cliHandle.Set("hostlist", []string{"swd1"})

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns error when number of hosts are equal in spread mirroring", func(t *testing.T) {
		testStr := "To enable spread mirroring, number of hosts should be more than number of primary segments per host."
		cliHandle := viper.New()
		basePort := 9000
		config := &gpctl_cli.InitConfig{PrimaryDataDirectories: []string{"/test1", "/test2", "/test3"}, HostList: []string{"swd1", "sdw2", "sdw3"}, Coordinator: gpctl_cli.Segment{Port: basePort}, MirrorDataDirectories: []string{"/mirror1", "/mirror2", "/mirror3"}, MirroringType: constants.SpreadMirroring}
		cliHandle.Set("mirroring-type", constants.SpreadMirroring)
		cliHandle.Set("primary-data-directories", []string{"/test1", "/test2", "/test3"})
		cliHandle.Set("hostlist", []string{"swd1", "sdw2", "sdw3"})

		err := gpctl_cli.ValidateExpansionConfigAndSetDefault(config, cliHandle)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})
}

func TestIsMultiHome(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("returns error when error getting all hostnames", func(t *testing.T) {
		testStr := "test Error"
		hostList := []string{"cdw"}
		cdw := mock_idl.NewMockHubClient(ctrl)
		request := &idl.GetAllHostNamesRequest{HostList: hostList}
		reply := &idl.GetAllHostNamesReply{}
		cdw.EXPECT().GetAllHostNames(gomock.Any(), request).Return(reply, fmt.Errorf(testStr))
		oldHubClient := gpctl_cli.HubClient
		gpctl_cli.HubClient = cdw
		defer func() { gpctl_cli.HubClient = oldHubClient }()

		_, _, _, err := gpctl_cli.IsMultiHome(hostList)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("Got:%v, Expected:%s", err, testStr)
		}
	})

	t.Run("returns false and other details properly when not a multi-home", func(t *testing.T) {
		hostList := []string{"sdw1", "sdw2", "sdw3"}
		expectedAddressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2", "sdw3": "sdw3"}
		expectedNameAddressMap := map[string][]string{"sdw1": {"sdw1"}, "sdw2": {"sdw2"}, "sdw3": {"sdw3"}}
		cdw := mock_idl.NewMockHubClient(ctrl)
		request := &idl.GetAllHostNamesRequest{HostList: hostList}
		reply := &idl.GetAllHostNamesReply{HostNameMap: expectedAddressNameMap}
		cdw.EXPECT().GetAllHostNames(gomock.Any(), request).Return(reply, nil)
		oldHubClient := gpctl_cli.HubClient
		gpctl_cli.HubClient = cdw
		defer func() { gpctl_cli.HubClient = oldHubClient }()

		isMultiHome, nameAddressMap, addressNameMap, err := gpctl_cli.IsMultiHome(hostList)
		if err != nil {
			t.Fatalf("Got:%v, Expected:no error", err)
		}
		if !reflect.DeepEqual(expectedAddressNameMap, addressNameMap) {
			t.Fatalf("AddressName Map Failed. Got:%v, Expected:%v", addressNameMap, expectedAddressNameMap)
		}
		if !reflect.DeepEqual(expectedNameAddressMap, nameAddressMap) {
			t.Fatalf("NameAddress Map Failed. Got:%v, Expected:%v", nameAddressMap, expectedNameAddressMap)
		}
		if isMultiHome {
			t.Fatalf("isMultihome Failed. Got:%v, Expected: false", isMultiHome)
		}
	})

	t.Run("returns true and other details properly when is a multi-home environment", func(t *testing.T) {
		hostList := []string{"sdw1-1", "sdw1-2", "sdw2-1", "sdw2-2"}
		expectedAddressNameMap := map[string]string{"sdw1-1": "sdw1", "sdw1-2": "sdw1", "sdw2-1": "sdw2", "sdw2-2": "sdw2"}
		expectedNameAddressMap := map[string][]string{"sdw1": {"sdw1-1", "sdw1-2"}, "sdw2": {"sdw2-1", "sdw2-2"}}
		cdw := mock_idl.NewMockHubClient(ctrl)
		request := &idl.GetAllHostNamesRequest{HostList: hostList}
		reply := &idl.GetAllHostNamesReply{HostNameMap: expectedAddressNameMap}
		cdw.EXPECT().GetAllHostNames(gomock.Any(), request).Return(reply, nil)
		oldHubClient := gpctl_cli.HubClient
		gpctl_cli.HubClient = cdw
		defer func() { gpctl_cli.HubClient = oldHubClient }()

		isMultiHome, nameAddressMap, addressNameMap, err := gpctl_cli.IsMultiHome(hostList)
		if err != nil {
			t.Fatalf("Got:%v, Expected:no error", err)
		}
		if !reflect.DeepEqual(expectedAddressNameMap, addressNameMap) {
			t.Fatalf("AddressName Map Failed. Got:%v, Expected:%v", addressNameMap, expectedAddressNameMap)
		}
		if !reflect.DeepEqual(expectedNameAddressMap, nameAddressMap) {
			t.Fatalf("NameAddress Map Failed. Got:%v, Expected:%v", nameAddressMap, expectedNameAddressMap)
		}
		if !isMultiHome {
			t.Fatalf("isMultihome Failed. Got:%v, Expected: true", isMultiHome)
		}
	})
}

// CheckIfPortConflict utility function to check if there's a port conflict in the generated segments array
func CheckIfPortConflict(segPairList *[]gpctl_cli.SegmentPair) error {
	var conflictList []int
	hostPortMap := map[string]map[int]bool{}
	for index, segPair := range *segPairList {
		if segPair.Primary != nil {
			if _, ok := hostPortMap[segPair.Primary.Hostname][segPair.Primary.Port]; ok {
				conflictList = append(conflictList, index)
			} else {
				hostPortMap[segPair.Primary.Hostname] = make(map[int]bool)
				hostPortMap[segPair.Primary.Hostname][segPair.Primary.Port] = true
			}
		}
		if segPair.Mirror != nil {
			if _, ok := hostPortMap[segPair.Mirror.Hostname][segPair.Mirror.Port]; ok {
				conflictList = append(conflictList, index)
			} else {
				hostPortMap[segPair.Mirror.Hostname] = make(map[int]bool)
				hostPortMap[segPair.Mirror.Hostname][segPair.Mirror.Port] = true
			}
		}
	}
	if len(conflictList) == 0 {
		return nil
	}
	return fmt.Errorf("port conflict in segPairList at index:%v", conflictList)

}

func TestExpandNonMultiHomePrimaryList(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("should expand primary array for a non multi home setup", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		primaryBasePort := 9000
		primaryBaseDataDirectories := []string{"/test", "/test"}
		hostList := []string{"sdw1", "sdw2"}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2"}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test/gpseg3"}},
		}

		segPairList = *gpctl_cli.ExpandNonMultiHomePrimaryList(&segPairList, primaryBasePort, primaryBaseDataDirectories, hostList, addressNameMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("should expand primary array for a non multi home setup with different mount-points", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		primaryBasePort := 9000
		primaryBaseDataDirectories := []string{"/test1", "/test2"}
		hostList := []string{"sdw1", "sdw2"}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2"}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test1/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test2/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test1/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test2/gpseg3"}},
		}

		segPairList = *gpctl_cli.ExpandNonMultiHomePrimaryList(&segPairList, primaryBasePort, primaryBaseDataDirectories, hostList, addressNameMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})
}

func TestExpandMultiHomePrimaryArray(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("should expand primary array for a multi home setup", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		primaryBasePort := 9000
		primaryBaseDataDirectories := []string{"/test", "/test"}
		uniqHostnameList := []string{"sdw1", "sdw2"}
		nameAddressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2"},
			"sdw2": {"sdw2-1", "sdw2-2"},
		}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-2", Port: 9001, DataDirectory: "/test/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test/gpseg3"}},
		}

		segPairList = *gpctl_cli.ExpandMultiHomePrimaryArray(&segPairList, primaryBasePort, primaryBaseDataDirectories, uniqHostnameList, nameAddressMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("should expand primary array for a multi home setup multi mount point", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		primaryBasePort := 9000
		primaryBaseDataDirectories := []string{"/test1", "/test2"}
		uniqHostnameList := []string{"sdw1", "sdw2"}
		nameAddressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2"},
			"sdw2": {"sdw2-1", "sdw2-2"},
		}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test1/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-2", Port: 9001, DataDirectory: "/test2/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test1/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test2/gpseg3"}},
		}

		segPairList = *gpctl_cli.ExpandMultiHomePrimaryArray(&segPairList, primaryBasePort, primaryBaseDataDirectories, uniqHostnameList, nameAddressMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})
}

func TestExpandNonMultiHomeGroupMirrorList(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("should expand mirror array for a non multi home setup", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair

		mirrorBasePort := 9000
		mirrorBaseDataDirectories := []string{"/test", "/test"}
		hostList := []string{"sdw1", "sdw2"}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2"}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test/gpseg0"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test/gpseg1"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test/gpseg2"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test/gpseg3"}},
		}
		segPairList = make([]gpctl_cli.SegmentPair, 4)

		segPairList = *gpctl_cli.ExpandNonMultiHomeGroupMirrorList(&segPairList, mirrorBasePort, mirrorBaseDataDirectories, hostList, addressNameMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("should expand mirror array for a non multi home setup with separate mount", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		mirrorBasePort := 9000
		mirrorBaseDataDirectories := []string{"/test1", "/test2"}
		hostList := []string{"sdw1", "sdw2"}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2"}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test1/gpseg0"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test2/gpseg1"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test1/gpseg2"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test2/gpseg3"}},
		}
		segPairList = make([]gpctl_cli.SegmentPair, 4)

		segPairList = *gpctl_cli.ExpandNonMultiHomeGroupMirrorList(&segPairList, mirrorBasePort, mirrorBaseDataDirectories, hostList, addressNameMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})
}

func TestExpandNonMultiHomeSpreadMirrorList(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("should expand mirror array for a non multi home setup - spread mirroring", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		mirrorBasePort := 9000
		mirrorBaseDataDirectories := []string{"/test", "/test"}
		hostList := []string{"sdw1", "sdw2", "sdw3"}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2", "sdw3": "sdw3"}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test/gpseg0"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3", Port: 9001, DataDirectory: "/test/gpseg1"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3", Port: 9000, DataDirectory: "/test/gpseg2"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test/gpseg3"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test/gpseg4"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test/gpseg5"}},
		}
		segPairList = make([]gpctl_cli.SegmentPair, 6)

		segPairList = *gpctl_cli.ExpandNonMultiHomeSpreadMirroring(&segPairList, mirrorBasePort, mirrorBaseDataDirectories, hostList, addressNameMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("should expand mirror array for a non multi home setup with separate mount - spread mirroring", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		mirrorBasePort := 9000
		mirrorBaseDataDirectories := []string{"/test1", "/test2"}
		hostList := []string{"sdw1", "sdw2", "sdw3"}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2", "sdw3": "sdw3"}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test1/gpseg0"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3", Port: 9001, DataDirectory: "/test2/gpseg1"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3", Port: 9000, DataDirectory: "/test1/gpseg2"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test2/gpseg3"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test1/gpseg4"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test2/gpseg5"}},
		}
		segPairList = make([]gpctl_cli.SegmentPair, 6)

		segPairList = *gpctl_cli.ExpandNonMultiHomeSpreadMirroring(&segPairList, mirrorBasePort, mirrorBaseDataDirectories, hostList, addressNameMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})
}

func TestExpandMultiHomeGroupMirrorList(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("should expand mirror array for a multi home setup - group mirroring", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		segPairList = make([]gpctl_cli.SegmentPair, 4)
		mirrorBasePort := 9000
		mirrorBaseDataDirectories := []string{"/test", "/test"}
		uniqHostnameList := []string{"sdw1", "sdw2"}
		nameAddressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2"},
			"sdw2": {"sdw2-1", "sdw2-2"},
		}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test/gpseg0"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test/gpseg1"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test/gpseg2"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-2", Port: 9001, DataDirectory: "/test/gpseg3"}},
		}

		segPairList = *gpctl_cli.ExpandMultiHomeGroupMirrorList(&segPairList, mirrorBasePort, mirrorBaseDataDirectories, uniqHostnameList, nameAddressMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})
	t.Run("should expand mirror array for a multi home setup - group mirroring with separate mount point", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		segPairList = make([]gpctl_cli.SegmentPair, 4)
		mirrorBasePort := 9000
		mirrorBaseDataDirectories := []string{"/test1", "/test2"}
		uniqHostnameList := []string{"sdw1", "sdw2"}
		nameAddressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2"},
			"sdw2": {"sdw2-1", "sdw2-2"},
		}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test1/gpseg0"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test2/gpseg1"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test1/gpseg2"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-2", Port: 9001, DataDirectory: "/test2/gpseg3"}},
		}

		segPairList = *gpctl_cli.ExpandMultiHomeGroupMirrorList(&segPairList, mirrorBasePort, mirrorBaseDataDirectories, uniqHostnameList, nameAddressMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})
}
func TestExpandMultiHomeSpreadMirrorList(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("should expand mirror array for a multi home setup - spread mirroring", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		segPairList = make([]gpctl_cli.SegmentPair, 6)
		mirrorBasePort := 9000
		mirrorBaseDataDirectories := []string{"/test", "/test"}
		uniqHostnameList := []string{"sdw1", "sdw2", "sdw3"}
		nameAddressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2"},
			"sdw2": {"sdw2-1", "sdw2-2"},
			"sdw3": {"sdw3-1", "sdw3-2"},
		}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test/gpseg0"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-2", Port: 9001, DataDirectory: "/test/gpseg1"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-2", Port: 9000, DataDirectory: "/test/gpseg2"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9001, DataDirectory: "/test/gpseg3"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test/gpseg4"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test/gpseg5"}},
		}

		segPairList = *gpctl_cli.ExpandMultiHomeSpreadMirrorList(&segPairList, mirrorBasePort, mirrorBaseDataDirectories, uniqHostnameList, nameAddressMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("should expand mirror array for a multi home setup - group mirroring with separate mount point", func(t *testing.T) {
		var segPairList []gpctl_cli.SegmentPair
		segPairList = make([]gpctl_cli.SegmentPair, 6)
		mirrorBasePort := 9000
		mirrorBaseDataDirectories := []string{"/test1", "/test2"}
		uniqHostnameList := []string{"sdw1", "sdw2", "sdw3"}
		nameAddressMap := map[string][]string{
			"sdw1": {"sdw1-1", "sdw1-2"},
			"sdw2": {"sdw2-1", "sdw2-2"},
			"sdw3": {"sdw3-1", "sdw3-2"},
		}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test1/gpseg0"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-2", Port: 9001, DataDirectory: "/test2/gpseg1"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-2", Port: 9000, DataDirectory: "/test1/gpseg2"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9001, DataDirectory: "/test2/gpseg3"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test1/gpseg4"}},
			{Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test2/gpseg5"}},
		}

		segPairList = *gpctl_cli.ExpandMultiHomeSpreadMirrorList(&segPairList, mirrorBasePort, mirrorBaseDataDirectories, uniqHostnameList, nameAddressMap)
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%+v Expected %+v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})
}
func TestExpandSegPairArray(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	t.Run("returns list of expansion for the given config when contains only primaries non multi-home", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test", "/test"},
			HostList:               []string{"sdw1", "sdw2"},
			MirrorDataDirectories:  nil,
			MirrorBasePort:         0,
			MirroringType:          "",
		}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1"}, "sdw2": {"sdw2"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test/gpseg3"}},
		}
		gpctl_cli.ContainsMirror = false

		segPairList := gpctl_cli.ExpandSegPairArray(config, false, nameAddressMap, addressNameMap)
		if len(segPairList) != 4 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("returns list of expansion for the given config when contains only primaries non multi-home different mount points", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test1", "/test2"},
			HostList:               []string{"sdw1", "sdw2"},
			MirrorDataDirectories:  nil,
			MirrorBasePort:         0,
			MirroringType:          "",
		}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1"}, "sdw2": {"sdw2"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test1/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test2/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test1/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test2/gpseg3"}},
		}
		gpctl_cli.ContainsMirror = false

		segPairList := gpctl_cli.ExpandSegPairArray(config, false, nameAddressMap, addressNameMap)
		if len(segPairList) != 4 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("returns list of expansion for the given config when contains only primaries non multi-home and given address list", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test", "/test"},
			HostList:               []string{"sdw1-1", "sdw2-1"},
			MirrorDataDirectories:  nil,
			MirrorBasePort:         0,
			MirroringType:          "",
		}
		addressNameMap := map[string]string{"sdw1-1": "sdw1", "sdw2-1": "sdw2"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1-1"}, "sdw2": {"sdw2-1"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9001, DataDirectory: "/test/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9001, DataDirectory: "/test/gpseg3"}},
		}
		gpctl_cli.ContainsMirror = false

		segPairList := gpctl_cli.ExpandSegPairArray(config, false, nameAddressMap, addressNameMap)
		if len(segPairList) != 4 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("returns list of expansion for the given config when contains only primaries on a multi-home", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test", "/test"},
			HostList:               []string{"sdw1-1", "sdw1-2", "sdw2-1", "sdw2-2"},
			MirrorDataDirectories:  nil,
			MirrorBasePort:         0,
			MirroringType:          "",
		}
		addressNameMap := map[string]string{"sdw1-1": "sdw1", "sdw1-2": "sdw1", "sdw2-1": "sdw2", "sdw2-2": "sdw2"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1-1", "sdw1-2"}, "sdw2": {"sdw2-1", "sdw2-2"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-2", Port: 9001, DataDirectory: "/test/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test/gpseg3"}},
		}
		gpctl_cli.ContainsMirror = false

		segPairList := gpctl_cli.ExpandSegPairArray(config, true, nameAddressMap, addressNameMap)
		if len(segPairList) != 4 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("returns list of expansion for the given config when contains only primaries on a multi-home with separate mount point", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test1", "/test2"},
			HostList:               []string{"sdw1-1", "sdw1-2", "sdw2-1", "sdw2-2"},
			MirrorDataDirectories:  nil,
			MirrorBasePort:         0,
			MirroringType:          "",
		}
		addressNameMap := map[string]string{"sdw1-1": "sdw1", "sdw1-2": "sdw1", "sdw2-1": "sdw2", "sdw2-2": "sdw2"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1-1", "sdw1-2"}, "sdw2": {"sdw2-1", "sdw2-2"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test1/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-2", Port: 9001, DataDirectory: "/test2/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test1/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test2/gpseg3"}},
		}
		gpctl_cli.ContainsMirror = false

		segPairList := gpctl_cli.ExpandSegPairArray(config, true, nameAddressMap, addressNameMap)
		if len(segPairList) != 4 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	// Tests with mirror
	t.Run("returns list of expansion for the given config contain primary and group mirrors non multi-home", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test", "/test"},
			HostList:               []string{"sdw1", "sdw2"},
			MirrorDataDirectories:  []string{"/mirror", "/mirror"},
			MirrorBasePort:         10000,
			MirroringType:          constants.GroupMirroring,
		}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1"}, "sdw2": {"sdw2"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test/gpseg0"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 10000, DataDirectory: "/mirror/gpseg0"}},

			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test/gpseg1"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 10001, DataDirectory: "/mirror/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test/gpseg2"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 10000, DataDirectory: "/mirror/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test/gpseg3"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 10001, DataDirectory: "/mirror/gpseg3"}},
		}
		gpctl_cli.ContainsMirror = true

		segPairList := gpctl_cli.ExpandSegPairArray(config, false, nameAddressMap, addressNameMap)
		if len(segPairList) != 4 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("returns list of expansion for the given config contain primary and spread mirrors non multi-home", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test", "/test"},
			HostList:               []string{"sdw1", "sdw2", "sdw3"},
			MirrorDataDirectories:  []string{"/mirror", "/mirror"},
			MirrorBasePort:         10000,
			MirroringType:          constants.SpreadMirroring,
		}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2", "sdw3": "sdw3"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1"}, "sdw2": {"sdw2"}, "sdw3": {"sdw3"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test/gpseg0"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 10000, DataDirectory: "/mirror/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test/gpseg1"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3", Port: 10001, DataDirectory: "/mirror/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test/gpseg2"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3", Port: 10000, DataDirectory: "/mirror/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test/gpseg3"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 10001, DataDirectory: "/mirror/gpseg3"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3", Port: 9000, DataDirectory: "/test/gpseg4"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 10000, DataDirectory: "/mirror/gpseg4"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3", Port: 9001, DataDirectory: "/test/gpseg5"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 10001, DataDirectory: "/mirror/gpseg5"}},
		}
		gpctl_cli.ContainsMirror = true

		segPairList := gpctl_cli.ExpandSegPairArray(config, false, nameAddressMap, addressNameMap)
		if len(segPairList) != 6 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("returns list of expansion for the given config contain primary and group mirrors non multi-home separate mount points", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test1", "/test2"},
			HostList:               []string{"sdw1", "sdw2"},
			MirrorDataDirectories:  []string{"/mirror1", "/mirror2"},
			MirrorBasePort:         10000,
			MirroringType:          constants.GroupMirroring,
		}
		addressNameMap := map[string]string{"sdw1": "sdw1", "sdw2": "sdw2"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1"}, "sdw2": {"sdw2"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9000, DataDirectory: "/test1/gpseg0"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 10000, DataDirectory: "/mirror1/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 9001, DataDirectory: "/test2/gpseg1"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 10001, DataDirectory: "/mirror2/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9000, DataDirectory: "/test1/gpseg2"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 10000, DataDirectory: "/mirror1/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2", Port: 9001, DataDirectory: "/test2/gpseg3"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1", Port: 10001, DataDirectory: "/mirror2/gpseg3"}},
		}
		gpctl_cli.ContainsMirror = true

		segPairList := gpctl_cli.ExpandSegPairArray(config, false, nameAddressMap, addressNameMap)
		if len(segPairList) != 4 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
	})

	t.Run("returns list of expansion for the given config contain primary and group mirrors non multi-home when given address", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test", "/test"},
			HostList:               []string{"sdw1-1", "sdw2-1"},
			MirrorDataDirectories:  []string{"/mirror", "/mirror"},
			MirrorBasePort:         10000,
			MirroringType:          constants.GroupMirroring,
		}
		addressNameMap := map[string]string{"sdw1-1": "sdw1", "sdw2-1": "sdw2"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1-1"}, "sdw2": {"sdw2-1"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test/gpseg0"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 10000, DataDirectory: "/mirror/gpseg0"}},

			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9001, DataDirectory: "/test/gpseg1"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 10001, DataDirectory: "/mirror/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test/gpseg2"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 10000, DataDirectory: "/mirror/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9001, DataDirectory: "/test/gpseg3"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 10001, DataDirectory: "/mirror/gpseg3"}},
		}
		gpctl_cli.ContainsMirror = true

		segPairList := gpctl_cli.ExpandSegPairArray(config, false, nameAddressMap, addressNameMap)
		if len(segPairList) != 4 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("returns list of expansion for the given config contain primary and spread mirrors non multi-home when given address", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test", "/test"},
			HostList:               []string{"sdw1-1", "sdw2-1", "sdw3-1"},
			MirrorDataDirectories:  []string{"/mirror", "/mirror"},
			MirrorBasePort:         10000,
			MirroringType:          constants.SpreadMirroring,
		}
		addressNameMap := map[string]string{"sdw1-1": "sdw1", "sdw2-1": "sdw2", "sdw3-1": "sdw3"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1-1"}, "sdw2": {"sdw2-1"}, "sdw3": {"sdw3-1"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test/gpseg0"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 10000, DataDirectory: "/mirror/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9001, DataDirectory: "/test/gpseg1"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-1", Port: 10001, DataDirectory: "/mirror/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test/gpseg2"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-1", Port: 10000, DataDirectory: "/mirror/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9001, DataDirectory: "/test/gpseg3"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 10001, DataDirectory: "/mirror/gpseg3"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-1", Port: 9000, DataDirectory: "/test/gpseg4"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 10000, DataDirectory: "/mirror/gpseg4"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-1", Port: 9001, DataDirectory: "/test/gpseg5"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 10001, DataDirectory: "/mirror/gpseg5"}},
		}
		gpctl_cli.ContainsMirror = true

		segPairList := gpctl_cli.ExpandSegPairArray(config, false, nameAddressMap, addressNameMap)
		if len(segPairList) != 6 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("returns list of expansion for the given config when contains primaries and group mirror on a multi-home", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test", "/test"},
			HostList:               []string{"sdw1-1", "sdw1-2", "sdw2-1", "sdw2-2"},
			MirrorDataDirectories:  []string{"/mirror", "/mirror"},
			MirrorBasePort:         10000,
			MirroringType:          constants.GroupMirroring,
		}
		addressNameMap := map[string]string{"sdw1-1": "sdw1", "sdw1-2": "sdw1", "sdw2-1": "sdw2", "sdw2-2": "sdw2"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1-1", "sdw1-2"}, "sdw2": {"sdw2-1", "sdw2-2"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test/gpseg0"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 10000, DataDirectory: "/mirror/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-2", Port: 9001, DataDirectory: "/test/gpseg1"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 10001, DataDirectory: "/mirror/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test/gpseg2"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 10000, DataDirectory: "/mirror/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test/gpseg3"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-2", Port: 10001, DataDirectory: "/mirror/gpseg3"}},
		}
		gpctl_cli.ContainsMirror = true

		segPairList := gpctl_cli.ExpandSegPairArray(config, true, nameAddressMap, addressNameMap)
		if len(segPairList) != 4 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})

	t.Run("returns list of expansion for the given config when contains primaries and spread mirror on a multi-home", func(t *testing.T) {
		config := gpctl_cli.InitConfig{
			PrimaryBasePort:        9000,
			PrimaryDataDirectories: []string{"/test", "/test"},
			HostList:               []string{"sdw1-1", "sdw1-2", "sdw2-1", "sdw2-2", "sdw3-1", "sdw3-2"},
			MirrorDataDirectories:  []string{"/mirror", "/mirror"},
			MirrorBasePort:         10000,
			MirroringType:          constants.SpreadMirroring,
		}
		addressNameMap := map[string]string{"sdw1-1": "sdw1", "sdw1-2": "sdw1", "sdw2-1": "sdw2", "sdw2-2": "sdw2", "sdw3-1": "sdw3", "sdw3-2": "sdw3"}
		nameAddressMap := map[string][]string{"sdw1": {"sdw1-1", "sdw1-2"}, "sdw2": {"sdw2-1", "sdw2-2"}, "sdw3": {"sdw3-1", "sdw3-2"}}
		expectedSegPairList := []gpctl_cli.SegmentPair{
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 9000, DataDirectory: "/test/gpseg0"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 10000, DataDirectory: "/mirror/gpseg0"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-2", Port: 9001, DataDirectory: "/test/gpseg1"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-2", Port: 10001, DataDirectory: "/mirror/gpseg1"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-1", Port: 9000, DataDirectory: "/test/gpseg2"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-2", Port: 10000, DataDirectory: "/mirror/gpseg2"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 9001, DataDirectory: "/test/gpseg3"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 10001, DataDirectory: "/mirror/gpseg3"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-1", Port: 9000, DataDirectory: "/test/gpseg4"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw1", Address: "sdw1-1", Port: 10000, DataDirectory: "/mirror/gpseg4"}},
			{Primary: &gpctl_cli.Segment{Hostname: "sdw3", Address: "sdw3-2", Port: 9001, DataDirectory: "/test/gpseg5"},
				Mirror: &gpctl_cli.Segment{Hostname: "sdw2", Address: "sdw2-2", Port: 10001, DataDirectory: "/mirror/gpseg5"}},
		}
		gpctl_cli.ContainsMirror = true

		segPairList := gpctl_cli.ExpandSegPairArray(config, true, nameAddressMap, addressNameMap)
		if len(segPairList) != 6 {
			t.Fatalf("Got segPairList length %d, expected length: 4", len(segPairList))
		}
		if !reflect.DeepEqual(segPairList, expectedSegPairList) {
			t.Fatalf("Got:%v, Want:%v", segPairList, expectedSegPairList)
		}
		err := CheckIfPortConflict(&segPairList)
		if err != nil {
			t.Fatalf("Got:%v, expected no error", err)
		}
	})
}
func TestRunInitClusterCmd(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	t.Run("returns error when length of args less than 1", func(t *testing.T) {
		testStr := "please provide config file for cluster initialization"
		cmd := cobra.Command{}
		err := gpctl_cli.RunInitClusterCmd(&cmd, nil)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got:%v, expected:%s", err, testStr)
		}
	})
	t.Run("returns error when length of args greater than 1", func(t *testing.T) {
		testStr := "more arguments than expected"
		cmd := cobra.Command{}
		args := []string{"/tmp/1", "/tmp/2"}
		err := gpctl_cli.RunInitClusterCmd(&cmd, args)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got:%v, expected:%s", err, testStr)
		}
	})
	t.Run("returns error cluster creation fails", func(t *testing.T) {
		testStr := "test-error"
		cmd := cobra.Command{}
		args := []string{"/tmp/1"}
		gpctl_cli.InitClusterService = func(inputConfigFile string, force, verbose bool) error {
			return fmt.Errorf(testStr)
		}
		defer resetCLIVars()
		err := gpctl_cli.RunInitClusterCmd(&cmd, args)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got:%v, expected:%s", err, testStr)
		}
	})
	t.Run("returns error cluster is created successfully", func(t *testing.T) {

		cmd := cobra.Command{}
		args := []string{"/tmp/1"}
		gpctl_cli.InitClusterService = func(inputConfigFile string, force, verbose bool) error {
			return nil
		}
		defer resetCLIVars()
		err := gpctl_cli.RunInitClusterCmd(&cmd, args)
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
		err := gpctl_cli.InitClusterService("/tmp/invalid_file", false, false)
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
		gpctl_cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, fmt.Errorf(testStr)
		}
		gpserviceCLI.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return mock_idl.NewMockHubClient(ctrl), nil
		}
		err := gpctl_cli.InitClusterService("/tmp/invalid_file", false, false)
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
		gpctl_cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		gpctl_cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return fmt.Errorf(testStr)
		}
		gpserviceCLI.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return mock_idl.NewMockHubClient(ctrl), nil
		}
		err := gpctl_cli.InitClusterService("/tmp/invalid_file", false, false)
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
		gpctl_cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		gpctl_cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return fmt.Errorf(testStr)
		}
		gpserviceCLI.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return mock_idl.NewMockHubClient(ctrl), nil
		}
		err := gpctl_cli.InitClusterService("/tmp/invalid_file", false, false)
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
		gpctl_cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		gpctl_cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return nil
		}
		gpserviceCLI.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, fmt.Errorf(testStr)
		}

		err := gpctl_cli.InitClusterService("/tmp/invalid_file", false, false)
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
		gpctl_cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		gpctl_cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return nil
		}
		gpserviceCLI.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().MakeCluster(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf(testStr))
			return hubClient, nil
		}

		err := gpctl_cli.InitClusterService("/tmp/invalid_file", false, false)
		if err == nil || !strings.Contains(err.Error(), testStr) {
			t.Fatalf("got %v, want %v", err, testStr)
		}
	})
	t.Run("returns error if stream receiver returns error", func(t *testing.T) {
		_, _, logfile := testhelper.SetupTestLogger()
		testStr := "Cluster creation failed"
		defer resetCLIVars()
		defer utils.ResetSystemFunctions()

		utils.System.Stat = func(name string) (os.FileInfo, error) {
			return nil, nil
		}
		gpctl_cli.LoadInputConfigToIdl = func(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
			return nil, nil
		}
		gpctl_cli.ValidateInputConfigAndSetDefaults = func(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
			return nil
		}
		gpserviceCLI.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().MakeCluster(gomock.Any(), gomock.Any()).Return(nil, nil)
			return hubClient, nil
		}
		gpctl_cli.ParseStreamResponse = func(stream gpserviceCLI.StreamReceiver) error {
			return fmt.Errorf(testStr)
		}
		err := gpctl_cli.InitClusterService("/tmp/invalid_file", false, false)
		if err != nil {
			t.Fatalf("unexpected error")
		}
		testutils.AssertLogMessage(t, logfile, testStr)

	})

}

func resetConfHostnames() {
	gpserviceCLI.Conf.Hostnames = []string{"cdw", "sdw1", "sdw2"}
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
			SegmentArray: []*idl.SegmentPair{
				{
					Primary: &idl.Segment{
						HostAddress:   "sdw1",
						HostName:      "sdw1",
						Port:          7002,
						DataDirectory: "/tmp/demo/1",
					},
				},
				{
					Primary: &idl.Segment{
						HostAddress:   "sdw1",
						HostName:      "sdw1",
						Port:          7003,
						DataDirectory: "/tmp/demo/2",
					},
				},
				{
					Primary: &idl.Segment{
						HostAddress:   "sdw2",
						HostName:      "sdw2",
						Port:          7004,
						DataDirectory: "/tmp/demo/3",
					},
				},
				{
					Primary: &idl.Segment{
						HostAddress:   "sdw2",
						HostName:      "sdw2",
						Port:          7005,
						DataDirectory: "/tmp/demo/4",
					},
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
	cliHandler.Set("segment-array", []idl.Segment{})
	cliHandler.Set("locale", idl.Locale{})

	initializeRequest()
	resetConfHostnames()
	t.Run("fails if coordinator segment is not provided in input config", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()

		expectedError := "no coordinator segment provided in input config file"
		request.GpArray.Coordinator = nil
		v := viper.New()
		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, v)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("fails if 0 primary segments are provided in input config file", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		expectedError := "no primary segments are provided in input config file"
		request.GpArray.SegmentArray = []*idl.SegmentPair{}
		v := viper.New()
		v.Set("coordinator", idl.Segment{})
		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, v)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})

	t.Run("fails if IsGpServicesEnabled returns error", func(t *testing.T) {
		defer resetCLIVars()
		defer resetConfHostnames()
		defer initializeRequest()
		gpserviceCLI.Conf.Hostnames = []string{"cdw", "sdw1"}
		expectedError := "following hostnames [sdw2 sdw2] do not have gp services configured. Please configure the services"

		gpctl_cli.IsGpServicesEnabled = func(req *idl.MakeClusterRequest) error {
			return fmt.Errorf(expectedError)
		}

		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got: %v, want: %v", err, expectedError)
		}
	})
	t.Run("succeeds with info if encoding is not provided", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		defer resetConfHostnames()
		gpctl_cli.CheckForDuplicatPortAndDataDirectory = func(primaries []*idl.Segment) error {
			return nil
		}
		request.ClusterParams.Encoding = ""

		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
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

		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
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

		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err != nil {
			t.Fatalf("got an unexpected error %v", err)
		}
		expectedLogMsg := `Coordinator max_connections not set, will set to value 150 from CommonConfig`
		testutils.AssertLogMessage(t, logfile, expectedLogMsg)
	})
	t.Run("fails if provided coordinator max_connection is less than 1", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		defer resetConfHostnames()
		request.ClusterParams.CoordinatorConfig["max_connections"] = "-1"
		expectedError := "COORDINATOR max_connections value -1 is too small. Should be more than 1"

		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
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

		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
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
		expectedLogMsg := "Coordinator max_connections not set, will set to value 300 from CommonConfig"

		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
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

		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
		if err != nil {
			t.Fatalf("got an unexpected error %v", err)
		}
		expectedLogMsg := `shared_buffers is not set in CommonConfig, will set to default value 128000kB`
		testutils.AssertLogMessage(t, logfile, expectedLogMsg)
	})
	t.Run("fails if port/directory duplicate check fails returns error", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeRequest()
		expectedError := "Got an error"
		gpctl_cli.CheckForDuplicatPortAndDataDirectory = func(primaries []*idl.Segment) error {
			return errors.New(expectedError)
		}

		err := gpctl_cli.ValidateInputConfigAndSetDefaults(request, cliHandler)
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
		err := gpctl_cli.CheckForDuplicatPortAndDataDirectory(primaries)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}

	})
	t.Run("fails if duplicate port entry is found for a host", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeData()
		expectedError := "duplicate port entry 7002 found for host sdw1"
		primaries[1].Port = 7002
		err := gpctl_cli.CheckForDuplicatPortAndDataDirectory(primaries)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("succeeds if no duplicate port/datadir entry is found for any of the hosts", func(t *testing.T) {
		defer resetCLIVars()
		defer initializeData()
		err := gpctl_cli.CheckForDuplicatPortAndDataDirectory(primaries)
		if err != nil {
			t.Fatalf("got an unexpected error")
		}
	})
}

func TestValidateSegmentFn(t *testing.T) {
	setupTest(t)
	defer teardownTest()

	_, _, logfile := testhelper.SetupTestLogger()

	t.Run("Returns error if hostname and address both are empty string", func(t *testing.T) {
		defer resetCLIVars()
		expectedError := "hostName has not been provided for the segment with port 5000 and data_directory /tmp/demo/gpseg1"
		err := gpctl_cli.ValidateSegment(&idl.Segment{
			HostName:      "",
			HostAddress:   "",
			Port:          5000,
			DataDirectory: "/tmp/demo/gpseg1",
			Dbid:          2,
			Contentid:     1,
		})
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})

	t.Run("Returns error if hostname is not provided", func(t *testing.T) {
		defer resetCLIVars()
		expectedError := "hostName has not been provided for the segment with port 5000 and data_directory /tmp/demo/gpseg1"
		err := gpctl_cli.ValidateSegment(&idl.Segment{
			HostName:      "",
			HostAddress:   "sdw1",
			Port:          5000,
			DataDirectory: "/tmp/demo/gpseg1",
			Dbid:          2,
			Contentid:     1,
		})
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("Warns if hostAddress is not provided and sets it to same value as hostname", func(t *testing.T) {
		defer resetCLIVars()
		expectedWarning := "hostAddress has not been provided, populating it with same as hostName sdw1 for the segment with port 5000 and data_directory /tmp/demo/gpseg1"
		err := gpctl_cli.ValidateSegment(&idl.Segment{
			HostName:      "sdw1",
			HostAddress:   "",
			Port:          5000,
			DataDirectory: "/tmp/demo/gpseg1",
			Dbid:          2,
			Contentid:     1,
		})
		if err != nil {
			t.Fatalf("got an unexpected error")
		}
		testutils.AssertLogMessage(t, logfile, expectedWarning)
	})

	t.Run("Returns error if valid port is not provided", func(t *testing.T) {
		defer resetCLIVars()
		expectedError := "invalid port has been provided for segment with hostname sdw1 and data_directory /tmp/demo/gpseg1"
		err := gpctl_cli.ValidateSegment(&idl.Segment{
			HostName:      "sdw1",
			HostAddress:   "sdw1",
			Port:          -189,
			DataDirectory: "/tmp/demo/gpseg1",
			Dbid:          2,
			Contentid:     1,
		})
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})

	t.Run("Returns error if valid dataDir is not provided", func(t *testing.T) {
		defer resetCLIVars()
		expectedError := "data_directory has not been provided for segment with hostname sdw1 and port 5000"
		err := gpctl_cli.ValidateSegment(&idl.Segment{
			HostName:      "sdw1",
			HostAddress:   "sdw1",
			Port:          5000,
			DataDirectory: "",
			Dbid:          2,
			Contentid:     1,
		})
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
}

func TestGetSystemLocaleFn(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	t.Run("returns error if locale command fails", func(t *testing.T) {
		defer resetCLIVars()
		utils.System.ExecCommand = exectest.NewCommand(CommandFailure)
		defer utils.ResetSystemFunctions()
		expectedError := "failed to get locale on this system:"
		_, err := gpctl_cli.GetSystemLocale()
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("succeeds if locale command succeeds", func(t *testing.T) {
		defer resetCLIVars()
		utils.System.ExecCommand = exectest.NewCommand(CommandSuccess)
		defer utils.ResetSystemFunctions()
		_, err := gpctl_cli.GetSystemLocale()
		if err != nil {
			t.Fatalf("got an unexpected error")
		}
	})
}

func TestSetDefaultLocaleFn(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	locale := &idl.Locale{}
	t.Run("returns error if GetSystemLocale fails", func(t *testing.T) {
		defer resetCLIVars()
		expectedError := "failure"
		gpctl_cli.GetSystemLocale = func() ([]byte, error) {
			return []byte(""), fmt.Errorf("failure")
		}
		err := gpctl_cli.SetDefaultLocale(locale)
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
	t.Run("succeeds if GetSystemLocale gives output successfully", func(t *testing.T) {
		defer resetCLIVars()
		gpctl_cli.GetSystemLocale = func() ([]byte, error) {
			return []byte("LANG=\"\"\nLC_COLLATE=\"C\"\nLC_CTYPE=\"en_US.UTF-8\"\nLC_MESSAGES=\"C\"\nLC_MONETARY=\"C\"\nLC_NUMERIC=\"C\"\nLC_TIME=\"C\"\nLC_ALL="), nil
		}
		err := gpctl_cli.SetDefaultLocale(locale)
		if err != nil {
			t.Fatalf("got an unexpected error")
		}
	})
}

func CommandSuccess() {
	os.Stdout.WriteString("Success")
	os.Exit(0)
}

func CommandFailure() {
	os.Stderr.WriteString("failure")
	os.Exit(1)
}

func TestIsGpServicesEnabledFn(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	coordinator := &idl.Segment{
		HostAddress:   "cdw",
		HostName:      "cdw",
		Port:          700,
		DataDirectory: "/tmp/coordinator/",
	}
	gparray := idl.GpArray{
		Coordinator: coordinator,
		SegmentArray: []*idl.SegmentPair{
			{
				Primary: &idl.Segment{
					HostAddress:   "sdw1",
					HostName:      "sdw1",
					Port:          7002,
					DataDirectory: "/tmp/demo/1",
				},
			},
			{
				Primary: &idl.Segment{
					HostAddress:   "sdw1",
					HostName:      "sdw1",
					Port:          7003,
					DataDirectory: "/tmp/demo/2",
				},
			},
			{
				Primary: &idl.Segment{
					HostAddress:   "sdw2",
					HostName:      "sdw2",
					Port:          7004,
					DataDirectory: "/tmp/demo/3",
				},
			},
			{
				Primary: &idl.Segment{
					HostAddress:   "sdw2",
					HostName:      "sdw2",
					Port:          7005,
					DataDirectory: "/tmp/demo/4",
				},
			},
		},
	}
	t.Run("fails if some of hosts do not have gp services configured", func(t *testing.T) {
		defer resetCLIVars()
		defer resetConfHostnames()
		gpserviceCLI.Conf.Hostnames = []string{"cdw", "sdw1"}
		expectedError := "following hostnames [sdw2] do not have gp services configured. Please configure the services"
		err := gpctl_cli.IsGpServicesEnabled(&idl.MakeClusterRequest{GpArray: &gparray})
		if err == nil || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("got %v, want %v", err, expectedError)
		}
	})
}

func TestInitCleanFn(t *testing.T) {

	t.Run("ConnectToHub Fails ", func(t *testing.T) {

		defer resetCLIVars()

		expectedErr := errors.New("test error")
		gpserviceCLI.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			return nil, expectedErr
		}

		err := gpctl_cli.InitCleanFn(false)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("got %#v, want %#v", err, expectedErr)
		}
	})

	t.Run("CleanInitCluster RPC succeeds", func(t *testing.T) {
		setupTest(t)
		defer teardownTest()
		defer resetCLIVars()

		_, _, logfile := testhelper.SetupTestLogger()

		gpserviceCLI.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().CleanInitCluster(gomock.Any(), gomock.Any())
			return hubClient, nil
		}

		err := gpctl_cli.InitCleanFn(false)

		if err != nil {
			t.Fatalf("unexpected error: err %v", err)
		}
		testutils.AssertLogMessage(t, logfile, "clean cluster command successful")
	})

	t.Run("CleanInitCluster RPC fails", func(t *testing.T) {
		setupTest(t)
		defer teardownTest()

		expectedErr := errors.New("clean cluster command failed: test_err")
		gpserviceCLI.ConnectToHub = func(conf *hub.Config) (idl.HubClient, error) {
			hubClient := mock_idl.NewMockHubClient(ctrl)
			hubClient.EXPECT().CleanInitCluster(gomock.Any(), gomock.Any()).Return(nil, expectedErr)
			return hubClient, nil
		}

		err := gpctl_cli.InitCleanFn(false)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("got %#v, want %#v", err, expectedErr)
		}

		expectedErrPrefix := "clean cluster command failed"
		if !strings.HasPrefix(err.Error(), expectedErrPrefix) {
			t.Fatalf("got %#v, want %#v", err, expectedErr)
		}
	})
}
