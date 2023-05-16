package agent

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	platform = utils.GetPlatform()
)

type Config struct {
	Port        int
	ServiceName string

	Credentials utils.Credentials
}

type Server struct {
	*Config

	mutex      sync.Mutex
	grpcServer *grpc.Server
	listener   net.Listener
}

func New(conf Config) *Server {
	return &Server{
		Config: &conf,
	}
}

func (s *Server) Stop(ctx context.Context, in *idl.StopAgentRequest) (*idl.StopAgentReply, error) {
	s.Shutdown()
	return &idl.StopAgentReply{}, nil
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.Port))
	if err != nil {
		gplog.Error("Could not listen on port %d: %s", s.Port, err.Error())
		return fmt.Errorf("Could not listen on port %d: %w", s.Port, err)
	}

	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// handle stuff here if needed
		return handler(ctx, req)
	}
	credentials, err := s.Credentials.LoadServerCredentials()
	if err != nil {
		gplog.Error("Could not load credentials: %s", err.Error())
		return fmt.Errorf("Could not load credentials: %w", err)
	}
	grpcServer := grpc.NewServer(
		grpc.Creds(credentials),
		grpc.UnaryInterceptor(interceptor),
	)

	s.mutex.Lock()
	s.grpcServer = grpcServer
	s.listener = listener
	s.mutex.Unlock()

	idl.RegisterAgentServer(grpcServer, s)
	reflection.Register(grpcServer)

	err = grpcServer.Serve(listener)
	if err != nil {
		gplog.Error("Failed to serve: %s", err.Error())
		return fmt.Errorf("Failed to serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
}

func (s *Server) Status(ctx context.Context, in *idl.StatusAgentRequest) (*idl.StatusAgentReply, error) {
	status, err := s.GetStatus()
	if err != nil {
		gplog.Error("Error getting status of Agent Service:%s", err.Error())
		return &idl.StatusAgentReply{}, nil
	}
	return &idl.StatusAgentReply{Status: status.Status, Uptime: status.Uptime, Pid: uint32(status.Pid)}, nil
}

func (s *Server) GetStatus() (*idl.ServiceStatus, error) {
	message, err := platform.GetServiceStatusMessage(fmt.Sprintf("%s_agent", s.ServiceName))
	if err != nil {
		gplog.Error("Error while getting Service Status Message:%s", err.Error())
		return nil, err
	}
	status := platform.ParseServiceStatusMessage(message)
	return &status, nil
}

func SetPlatform(p utils.Platform) {
	platform = p
}

func ResetPlatform() {
	platform = utils.GetPlatform()
}
