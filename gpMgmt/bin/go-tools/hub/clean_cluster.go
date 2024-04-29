package hub

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/constants"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
)

/*
function to convert entries in entries.txt file into hashmap
*/
func CreateHostDataDirMap(entries []string, file string) (error, map[string][]string) {

	hostDataDirMap := make(map[string][]string)

	for _, entry := range entries {
		fields := strings.Fields(entry)
		if len(fields) == 2 {
			host := fields[0]
			dataDir := fields[1]
			hostDataDirMap[host] = append(hostDataDirMap[host], dataDir)
		} else {
			gplog.Debug(fmt.Sprintf("Invalid entries in %s", file))
			return errors.New("invalid entries in map"), nil
		}
	}
	return nil, hostDataDirMap
}

/*
rpc to cleanup the data directories in case gp init cluster fails.
*/
func (s *Server) CleanCluster(context.Context, *idl.CleanClusterRequest) (*idl.CleanClusterReply, error) {
	var err error

	fileName := filepath.Join(s.LogDir, constants.CleanFileName)

	_, err = utils.System.Stat(fileName)
	if err != nil {
		fmt.Printf("Cluster is clean")
		return &idl.CleanClusterReply{}, nil
	}

	//Read entries from fileName
	entries, err := utils.ReadEntriesFromFile(fileName)

	if err != nil {
		gplog.Debug(fmt.Sprintf("Error reading file %s", fileName))
		return &idl.CleanClusterReply{}, err
	}
	// Create hostDataDirMap from entries
	err, hostDataDirMap := CreateHostDataDirMap(entries, fileName)
	if err != nil {
		return &idl.CleanClusterReply{}, err
	}

	request := func(conn *Connection) error {
		var wg sync.WaitGroup

		hostname := hostDataDirMap[conn.Hostname]
		errs := make(chan error, len(hostname))

		for _, dir := range hostname {
			dir := dir
			wg.Add(1)

			go func(dirname string) {
				defer wg.Done()

				gplog.Debug(fmt.Sprintf("Removing Data Directories: %s", dir))
				_, err := conn.AgentClient.RemoveDirectory(context.Background(), &idl.RemoveDirectoryRequest{
					DataDirectory: dir,
				})
				if err != nil {
					errs <- utils.FormatGrpcError(err)
				}
			}(dir)
		}

		wg.Wait()
		close(errs)

		var err error
		for e := range errs {
			err = errors.Join(err, e)
		}
		return err
	}

	defer os.Remove(fileName)

	return &idl.CleanClusterReply{}, ExecuteRPC(s.Conns, request)
}
