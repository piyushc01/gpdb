package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
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
	CheckDirEmpty             = CheckDirEmptyFn
	CheckFileOwnerGroup       = CheckFileOwnerGroupFn
	CheckExecutable           = CheckExecutableFn
	OsIsNotExist              = os.IsNotExist
	GetAllNonEmptyDir         = GetAllNonEmptyDirFn
	CheckFilePermissions      = CheckFilePermissionsFn
	GetAllAvailableLocales    = GetAllAvailableLocalesFn
	ValidateLocaleSettings    = ValidateLocaleSettingsFn
	ValidatePorts             = ValidatePortsFn
	VerifyPgVersion           = ValidatePgVersionFn
	GetMaxFilesFromLimitsFile = GetMaxFilesFromLimitsFileFn
)

func (s *Server) ValidateHostEnv(ctx context.Context, request *idl.ValidateHostEnvRequest) (*idl.ValidateHostEnvReply, error) {
	gplog.Debug("Starting ValidateHostEnvFn for request:%v", request)
	dirList := request.DirectoryList
	locale := request.Locale
	socketAddressList := request.SocketAddressList
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
	err = ValidatePorts(socketAddressList)
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
	curUser, err := utils.System.CurrentUser()
	if err != nil {
		warnMsg := fmt.Sprintf("error getting current user: %v", err)
		warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
		gplog.Warn(warnMsg)
		return warnings
	}

	out, err := utils.System.ExecCommand("bash", "-c", "source ~/.bashrc;ulimit -n").CombinedOutput()
	if err != nil {
		warnMsg := fmt.Sprintf("error fetching open file limit values:%v", err)
		warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
		gplog.Warn(warnMsg)
		return warnings
	}

	ulimitVal, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		warnMsg := fmt.Sprintf("could not convert the ulimit value: %v", err)
		warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
		gplog.Warn(warnMsg)
		return warnings
	}
	//curPlatform := utils.GetPlatform()
	if platform.GetPlatformOS() == constants.PlatformDarwin {
		if ulimitVal < constants.OsOpenFiles {
			// In case of macOS, no limits file are present, return error
			warnMsg := fmt.Sprintf("Host open file limit is %d should be >= %d", ulimitVal, constants.OsOpenFiles)
			warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
			gplog.Warn(warnMsg)
			return warnings
		}
		return warnings
	}
	// Continue checking limits further for Linux platform
	if ulimitVal < constants.OsOpenFiles {
		// Sometime not able to get the correct value of open files limit using `ulimit -n`
		// This happens because agent is running as a service and for services settings are applied through service file
		// For linux platform, Check in limits.d/limits.conf and limits.conf file
		// limits.d/*limit.conf settings override limits.conf

		// look for entries like o extract nofile limit:
		// * soft nofile 65536
		// gpadmin soft nofile 65536
		files, err := utils.System.ReadDir(constants.SecurityLimitsdDir)

		if err != nil && !OsIsNotExist(err) {
			warnMsg := fmt.Sprintf("error opening directory: %v", err)
			warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
			gplog.Warn(warnMsg)
			return warnings
		}

		// In case of multiple files, files are sourced in ascending alphabetical order.
		// To get effective value, sort files in descending alphabetical order
		// on first detection of soft limit, extract value and return
		for _, file := range files {
			if file.IsDir() {
				// skip if it's a directory
				continue
			}
			// open file, read contents and get the entry for max open files limit for the user
			fileLimit, err := GetMaxFilesFromLimitsFile(filepath.Join(constants.SecurityLimitsdDir, file.Name()), curUser)
			if err != nil {
				warnMsg := fmt.Sprintf(" %v", err)
				warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
				gplog.Warn(warnMsg)
				return warnings
			}
			if fileLimit < constants.OsOpenFiles && fileLimit != -1 {
				warnMsg := fmt.Sprintf("Host open file limit is %d should be >= %d", fileLimit, constants.OsOpenFiles)
				warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
				gplog.Warn(warnMsg)
				return warnings
			}
		}

		// in case not in limits.d/* files, check security/limits.conf file
		fileLimit, err := GetMaxFilesFromLimitsFile(constants.SecurityLimitsConf, curUser)
		if err != nil {
			warnMsg := fmt.Sprintf(" %v", err)
			warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
			gplog.Warn(warnMsg)
			return warnings
		}
		if fileLimit < constants.OsOpenFiles && fileLimit != -1 {

			warnMsg := fmt.Sprintf("Host open file limit is %d should be >= %d", fileLimit, constants.OsOpenFiles)
			warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
			gplog.Warn(warnMsg)
			return warnings
		}
		// Check if limit not defined in limits.conf or limits.d/* report warning based on value from ulimit
		if fileLimit == -1 && ulimitVal < constants.OsOpenFiles {
			warnMsg := fmt.Sprintf("Host open file limit is %d should be >= %d", ulimitVal, constants.OsOpenFiles)
			warnings = append(warnings, &idl.LogMessage{Message: warnMsg, Level: idl.LogLevel_WARNING})
			gplog.Warn(warnMsg)
			return warnings
		}
	}

	return warnings
}

