package cli

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
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
type SegmentPair struct {
	Primary Segment
	Mirror  Segment
}

type InitConfig struct {
	DbName                 string            `mapstructure:"db-name"`
	Encoding               string            `mapstructure:"encoding"`
	HbaHostnames           bool              `mapstructure:"hba-hostnames"`
	DataChecksums          bool              `mapstructure:"data-checksums"`
	SuPassword             string            `mapstructure:"su-password"` //TODO set to default if not provided
	Locale                 Locale            `mapstructure:"locale"`
	CommonConfig           map[string]string `mapstructure:"common-config"`
	CoordinatorConfig      map[string]string `mapstructure:"coordinator-config"`
	SegmentConfig          map[string]string `mapstructure:"segment-config"`
	Coordinator            Segment           `mapstructure:"coordinator"`
	PrimarySegmentsArray   []Segment         `mapstructure:"primary-segments-array"`
	PrimaryBasePort        int               `mapstructure:"primary-base-port"`
	PrimaryDataDirectories []string          `mapstructure:"primary-data-directories"`
	HostList               []string          `mapstructure:"hostlist"`
	MirrorDataDirectories  []string          `mapstructure:"mirror-data-directories"`
	MirrorBasePort         int               `mapstructure:"mirror-base-port"`
	MirroringType          string            `mapstructure:"mirroring-type"`
}

var (
	InitClusterService                   = InitClusterServiceFn
	LoadInputConfigToIdl                 = LoadInputConfigToIdlFn
	ValidateInputConfigAndSetDefaults    = ValidateInputConfigAndSetDefaultsFn
	CheckForDuplicatPortAndDataDirectory = CheckForDuplicatePortAndDataDirectoryFn
	ParseStreamResponse                  = ParseStreamResponseFn
	GetSystemLocale                      = GetSystemLocaleFn
	SetDefaultLocale                     = SetDefaultLocaleFn
	IsGpServicesEnabled                  = IsGpServicesEnabledFn
)
var cliForceFlag bool
var containsMirror bool

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

	// Check if expansion is required

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
	// Expansion config parameters
	/*
		cliHandler.SetDefault("hostlist", []string{})
		cliHandler.SetDefault("primary-data-directories", []string{})
		cliHandler.SetDefault("primary-base-port", 0)
		cliHandler.SetDefault("mirror-data-directories", []string{})
		cliHandler.SetDefault("mirror-base-port", 0)
		cliHandler.SetDefault("mirroring-type", "")
	*/
	if err := cliHandler.ReadInConfig(); err != nil {
		return &idl.MakeClusterRequest{}, fmt.Errorf("while reading config file: %w", err)
	}

	var config InitConfig
	if err := cliHandler.UnmarshalExact(&config); err != nil {
		return &idl.MakeClusterRequest{}, fmt.Errorf("while unmarshaling config file: %w", err)
	}
	if AnyExpansionConfigPresent(cliHandler) {
		// Validate expansion config
		err := ValidateExpansionConfigAndSetDefault(&config, cliHandler)
		if err != nil {
			return &idl.MakeClusterRequest{}, err
		}
		//Expand details to config for primary
		ExpandPrimaryArray(config)
		if AnyExpansionMirrorConfigPresent(cliHandler) {
			// Expand configuration for mirror
		}
	}

	return CreateMakeClusterReq(&config, force, verbose), nil
}

func AnyExpansionConfigPresent(cliHandle *viper.Viper) bool {
	expansionKeys := []string{"hostlist", "primary-base-port", "primary-data-directories", "mirroring-type", "mirror-base-port", "mirror-data-directories"}
	for _, key := range expansionKeys {
		if cliHandle.IsSet(key) {
			return true
		}
	}
	return false
}

