package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/greenplum-db/gpdb/gp/utils/postgres"
)

func (s *Server) RemoveDirectory(ctx context.Context, req *idl.RemoveDirectoryRequest) (*idl.RemoveDirectoryReply, error) {

	_, err := utils.System.Stat(req.DataDirectory)
	if err != nil {
		if !os.IsNotExist(err) {
			return &idl.RemoveDirectoryReply{}, utils.LogAndReturnError(fmt.Errorf("unexpected error :%w", err))
		}
	}

	//Check to see if there are postgres processes running, if true call gpstop.
	pgCtlStatusOptions := postgres.PgCtlStatus{
		PgData: req.DataDirectory,
	}

	_, err = utils.RunGpCommand(&pgCtlStatusOptions, s.GpHome)
	if err == nil {
		pgCtlStopOptions := postgres.PgCtlStop{
			PgData: req.DataDirectory,
			Mode:   "immediate",
		}
		out, err := utils.RunGpCommand(&pgCtlStopOptions, s.GpHome)
		if err != nil {
			gplog.Error("executing pg_ctl stop: %s, %v", out, err)
		}
	}

	err = utils.RemoveContents(req.DataDirectory)
	if err != nil {
		if !os.IsNotExist(err) {
			return &idl.RemoveDirectoryReply{}, fmt.Errorf("could not remove directory %s : %w", req.DataDirectory, err)
		}
	}

	return &idl.RemoveDirectoryReply{}, nil
}
