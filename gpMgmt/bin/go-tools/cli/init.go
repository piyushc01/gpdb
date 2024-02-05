package cli

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/constants"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Locale struct {
	LcAll      string `mapstructure:"lc-all"`
	LcCollate  string `mapstructure:"lc-collate"`
	LcCtype    string `mapstructure:"lc-ctype"`
	LcMessages string `mapstructure:"lc-messages"`
	LcMonetary string `mapstructure:"lc-monetary"`
	LcNumeric  string `mapstructure:"lc-numeric"`
	LcTime     string `mapstructure:"lc-time"`
}

type Segment struct {
	Hostname      string `mapstructure:"hostname"`
	Address       string `mapstructure:"address"`
	Port          int    `mapstructure:"port"`
	DataDirectory string `mapstructure:"data-directory"`
}

type InitConfig struct {
	DbName               string            `mapstructure:"db-name"`
	Encoding             string            `mapstructure:"encoding"`
	HbaHostnames         bool              `mapstructure:"hba-hostnames"`
	DataChecksums        bool              `mapstructure:"data-checksums"`
	SuPassword           string            `mapstructure:"su-password"` //TODO set to default if not provided
	Locale               Locale            `mapstructure:"locale"`
	CommonConfig         map[string]string `mapstructure:"common-config"`
	CoordinatorConfig    map[string]string `mapstructure:"coordinator-config"`
	SegmentConfig        map[string]string `mapstructure:"segment-config"`
	Coordinator          Segment           `mapstructure:"coordinator"`
	PrimarySegmentsArray []Segment         `mapstructure:"primary-segments-array"`
}

var (
	InitClusterService                   = InitClusterServiceFn
	LoadInputConfigToIdl                 = LoadInputConfigToIdlFn
	ValidateInputConfigAndSetDefaults    = ValidateInputConfigAndSetDefaultsFn
	CheckForDuplicatPortAndDataDirectory = CheckForDuplicatePortAndDataDirectoryFn
	ParseStreamResponse                  = ParseStreamResponseFn
	GetSystemLocale                      = GetSystemLocaleFn
	SetDefaultLocale                     = SetDefaultLocaleFn
)
var cliForceFlag bool

// initCmd adds support for command "gp init <config-file> [--force]
func initCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:     "init",
		Short:   "Initialize cluster, segments",
		PreRunE: InitializeCommand,
		RunE:    RunInitClusterCmd,
	}
	initCmd.PersistentFlags().BoolVar(&cliForceFlag, "force", false, "Create cluster forcefully by overwriting existing directories")
	initCmd.AddCommand(initClusterCmd())
	return initCmd
}

// initClusterCmd adds support for command "gp init cluster <config-file> [--force]
func initClusterCmd() *cobra.Command {
	initClusterCmd := &cobra.Command{
		Use:     "cluster",
		Short:   "Initialize the cluster",
		PreRunE: InitializeCommand,
		RunE:    RunInitClusterCmd,
	}

	return initClusterCmd
}

// RunInitClusterCmd driving function gets called from cobra on gp init cluster command
func RunInitClusterCmd(cmd *cobra.Command, args []string) error {
	// initial basic cli validations
	if len(args) == 0 {
		return fmt.Errorf("please provide config file for cluster initialization")
	}
	if len(args) > 1 {
		return fmt.Errorf("more arguments than expected")
	}

	// Call for further input config validation and cluster creation
	err := InitClusterService(args[0], cliForceFlag, Verbose)
	if err != nil {
		return err
	}
	gplog.Info("Cluster initialized successfully")

	return nil
}

/*
InitClusterServiceFn does input config file validation followed by actual cluster creation
*/
func InitClusterServiceFn(inputConfigFile string, force, verbose bool) error {
	if _, err := utils.System.Stat(inputConfigFile); err != nil {
		return err
	}
	// Viper instance to read the input config
	cliHandler := viper.New()

	// Load cluster-request from the config file
	clusterReq, err := LoadInputConfigToIdl(inputConfigFile, cliHandler, force, verbose)
	if err != nil {
		return err
	}

	// Validate give input configuration
	if err := ValidateInputConfigAndSetDefaults(clusterReq, cliHandler); err != nil {
		return err
	}

	// Make call to MakeCluster RPC and wait for results
	client, err := ConnectToHub(Conf)
	if err != nil {
		return err
	}

	// Call RPC on Hub to create the cluster
	stream, err := client.MakeCluster(context.Background(), clusterReq)
	if err != nil {
		return utils.FormatGrpcError(err)
	}

	err = ParseStreamResponse(stream)
	if err != nil {
		return err
	}

	return nil
}

