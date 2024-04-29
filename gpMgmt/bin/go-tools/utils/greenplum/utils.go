package greenplum

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/constants"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/greenplum-db/gpdb/gp/utils/postgres"
)

var newDBConnFromEnvironment = dbconn.NewDBConnFromEnvironment

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
func GetCoordinatorConn(datadir, dbname string, utility ...bool) (*dbconn.DBConn, error) {
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
	conn := newDBConnFromEnvironment(dbname)
	conn.Port = port

	err = conn.Connect(1, utility...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func TriggerFtsProbe(coordinatorDataDir string) error {
	conn, err := GetCoordinatorConn(coordinatorDataDir, "")
	if err != nil {
		return err
	}
	defer conn.Close()

	query := "SELECT gp_request_fts_probe_scan()"
	_, err = conn.Exec(query)
	gplog.Debug("Executing query %q", query)
	if err != nil {
		return fmt.Errorf("triggering FTS probe: %w", err)
	}

	return nil
}

// used only for testing
func SetNewDBConnFromEnvironment(customFunc func(dbname string) *dbconn.DBConn) {
	newDBConnFromEnvironment = customFunc
}

func ResetNewDBConnFromEnvironment() {
	newDBConnFromEnvironment = dbconn.NewDBConnFromEnvironment
}

func GetUserInput() bool {
	// Create a channel to receive user input
	userInput := make(chan string)

	// Create a timer channel that sends a signal after the timeout duration
	timer := time.After(10 * time.Second)

	// Create a goroutine to read user input
	go func() {
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("ReadString error")
		}
		userInput <- strings.ToLower(strings.TrimSpace(input))
	}()

	// Wait for either user input or timeout
	select {
	case input := <-userInput:
		if input == "yes" || input == "y" {
			fmt.Println("You entered 'yes'.")
			return true
		} else if input == "no" || input == "n" {
			fmt.Println("You entered 'no'.")
			return false
		} else {
			fmt.Println("Invalid input. Please enter 'yes' or 'no'.")
		}
	case <-timer:
		fmt.Println("no user input received. Default input is 'no'.")
		return false
	}
	return false
}
