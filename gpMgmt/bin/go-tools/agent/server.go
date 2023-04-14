package agent

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sync"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	platform = utils.GetOS()
)

type Config struct {
	Port        int
	ServiceName string

	*utils.Credentials
}

type Server struct {
	*Config

	mu     sync.Mutex
	server *grpc.Server
	lis    net.Listener
	stop   chan int
}

func New(conf Config) *Server {
	return &Server{
		Config: &conf,
	}
}

func (s *Server) Stop(ctx context.Context, in *idl.StopAgentRequest) (*idl.StopAgentReply, error) {
	gplog.Debug("Entering function:Stop")
	s.Shutdown()
	gplog.Debug("Exiting function:Stop")
	return &idl.StopAgentReply{}, nil
}

func (s *Server) Start() error {
	gplog.Debug("Entering function:Start")
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.Port))
	if err != nil {
		gplog.Error("Could not listen on port %d: %s", s.Port, err.Error())
		return fmt.Errorf("Could not listen on port %d: %w", s.Port, err)
	}

	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// handle stuff here if needed
		gplog.Debug("Exiting function:Start")
		return handler(ctx, req)
	}
	credentials, err := s.LoadServerCredentials()
	if err != nil {
		gplog.Error("Could not load credentials: %s", err.Error())
		return fmt.Errorf("Could not load credentials: %w", err)
	}
	server := grpc.NewServer(
		grpc.Creds(credentials),
		grpc.UnaryInterceptor(interceptor),
	)

	s.mu.Lock()
	s.server = server
	s.lis = lis
	s.mu.Unlock()

	idl.RegisterAgentServer(server, s)
	reflection.Register(server)

	err = server.Serve(lis)
	if err != nil {
		gplog.Error("Failed to serve: %s", err.Error())
		return fmt.Errorf("Failed to serve: %w", err)
	}
	gplog.Debug("Exiting function:Start")
	return nil
}

func (s *Server) Shutdown() {
	gplog.Debug("Entering function:Shutdown")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		s.server.Stop()
	}
	gplog.Debug("Exiting function:Shutdown")
}

func (s *Server) Status(ctx context.Context, in *idl.StatusAgentRequest) (*idl.StatusAgentReply, error) {
	gplog.Debug("Entering function:Status")
	status, err := s.GetStatus()
	if err != nil {
		gplog.Error("Error getting status of Agent Service:%s", err.Error())
		return &idl.StatusAgentReply{}, nil
	}
	gplog.Debug("Exiting function:Status")
	return &idl.StatusAgentReply{Status: status.Status, Uptime: status.Uptime, Pid: uint32(status.Pid)}, nil
}

func (s *Server) GetStatus() (*idl.ServiceStatus, error) {
	gplog.Debug("Entering function:GetStatus")
	service := ""
	if runtime.GOOS == "linux" {
		service = fmt.Sprintf("%s_agent", s.ServiceName)
	} else if runtime.GOOS == "darwin" {
		service = fmt.Sprintf("%s.agent", s.ServiceName)
	}
	message, err := platform.GetServiceStatusMessage(service)
	if err != nil {
		gplog.Error("Error while getting Service Status Message:%s", err.Error())
		return nil, err
	}
	status := platform.ParseServiceStatusMessage(message)
	gplog.Debug("Exiting function:GetStatus")
	return &status, nil
}
