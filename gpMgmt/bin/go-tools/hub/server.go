package hub

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/greenplum-db/gpdb/gp/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/reflection"
	grpcStatus "google.golang.org/grpc/status"
)

var (
	platform    = utils.GetOS()
	DialTimeout = 3 * time.Second
)

type Dialer func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)

type Config struct {
	Port        int      `json:"hubPort"`
	AgentPort   int      `json:"agentPort"`
	Hostnames   []string `json:"hostnames"`
	LogDir      string   `json:"hubLogDir"` // log directory for the hub itself; utilities might go somewhere else
	ServiceName string   `json:"serviceName"`
	GpHome      string   `json:"gphome"`

	*utils.Credentials
}

type Server struct {
	*Config
	conns      []*Connection
	grpcDialer Dialer

	mu     sync.Mutex
	server *grpc.Server
	lis    net.Listener
	finish chan struct{}
}

type Connection struct {
	Conn          *grpc.ClientConn
	AgentClient   idl.AgentClient
	Hostname      string
	CancelContext func()
}

func New(conf *Config, grpcDialer Dialer) *Server {
	h := &Server{
		Config:     conf,
		grpcDialer: grpcDialer,
		finish:     make(chan struct{}, 1),
	}
	return h
}

func (s *Server) Start() error {
	gplog.Debug("Entering function:Start")
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.Port)) // TODO: make this "hostname:port" so it can be started from somewhere other than the coordinator host
	if err != nil {
		gplog.Error("Could not listen on port %d: %s", s.Port, err.Error())
		return fmt.Errorf("Could not listen on port %d: %w", s.Port, err)
	}

	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// handle stuff here if needed
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

	idl.RegisterHubServer(server, s)
	reflection.Register(server)

	//sigChan := make(chan os.Signal, 1)
	//signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		select {
		//case sig := <-sigChan:
		//	gplog.Info("Received signal %v", sig)
		case <-s.finish:
			gplog.Info("Received stop command, attempting graceful shutdown")
			s.server.GracefulStop()
			gplog.Info("gRPC server has shut down")
		}
		cancel()
		wg.Done()
	}()

	err = server.Serve(lis)
	if err != nil {
		gplog.Error("Failed to serve: %s", err.Error())
		return fmt.Errorf("Failed to serve: %w", err)
	}
	wg.Wait()
	gplog.Debug("Exiting function:Start")
	return nil
}

func (s *Server) Stop(ctx context.Context, in *idl.StopHubRequest) (*idl.StopHubReply, error) {
	gplog.Debug("Entering function:Stop")
	s.Shutdown()
	gplog.Debug("Exiting function:Stop")
	return &idl.StopHubReply{}, nil
}

func (s *Server) Shutdown() {
	gplog.Debug("Entering function:Shutdown")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		s.finish <- struct{}{}
	}
	gplog.Debug("Exiting function:Shutdown")
}

func (s *Server) StartAgents(ctx context.Context, in *idl.StartAgentsRequest) (*idl.StartAgentsReply, error) {
	gplog.Debug("Entering function:StartAgents")
	err := s.StartAllAgents()
	if err != nil {
		gplog.Error("Error in starting all agents: %s", err.Error())
		return &idl.StartAgentsReply{}, err
	}
	err = s.DialAllAgents()
	if err != nil {
		gplog.Error("Error connecting agents: %s", err.Error())
		return &idl.StartAgentsReply{}, err
	}

	gplog.Debug("Exiting function:StartAgents")
	return &idl.StartAgentsReply{}, nil
}

func (s *Server) StartAllAgents() error {
	gplog.Debug("Entering function:StartAllAgents")
	var err error

	remoteCmd := make([]string, 0)
	for _, host := range s.Hostnames {
		remoteCmd = append(remoteCmd, "-h", host)
	}
	remoteCmd = append(remoteCmd, platform.GetStartAgentCmd(s.ServiceName)...)
	err = exec.Command("/bin/bash", "-c", fmt.Sprintf("source %s/greenplum_path.sh; gpssh %s", s.GpHome, strings.Join(remoteCmd, " "))).Run()
	if err != nil {
		gplog.Error("Could not start agent: %s", err.Error())
		return fmt.Errorf("Could not start agent: %w", err)
	}

	gplog.Debug("Exiting function:StartAllAgents")
	return nil
}