func AnyExpansionMirrorConfigPresent(cliHandle *viper.Viper) bool {
	expansionKeys := []string{"mirroring-type", "mirror-base-port", "mirror-data-directories"}
	for _, key := range expansionKeys {
		if cliHandle.IsSet(key) {
			return true
		}
	}
	return false
}
func ValidateExpansionConfigAndSetDefault(config *InitConfig, cliHandle *viper.Viper) error {
	// Check if mandatory primary expansion parameters are provided
	if len(config.PrimaryDataDirectories) < 1 {
		strErr := fmt.Sprintf("Empty primary-data-direcotiers provided")
		gplog.Error(strErr)
		return fmt.Errorf(strErr)
	}
	if len(config.HostList) < 1 {
		strErr := fmt.Sprintf("Empty hostlist provided")
		gplog.Error(strErr)
		return fmt.Errorf(strErr)
	}
	if config.PrimaryBasePort < 1 {
		defaultPrimaryBasePort := config.Coordinator.Port + 2
		gplog.Warn("No primary-base-port value provided. Setting default to:%d", defaultPrimaryBasePort)
		config.PrimaryBasePort = defaultPrimaryBasePort
	}

	// Check if mandatory mirror expansion parameters are provided
	if AnyExpansionMirrorConfigPresent(cliHandle) {
		containsMirror = true
		if len(config.PrimaryDataDirectories) != len(config.MirrorDataDirectories) {
			strErr := "number of primary-data-directories should be equal to number of mirror-data-directories"
			gplog.Error(strErr)
			return fmt.Errorf(strErr)
		}
		if config.MirrorBasePort < 1 {
			defaultMirrorBasePort := config.Coordinator.Port + 1002
			gplog.Warn("No primary-base-port value provided. Setting default to:%d", defaultMirrorBasePort)
			config.MirrorBasePort = defaultMirrorBasePort
		}

		if config.MirroringType == "" {
			// Default is spread mirroring
			config.MirroringType = "spread"
		} else if strings.ToLower(config.MirroringType) != "spread" && strings.ToLower(config.MirroringType) != "group" {
			strErr := fmt.Sprintf("mirroring-Type: %s is not supported. Only 'group' or 'spread' mirroring is supported",
				config.MirroringType)
			gplog.Error(strErr)
			return fmt.Errorf(strErr)
		}
		config.MirroringType = strings.ToLower(config.MirroringType)
	} else {
		// Mirror-less configuration
		strErr := fmt.Sprintf("No mirror-data-direcotiers provided. Will create a mirrorless cluster")
		gplog.Warn(strErr)
		containsMirror = false
	}

	// If provided expansion config wand primary/mirrors array is also defined
	if len(config.PrimarySegmentsArray) > 0 {
		strErr := "segments-array list should be empty when configuration contains primary-base-directories and hostlist"
		gplog.Error(strErr)
		return fmt.Errorf(strErr)
	}

	// TODO Check related to multi-homing
	return nil
}

