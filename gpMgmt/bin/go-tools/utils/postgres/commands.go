package postgres

import (
	"os/exec"

	"github.com/greenplum-db/gpdb/gp/utils"
)

const (
	initdbUtility   = "initdb"
	pgCtlUtility    = "pg_ctl"
	postgresUtility = "postgres"
)

type Initdb struct {
	PgData        string `flag:"--pgdata"`
	Encoding      string `flag:"--encoding"`
	Locale        string `flag:"--locale"`
	LcCollate     string `flag:"--lc-collate"`
	LcCtype       string `flag:"--lc-ctype"`
	LcMessages    string `flag:"--lc-messages"`
	LcMonetory    string `flag:"--lc-monetary"`
	LcNumeric     string `flag:"--lc-numeric"`
	LcTime        string `flag:"--lc-time"`
	DataChecksums bool   `flag:"--data-checksums"`
}

func (cmd *Initdb) BuildExecCommand(gphome string) *exec.Cmd {
	utility := utils.GetGpUtilityPath(gphome, initdbUtility)
	args := utils.GenerateArgs(cmd)

	return utils.System.ExecCommand(utility, args...)
}

type PgCtlStart struct {
	PgData  string `flag:"--pgdata"`
	Timeout int    `flag:"--timeout"`
	Wait    bool   `flag:"--wait"`
	NoWait  bool   `flag:"--no-wait"`
	Logfile string `flag:"--log"`
	Options string `flag:"--options"`
}

func (cmd *PgCtlStart) BuildExecCommand(gphome string) *exec.Cmd {
	utility := utils.GetGpUtilityPath(gphome, pgCtlUtility)
	args := append([]string{"start"}, utils.GenerateArgs(cmd)...)

	return utils.System.ExecCommand(utility, args...)
}

type PgCtlStop struct {
	PgData  string `flag:"--pgdata"`
	Timeout int    `flag:"--timeout"`
	Wait    bool   `flag:"--wait"`
	NoWait  bool   `flag:"--no-wait"`
	Mode    string `flag:"--mode"`
}

func (cmd *PgCtlStop) BuildExecCommand(gphome string) *exec.Cmd {
	utility := utils.GetGpUtilityPath(gphome, pgCtlUtility)
	args := append([]string{"stop"}, utils.GenerateArgs(cmd)...)

	return utils.System.ExecCommand(utility, args...)
}

type Postgres struct {
	GpVersion bool `flag:"--gp-version"`
}

func (cmd *Postgres) BuildExecCommand(gphome string) *exec.Cmd {
	utililty := utils.GetGpUtilityPath(gphome, postgresUtility)
	args := utils.GenerateArgs(cmd)
	return utils.System.ExecCommand(utililty, args...)
}
