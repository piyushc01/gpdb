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
	PgData         string
	Encoding       string
	Locale         string
	LcCollate      string
	LcCtype        string
	LcMessages     string
	LcMonetory     string
	LcNumeric      string
	LcTime         string
	SharedBuffers  string
	MaxConnections int
	DataChecksums  bool
}

func (cmd *Initdb) BuildExecCommand(gphome string) *exec.Cmd {
	utility := utils.GetGphomeUtilityPath(gphome, initdbUtility)
	args := []string{}

	args = utils.AppendIfNotEmpty(args, "--pgdata", cmd.PgData)
	args = utils.AppendIfNotEmpty(args, "--encoding", cmd.Encoding)
	args = utils.AppendIfNotEmpty(args, "--locale", cmd.Locale)
	args = utils.AppendIfNotEmpty(args, "--lc-collate", cmd.LcCollate)
	args = utils.AppendIfNotEmpty(args, "--lc-ctype", cmd.LcCtype)
	args = utils.AppendIfNotEmpty(args, "--lc-messages", cmd.LcMessages)
	args = utils.AppendIfNotEmpty(args, "--lc-monetary", cmd.LcMonetory)
	args = utils.AppendIfNotEmpty(args, "--lc-numeric", cmd.LcNumeric)
	args = utils.AppendIfNotEmpty(args, "--lc-time", cmd.LcTime)
	args = utils.AppendIfNotEmpty(args, "--max_connections", cmd.MaxConnections)
	args = utils.AppendIfNotEmpty(args, "--shared_buffers", cmd.SharedBuffers)
	args = utils.AppendIfNotEmpty(args, "--data-checksums", cmd.DataChecksums)

	return utils.System.ExecCommand(utility, args...)
}

type PgCtlStart struct {
	PgData  string
	Timeout int
	Wait    bool
	NoWait  bool
	Logfile string
	Options string
}

func (cmd *PgCtlStart) BuildExecCommand(gphome string) *exec.Cmd {
	utility := utils.GetGphomeUtilityPath(gphome, pgCtlUtility)
	args := []string{"start"}

	args = utils.AppendIfNotEmpty(args, "--pgdata", cmd.PgData)
	args = utils.AppendIfNotEmpty(args, "--timeout", cmd.Timeout)
	args = utils.AppendIfNotEmpty(args, "--wait", cmd.Wait)
	args = utils.AppendIfNotEmpty(args, "--no-wait", cmd.NoWait)
	args = utils.AppendIfNotEmpty(args, "--log", cmd.Logfile)
	args = utils.AppendIfNotEmpty(args, "--options", cmd.Options)

	return utils.System.ExecCommand(utility, args...)
}

type PgCtlStop struct {
	PgData  string
	Timeout int
	Wait    bool
	NoWait  bool
	Mode    string
}

func (cmd *PgCtlStop) BuildExecCommand(gphome string) *exec.Cmd {
	utility := utils.GetGphomeUtilityPath(gphome, pgCtlUtility)
	args := []string{"stop"}

	args = utils.AppendIfNotEmpty(args, "--pgdata", cmd.PgData)
	args = utils.AppendIfNotEmpty(args, "--timeout", cmd.Timeout)
	args = utils.AppendIfNotEmpty(args, "--wait", cmd.Wait)
	args = utils.AppendIfNotEmpty(args, "--no-wait", cmd.NoWait)
	args = utils.AppendIfNotEmpty(args, "--mode", cmd.Mode)

	return utils.System.ExecCommand(utility, args...)
}

type PostgresGpVersion struct {
	GpVersion bool
}

func (cmd *PostgresGpVersion) BuildExecCommand(gphome string) *exec.Cmd {
	utililty := utils.GetGphomeUtilityPath(gphome, postgresUtility)
	args := []string{"--gp-version"}
	return utils.System.ExecCommand(utililty, args...)
}
