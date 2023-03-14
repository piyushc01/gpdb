package agent

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	s.Shutdown()
	return &idl.StopAgentReply{}, nil
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.Port))
	if err != nil {
		return fmt.Errorf("Could not listen on port %d: %w", s.Port, err)
	}

	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// handle stuff here if needed
		return handler(ctx, req)
	}
	credentials, err := s.LoadServerCredentials()
	if err != nil {
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
		return fmt.Errorf("Failed to serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		s.server.Stop()
	}
}

func (s *Server) Status(ctx context.Context, in *idl.StatusAgentRequest) (*idl.StatusAgentReply, error) {
	status, err := s.GetStatus()
	if err != nil {
		return &idl.StatusAgentReply{}, nil
	}
	return &idl.StatusAgentReply{Status: status.Status, Uptime: status.Uptime, Pid: uint32(status.Pid)}, nil
}

func (s *Server) GetStatus() (*idl.ServiceStatus, error) {
	message, err := utils.GetServiceStatusMessage(fmt.Sprintf("%s_agent", s.ServiceName))
	if err != nil {
		return nil, err
	}
	status := utils.ParseServiceStatusMessage(message)
	return &status, nil
}
