package hub

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/greenplum-db/gpdb/gp/utils/greenplum"
	"github.com/greenplum-db/gpdb/gp/utils/postgres"
)

var execOnDatabaseFunc = ExecOnDatabase

func (s *Server) MakeCluster(request *idl.MakeClusterRequest, stream idl.Hub_MakeClusterServer) error {
	var err error
	var shutdownCoordinator bool

	// shutdown the coordinator segment if any error occurs
	defer func() {
		if err != nil && shutdownCoordinator {
			streamLogMsg(stream, &idl.LogMessage{Message: "Not able to create the the cluster, proceeding to shutdown the coordinator segment", Level: idl.LogLevel_WARNING})
			err := s.StopCoordinator(stream, request.GpArray.Coordinator.DataDirectory)
			if err != nil {
				gplog.Error(err.Error())
			}
		}
	}()

	err = s.DialAllAgents()
	if err != nil {
		return err
	}

	streamLogMsg(stream, &idl.LogMessage{Message: "Starting MakeCluster", Level: idl.LogLevel_INFO})
	err = s.ValidateEnvironment(stream, request)
	if err != nil {
		gplog.Error("Error during validation:%v", err)
		return err
	}

	streamLogMsg(stream, &idl.LogMessage{Message: "Creating coordinator segment", Level: idl.LogLevel_INFO})
	err = s.CreateAndStartCoordinator(request.GpArray.Coordinator, request.ClusterParams)
	if err != nil {
		return err
	}
	streamLogMsg(stream, &idl.LogMessage{Message: "Successfully created coordinator segment", Level: idl.LogLevel_INFO})

	shutdownCoordinator = true

	streamLogMsg(stream, &idl.LogMessage{Message: "Starting to register primary segments with the coordinator", Level: idl.LogLevel_INFO})

	conn, err := greenplum.ConnectDatabase(request.GpArray.Coordinator.HostName, int(request.GpArray.Coordinator.Port))
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	err = conn.Connect(1, true)
	if err != nil {
		return err
	}

	err = greenplum.RegisterCoordinator(request.GpArray.Coordinator, conn)
	if err != nil {
		return err
	}

	err = greenplum.RegisterPrimaries(request.GpArray.Primaries, conn)
	if err != nil {
		return err
	}
	streamLogMsg(stream, &idl.LogMessage{Message: "Successfully registered primary segments with the coordinator", Level: idl.LogLevel_INFO})

	gpArray := greenplum.NewGpArray()
	err = gpArray.ReadGpSegmentConfig(conn)
	if err != nil {
		return err
	}
	conn.Close()

	primarySegs, err := gpArray.GetPrimarySegments()
	if err != nil {
		return err
	}

	var coordinatorAddrs []string
	if request.ClusterParams.HbaHostnames {
		coordinatorAddrs = append(coordinatorAddrs, request.GpArray.Coordinator.HostAddress)
	} else {
		addrs, err := utils.GetHostAddrsNoLoopback()
		if err != nil {
			return err
		}

		coordinatorAddrs = append(coordinatorAddrs, addrs...)
	}

	streamLogMsg(stream, &idl.LogMessage{Message: "Creating primary segments", Level: idl.LogLevel_INFO})
	err = s.CreateSegments(stream, primarySegs, request.ClusterParams, coordinatorAddrs)
	if err != nil {
		return err
	}
	streamLogMsg(stream, &idl.LogMessage{Message: "Successfully created primary segments", Level: idl.LogLevel_INFO})

	shutdownCoordinator = false

	streamLogMsg(stream, &idl.LogMessage{Message: "Restarting the Greenplum cluster in production mode", Level: idl.LogLevel_INFO})
	err = s.StopCoordinator(stream, request.GpArray.Coordinator.DataDirectory)
	if err != nil {
		return err
	}

	gpstartOptions := &greenplum.GpStart{
		DataDirectory: request.GpArray.Coordinator.DataDirectory,
		Verbose:       request.Verbose,
	}
	cmd := utils.NewGpSourcedCommand(gpstartOptions, s.GpHome)
	err = streamExecCommand(stream, cmd, s.GpHome)
	if err != nil {
		return fmt.Errorf("executing gpstart: %w", err)
	}
	streamLogMsg(stream, &idl.LogMessage{Message: "Completed restart of Greenplum cluster in production mode", Level: idl.LogLevel_INFO})

	streamLogMsg(stream, &idl.LogMessage{Message: "Creating core GPDB extensions", Level: idl.LogLevel_INFO})
	err = CreateGpToolkitExt(conn)
	if err != nil {
		return err
	}
	streamLogMsg(stream, &idl.LogMessage{Message: "Successfully created core GPDB extensions", Level: idl.LogLevel_INFO})

	streamLogMsg(stream, &idl.LogMessage{Message: "Importing system collations", Level: idl.LogLevel_INFO})
	err = ImportCollation(conn)
	if err != nil {
		return err
	}

	if request.ClusterParams.DbName != "" {
		streamLogMsg(stream, &idl.LogMessage{Message: fmt.Sprintf("Creating database %q", request.ClusterParams.DbName), Level: idl.LogLevel_INFO})
		err = CreateDatabase(conn, request.ClusterParams.DbName)
		if err != nil {
			return err
		}
	}

	streamLogMsg(stream, &idl.LogMessage{Message: "Setting Greenplum superuser password", Level: idl.LogLevel_INFO})
	err = SetGpUserPasswd(conn, request.ClusterParams.SuPassword)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) ValidateEnvironment(stream idl.Hub_MakeClusterServer, request *idl.MakeClusterRequest) error {
	var replies []*idl.LogMessage

	gparray := request.GpArray
	hostDirMap := make(map[string][]string)
	hostPortMap := make(map[string][]int32)
	hostAddressMap := make(map[string]map[string]bool)

	// Add coordinator to the map
	hostDirMap[gparray.Coordinator.HostName] = append(hostDirMap[gparray.Coordinator.HostName], gparray.Coordinator.DataDirectory)
	hostPortMap[gparray.Coordinator.HostName] = append(hostPortMap[gparray.Coordinator.HostName], gparray.Coordinator.Port)
	hostAddressMap[gparray.Coordinator.HostName] = make(map[string]bool)
	hostAddressMap[gparray.Coordinator.HostName][gparray.Coordinator.HostAddress] = true

	// Add primaries to the map
	for _, seg := range gparray.Primaries {
		hostDirMap[seg.HostName] = append(hostDirMap[seg.HostName], seg.DataDirectory)
		hostPortMap[seg.HostName] = append(hostPortMap[seg.HostName], seg.Port)

		if hostAddressMap[seg.HostName] == nil {
			hostAddressMap[seg.HostName] = make(map[string]bool)
		}
		hostAddressMap[seg.HostName][seg.HostAddress] = true
	}
	gplog.Debug("Host-Address-Map:[%v]", hostAddressMap)

	// Get local gpVersion

	localPgVersion, err := greenplum.GetPostgresGpVersion(s.GpHome)
	if err != nil {
		gplog.Error("fetching postgres gp-version:%v", err)
		return err
	}

	progressLabel := "Validating Hosts:"
	progressTotal := len(hostDirMap)
	streamProgressMsg(stream, progressLabel, progressTotal)
	validateFn := func(conn *Connection) error {
		dirList := hostDirMap[conn.Hostname]
		portList := hostPortMap[conn.Hostname]
		var addressList []string
		for address := range hostAddressMap[conn.Hostname] {
			addressList = append(addressList, address)
		}
		gplog.Debug("AddressList:[%v]", addressList)

		validateReq := idl.ValidateHostEnvRequest{
			DirectoryList:   dirList,
			Locale:          request.ClusterParams.Locale,
			PortList:        portList,
			Forced:          request.ForceFlag,
			HostAddressList: addressList,
			GpVersion:       localPgVersion,
		}
		reply, err := conn.AgentClient.ValidateHostEnv(context.Background(), &validateReq)
		if err != nil {
			return utils.FormatGrpcError(err)
		}

		streamProgressMsg(stream, progressLabel, progressTotal)
		replies = append(replies, reply.Messages...)

		return nil
	}

	err = ExecuteRPC(s.Conns, validateFn)
	if err != nil {
		return err
	}

	for _, msg := range replies {
		streamLogMsg(stream, msg)
	}

	return nil
}

