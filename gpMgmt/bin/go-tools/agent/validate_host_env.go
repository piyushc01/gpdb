package agent

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/greenplum-db/gpdb/gp/constants"
	"github.com/greenplum-db/gpdb/gp/utils/greenplum"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
)

var (
	CheckDirEmpty          = CheckDirEmptyFn
	CheckFileOwnerGroup    = CheckFileOwnerGroupFn
	CheckExecutable        = CheckExecutableFn
	OsIsNotExist           = os.IsNotExist
	GetAllNonEmptyDir      = GetAllNonEmptyDirFn
	CheckFilePermissions   = CheckFilePermissionsFn
	ValidateLocaleSettings = ValidateLocaleSettingsFn
	ValidatePortList       = ValidatePortListFn
	VerifyPgVersion        = ValidatePgVersionFn
)

func (s *Server) ValidateHostEnv(ctx context.Context, request *idl.ValidateHostEnvRequest) (*idl.ValidateHostEnvReply, error) {
	gplog.Debug("Starting ValidateHostEnvFn for request:%v", request)
	dirList := request.DirectoryList
	locale := request.Locale
	portList := request.PortList
	forced := request.Forced

	// Check if user is non-root
	if utils.System.Getuid() == 0 {
		userInfo, err := utils.System.CurrentUser()
		if err != nil {
			gplog.Error("failed to get user name Error:%v. Current user is a root user. Can't create cluster under root", err)
			return &idl.ValidateHostEnvReply{}, fmt.Errorf("failed to get user name Error:%v. Current user is a root user. Can't create cluster under root", err)
		}
		return &idl.ValidateHostEnvReply{}, fmt.Errorf("user:%s is a root user, Can't create cluster under root user", userInfo.Name)
	}
	gplog.Debug("Done with checking user is non root")

	//Check for PGVersion
	pgVersionErr := VerifyPgVersion(request.GpVersion, s.GpHome)
	if pgVersionErr != nil {
		gplog.Error("Postgres gp-version validation failed:%v", pgVersionErr)
		return &idl.ValidateHostEnvReply{}, pgVersionErr
	}

	// Check for each directory, if directory is empty
	nonEmptyDirList := GetAllNonEmptyDir(dirList)
	gplog.Debug("Got the list of all non-empty directories")

	if len(nonEmptyDirList) > 0 && !forced {
		return &idl.ValidateHostEnvReply{}, fmt.Errorf("directory not empty:%v", nonEmptyDirList)
	}
	if forced && len(nonEmptyDirList) > 0 {

		gplog.Debug("Forced init. Deleting non-empty directories:%s", dirList)
		for _, dir := range nonEmptyDirList {
			err := utils.System.RemoveAll(dir)
			if err != nil {
				return &idl.ValidateHostEnvReply{}, fmt.Errorf("delete not empty dir:%s, error:%v", dir, err)
			}
		}
	}

	// Validate permission to initdb ? Error will be returned upon running
	gplog.Debug("Checking initdb for permissions")
	initdbPath := filepath.Join(s.GpHome, "bin", "initdb")
	err := CheckFilePermissions(initdbPath)
	if err != nil {
		return &idl.ValidateHostEnvReply{}, err
	}

	// Validate that the different locale settings are available on the system
	err = ValidateLocaleSettings(locale)
	if err != nil {
		gplog.Info("Got error while validating locale %v", err)
		return &idl.ValidateHostEnvReply{}, err
	}

	// Check if port in use
	err = ValidatePortList(portList)
	if err != nil {
		return &idl.ValidateHostEnvReply{}, err
	}

	// Any checks to raise warnings
	var warnings []*idl.LogMessage

	// check coordinator open file values
	warnings = CheckOpenFilesLimit()
	addressWarnings := CheckHostAddressInHostsFile(request.HostAddressList)
	warnings = append(warnings, addressWarnings...)
	return &idl.ValidateHostEnvReply{Messages: warnings}, nil
}

func ValidatePgVersionFn(expectedVersion string, gpHome string) error {
	localPgVersion, err := greenplum.GetPostgresGpVersion(gpHome)
	if err != nil {
		return err
	}

	if expectedVersion != localPgVersion {
		return fmt.Errorf("postgres gp-version does not matches with coordinator postgres gp-version."+
			"Coordinator version:'%s', Current version:'%s'", expectedVersion, localPgVersion)
	}
	return nil

}
func CheckOpenFilesLimit() []*idl.LogMessage {
	var warnings []*idl.LogMessage
	out, err := utils.System.ExecCommand("ulimit", "-n").CombinedOutput()
	if err != nil {
		warnMsg := fmt.Sprintf("error fetching open file limit values:%v", err)
		warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
		gplog.Warn(warnMsg)
		return warnings
	}

	openFileLimit, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		warnMsg := fmt.Sprintf("could not convert the ulimit value: %v", err)
		warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
		gplog.Warn(warnMsg)
		return warnings
	}
	if openFileLimit < constants.OsOpenFiles {
		warnMsg := fmt.Sprintf("Coordinator open file limit is %d should be >= %d", openFileLimit, constants.OsOpenFiles)
		warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
		gplog.Warn(warnMsg)
	}

	return warnings
}