/*
LoadInputConfigToIdlFn reads config file and populates RPC IDL request structure
*/
func LoadInputConfigToIdlFn(inputConfigFile string, cliHandler *viper.Viper, force bool, verbose bool) (*idl.MakeClusterRequest, error) {
	cliHandler.SetConfigFile(inputConfigFile)

	cliHandler.SetDefault("common-config", make(map[string]string))
	cliHandler.SetDefault("coordinator-config", make(map[string]string))
	cliHandler.SetDefault("segment-config", make(map[string]string))
	cliHandler.SetDefault("data-checksums", true)

	if err := cliHandler.ReadInConfig(); err != nil {
		return &idl.MakeClusterRequest{}, fmt.Errorf("while reading config file: %w", err)
	}

	var config InitConfig
	if err := cliHandler.UnmarshalExact(&config); err != nil {
		return &idl.MakeClusterRequest{}, fmt.Errorf("while unmarshaling config file: %w", err)
	}

	return CreateMakeClusterReq(&config, force, verbose), nil
}

/*
CreateMakeClusterReq helper function to populate cluster request from the config
*/
func CreateMakeClusterReq(config *InitConfig, forceFlag bool, verbose bool) *idl.MakeClusterRequest {
	var primarySegs []*idl.Segment
	for _, seg := range config.PrimarySegmentsArray {
		primarySegs = append(primarySegs, SegmentToIdl(seg))
	}

	return &idl.MakeClusterRequest{
		GpArray: &idl.GpArray{
			Coordinator: SegmentToIdl(config.Coordinator),
			Primaries:   primarySegs,
		},
		ClusterParams: ClusterParamsToIdl(config),
		ForceFlag:     forceFlag,
		Verbose:       verbose,
	}
}

func SegmentToIdl(seg Segment) *idl.Segment {
	return &idl.Segment{
		Port:          int32(seg.Port),
		DataDirectory: seg.DataDirectory,
		HostName:      seg.Hostname,
		HostAddress:   seg.Address,
	}
}

func ClusterParamsToIdl(config *InitConfig) *idl.ClusterParams {
	return &idl.ClusterParams{
		CoordinatorConfig: config.CoordinatorConfig,
		SegmentConfig:     config.SegmentConfig,
		CommonConfig:      config.CommonConfig,
		Locale: &idl.Locale{
			LcAll:      config.Locale.LcAll,
			LcCollate:  config.Locale.LcCollate,
			LcCtype:    config.Locale.LcCtype,
			LcMessages: config.Locale.LcMessages,
			LcMonetory: config.Locale.LcMonetary,
			LcNumeric:  config.Locale.LcNumeric,
			LcTime:     config.Locale.LcTime,
		},
		HbaHostnames:  config.HbaHostnames,
		Encoding:      config.Encoding,
		SuPassword:    config.SuPassword,
		DbName:        config.DbName,
		DataChecksums: config.DataChecksums,
	}
}