func CreateSingleSegment(conn *Connection, seg *idl.Segment, clusterParams *idl.ClusterParams, coordinatorAddrs []string) error {
	pgConfig := make(map[string]string)
	maps.Copy(pgConfig, clusterParams.CommonConfig)
	if seg.Contentid == -1 {
		maps.Copy(pgConfig, clusterParams.CoordinatorConfig)
	} else {
		maps.Copy(pgConfig, clusterParams.SegmentConfig)
	}

	makeSegmentReq := &idl.MakeSegmentRequest{
		Segment:          seg,
		Locale:           clusterParams.Locale,
		Encoding:         clusterParams.Encoding,
		SegConfig:        pgConfig,
		CoordinatorAddrs: coordinatorAddrs,
		HbaHostNames:     clusterParams.HbaHostnames,
		DataChecksums:    clusterParams.DataChecksums,
	}

	_, err := conn.AgentClient.MakeSegment(context.Background(), makeSegmentReq)
	if err != nil {
		return utils.FormatGrpcError(err)
	}

	return nil
}

func (s *Server) CreateAndStartCoordinator(seg *idl.Segment, clusterParams *idl.ClusterParams) error {
	coordinatorConn := getConnByHost(s.Conns, []string{seg.HostName})

	seg.Contentid = -1
	seg.Dbid = 1
	request := func(conn *Connection) error {
		err := CreateSingleSegment(conn, seg, clusterParams, []string{})
		if err != nil {
			return err
		}

		startSegReq := &idl.StartSegmentRequest{
			DataDir: seg.DataDirectory,
			Wait:    true,
			Options: "-c gp_role=utility",
		}
		_, err = conn.AgentClient.StartSegment(context.Background(), startSegReq)

		return utils.FormatGrpcError(err)
	}

	return ExecuteRPC(coordinatorConn, request)
}