func GetMaxFilesFromLimitsFileFn(fileName string, curUser *user.User) (int, error) {
	fd, err := utils.System.Open(fileName)
	if err != nil {
		warnMsg := fmt.Sprintf("error opening file: %v", err)
		gplog.Warn(warnMsg)
		return -1, fmt.Errorf(warnMsg)
	}
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		// Form the regex and check if nofile is set
		re := regexp.MustCompile(fmt.Sprintf("(^\\s*%s|\\*)\\s+soft\\s+nofile\\s+(\\d+)\\s*$", curUser.Name))
		match := re.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		//if there's a regex match, check integer value
		intMaxFiles, err := strconv.Atoi(match[2])
		if err != nil {
			warnMsg := fmt.Sprintf("error converting max files limit value: %v", err)
			gplog.Warn(warnMsg)
			return -1, fmt.Errorf(warnMsg)
		}
		gplog.Debug("GetMaxFilesFromLimitsFile for file:%s, returned value:%d", fileName, intMaxFiles)
		return intMaxFiles, nil

	}
	if scanner.Err() != nil {
		warnMsg := fmt.Sprintf("error reading the file: %v", err)
		gplog.Warn(warnMsg)
		return -1, fmt.Errorf(warnMsg)
	}
	gplog.Debug("GetMaxFilesFromLimitsFile for file:%s, No value found", fileName)
	return -1, nil
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

func ValidatePortsFn(socketAddressList []string) error {
	gplog.Debug("Started with ValidatePorts")
	var usedSocketAddressList []string
	for _, socketAddress := range socketAddressList {
		listener, err := net.Listen("tcp", socketAddress)
		if err != nil {
			usedSocketAddressList = append(usedSocketAddressList, socketAddress)
		} else {
			_ = listener.Close()
		}
	}
	if len(usedSocketAddressList) > 0 {
		gplog.Error("ports already in use: %v, check if cluster already running", usedSocketAddressList)
		return fmt.Errorf("ports already in use: %v, check if cluster already running", usedSocketAddressList)
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
		fmt.Printf("StatUID:%d, StatGID:%d\nSysUID:%d SysGID:%d\n", stat.Uid, stat.Gid, systemUid, systemGid)
		return fmt.Errorf("file %s is neither owned by the user nor by group", filePath)
	}
	return nil
}

func CheckExecutableFn(FileMode os.FileMode) bool {
	return FileMode&0111 != 0
}

func GetAllAvailableLocalesFn() (string, error) {
	cmd := utils.System.ExecCommand("/usr/bin/locale", "-a")
	availableLocales, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("failed to get the available locales on this system: %w", err)
	}
	return string(availableLocales), nil
}

// Simplified version of _nl_normalize_codeset from glibc
// https://sourceware.org/git/?p=glibc.git;a=blob;f=intl/l10nflist.c;h=078a450dfec21faf2d26dc5d0cb02158c1f23229;hb=1305edd42c44fee6f8660734d2dfa4911ec755d6#l294
// Input parameter - string with locale define as [language[_territory][.codeset][@modifier]]
func NormalizeCodesetInLocale(locale string) string {
	localeSplit := strings.Split(locale, ".")
	languageAndTerritory := localeSplit[0]

	codesetAndModifier := strings.Split(localeSplit[1], "@")

	codeset := codesetAndModifier[0]

	modifier := ""

	if len(codesetAndModifier) == 2 {
		modifier = codesetAndModifier[1]
	}

	digitPattern := regexp.MustCompile(`^[0-9]+$`)
	if digitPattern.MatchString(codeset) {
		codeset = "iso" + codeset
	} else {
		codeset = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, codeset)
		codeset = strings.ToLower(codeset)
	}

	result := fmt.Sprintf("%s%s%s", languageAndTerritory, dotIfNotEmpty(codeset), atIfNotEmpty(modifier))
	return result
}

func dotIfNotEmpty(s string) string {
	if s != "" {
		return "." + s
	}
	return ""
}

func atIfNotEmpty(s string) string {
	if s != "" {
		return "@" + s
	}
	return ""
}

func IsLocaleAvailable(locale_type string, allAvailableLocales string) bool {
	locales := strings.Split(allAvailableLocales, "\n")
	normalizedLocale := NormalizeCodesetInLocale(locale_type)

	for _, v := range locales {
		if locale_type == v || normalizedLocale == v {
			return true
		}
	}
	return false
}

func ValidateLocaleSettingsFn(locale *idl.Locale) error {
	systemLocales, err := GetAllAvailableLocales()
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