/*
ValidateInputConfigAndSetDefaultsFn performs various validation checks on the configuration
*/
func ValidateInputConfigAndSetDefaultsFn(request *idl.MakeClusterRequest, cliHandler *viper.Viper) error {
	//Check if coordinator details are provided
	if !cliHandler.IsSet("coordinator") {
		return fmt.Errorf("no coordinator segments are provided in input config file")
	}

	//Check if primary segment details are provided
	if !cliHandler.IsSet("primary-segments-array") {
		return fmt.Errorf("no primary segments are provided in input config file")
	}

	//Check if locale is provided, if not set it to system locale
	if !cliHandler.IsSet("locale") {
		gplog.Warn("locale is not provided, setting it to system locale")
		err := SetDefaultLocale(request.ClusterParams.Locale)
		if err != nil {
			return err
		}
	}
	// Check if length of Gparray.PimarySegments is 0
	if len(request.GpArray.Primaries) == 0 {
		return fmt.Errorf("no primary segments are provided in input config file")
	}

	hostnames = []string{}
	hostnames = append(hostnames, request.GpArray.Coordinator.HostName)
	for _, seg := range request.GpArray.Primaries {
		hostnames = append(hostnames, seg.HostName)
	}

	diff := utils.GetListDifference(hostnames, Conf.Hostnames)
	if len(diff) != 0 {
		return fmt.Errorf("following hostnames %s do not have gp services configured. Please configure the services", diff)
	}

	if request.ClusterParams.Encoding == "" {
		gplog.Info(fmt.Sprintf("Could not find encoding in cluster config, defaulting to %v", constants.DefaultEncoding))
		request.ClusterParams.Encoding = "UTF-8"
	}

	if request.ClusterParams.Encoding == "SQL_ASCII" {
		return fmt.Errorf("SQL_ASCII is no longer supported as a server encoding")
	}

	if _, ok := request.ClusterParams.CoordinatorConfig["max_connections"]; !ok {
		// Check if common-config has max-connections defined
		if _, ok := request.ClusterParams.CommonConfig["max_connections"]; !ok {
			gplog.Info("COORDINATOR max_connections not set, will set to default value %v", constants.DefaultQdMaxConnect)
			request.ClusterParams.CoordinatorConfig["max_connections"] = strconv.Itoa(constants.DefaultQdMaxConnect)
		} else {
			gplog.Info("COORDINATOR max_connections set to value: %s", request.ClusterParams.CommonConfig["max_connections"])
			request.ClusterParams.CoordinatorConfig["max_connections"] = request.ClusterParams.CommonConfig["max_connections"]
		}
	}
	coordinatorMaxConnect, err := strconv.Atoi(request.ClusterParams.CoordinatorConfig["max_connections"])
	if err != nil {
		return fmt.Errorf("invalid value %s for max_connections, must be an integer. error: %v",
			request.ClusterParams.CoordinatorConfig["max_connections"], err)
	}

	if coordinatorMaxConnect < 1 {
		return fmt.Errorf("COORDINATOR max_connections value %d is too small. Should be more than 1. ", coordinatorMaxConnect)
	}
	if _, ok := request.ClusterParams.SegmentConfig["max_connections"]; !ok {
		request.ClusterParams.SegmentConfig["max_connections"] = strconv.Itoa(coordinatorMaxConnect * constants.QeConnectFactor)
	}

	// check for shared_buffers if not provided in config then set the COORDINATOR_SHARED_BUFFERS and QE_SHARED_BUFFERS to DEFAULT_BUFFERS (128000 kB)
	if _, ok := request.ClusterParams.CommonConfig["shared_buffers"]; !ok {
		gplog.Info(fmt.Sprintf("shared_buffers is not set, will set to default value %v", constants.DefaultBuffer))
		request.ClusterParams.CommonConfig["shared_buffers"] = constants.DefaultBuffer
	}

	// check coordinator open file values
	out, err := utils.System.ExecCommand("ulimit", "-n").CombinedOutput()
	if err != nil {
		return err
	}

	coordinatorOpenFileLimit, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return fmt.Errorf("could not convert the ulimit value: %w", err)
	}

	if coordinatorOpenFileLimit < constants.OsOpenFiles {
		gplog.Warn(fmt.Sprintf("Coordinator open file limit is %d should be >= %d", coordinatorOpenFileLimit, constants.OsOpenFiles))
	}

	// validate details of coordinator
	err = ValidateSegment(request.GpArray.Coordinator)
	if err != nil {
		return err
	}

	// validate the details of primary segments
	for _, segment := range request.GpArray.Primaries {
		err = ValidateSegment(segment)
		if err != nil {
			return err
		}
	}

	err = CheckForDuplicatPortAndDataDirectory(request.GpArray.Primaries)
	if err != nil {
		return err
	}
	return nil
}