func ExpandPrimaryArray(config InitConfig) []SegmentPair {
	segPairList := []SegmentPair{}

	/*
		hostNameAddressMap := make(map[string][]string)
		hostAddressNameMap := make(map[string]string)
		// Get hostname for the given hostlist
		for _, address := range config.HostList {
			err := ExecuteRPC(s.Conns, validateFn)
			if err != nil {
				return err
			}

			for _, msg := range replies {
				stream.StreamLogMsg(msg.Message, msg.Level)
			}
		}
	*/
	// TODO check if multi-home

	// non-multi-home
	segNum := 0
	for _, hostAddress := range config.HostList {
		for segIdx, directory := range config.PrimaryDataDirectories {
			seg := Segment{
				Hostname:      hostAddress, //TODO should be hostname
				Address:       hostAddress,
				Port:          config.PrimaryBasePort + segIdx,
				DataDirectory: filepath.Join(directory, fmt.Sprintf("gpseg-%d", segNum)),
			}
			segPairList = append(segPairList, SegmentPair{Primary: seg})
		}
	}
	if containsMirror {
		// Configure mirrors
		hostListLen := len(config.HostList)
		if config.MirroringType == "group" {
			// Perform group mirroring
			segNum = 0
			for hostIdx, _ := range config.HostList {
				for segIdx, directory := range config.MirrorDataDirectories {
					hostAddress := config.HostList[(hostIdx+1)%hostListLen]
					seg := Segment{
						Hostname:      hostAddress,
						Address:       hostAddress,
						Port:          config.MirrorBasePort + segIdx,
						DataDirectory: filepath.Join(directory, fmt.Sprintf("gpseg-%d", segNum)),
					}
					segPairList[segNum].Mirror = seg
					segNum++
				}
			}

		} else {
			// Perform spread mirroring
			segmentsPerHost := len(config.MirrorDataDirectories)
			segNum := 0
			for hostIndex, _ := range config.HostList {
				mirrorHostIndex := (hostIndex + 1) % hostListLen
				for localSeg := 0; localSeg < segmentsPerHost; localSeg++ {
					hostAddress := config.HostList[mirrorHostIndex]
					seg := Segment{
						Hostname:      hostAddress, //TODO fetch name from address
						Address:       hostAddress,
						Port:          config.MirrorBasePort + localSeg,
						DataDirectory: filepath.Join(config.MirrorDataDirectories[localSeg], fmt.Sprintf("gpseg-%d", segNum)),
					}
					segPairList[segNum].Mirror = seg
					segNum++
					mirrorHostIndex = (mirrorHostIndex + 1) % hostListLen
				}
			}
		}
	}
	// TODO Populate segPairList to config
	return segPairList
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
	// validate details of coordinator
	err := ValidateSegment(request.GpArray.Coordinator)
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

	// check for conflicting port and data-dir on a host
	err = CheckForDuplicatPortAndDataDirectory(request.GpArray.Primaries)
	if err != nil {
		return err
	}

	// check if gp services enabled on hosts
	err = IsGpServicesEnabled(request.GpArray)
	if err != nil {
		return err
	}

	if request.ClusterParams.Encoding == "" {
		gplog.Info(fmt.Sprintf("Could not find encoding in cluster config, defaulting to %v", constants.DefaultEncoding))
		request.ClusterParams.Encoding = "UTF-8"
	}

	if request.ClusterParams.Encoding == "SQL_ASCII" {
		return fmt.Errorf("SQL_ASCII is no longer supported as a server encoding")
	}

	// Validate max_connections
	err = ValidateMaxConnections(request.ClusterParams)
	if err != nil {
		return err
	}

	// if shared_buffers not provided in config then set the COORDINATOR_SHARED_BUFFERS and QE_SHARED_BUFFERS to DEFAULT_BUFFERS (128000 kB)
	CheckAndSetDefaultConfigParams(request.ClusterParams, "shared_buffers", constants.DefaultBuffer)

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
		//TODO Call RPC to get the hostname from hostAddress and populate here as segment.HostName
		return fmt.Errorf("hostName has not been provided for the segment with port %v and data_directory %v", segment.Port, segment.DataDirectory)
	} else if segment.HostAddress == "" {
		segment.HostAddress = segment.HostName
		gplog.Warn("hostAddress has not been provided, populating it with same as hostName %v for the segment with port %v and data_directory %v", segment.HostName, segment.Port, segment.DataDirectory)
	}

	if segment.Port <= 0 {
		return fmt.Errorf("invalid port has been provided for segment with hostname %v and data_directory %v", segment.HostName, segment.DataDirectory)
	}

	if segment.DataDirectory == "" {
		return fmt.Errorf("data_directory has not been provided for segment with hostname %v and port %v", segment.HostName, segment.Port)
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
		if _, ok := hostToDataDirectory[primary.HostName][primary.DataDirectory]; ok {
			return fmt.Errorf("duplicate data directory entry %v found for host %v", primary.DataDirectory, primary.HostAddress)
		}
		hostToDataDirectory[primary.HostName][primary.DataDirectory] = true

		// Check for port
		if _, ok := hostToPort[primary.HostAddress]; !ok {
			hostToPort[primary.HostAddress] = make(map[int32]bool)
		}
		if _, ok := hostToPort[primary.HostName][primary.Port]; ok {
			return fmt.Errorf("duplicate port entry %v found for host %v", primary.Port, primary.HostName)
		}
		hostToPort[primary.HostAddress][primary.Port] = true
	}
	return nil
}

/*
GetSystemLocaleFn returns system locales
*/
func GetSystemLocaleFn() ([]byte, error) {
	cmd := utils.System.ExecCommand("/usr/bin/locale")
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
	err = v.ReadConfig(bytes.NewBuffer(systemLocale))
	if err != nil {
		return err
	}

	locale.LcAll = strings.Trim(v.GetString("LC_ALL"), "\"")
	locale.LcCollate = strings.Trim(v.GetString("LC_COLLATE"), "\"")
	locale.LcCtype = strings.Trim(v.GetString("LC_CTYPE"), "\"")
	locale.LcMessages = strings.Trim(v.GetString("LC_MESSAGES"), "\"")
	locale.LcMonetory = strings.Trim(v.GetString("LC_MONETARY"), "\"")
	locale.LcNumeric = strings.Trim(v.GetString("LC_NUMERIC"), "\"")
	locale.LcTime = strings.Trim(v.GetString("LC_TIME"), "\"")

	return nil
}

