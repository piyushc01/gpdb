package postgres

import (
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/utils"
)

const (
	postgresqlConfFile       = "postgresql.conf"
	postgresInternalConfFile = "internal.auto.conf"
	pgHbaConfFile            = "pg_hba.conf"
)

func UpdatePostgresqlConf(pgdata string, configParams map[string]string, overwrite bool) error {
	gplog.Info("Updating %s for data directory %s with: %s", postgresqlConfFile, pgdata, configParams)
	err := updateConfFile(postgresqlConfFile, pgdata, configParams, overwrite)
	if err != nil {
		return err
	}

	gplog.Info("Successfully updated %s for data directory %s", postgresqlConfFile, pgdata)
	return nil
}

func CreatePostgresInternalConf(pgdata string, dbid int) error {
	postgresInternalConfFilePath := filepath.Join(pgdata, postgresInternalConfFile)
	file, err := utils.System.Create(postgresInternalConfFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	contents := fmt.Sprintf("gp_dbid = %d", dbid)
	_, err = file.WriteString(contents)
	if err != nil {
		return err
	}

	gplog.Info("Successfully created %s for data directory %s", postgresInternalConfFile, pgdata)
	return nil
}

func BuildCoordinatorPgHbaConf(pgdata string, addrs []string) error {
	pgHbaFilePath := filepath.Join(pgdata, pgHbaConfFile)

	file, err := utils.System.Open(pgHbaFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	updatedLines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			updatedLines = append(updatedLines, line)
		}
	}

	user, err := utils.System.CurrentUser()
	if err != nil {
		return err
	}

	localAccessString := "local\t%s\t%s\tident"

	// Add access entries
	line := fmt.Sprintf(localAccessString, "all", user.Username)
	updatedLines = append(updatedLines, line)
	addPgHbaEntries(&updatedLines, []string{"localhost"}, "all", user.Username)
	addPgHbaEntries(&updatedLines, addrs, "all", user.Username)

	// Add replication entries
	line = fmt.Sprintf(localAccessString, "replication", user.Username)
	updatedLines = append(updatedLines, line)
	addPgHbaEntries(&updatedLines, []string{"samehost"}, "replication", user.Username)
	addPgHbaEntries(&updatedLines, addrs, "replication", user.Username)

	err = utils.WriteLinesToFile(pgHbaFilePath, updatedLines)
	if err != nil {
		return err
	}

	gplog.Info("Successfully updated %s for data directory %s", pgHbaConfFile, pgdata)
	return nil
}

func UpdateSegmentPgHbaConf(pgdata string, coordinatorAddrs []string, addrs []string) error {
	pgHbaFilePath := filepath.Join(pgdata, pgHbaConfFile)

	file, err := utils.System.Open(pgHbaFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	updatedLines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		updatedLines = append(updatedLines, line)
	}

	user, err := utils.System.CurrentUser()
	if err != nil {
		return err
	}

	// Add coordinator entries
	addPgHbaEntries(&updatedLines, coordinatorAddrs, "all", "all")

	// Add access entries
	addPgHbaEntries(&updatedLines, addrs, "all", user.Username)

	err = utils.WriteLinesToFile(pgHbaFilePath, updatedLines)
	if err != nil {
		return err
	}

	gplog.Info("Successfully updated %s for data directory %s", pgHbaConfFile, pgdata)
	return nil
}

func updateConfFile(filename, pgdata string, configParams map[string]string, overwrite bool) error {
	var line string
	confFilePath := filepath.Join(pgdata, filename)

	file, err := utils.System.Open(confFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	updatedLines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = scanner.Text()

		for key, value := range configParams {
			pattern, err := regexp.Compile(fmt.Sprintf("^%s[\\s=]+", key))
			if err != nil {
				return err
			}

			if pattern.MatchString(line) {
				if !overwrite {
					updatedLines = append(updatedLines, fmt.Sprintf("# %s", line))
				}
				line = fmt.Sprintf("%s = %s", key, quoteString(value))
				delete(configParams, key)
			}
		}

		updatedLines = append(updatedLines, line)
	}

	// Add the remaining entries
	for key, value := range configParams {
		line := fmt.Sprintf("%s = %s", key, quoteString(value))
		updatedLines = append(updatedLines, line)
	}

	err = utils.WriteLinesToFile(confFilePath, updatedLines)
	if err != nil {
		return err
	}

	return nil
}

func addPgHbaEntries(existingEntries *[]string, addrs []string, accessType string, user string) {
	var hostAccessString = "host\t%s\t%s\t%s\ttrust"

	for _, addr := range addrs {
		line := fmt.Sprintf(hostAccessString, accessType, user, addr)
		if !slices.Contains(*existingEntries, line) {
			*existingEntries = append(*existingEntries, line)
		}
	}
}

/*
Enclose string values inside quotes
*/
func quoteString(value string) string {
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return value
	} else {
		return fmt.Sprintf("'%s'", value)
	}
}