/*
ValidateSegment checks if valid values have been provided for the segment hostname, address, port and data-directory.
If both hostname and address are not provided then the function returns an error.
If one of hostname or address is not provided it is populated with other non-empty value.
*/
func ValidateSegment(segment *idl.Segment) error {
	if segment.HostName == "" && segment.HostAddress == "" {
		return fmt.Errorf("neither hostName nor hostAddress is provided for the segment with port %v and data_directory %v", segment.Port, segment.DataDirectory)
	} else if segment.HostName == "" {
		segment.HostName = segment.HostAddress
		gplog.Warn("hostName has not been provided, populating it with same as hostAddress %v for the segment with port %v and data_directory %v", segment.HostAddress, segment.Port, segment.DataDirectory)
	} else if segment.HostAddress == "" {
		segment.HostAddress = segment.HostName
		gplog.Warn("hostAddress has not been provided, populating it with same as hostName %v for the segment with port %v and data_directory %v", segment.HostName, segment.Port, segment.DataDirectory)
	}

	if segment.Port <= 0 {
		return fmt.Errorf("invalid port has been provided for segment with hostname %v and data_directory %v", segment.HostName, segment.DataDirectory)
	}

	if segment.DataDirectory == "" {
		return fmt.Errorf("data_directory has not been provided for segment with hostname %v and data_directory %v", segment.HostName, segment.DataDirectory)
	}
	return nil
}

/*
CheckForDuplicatePortAndDataDirectoryFn checks for duplicate data-directories and ports on host.
In case of data-directories, look for unique host-names.
For checking duplicate port, checking if address is unique. A host can use same the port for a different address.
*/
func CheckForDuplicatePortAndDataDirectoryFn(primaries []*idl.Segment) error {
	hostToDataDirectory := make(map[string]map[string]bool)
	hostToPort := make(map[string]map[int32]bool)
	for _, primary := range primaries {
		//Check for data-directory
		if _, ok := hostToDataDirectory[primary.HostName]; !ok {
			hostToDataDirectory[primary.HostName] = make(map[string]bool)
		}
		if hostToDataDirectory[primary.HostName][primary.DataDirectory] {
			return fmt.Errorf("duplicate data directory entry %v found for host %v", primary.DataDirectory, primary.HostAddress)
		}
		hostToDataDirectory[primary.HostAddress][primary.DataDirectory] = true

		// Check for port
		if _, ok := hostToPort[primary.HostAddress]; !ok {
			hostToPort[primary.HostAddress] = make(map[int32]bool)
		}
		if hostToPort[primary.HostName][primary.Port] {
			return fmt.Errorf("duplicate port entry %v found for host %v", primary.Port, primary.HostName)
		}
		hostToPort[primary.HostName][primary.Port] = true
	}
	return nil
}

/*
GetSystemLocaleFn returns system locales
*/
func GetSystemLocaleFn() ([]byte, error) {
	cmd := utils.System.ExecCommand(fmt.Sprintf("/usr/bin/locale"))
	output, err := cmd.Output()

	if err != nil {
		return []byte(""), fmt.Errorf("failed to get locale on this system: %w", err)
	}

	return output, nil
}

/*
SetDefaultLocaleFn populates the locale struct with system locales
*/
func SetDefaultLocaleFn(locale *idl.Locale) error {
	systemLocale, err := GetSystemLocale()
	if err != nil {
		return err
	}
	v := viper.New()
	v.SetConfigType("properties")
	v.ReadConfig(bytes.NewBuffer(systemLocale))

	locale.LcAll = strings.Trim(v.GetString("LC_ALL"), "\"")
	locale.LcCollate = strings.Trim(v.GetString("LC_COLLATE"), "\"")
	locale.LcCtype = strings.Trim(v.GetString("LC_CTYPE"), "\"")
	locale.LcMessages = strings.Trim(v.GetString("LC_MESSAGES"), "\"")
	locale.LcMonetory = strings.Trim(v.GetString("LC_MONETARY"), "\"")
	locale.LcNumeric = strings.Trim(v.GetString("LC_NUMERIC"), "\"")
	locale.LcTime = strings.Trim(v.GetString("LC_TIME"), "\"")

	return nil
}