func CheckHostAddressInHostsFile(hostAddressList []string) []*idl.LogMessage {
	var warnings []*idl.LogMessage
	gplog.Debug("CheckHostAddressInHostsFile checking for address:%v", hostAddressList)
	content, err := utils.System.ReadFile(constants.EtcHostsFilepath)
	if err != nil {
		warnMsg := fmt.Sprintf("error reading file %s error:%v", constants.EtcHostsFilepath, err)
		gplog.Warn(warnMsg)
		warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
		return warnings
	}

	lines := strings.Split(string(content), "\n")
	for _, hostAddress := range hostAddressList {
		for _, line := range lines {
			hosts := strings.Split(line, " ")
			if slices.Contains(hosts, hostAddress) && slices.Contains(hosts, "localhost") {
				warnMsg := fmt.Sprintf("HostAddress %s is assigned localhost in %s."+
					"This will cause segment->coordinator communication failures."+
					"Remote %s from local host line in /etc/hosts",
					hostAddress, constants.EtcHostsFilepath, constants.EtcHostsFilepath)

				warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
				gplog.Warn(warnMsg)
				break
			}
		}
	}

	return warnings
}

func ValidatePortListFn(portList []int32) error {
	gplog.Debug("Started with ValidatePortList")
	var usedPortList []int
	for _, port := range portList {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			usedPortList = append(usedPortList, int(port))
		} else {
			_ = listener.Close()
		}
	}
	if len(usedPortList) > 0 {
		gplog.Error("ports already in use:%v, check if cluster already running", usedPortList)
		return fmt.Errorf("ports already in use:%v, check if cluster already running", usedPortList)
	}
	return nil
}

func GetAllNonEmptyDirFn(dirList []string) []string {
	var nonEmptyDir []string
	for _, dir := range dirList {
		isEmpty, err := CheckDirEmpty(dir)
		if err != nil {
			gplog.Error("Directory:%s Error checking if empty:%s", dir, err.Error())
			nonEmptyDir = append(nonEmptyDir, dir)
		} else if !isEmpty {
			// Directory not empty
			nonEmptyDir = append(nonEmptyDir, dir)
		}
	}
	return nonEmptyDir
}
func CheckDirEmptyFn(dirPath string) (bool, error) {
	// check if dir exists
	file, err := os.Open(dirPath)
	if OsIsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("error opening file:%v", err)
	}
	defer file.Close()
	_, err = file.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, nil
}

func CheckFilePermissionsFn(filePath string) error {
	fileInfo, err := utils.System.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error getting file info:%v", err)
	}
	// Get current user-id, group-id and checks against initdb file
	err = CheckFileOwnerGroup(filePath, fileInfo)
	if err != nil {
		return err
	}

	// Check if the file has execute permission
	if !CheckExecutable(fileInfo.Mode()) {
		return fmt.Errorf("file %s does not have execute permissions", filePath)
	}
	return nil
}

func CheckFileOwnerGroupFn(filePath string, fileInfo os.FileInfo) error {
	systemUid := utils.System.Getuid()
	systemGid := utils.System.Getgid()
	// Fetch file info: file owner, group ID
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("error converting fileinfo:%v", ok)
	}

	if int(stat.Uid) != systemUid && int(stat.Gid) != systemGid {
		return fmt.Errorf("file %s is neither owned by the user nor by group", filePath)
	}
	return nil
}

func CheckExecutableFn(FileMode os.FileMode) bool {
	return FileMode&0111 != 0
}

func getAllAvailableLocales() (string, error) {
	cmd := utils.System.ExecCommand("/usr/bin/locale", "-a")
	availableLocales, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to get the available locales on this system: %w", err)
	}
	return string(availableLocales), nil
}

func IsLocaleAvailable(locale_type string, allAvailableLocales string) bool {
	locales := strings.Split(allAvailableLocales, "\n")

	for _, v := range locales {
		if locale_type == v {
			return true
		}
	}
	return false
}

func ValidateLocaleSettingsFn(locale *idl.Locale) error {
	systemLocales, err := getAllAvailableLocales()
	if err != nil {
		return err
	}
	localeMap := make(map[string]bool)
	localeMap[locale.LcMonetory] = true
	localeMap[locale.LcAll] = true
	localeMap[locale.LcNumeric] = true
	localeMap[locale.LcTime] = true
	localeMap[locale.LcCollate] = true
	localeMap[locale.LcMessages] = true
	localeMap[locale.LcCtype] = true

	for lc := range localeMap {
		// TODO normalize codeset in locale and the check for the availability
		if !IsLocaleAvailable(lc, systemLocales) {
			return fmt.Errorf("locale value '%s' is not a valid locale", lc)
		}
	}

	return nil
}
