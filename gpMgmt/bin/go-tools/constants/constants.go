package constants

import (
	"path/filepath"

	"github.com/greenplum-db/gp-common-go-libs/operating"
)

const (
	DefaultHubPort      = 4242
	DefaultAgentPort    = 8000
	DefaultServiceName  = "gp"
	ConfigFileName      = "gp.conf"
	ShellPath           = "/bin/bash"
	GpSSH               = "gpssh"
	MaxRetries          = 10
	PlatformDarwin      = "darwin"
	PlatformLinux       = "linux"
	DefaultQdMaxConnect = 150
	QeConnectFactor     = 3
	DefaultBuffer       = "128000kB"
	OsOpenFiles         = 65535
	DefaultDbName       = "template1"
	DefaultEncoding     = "UTF-8"
	RolePrimary         = "p"
	EtcHostsFilepath    = "/etc/hosts"
	SecurityLimitsConf  = "/etc/security/limits.conf"
	SecurityLimitsdDir  = "/etc/security/limits.d/"
)

func GetDefaultHubLogDir() string {
	currentUser, _ := operating.System.CurrentUser()

	return filepath.Join(currentUser.HomeDir, "gpAdminLogs")
}