func (s *Server) StopCoordinator(stream idl.Hub_MakeClusterServer, pgdata string) error {
	streamLogMsg(stream, &idl.LogMessage{Message: "Shutting down coordinator segment", Level: idl.LogLevel_INFO})
	pgCtlStopCmd := &postgres.PgCtlStop{
		PgData: pgdata,
	}

	out, err := utils.RunExecCommand(pgCtlStopCmd, s.GpHome)
	if err != nil {
		return fmt.Errorf("executing pg_ctl stop: %s, %w", out, err)
	}
	streamLogMsg(stream, &idl.LogMessage{Message: "Successfully shut down coordinator segment", Level: idl.LogLevel_INFO})

	return nil
}

func (s *Server) CreateSegments(stream idl.Hub_MakeClusterServer, segs []greenplum.Segment, clusterParams *idl.ClusterParams, coordinatorAddrs []string) error {
	hostSegmentMap := map[string][]*idl.Segment{}
	for _, seg := range segs {
		segReq := &idl.Segment{
			Port:          int32(seg.Port),
			DataDirectory: seg.DataDirectory,
			HostName:      seg.HostName,
			HostAddress:   seg.HostAddress,
			Contentid:     int32(seg.ContentId),
			Dbid:          int32(seg.Dbid),
		}

		if _, ok := hostSegmentMap[seg.HostName]; !ok {
			hostSegmentMap[seg.HostName] = []*idl.Segment{segReq}
		} else {
			hostSegmentMap[seg.HostName] = append(hostSegmentMap[seg.HostName], segReq)
		}
	}

	progressLabel := "Initializing segments:"
	progressTotal := len(segs)
	streamProgressMsg(stream, progressLabel, progressTotal)

	request := func(conn *Connection) error {
		var wg sync.WaitGroup

		segs := hostSegmentMap[conn.Hostname]
		errs := make(chan error, len(segs))
		for _, seg := range segs {
			seg := seg
			wg.Add(1)
			go func(seg *idl.Segment) {
				defer wg.Done()

				err := CreateSingleSegment(conn, seg, clusterParams, coordinatorAddrs)
				if err != nil {
					errs <- err
				} else {
					streamProgressMsg(stream, progressLabel, progressTotal)
				}
			}(seg)
		}

		wg.Wait()
		close(errs)

		var err error
		for e := range errs {
			err = errors.Join(err, e)
		}
		return err
	}

	return ExecuteRPC(s.Conns, request)
}

func ExecOnDatabase(conn *dbconn.DBConn, dbname string, query string) error {
	conn.DBName = dbname
	if err := conn.Connect(1); err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.Exec(query); err != nil {
		return err
	}

	return nil
}

func CreateGpToolkitExt(conn *dbconn.DBConn) error {
	createExtensionQuery := "CREATE EXTENSION gp_toolkit"

	for _, dbname := range []string{"template1", "postgres"} {
		if err := execOnDatabaseFunc(conn, dbname, createExtensionQuery); err != nil {
			return err
		}
	}

	return nil
}

func ImportCollation(conn *dbconn.DBConn) error {
	importCollationQuery := "SELECT pg_import_system_collations('pg_catalog'); ANALYZE;"

	if err := execOnDatabaseFunc(conn, "postgres", "ALTER DATABASE template0 ALLOW_CONNECTIONS on"); err != nil {
		return err
	}

	if err := execOnDatabaseFunc(conn, "template0", importCollationQuery); err != nil {
		return err
	}
	if err := execOnDatabaseFunc(conn, "template0", "VACUUM FREEZE"); err != nil {
		return err
	}

	if err := execOnDatabaseFunc(conn, "postgres", "ALTER DATABASE template0 ALLOW_CONNECTIONS off"); err != nil {
		return err
	}

	for _, dbname := range []string{"template1", "postgres"} {
		if err := execOnDatabaseFunc(conn, dbname, importCollationQuery); err != nil {
			return err
		}

		if err := execOnDatabaseFunc(conn, dbname, "VACUUM FREEZE"); err != nil {
			return err
		}
	}

	return nil
}

func CreateDatabase(conn *dbconn.DBConn, dbname string) error {
	createDbQuery := fmt.Sprintf("CREATE DATABASE %s", dbname)
	if err := execOnDatabaseFunc(conn, "template1", createDbQuery); err != nil {
		return err
	}

	return nil
}

func SetGpUserPasswd(conn *dbconn.DBConn, passwd string) error {
	user, err := utils.System.CurrentUser()
	if err != nil {
		return err
	}

	alterPasswdQuery := fmt.Sprintf("ALTER USER %s WITH PASSWORD '%s'", user.Username, passwd)
	if err := execOnDatabaseFunc(conn, "template1", alterPasswdQuery); err != nil {
		return err
	}

	return nil
}

func SetExecOnDatabase(customFunc func(*dbconn.DBConn, string, string) error) {
	execOnDatabaseFunc = customFunc
}

func ResetExecOnDatabase() {
	execOnDatabaseFunc = ExecOnDatabase
}
