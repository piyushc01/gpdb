package cli

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/constants"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
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
	Primary *Segment `mapstructure:"primary"`
	Mirror  *Segment `mapstructure:"mirror"`
}

type InitConfig struct {
	DbName            string            `mapstructure:"db-name"`
	Encoding          string            `mapstructure:"encoding"`
	HbaHostnames      bool              `mapstructure:"hba-hostnames"`
	DataChecksums     bool              `mapstructure:"data-checksums"`
	SuPassword        string            `mapstructure:"su-password"` //TODO set to default if not provided
	Locale            Locale            `mapstructure:"locale"`
	CommonConfig      map[string]string `mapstructure:"common-config"`
	CoordinatorConfig map[string]string `mapstructure:"coordinator-config"`
	SegmentConfig     map[string]string `mapstructure:"segment-config"`
	Coordinator       Segment           `mapstructure:"coordinator"`
	SegmentArray      []SegmentPair     `mapstructure:"segment-array"`

	//Expansion config parameters
	//PrimarySegmentsArray   []Segment `mapstructure:"primary-segments-array"`
	PrimaryBasePort        int      `mapstructure:"primary-base-port"`
	PrimaryDataDirectories []string `mapstructure:"primary-data-directories"`
	HostList               []string `mapstructure:"hostlist"`
	MirrorDataDirectories  []string `mapstructure:"mirror-data-directories"`
	MirrorBasePort         int      `mapstructure:"mirror-base-port"`
	MirroringType          string   `mapstructure:"mirroring-type"`
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
var ContainsMirror bool
var HubClient idl.HubClient

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
	_, err := utils.System.Stat(inputConfigFile)
	if err != nil {
		return err
	}
	// Viper instance to read the input config
	cliHandler := viper.New()

	// Make call to MakeCluster RPC and wait for results
	HubClient, err = ConnectToHub(Conf)
	if err != nil {
		return err
	}

	// Load cluster-request from the config file
	clusterReq, err := LoadInputConfigToIdl(inputConfigFile, cliHandler, force, verbose)
	if err != nil {
		return err
	}

	// Validate give input configuration
	if err := ValidateInputConfigAndSetDefaults(clusterReq, cliHandler); err != nil {
		return err
	}

	// Call RPC on Hub to create the cluster
	stream, err := HubClient.MakeCluster(context.Background(), clusterReq)
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

	if AnyExpansionConfigPresent(cliHandler) {
		// Validate expansion config
		err := ValidateExpansionConfigAndSetDefault(&config, cliHandler)
		if err != nil {
			return &idl.MakeClusterRequest{}, err
		}

		// Check for multi-home
		isMultHome, NameAddressMap, AddressNameMap, err := IsMultiHome(config.HostList)
		if err != nil {
			gplog.Error("multihome detection failed, error:%v", err)
			return &idl.MakeClusterRequest{}, err
		}
		//Expand details to config for primary
		segmentPairArray := ExpandSegPairArray(config, isMultHome, NameAddressMap, AddressNameMap)
		config.SegmentArray = segmentPairArray
	}

	return CreateMakeClusterReq(&config, force, verbose), nil
}

