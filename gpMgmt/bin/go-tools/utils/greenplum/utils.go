package greenplum

import (
	"fmt"
	"strings"

	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/greenplum-db/gpdb/gp/utils/postgres"
)

func GetPostgresGpVersion(gphome string) (string, error) {
	pgGpVersionCmd := &postgres.PostgresGpVersion{GpVersion: true}
	out, err := utils.RunExecCommand(pgGpVersionCmd, gphome)
	if err != nil {
		return "", fmt.Errorf("fetching postgres gp-version: %w", err)
	}

	return strings.TrimSpace(out.String()), nil
}
