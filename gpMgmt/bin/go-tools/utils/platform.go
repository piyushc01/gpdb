package utils

import (
	"fmt"
	"github.com/greenplum-db/gpdb/gp/idl"
	"os"
	"os/exec"
	"runtime"
)

type OS interface {
	GenerateServiceFileContents(which string, gphome string, serviceName string) string
	GetDefaultServiceDir() string
	CreateAndInstallHubServiceFile(gphome string, serviceDir string, serviceName string) error
	CreateAndInstallAgentServiceFile(hostnames []string, gphome string, serviceDir string, serviceName string) error
	GetStartHubCmd(serviceName string) *exec.Cmd
	GetStartAgentCmd(serviceName string) []string
	GetServiceStatusMessage(serviceName string) (string, error)
	ParseServiceStatusMessage(message string) idl.ServiceStatus
	DisplayServiceStatus(statuses []*idl.ServiceStatus)
}

func GetOS() OS {
	switch runtime.GOOS {
	case "darwin":
		return NewDarwinOS()
	case "linux":
		return NewLinuxOS()
	default:
		panic("Unsupported OS")
	}
}

func writeServiceFile(filename string, contents string) error {
	handle, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("Could not create systemd service file %s: %w\n", filename, err)
	}
	defer handle.Close()

	_, err = handle.WriteString(contents)
	if err != nil {
		return fmt.Errorf("Could not write to systemd service file %s: %w\n", filename, err)
	}
	return nil
}
