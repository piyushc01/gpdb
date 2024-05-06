package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/constants"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	ReadFile        = os.ReadFile
	Unmarshal       = json.Unmarshal
	DialContextFunc = grpc.DialContext
	ConnectToHub    = ConnectToHubFunc

	ConfigFilePath string
	Conf           *hub.Config

	Verbose bool
)

func RootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "gp",
	}

	root.PersistentFlags().StringVar(&ConfigFilePath, "config-file", filepath.Join(os.Getenv("GPHOME"), constants.ConfigFileName), `Path to gp configuration file`)
	root.PersistentFlags().BoolVar(&Verbose, "verbose", false, `Provide verbose output`)

	root.AddCommand(
		agentCmd(),
		configureCmd(),
		hubCmd(),
		startCmd(),
		statusCmd(),
		stopCmd(),
		initCmd(),
	)

	return root
}

// InitializeCommand Performs general setup needed for most commands
// Public, so it can be mocked out in testing
func InitializeCommand(cmd *cobra.Command, args []string) error {
	// TODO: Add a new constructor to gplog to allow initializing with a custom logfile path directly
	Conf = &hub.Config{}
	if _, err := os.Stat(ConfigFilePath); errors.Is(err, os.ErrNotExist) {
		// Config file does not exist

		gplog.Debug("gp config file does not exists. Starting with self-sufficient mode")
		// create default config file with no-TLS
		Conf.Port = constants.DefaultHubPort
		Conf.AgentPort = constants.DefaultAgentPort
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		Conf.LogDir = filepath.Join(homeDir, "gpAdminLogs")
		// TODO Populate hostlist later by reading the given config file before copy starts
		Conf.IsSecure = false
		Conf.GpHome = os.Getenv("GPHOME")
		Conf.ServiceName = "gp"
		Conf.IsConfigured = false
		// copy config file to all the hosts - gpsync
		// Start gp services using SSH
		// Go ahead with cluster creation

	} else {
		err = Conf.Load(ConfigFilePath)
		if err != nil {
			return err
		}
	}
	hubLogDir = Conf.LogDir

	err := InitializeLogger(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

func InitializeLogger(cmd *cobra.Command, args []string) error {
	// CommandPath lists the names of the called command and all of its parent commands, so this
	// turns e.g. "gp stop hub" into "gp_stop_hub" to generate a unique log file name for each command.
	logName := strings.ReplaceAll(cmd.CommandPath(), " ", "_")
	gplog.InitializeLogging(logName, hubLogDir)

	if Verbose {
		gplog.SetVerbosity(gplog.LOGVERBOSE)
	}

	return nil
}

func ConnectToHubFunc(conf *hub.Config) (idl.HubClient, error) {
	var conn *grpc.ClientConn

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var credentials credentials.TransportCredentials
	opts := []grpc.DialOption{grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithReturnConnectionError()}
	if conf.IsSecure {
		var err error
		credentials, err = conf.Credentials.LoadClientCredentials()
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	address := fmt.Sprintf("localhost:%d", conf.Port)
	conn, err := DialContextFunc(ctx, address, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to hub on port %d: %w", conf.Port, err)
	}

	return idl.NewHubClient(conn), nil
}