// IsMultiHome checks if it is a multi-home environment
// Makes a call to resolve all addresses to hostname and returns a map of address vs hostnames
// In case of error, map will be empty, and returns false
func IsMultiHome(hostlist []string) (isMultiHome bool, NameAddress map[string][]string, AddressNameMap map[string]string, err error) {
	// get a list of hostnames against each address
	request := idl.GetAllHostNamesRequest{HostList: hostlist}
	reply, err := HubClient.GetAllHostNames(context.Background(), &request)
	if err != nil {
		gplog.Error("failed names of the host against address:%v", err)
		return false, nil, nil, err
	}
	isMultiHome = false
	NameAddressMap := make(map[string][]string)
	// convert address-name map to name-address map
	for address, hostname := range reply.HostNameMap {
		NameAddressMap[hostname] = append(NameAddressMap[hostname], address)
	}
	if len(hostlist) > len(NameAddressMap) {
		isMultiHome = true
	}
	return isMultiHome, NameAddressMap, reply.HostNameMap, nil
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
		strErr := "primary-data-directories not specified. Please specify primary-data-directories to continue"
		gplog.Error(strErr)
		return fmt.Errorf(strErr)
	}
	if len(config.HostList) < 1 {
		strErr := "hostlist not specified. Please specify hostlist to continue"
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
		ContainsMirror = true
		if len(config.PrimaryDataDirectories) != len(config.MirrorDataDirectories) {
			strErr := "number of primary-data-directories should be equal to number of mirror-data-directories"
			gplog.Error(strErr)
			return fmt.Errorf(strErr)
		}
		if config.MirrorBasePort < 1 {
			defaultMirrorBasePort := config.Coordinator.Port + 1002
			gplog.Warn("No mirror-base-port value provided. Setting default to:%d", defaultMirrorBasePort)
			config.MirrorBasePort = defaultMirrorBasePort
		}
		//TODO Check if spread mirroring, num hosts should be greater than number of primaries per host so that we can spread segments
		if config.MirroringType == constants.SpreadMirroringStr && len(config.MirrorDataDirectories) > len(config.HostList) {
			strErr := fmt.Sprintf("To enable spread mirroring, number of hosts should be more than number of primary segments per host. "+
				"Current number of hosts is: %d and number of primaries per host is:%d", len(config.HostList), len(config.MirrorDataDirectories))
			gplog.Error(strErr)
			return fmt.Errorf(strErr)
		}
		//TODO Check if the primary and mirror range overlaps

		if config.MirroringType == "" {
			// Default is group mirroring
			config.MirroringType = constants.GroupMirroringStr
			gplog.Warn("Mirroring type not specified. Setting default as 'group' mirroring")
		} else if strings.ToLower(config.MirroringType) != constants.SpreadMirroringStr && strings.ToLower(config.MirroringType) != constants.GroupMirroringStr {
			strErr := fmt.Sprintf("mirroring-Type: %s is not supported. Only 'group' or 'spread' mirroring is supported",
				config.MirroringType)
			gplog.Error(strErr)
			return fmt.Errorf(strErr)
		}
		config.MirroringType = strings.ToLower(config.MirroringType)
	} else {
		// Mirror-less configuration
		strErr := "No mirror-data-direcotiers provided. Will create mirrorless cluster"
		gplog.Warn(strErr)
		ContainsMirror = false
	}

	// If provided expansion config and primary/mirrors array is also defined
	if len(config.SegmentArray) > 0 {
		strErr := "segments-array list should be empty when configuration contains primary-base-directories and hostlist"
		gplog.Error(strErr)
		return fmt.Errorf(strErr)
	}

	// TODO Check related to multi-homing
	return nil
}

func ExpandNonMultiHomePrimaryList(segPairList *[]SegmentPair, PrimaryBasePort int, PrimaryDataDirectories []string, hostList []string, addressNameMap map[string]string) *[]SegmentPair {
	segNum := 0
	for _, hostAddress := range hostList {
		for segIdx, directory := range PrimaryDataDirectories {
			seg := Segment{
				Hostname:      addressNameMap[hostAddress],
				Address:       hostAddress,
				Port:          PrimaryBasePort + segIdx,
				DataDirectory: filepath.Join(directory, fmt.Sprintf("gpseg-%d", segNum)),
			}
			*segPairList = append(*segPairList, SegmentPair{Primary: &seg})
			segNum++
		}
	}
	return segPairList
}
func ExpandNonMultiHomeGroupMirrorList(segPairList *[]SegmentPair, mirrorBasePort int, mirrorDataDirectories []string, hostList []string, addressNameMap map[string]string) *[]SegmentPair {
	segNum := 0
	hostListLen := len(hostList)
	for hostIdx := range hostList {
		for segIdx, directory := range mirrorDataDirectories {
			hostAddress := hostList[(hostIdx+1)%hostListLen]
			seg := Segment{
				Hostname:      addressNameMap[hostAddress],
				Address:       hostAddress,
				Port:          mirrorBasePort + segIdx,
				DataDirectory: filepath.Join(directory, fmt.Sprintf("gpseg-%d", segNum)),
			}
			(*segPairList)[segNum].Mirror = &seg
			segNum++
		}
	}
	return segPairList
}

func ExpandNonMultiHomeSpreadMirroring(segPairList *[]SegmentPair, mirrorBasePort int, mirrorDataDirectories []string, hostList []string, addressNameMap map[string]string) *[]SegmentPair {
	segmentsPerHost := len(mirrorDataDirectories)
	hostListLen := len(hostList)
	segNum := 0
	for hostIndex := range hostList {
		mirrorHostIndex := (hostIndex + 1) % hostListLen
		for localSeg := 0; localSeg < segmentsPerHost; localSeg++ {
			hostAddress := hostList[mirrorHostIndex]
			seg := Segment{
				Hostname:      addressNameMap[hostAddress],
				Address:       hostAddress,
				Port:          mirrorBasePort + localSeg,
				DataDirectory: filepath.Join(mirrorDataDirectories[localSeg], fmt.Sprintf("gpseg-%d", segNum)),
			}
			(*segPairList)[segNum].Mirror = &seg
			segNum++
			mirrorHostIndex = (mirrorHostIndex + 1) % hostListLen
		}
	}
	return segPairList
}