func (s *Server) DialAllAgents() error {
	gplog.Debug("Entering function:DialAllAgents")
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conns != nil {
		err := EnsureConnectionsAreReady(s.conns)
		if err != nil {
			gplog.Error("Could not ensure connections were ready: %s", err.Error())
			return fmt.Errorf("Could not ensure connections were ready: %w", err)
		}

		gplog.Debug("Exiting function:DialAllAgents")
		return nil
	}

	for _, host := range s.Hostnames {
		ctx, cancelFunc := context.WithTimeout(context.Background(), DialTimeout)

		credentials, err := s.Credentials.LoadClientCredentials()
		if err != nil {
			gplog.Error("Could not load credentials: %s", err.Error())
			return fmt.Errorf("Could not load credentials: %w", err)
		}

		address := fmt.Sprintf("%s:%d", host, s.AgentPort)
		// address := fmt.Sprintf("localhost:%d", s.AgentPort)
		conn, err := s.grpcDialer(ctx, address,
			grpc.WithBlock(),
			grpc.WithTransportCredentials(credentials),
		)
		if err != nil {
			cancelFunc()
			gplog.Error("Could not connect to agent: %s", err.Error())
			return fmt.Errorf("Could not connect to agent on host %s: %w", host, err)
		}
		s.conns = append(s.conns, &Connection{
			Conn:          conn,
			AgentClient:   idl.NewAgentClient(conn),
			Hostname:      host,
			CancelContext: cancelFunc,
		})
	}

	err := EnsureConnectionsAreReady(s.conns)
	if err != nil {
		gplog.Error("Could not ensure connections were ready: %s", err.Error())
		return fmt.Errorf("Could not ensure connections were ready: %w", err)
	}

	gplog.Debug("Exiting function:DialAllAgents")
	return nil
}

func (s *Server) StopAgents(ctx context.Context, in *idl.StopAgentsRequest) (*idl.StopAgentsReply, error) {
	gplog.Debug("Entering function:StopAgents")
	request := func(conn *Connection) error {
		_, err := conn.AgentClient.Stop(context.Background(), &idl.StopAgentRequest{})
		if err == nil { // no error -> didn't stop
			gplog.Error("Failed to stop agent on host %s", conn.Hostname)
			return fmt.Errorf("Failed to stop agent on host %s", conn.Hostname)
		}

		errStatus := grpcStatus.Convert(err)
		if errStatus.Code() != codes.Unavailable {
			gplog.Error("Failed to stop agent on host %s: %s", conn.Hostname, err.Error())
			return fmt.Errorf("Failed to stop agent on host %s: %w", conn.Hostname, err)
		}

		gplog.Debug("Entering function:StopAgents")
		return nil
	}

	err := s.DialAllAgents()
	if err != nil {
		gplog.Error("Error dialing agents:%s", err.Error())
		return &idl.StopAgentsReply{}, err
	}
	err = ExecuteRPC(s.conns, request)
	if err != nil {
		gplog.Error("Error executing RPC on connection. Error:%s", err.Error())
	}
	s.conns = nil
	return &idl.StopAgentsReply{}, err
}

func (s *Server) StatusAgents(ctx context.Context, in *idl.StatusAgentsRequest) (*idl.StatusAgentsReply, error) {
	gplog.Debug("Entering function:StatusAgents")
	statusChan := make(chan idl.ServiceStatus, len(s.conns))

	request := func(conn *Connection) error {
		status, err := conn.AgentClient.Status(context.Background(), &idl.StatusAgentRequest{})
		if err != nil { // no error -> didn't stop
			gplog.Error("Failed to get agent status on host %s", conn.Hostname)
			return fmt.Errorf("Failed to get agent status on host %s", conn.Hostname)
		}
		s := idl.ServiceStatus{
			Host:   conn.Hostname,
			Status: status.Status,
			Uptime: status.Uptime,
			Pid:    status.Pid,
		}
		statusChan <- s
		gplog.Debug("Exiting function:StatusAgents")
		return nil
	}

	err := s.DialAllAgents()
	if err != nil {
		gplog.Error("Error while dialing all agents:%s", err.Error())
		return &idl.StatusAgentsReply{}, err
	}
	err = ExecuteRPC(s.conns, request)
	if err != nil {
		gplog.Error("Error in executing RPC on connection: %s", err.Error())
		return &idl.StatusAgentsReply{}, err
	}
	close(statusChan)

	statuses := make([]*idl.ServiceStatus, 0)
	for status := range statusChan {
		s := status
		statuses = append(statuses, &s)
	}

	gplog.Debug("Exiting function:StatusAgents")
	return &idl.StatusAgentsReply{Statuses: statuses}, err
}

func EnsureConnectionsAreReady(conns []*Connection) error {
	gplog.Debug("Entering function:EnsureConnectionsAreReady")
	hostnames := []string{}
	for _, conn := range conns {
		if conn.Conn.GetState() != connectivity.Ready {
			hostnames = append(hostnames, conn.Hostname)
		}
	}

	if len(hostnames) > 0 {
		gplog.Error("unready hosts: %s", strings.Join(hostnames, ","))
		return fmt.Errorf("unready hosts: %s", strings.Join(hostnames, ","))
	}

	gplog.Debug("Exiting function:EnsureConnectionsAreReady")
	return nil
}

func ExecuteRPC(agentConns []*Connection, executeRequest func(conn *Connection) error) error {
	gplog.Debug("Entering function:ExecuteRPC")
	var wg sync.WaitGroup
	errs := make(chan error, len(agentConns))

	for _, conn := range agentConns {
		conn := conn
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := executeRequest(conn)
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	var err error
	for e := range errs {
		err = e
		break
	}

	gplog.Debug("Exiting function:ExecuteRPC")
	return err
}
