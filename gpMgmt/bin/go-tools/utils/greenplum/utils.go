package greenplum

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/constants"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/greenplum-db/gpdb/gp/utils/postgres"
)

func GetPostgresGpVersion(gpHome string) (string, error) {
	pgGpVersionCmd := &postgres.Postgres{GpVersion: true}
	out, err := utils.RunGpCommand(pgGpVersionCmd, gpHome)
	if err != nil {
		return "", fmt.Errorf("fetching postgres gp-version: %w", err)
	}

	return strings.TrimSpace(out.String()), nil
}

func GetDefaultHubLogDir() string {
	currentUser, _ := utils.System.CurrentUser()

	return filepath.Join(currentUser.HomeDir, "gpAdminLogs")
}

// GetCoordinatorConn creates a connection object for the coordinator segment
// given only its data directory. This function is expected to be called on the
// coordinator host only. By default it creates a non utility mode connection
// and uses the 'template1' database if no database is provided
func GetCoordinatorConn(ctx context.Context, datadir, dbname string, utility ...bool) (*utils.DBConnWithContext, error) {
	value, err := postgres.GetConfigValue(datadir, "port")
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}

	if dbname == "" {
		dbname = constants.DefaultDatabase
	}
	conn := utils.NewDBConnWithContext(ctx, dbname)
	conn.DB.Port = port

	err = conn.DB.Connect(1, utility...)
	if err != nil {
		return nil, err
	}

	return &conn, nil
}

func TriggerFtsProbe(coordinatorDataDir string) error {
	conn, err := GetCoordinatorConn(context.Background(), coordinatorDataDir, "")
	if err != nil {
		return err
	}
	defer conn.DB.Close()

	query := "SELECT gp_request_fts_probe_scan()"
	_, err = conn.DB.Exec(query)
	gplog.Debug("Executing query %q", query)
	if err != nil {
		return fmt.Errorf("triggering FTS probe: %w", err)
	}

	return nil
}
