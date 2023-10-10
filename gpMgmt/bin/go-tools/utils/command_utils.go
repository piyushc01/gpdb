package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
)

type CommandBuilder interface {
	BuildExecCommand(gphome string) *exec.Cmd
}

func NewExecCommand(cmdBuilder CommandBuilder, gphome string) *exec.Cmd {
	return cmdBuilder.BuildExecCommand(gphome)
}

func NewGpSourcedCommand(cmdBuilder CommandBuilder, gphome string) *exec.Cmd {
	cmd := cmdBuilder.BuildExecCommand(gphome)
	gpSourceFilePath := filepath.Join(gphome, "greenplum_path.sh")

	return System.ExecCommand("bash", "-c", fmt.Sprintf("source %s && %s", gpSourceFilePath, cmd.String()))
}

func runCommand(cmd *exec.Cmd) (*bytes.Buffer, error) {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	gplog.Debug("Executing command: %s", cmd.String())
	err := cmd.Run()

	if err != nil {
		return stderr, err
	} else {
		return stdout, err
	}
}

func RunExecCommand(cmdBuilder CommandBuilder, gphome string) (*bytes.Buffer, error) {
	return runCommand(NewExecCommand(cmdBuilder, gphome))
}

func RunGpSourcedCommand(cmdBuilder CommandBuilder, gphome string) (*bytes.Buffer, error) {
	return runCommand(NewGpSourcedCommand(cmdBuilder, gphome))
}

func GetGphomeUtilityPath(gphome, utility string) string {
	return path.Join(gphome, "bin", utility)
}

func AppendIfNotEmpty(args []string, flag string, value interface{}) []string {
	switch value := value.(type) {
	case int:
		if value != 0 {
			args = append(args, flag, strconv.Itoa(value))
		}
	case float64:
		if value != 0 {
			args = append(args, flag, strconv.FormatFloat(value, 'f', -1, 64))
		}
	case string:
		if value != "" {
			args = append(args, flag, value)
		}
	case bool:
		if value {
			args = append(args, flag)
		}
	}

	return args
}