func ExpandMultiHomePrimaryArray(segPairList *[]SegmentPair, primaryBasePort int, primaryDataDirectories []string, hostnameArray []string, nameAddressMap map[string][]string) *[]SegmentPair {
	segNum := 0
	for _, hostname := range hostnameArray {
		addressList := nameAddressMap[hostname]
		for idx, directory := range primaryDataDirectories {
			seg := Segment{
				Hostname:      hostname,
				Address:       addressList[idx%len(addressList)],
				Port:          primaryBasePort + idx,
				DataDirectory: filepath.Join(directory, fmt.Sprintf("gpseg-%d", segNum)),
			}
			*segPairList = append(*segPairList, SegmentPair{Primary: &seg})
			segNum++
		}
	}
	return segPairList
}

func ExpandMultiHomeGroupMirrorList(segPairList *[]SegmentPair, mirrorBasePort int, mirrorDataDirectories []string, hostnameArray []string, nameAddressMap map[string][]string) *[]SegmentPair {
	uniqueHostCount := len(hostnameArray)
	segNum := 0
	// Group mirroring
	for idx := 0; idx < uniqueHostCount; idx++ {
		hostname := hostnameArray[(idx+1)%uniqueHostCount]
		addressList := nameAddressMap[hostname]
		for segIdx, directory := range mirrorDataDirectories {
			seg := Segment{
				Hostname:      hostname,
				Address:       addressList[segIdx%len(addressList)],
				Port:          mirrorBasePort + segIdx,
				DataDirectory: filepath.Join(directory, fmt.Sprintf("gpseg-%d", segNum)),
			}
			(*segPairList)[segNum].Mirror = &seg
			segNum++
		}
	}
	return segPairList
}

func ExpandMultiHomeSpreadMirrorList(segPairList *[]SegmentPair, mirrorBasePort int, mirrorDataDirectories []string, hostnameArray []string, nameAddressMap map[string][]string) *[]SegmentPair {
	segNum := 0
	uniqueHostCount := len(hostnameArray)
	for hostnameIdx := range hostnameArray {
		for segIdx, directory := range mirrorDataDirectories {
			nxtHostIdx := (hostnameIdx + segIdx + 1) % uniqueHostCount
			// if the current hostname and mirror hostname is same, move to next
			if nxtHostIdx == hostnameIdx {
				nxtHostIdx = (nxtHostIdx + 1) % uniqueHostCount
			}
			nxtHostName := hostnameArray[nxtHostIdx]
			addressList := nameAddressMap[nxtHostName]
			addressCnt := len(addressList)
			seg := Segment{
				Address:       addressList[(hostnameIdx+segIdx)%addressCnt],
				Hostname:      nxtHostName,
				Port:          mirrorBasePort + segIdx,
				DataDirectory: filepath.Join(directory, fmt.Sprintf("gpseg-%d", segNum)),
			}
			(*segPairList)[segNum].Mirror = &seg
			segNum++
		}
	}
	return segPairList
}

/*
ExpandSegPairArray expands primary and mirror configuration from the given configuration
Returns an array of segmentPair to be updated in the MakeCluster request
*/
func ExpandSegPairArray(config InitConfig, multiHome bool, nameAddressMap map[string][]string, addressNameMap map[string]string) []SegmentPair {
	var segPairList []SegmentPair

	slices.Sort(config.HostList)

	if multiHome {
		// Build a list of unit hosts
		var hostnameArray []string
		for hostname := range nameAddressMap {
			hostnameArray = append(hostnameArray, hostname)
		}
		slices.Sort(hostnameArray)

		// Expand Primaries
		segPairList = *ExpandMultiHomePrimaryArray(&segPairList, config.PrimaryBasePort, config.PrimaryDataDirectories, hostnameArray, nameAddressMap)

		// Add mirrors to this expansion
		if ContainsMirror {
			if config.MirroringType == constants.GroupMirroringStr {
				segPairList = *ExpandMultiHomeGroupMirrorList(&segPairList, config.MirrorBasePort, config.MirrorDataDirectories, hostnameArray, nameAddressMap)
			} else {
				// Spread mirroring
				segPairList = *ExpandMultiHomeSpreadMirrorList(&segPairList, config.MirrorBasePort, config.MirrorDataDirectories, hostnameArray, nameAddressMap)
			}
		}
		return segPairList
	} else {
		// non-multi-home setup,
		// Create Primary segments
		segPairList = *ExpandNonMultiHomePrimaryList(&segPairList, config.PrimaryBasePort, config.PrimaryDataDirectories, config.HostList, addressNameMap)

		if ContainsMirror {
			if config.MirroringType == constants.GroupMirroringStr {
				// Perform group mirroring
				segPairList = *ExpandNonMultiHomeGroupMirrorList(&segPairList, config.MirrorBasePort, config.MirrorDataDirectories, config.HostList, addressNameMap)
			} else {
				// Perform spread mirroring
				segPairList = *ExpandNonMultiHomeSpreadMirroring(&segPairList, config.MirrorBasePort, config.MirrorDataDirectories, config.HostList, addressNameMap)
			}
		}
		//End of non-multi-home expansion
	}
	return segPairList
}

