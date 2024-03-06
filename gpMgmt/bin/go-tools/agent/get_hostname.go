package agent

import (
	"context"
	"fmt"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"os"
)

func (s *Server) GetHostName(ctx context.Context, request *idl.GetHostNameRequest) (*idl.GetHostNameReply, error) {
	hostname, err := os.Hostname()
	if err != nil {
		strErr := fmt.Sprintf("error getting hostname:%v", err)
		gplog.Error(strErr)
		return &idl.GetHostNameReply{}, err
	}
	return &idl.GetHostNameReply{Hostname: hostname}, nil
}
