package hub

import (
	"context"
	"fmt"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"google.golang.org/grpc"
	"sync"
)

type RpcReply struct {
	hostname, address string
}

func (s *Server) GetAllHostNames(ctx context.Context, request *idl.GetAllHostNamesRequest) (*idl.GetAllHostNamesReply, error) {
	gplog.Debug("Starting with rpc GetAllHostNames")
	addressConnectionMap := make(map[string]idl.AgentClient)
	ctx, cancelFunc := context.WithTimeout(context.Background(), DialTimeout)

	credentials, err := s.Credentials.LoadClientCredentials()
	if err != nil {
		cancelFunc()
		return &idl.GetAllHostNamesReply{}, err
	}
	for _, address := range request.HostList {
		// Dial to the address
		if _, ok := addressConnectionMap[address]; !ok {
			remoteAddress := fmt.Sprintf("%s:%d", address, s.AgentPort)
			opts := []grpc.DialOption{
				grpc.WithBlock(),
				grpc.WithTransportCredentials(credentials),
				grpc.WithReturnConnectionError(),
			}
			if s.grpcDialer != nil {
				opts = append(opts, grpc.WithContextDialer(s.grpcDialer))
			}
			conn, err := grpc.DialContext(ctx, remoteAddress, opts...)
			if err != nil {
				cancelFunc()
				return &idl.GetAllHostNamesReply{}, fmt.Errorf("could not connect to agent on host %s: %w", address, err)
			}
			addressConnectionMap[address] = idl.NewAgentClient(conn)
		}
	}
	var wg sync.WaitGroup
	errs := make(chan error, len(addressConnectionMap))
	replies := make(chan RpcReply)
	for address, conn := range addressConnectionMap {
		wg.Add(1)
		go func() {
			defer wg.Done()
			request := idl.GetHostNameRequest{}
			reply, err := conn.GetHostName(context.Background(), &request)
			if err != nil {
				errs <- fmt.Errorf("host: %s, %w", address, err)
			}
			replies <- RpcReply{address: address, hostname: reply.Hostname}
		}()
	}
	wg.Wait()

	// Check for errors
	if len(errs) > 0 {
		for e := range errs {
			err = e
			break
		}
		return &idl.GetAllHostNamesReply{}, err
	}

	// Extract replies and populate reply
	hostNameMap := make(map[string]string)
	for reply := range replies {
		hostNameMap[reply.address] = reply.hostname
	}

	return &idl.GetAllHostNamesReply{HostNameMap: hostNameMap}, nil
}
