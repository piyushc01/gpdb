package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
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
	platform       = utils.GetPlatform()
	DialTimeout    = 3 * time.Second
	ReadFile       = os.ReadFile
	OpenFile       = os.OpenFile
	Unmarshal      = json.Unmarshal
	MasrshalIndent = json.MarshalIndent
	ExecCommand    = exec.Command
)

type Dialer func(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error)

type Config struct {
	Port        int      `json:"hubPort"`
	AgentPort   int      `json:"agentPort"`
	Hostnames   []string `json:"hostnames"`
	LogDir      string   `json:"hubLogDir"` // log directory for the hub itself; utilities might go somewhere else
	ServiceName string   `json:"serviceName"`
	GpHome      string   `json:"gphome"`

	Credentials utils.Credentials
}

type Server struct {
	*Config
	conns      []*Connection
	grpcDialer Dialer

	mutex      sync.Mutex
	grpcServer *grpc.Server
	listener   net.Listener
	finish     chan struct{}
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
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.Port)) // TODO: make this "hostname:port" so it can be started from somewhere other than the coordinator host
	if err != nil {
		return fmt.Errorf("Could not listen on port %d: %w", s.Port, err)
	}

	interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// handle stuff here if needed
		return handler(ctx, req)
	}

	credentials, err := s.Credentials.LoadServerCredentials()
	if err != nil {
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

	idl.RegisterHubServer(grpcServer, s)
	reflection.Register(grpcServer)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		<-s.finish
		gplog.Info("Received stop command, attempting graceful shutdown")
		s.grpcServer.GracefulStop()
		gplog.Info("gRPC server has shut down")
		cancel()
		wg.Done()
	}()

	err = grpcServer.Serve(listener)
	if err != nil {
		return fmt.Errorf("Failed to serve: %w", err)
	}
	wg.Wait()
	return nil
}

func (s *Server) Stop(ctx context.Context, in *idl.StopHubRequest) (*idl.StopHubReply, error) {
	s.Shutdown()
	return &idl.StopHubReply{}, nil
}

func (s *Server) Shutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.grpcServer != nil {
		s.finish <- struct{}{}
	}
}

func (s *Server) StartAgents(ctx context.Context, in *idl.StartAgentsRequest) (*idl.StartAgentsReply, error) {
	err := s.StartAllAgents()
	if err != nil {
		return &idl.StartAgentsReply{}, fmt.Errorf("Could not start agents: %w", err)
	}
	err = s.DialAllAgents()
	if err != nil {
		return &idl.StartAgentsReply{}, fmt.Errorf("Could not dial agents: %w", err)
	}

	return &idl.StartAgentsReply{}, nil
}

func (s *Server) StartAllAgents() error {
	var outb, errb bytes.Buffer

	remoteCmd := make([]string, 0)
	for _, host := range s.Hostnames {
		remoteCmd = append(remoteCmd, "-h", host)
	}
	remoteCmd = append(remoteCmd, platform.GetStartAgentCommandString(s.ServiceName)...)

	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("source %s/greenplum_path.sh; gpssh %s", s.GpHome, strings.Join(remoteCmd, " ")))
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Could not start agent: %w", err)
	}
	//there are chances in most of the cases the command returns nil err even if there is error in stdout.
	//to overcome this we have added check to handle both
	if len(errb.String()) > 0 || strings.Contains(outb.String(), "ERROR") {
		errString := outb.String()
		if len(errb.String()) > 0 {
			errString = errb.String()
		}
		return fmt.Errorf("Could not start agent: %s", errString)
	}

	return nil
}

func (s *Server) DialAllAgents() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.conns != nil {
		err := EnsureConnectionsAreReady(s.conns)
		if err != nil {
			return fmt.Errorf("Could not ensure connections were ready: %w", err)
		}

		return nil
	}

	for _, host := range s.Hostnames {
		ctx, cancelFunc := context.WithTimeout(context.Background(), DialTimeout)

		credentials, err := s.Credentials.LoadClientCredentials()
		if err != nil {
			cancelFunc()
			return fmt.Errorf("Could not load credentials: %w", err)
		}

		address := fmt.Sprintf("%s:%d", host, s.AgentPort)
		conn, err := s.grpcDialer(ctx, address,
			grpc.WithBlock(),
			grpc.WithTransportCredentials(credentials),
			grpc.WithReturnConnectionError(),
		)
		if err != nil {
			cancelFunc()
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
		return fmt.Errorf("Could not ensure connections were ready: %w", err)
	}

	return nil
}

