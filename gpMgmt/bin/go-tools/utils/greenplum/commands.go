package greenplum

import (
	"os/exec"

	"github.com/greenplum-db/gpdb/gp/utils"
)

const (
	gpstart = "gpstart"
	gpstop  = "gpstop"
)

type GpStop struct {
	DataDirectory   string `flag:"-d"`
	CoordinatorOnly bool   `flag:"--coordinator_only"`
	Verbose         bool   `flag:"-v"`
}

func (cmd *GpStop) BuildExecCommand(gpHome string) *exec.Cmd {
	utility := utils.GetGpUtilityPath(gpHome, gpstop)
	args := append([]string{"-a"}, utils.GenerateArgs(cmd)...)

	return utils.System.ExecCommand(utility, args...)
}

type GpStart struct {
	DataDirectory string `flag:"-d"`
	Verbose       bool   `flag:"-v"`
}

func (cmd *GpStart) BuildExecCommand(gpHome string) *exec.Cmd {
	utility := utils.GetGpUtilityPath(gpHome, gpstart)
	args := append([]string{"-a"}, utils.GenerateArgs(cmd)...)

	return utils.System.ExecCommand(utility, args...)
}
