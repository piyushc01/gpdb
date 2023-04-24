package testutils

import (
	"github.com/greenplum-db/gpdb/gp/idl"
	"os/exec"
)

type MockOs struct {
	RetStatus  idl.ServiceStatus
	ServiceStatusMessage string
	Err error
	ServiceFileContent string
	DefServiceDir string
	StartCmd *exec.Cmd
}


func (s* MockOs)GetServiceStatusMessage(serviceName string) (string, error){
	return s.ServiceStatusMessage, s.Err
}
func (s* MockOs)GenerateServiceFileContents(which string, gphome string, serviceName string) string{
	return s.ServiceFileContent
}
func (s* MockOs)GetDefaultServiceDir() string{
	return s.DefServiceDir
}
func (s* MockOs)CreateAndInstallHubServiceFile(gphome string, serviceDir string, serviceName string) error{
	return s.Err
}
func (s* MockOs)CreateAndInstallAgentServiceFile(hostnames []string, gphome string, serviceDir string, serviceName string) error{
	return s.Err
}
func (s* MockOs)GetStartHubCmd(serviceName string) *exec.Cmd{
	return  s.StartCmd
}
func (s* MockOs)GetStartAgentCmd(serviceName string) []string{
	return nil
}
func (s* MockOs)ParseServiceStatusMessage(message string) idl.ServiceStatus{
	return s.RetStatus
}
func (s* MockOs)DisplayServiceStatus(statuses []*idl.ServiceStatus){
	return
}
