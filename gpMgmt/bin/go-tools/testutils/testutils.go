package testutils

import (
	"os/exec"

	"github.com/greenplum-db/gpdb/gp/idl"
)

type MockPlatform struct {
	RetStatus            idl.ServiceStatus
	ServiceStatusMessage string
	Err                  error
	ServiceFileContent   string
	DefServiceDir        string
	StartCmd             *exec.Cmd
}

func (p *MockPlatform) GetServiceStatusMessage(serviceName string) (string, error) {
	return p.ServiceStatusMessage, p.Err
}
func (p *MockPlatform) GenerateServiceFileContents(which string, gphome string, serviceName string) string {
	return p.ServiceFileContent
}
func (p *MockPlatform) GetDefaultServiceDir() string {
	return p.DefServiceDir
}
func (p *MockPlatform) CreateAndInstallHubServiceFile(gphome string, serviceDir string, serviceName string) error {
	return p.Err
}
func (p *MockPlatform) CreateAndInstallAgentServiceFile(hostnames []string, gphome string, serviceDir string, serviceName string) error {
	return p.Err
}
func (p *MockPlatform) GetStartHubCommand(serviceName string) *exec.Cmd {
	return p.StartCmd
}
func (p *MockPlatform) GetStartAgentCommandString(serviceName string) []string {
	return nil
}
func (p *MockPlatform) ParseServiceStatusMessage(message string) idl.ServiceStatus {
	return p.RetStatus
}
func (p *MockPlatform) DisplayServiceStatus(statuses []*idl.ServiceStatus) {
	return
}
func (p *MockPlatform) EnableUserLingering(hostnames []string, gphome string, serviceUser string) error {
	return nil
}
