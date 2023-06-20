package testutils

import (
	"errors"
	"github.com/greenplum-db/gpdb/gp/hub"
	"google.golang.org/grpc/credentials"
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
	ConfigFileData       []byte
}

func (p *MockPlatform) CreateServiceDir(hostnames []string, serviceDir string, gphome string) error {
	return nil
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
}
func (p *MockPlatform) EnableUserLingering(hostnames []string, gphome string, serviceUser string) error {
	return nil
}
func (p *MockPlatform) ReadFile(configFilePath string) (config *hub.Config, err error) {
	return nil, err
}

type MockCredentials struct {
	TlsConnection credentials.TransportCredentials
	err           error
}

func (s *MockCredentials) LoadServerCredentials() (credentials.TransportCredentials, error) {
	return s.TlsConnection, s.err
}

func (s *MockCredentials) LoadClientCredentials() (credentials.TransportCredentials, error) {
	return s.TlsConnection, s.err
}

func (s *MockCredentials) GetClientServerCredsPath() (CACertPath string, CAKeyPath string, ServerCertPath string, ServerKeyPath string) {
	return "test0", "test1", "test2", "test3"
}

func (s *MockCredentials) SetCredsError(errMsg string) {
	s.err = errors.New(errMsg)
}
func (s *MockCredentials) ResetCredsError() {
	s.err = nil
}