func (s *Server) StopAgents(ctx context.Context, in *idl.StopAgentsRequest) (*idl.StopAgentsReply, error) {
	request := func(conn *Connection) error {
		_, err := conn.AgentClient.Stop(context.Background(), &idl.StopAgentRequest{})
		if err == nil { // no error -> didn't stop
			return fmt.Errorf("Failed to stop agent on host %s", conn.Hostname)
		}

		errStatus := grpcStatus.Convert(err)
		if errStatus.Code() != codes.Unavailable {
			return fmt.Errorf("Failed to stop agent on host %s: %w", conn.Hostname, err)
		}

		return nil
	}

	err := s.DialAllAgents()
	if err != nil {
		return &idl.StopAgentsReply{}, err
	}

	err = ExecuteRPC(s.conns, request)
	s.conns = nil

	return &idl.StopAgentsReply{}, err
}

func (s *Server) StatusAgents(ctx context.Context, in *idl.StatusAgentsRequest) (*idl.StatusAgentsReply, error) {
	statusChan := make(chan *idl.ServiceStatus, len(s.conns))

	request := func(conn *Connection) error {
		status, err := conn.AgentClient.Status(context.Background(), &idl.StatusAgentRequest{})
		if err != nil { // no error -> didn't stop
			return fmt.Errorf("Failed to get agent status on host %s", conn.Hostname)
		}
		s := idl.ServiceStatus{
			Host:   conn.Hostname,
			Status: status.Status,
			Uptime: status.Uptime,
			Pid:    status.Pid,
		}
		statusChan <- &s
		return nil
	}

	err := s.DialAllAgents()
	if err != nil {
		return &idl.StatusAgentsReply{}, err
	}
	err = ExecuteRPC(s.conns, request)
	if err != nil {
		return &idl.StatusAgentsReply{}, err
	}
	close(statusChan)

	statuses := make([]*idl.ServiceStatus, 0)
	for status := range statusChan {
		statuses = append(statuses, status)
	}

	return &idl.StatusAgentsReply{Statuses: statuses}, err
}

func EnsureConnectionsAreReady(conns []*Connection) error {
	hostnames := []string{}
	for _, conn := range conns {
		if conn.Conn.GetState() != connectivity.Ready {
			hostnames = append(hostnames, conn.Hostname)
		}
	}

	if len(hostnames) > 0 {
		return fmt.Errorf("unready hosts: %s", strings.Join(hostnames, ","))
	}

	return nil
}

func ExecuteRPC(agentConns []*Connection, executeRequest func(conn *Connection) error) error {
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

	return err
}

func (conf *Config) Load(ConfigFilePath string) error {
	//Loads config from the configFilePath
	conf.Credentials = &utils.GpCredentials{}
	contents, err := ReadFile(ConfigFilePath)
	if err != nil {
		return err
	}
	err = Unmarshal(contents, &conf)
	if err != nil {
		return err
	}
	return nil
}

func (conf *Config) Write(ConfigFilePath string) error {
	// Updates config to the conf file
	configHandle, err := OpenFile(ConfigFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("Could not create configuration file %s: %w\n", ConfigFilePath, err)
	}
	defer configHandle.Close()
	configContents, err := MasrshalIndent(conf, "", "\t")
	if err != nil {
		return fmt.Errorf("Could not parse configuration file %s: %w\n", ConfigFilePath, err)
	}
	_, err = configHandle.Write(configContents)
	if err != nil {
		return fmt.Errorf("Could not write to configuration file %s: %w\n", ConfigFilePath, err)
	}
	gplog.Debug("Wrote configuration file to %s", ConfigFilePath)

	err = CopyConfigFileToAgents(conf, ConfigFilePath)
	if err != nil {
		return fmt.Errorf("Could not copy config file to hosts:%w", err)
	}
	return err
}

var CopyConfigFileToAgents = func(conf *Config, ConfigFilePath string) error {
	hostList := make([]string, 0)
	for _, host := range conf.Hostnames {
		hostList = append(hostList, "-h", host)
	}
	if len(hostList) < 1 {
		return fmt.Errorf("Hostlist should not be empty. No hosts to copy files.")
	}
	remoteCmd := append(hostList, ConfigFilePath, fmt.Sprintf("=:%s", ConfigFilePath))
	cmd := exec.Command("/bin/bash", "-c", fmt.Sprintf("source %s/greenplum_path.sh; gpsync %s", conf.GpHome, strings.Join(remoteCmd, " ")))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Could not copy gp.conf file to segment hosts: %w", err)
	}
	return nil
}