/*
IsGpServicesEnabledFn returns error if any of the hosts from config does not have gp services enabled
*/
func IsGpServicesEnabledFn(gpArray *idl.GpArray) error {
	hostnames = []string{}
	hostnames = append(hostnames, gpArray.Coordinator.HostName)
	for _, seg := range gpArray.Primaries {
		hostnames = append(hostnames, seg.HostName)
	}

	diff := utils.GetListDifference(hostnames, Conf.Hostnames)
	if len(diff) != 0 {
		return fmt.Errorf("following hostnames %s do not have gp services configured. Please configure the services", diff)
	}
	return nil
}

/*
ValidateMaxConnections sets the default value of max_connections if not provided in config. Also returns error if valid value is not provided
if max_connections not defined in CommonConfig set it to default value
if max_connections not defined in CoordinatorConfig set to CommonConfig value
if max_connections not defined in SegmentConfig set it to strconv.Atoi(clusterParams.CommonConfig["max_connections"])*constants.QeConnectFactor
*/
func ValidateMaxConnections(clusterParams *idl.ClusterParams) error {
	if _, ok := clusterParams.CommonConfig["max_connections"]; !ok {
		gplog.Info(" max_connections not set, will set to default value %v", constants.DefaultQdMaxConnect)
		clusterParams.CommonConfig["max_connections"] = strconv.Itoa(constants.DefaultQdMaxConnect)
	}

	if _, ok := clusterParams.CoordinatorConfig["max_connections"]; !ok {
		// Check if common-config has max-connections defined
		gplog.Info(" Coordinator max_connections not set, will set to value %v from CommonConfig", clusterParams.CommonConfig["max_connections"])
		clusterParams.CoordinatorConfig["max_connections"] = clusterParams.CommonConfig["max_connections"]
	}
	coordinatorMaxConnect, err := strconv.Atoi(clusterParams.CoordinatorConfig["max_connections"])
	if err != nil {
		return fmt.Errorf("invalid value %s for max_connections, must be an integer. error: %v",
			clusterParams.CoordinatorConfig["max_connections"], err)
	}

	if coordinatorMaxConnect < 1 {
		return fmt.Errorf("COORDINATOR max_connections value %d is too small. Should be more than 1. ", coordinatorMaxConnect)
	}

	// if max_connections not defined in SegmentConfig, set to commonConfigMaxConnections*QeConnectFactor
	if _, ok := clusterParams.SegmentConfig["max_connections"]; !ok {
		maxConnections, err := strconv.Atoi(clusterParams.CommonConfig["max_connections"])
		if err != nil {
			return fmt.Errorf("invalid value %s for max_connections, must be an integer. error: %v",
				clusterParams.CommonConfig["max_connections"], err)
		}
		segmentConfigMaxConnections := maxConnections * constants.QeConnectFactor
		gplog.Info(" Segment max_connections not set, will set to value %v", segmentConfigMaxConnections)
		clusterParams.SegmentConfig["max_connections"] = strconv.Itoa(segmentConfigMaxConnections)
	}
	return nil
}

/*
CheckAndSetDefaultConfigParams sets the default value for parameters not defined in config
if configParam is not defined in CommonConfig , the value will be set to defaultValue provided
if configParam is not defined in CoordinatorConfig or SegmentConfig, the value will be set to same as configParam from CommonConfig
*/
func CheckAndSetDefaultConfigParams(clusterParams *idl.ClusterParams, configParam string, defaultValue string) {
	if _, ok := clusterParams.CommonConfig[configParam]; !ok {
		gplog.Info(fmt.Sprintf("%v is not set in CommonConfig, will set to default value %v", configParam, defaultValue))
		clusterParams.CommonConfig[configParam] = defaultValue
	}

	if _, ok := clusterParams.CoordinatorConfig[configParam]; !ok {
		// Check if common-config has configParam defined
		gplog.Info(" Coordinator %v not set, will set to value %v from CommonConfig", configParam, clusterParams.CommonConfig[configParam])
		clusterParams.CoordinatorConfig[configParam] = clusterParams.CommonConfig[configParam]
	}
	if _, ok := clusterParams.SegmentConfig[configParam]; !ok {
		// Check if common-config has configParam defined
		gplog.Info(" Segment %v not set, will set to value %v from CommonConfig", configParam, clusterParams.CommonConfig[configParam])
		clusterParams.SegmentConfig[configParam] = clusterParams.CommonConfig[configParam]
	}
}