/*
CreateMakeClusterReq helper function to populate cluster request from the config
*/
func CreateMakeClusterReq(config *InitConfig, forceFlag bool, verbose bool) *idl.MakeClusterRequest {
	var segmentPairs []*idl.SegmentPair
	for _, pair := range config.SegmentArray {
		segmentPairs = append(segmentPairs, SegmentPairToIdl(&pair))
	}

	return &idl.MakeClusterRequest{
		GpArray: &idl.GpArray{
			Coordinator:  SegmentToIdl(&config.Coordinator),
			SegmentArray: segmentPairs,
		},
		ClusterParams: ClusterParamsToIdl(config),
		ForceFlag:     forceFlag,
		Verbose:       verbose,
	}
}

func SegmentToIdl(seg *Segment) *idl.Segment {
	if seg == nil {
		return nil
	}
	return &idl.Segment{
		Port:          int32(seg.Port),
		DataDirectory: seg.DataDirectory,
		HostName:      seg.Hostname,
		HostAddress:   seg.Address,
	}
}

func SegmentPairToIdl(pair *SegmentPair) *idl.SegmentPair {
	return &idl.SegmentPair{
		Primary: SegmentToIdl(pair.Primary),
		Mirror:  SegmentToIdl(pair.Mirror),
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
	if !cliHandler.IsSet("segment-array") && !cliHandler.IsSet("primary-data-directories") {
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

	if len(request.GetPrimarySegments()) == 0 {
		return fmt.Errorf("no primary segments are provided in input config file")
	}
	// validate details of coordinator
	err := ValidateSegment(request.GpArray.Coordinator)
	if err != nil {
		return err
	}

	// validate the details of primary segments
	for _, seg := range request.GetPrimarySegments() {
		err = ValidateSegment(seg)
		if err != nil {
			return err
		}
	}

	// check for conflicting port and data-dir on a host
	err = CheckForDuplicatPortAndDataDirectory(request.GetPrimarySegments())
	if err != nil {
		return err
	}

	// check if gp services enabled on hosts
	err = IsGpServicesEnabled(request)
	if err != nil {
		return err
	}

	if request.ClusterParams.Encoding == "" {
		gplog.Info(fmt.Sprintf("Could not find encoding in cluster config, defaulting to %v", constants.DefaultEncoding))
		request.ClusterParams.Encoding = constants.DefaultEncoding
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
If hostname is not provided then the function returns an error.
If address is not provided then it is populated with the hostname value
*/
func ValidateSegment(segment *idl.Segment) error {
	if segment.HostName == "" {
		//TODO Call RPC to get the hostname from hostAddress and populate here as segment.HostName
		return fmt.Errorf("hostName has not been provided for the segment with port %v and data_directory %v", segment.Port, segment.DataDirectory)
	}

	if segment.HostAddress == "" {
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
func IsGpServicesEnabledFn(req *idl.MakeClusterRequest) error {
	hostnames = []string{req.GpArray.Coordinator.HostName}
	for _, seg := range req.GetPrimarySegments() {
		hostnames = append(hostnames, seg.HostName)
	}

	// remove any duplicate entries
	slices.Sort(hostnames)
	hostnames = slices.Compact(hostnames)

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
		gplog.Info("max_connections not set, will set to default value %v", constants.DefaultQdMaxConnect)
		clusterParams.CommonConfig["max_connections"] = strconv.Itoa(constants.DefaultQdMaxConnect)
	}

	if _, ok := clusterParams.CoordinatorConfig["max_connections"]; !ok {
		// Check if common-config has max-connections defined
		gplog.Info("Coordinator max_connections not set, will set to value %v from CommonConfig", clusterParams.CommonConfig["max_connections"])
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
		gplog.Info("Segment max_connections not set, will set to value %v", segmentConfigMaxConnections)
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
		gplog.Info("Coordinator %v not set, will set to value %v from CommonConfig", configParam, clusterParams.CommonConfig[configParam])
		clusterParams.CoordinatorConfig[configParam] = clusterParams.CommonConfig[configParam]
	}
	if _, ok := clusterParams.SegmentConfig[configParam]; !ok {
		// Check if common-config has configParam defined
		gplog.Info("Segment %v not set, will set to value %v from CommonConfig", configParam, clusterParams.CommonConfig[configParam])
		clusterParams.SegmentConfig[configParam] = clusterParams.CommonConfig[configParam]
	}
}
