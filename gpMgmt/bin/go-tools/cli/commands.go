package cli

import (
	"context"
	"encoding/json"
	"fmt"
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
	ConnectToHub    = ConnectToHubFn

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

	root.AddCommand(agentCmd())
	root.AddCommand(hubCmd())
	root.AddCommand(configureCmd())
	root.AddCommand(startCmd())
	root.AddCommand(statusCmd())
	root.AddCommand(stopCmd())

	return root
}

// Performs general setup needed for most commands
// Public, so it can be mocked out in testing
func InitializeCommand(cmd *cobra.Command, args []string) error {
	// TODO: Add a new constructor to gplog to allow initializing with a custom logfile path directly
	Conf = &hub.Config{}
	err := Conf.Load(ConfigFilePath)
	if err != nil {
		return fmt.Errorf("Error parsing config file: %s\n", err.Error())
	}
	hubLogDir = Conf.LogDir
	InitializeGplog(cmd, args)

	return nil
}

func InitializeGplog(cmd *cobra.Command, args []string) {
	// CommandPath lists the names of the called command and all of its parent commands, so this
	// turns e.g. "gp stop hub" into "gp_stop_hub" to generate a unique log file name for each command.
	logName := strings.ReplaceAll(cmd.CommandPath(), " ", "_")
	gplog.SetLogFileNameFunc(func(program string, logdir string) string {
		// TODO: Check if this path exists or not, otherwise gplog panics
		return filepath.Join(hubLogDir, fmt.Sprintf("%s.log", logName))
	})
	gplog.InitializeLogging(logName, "")

	timeFormat := time.Now().Format("2006-01-02 15:04:05.000000")
	hostname, _ := os.Hostname()
	gplog.SetLogPrefixFunc(func(level string) string {
		return fmt.Sprintf("%s %s  [%s] ", timeFormat, hostname, level) // TODO: decide what prefix we want, assuming we want one, but we *definitely* don't want the legacy one
	})
	if Verbose {
		gplog.SetVerbosity(gplog.LOGDEBUG)
	}
}

func ConnectToHubFn(conf *hub.Config) (idl.HubClient, error) {
	var conn *grpc.ClientConn

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	credentials, err := conf.Credentials.LoadClientCredentials()
	if err != nil {
		return nil, fmt.Errorf("Could not load hub credentials: %w", err)
	}

	address := fmt.Sprintf("localhost:%d", conf.Port)
	conn, err = DialContextFunc(ctx, address,
		grpc.WithTransportCredentials(credentials),
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithReturnConnectionError(),
	)
	if err != nil {
		return nil, fmt.Errorf("Could not connect to hub on port %d: %w", conf.Port, err)
	}

	return idl.NewHubClient(conn), nil
}
