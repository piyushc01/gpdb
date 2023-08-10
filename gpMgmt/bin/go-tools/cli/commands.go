package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	ReadFile    = os.ReadFile
	Unmarshal   = json.Unmarshal
	DialContext = grpc.DialContext

	ConfigFilePath        string
	conf                  *hub.Config
	defaultConfigFilePath string = "%s/gp.conf"

	verbose bool
)

func RootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "gp",
	}

	root.PersistentFlags().StringVar(&ConfigFilePath, "config-file", fmt.Sprintf(defaultConfigFilePath, os.Getenv("GPHOME")), `Path to gp configuration file`)
	root.PersistentFlags().BoolVar(&verbose, "verbose", false, `Provide verbose output`)

	root.AddCommand(agentCmd())
	root.AddCommand(hubCmd())
	root.AddCommand(installCmd())
	root.AddCommand(startCmd())
	root.AddCommand(statusCmd())
	root.AddCommand(stopCmd())

	return root
}

/*
 * Various helper functions used by multiple CLI commands
 */

// Performs general setup needed for most commands
// Public, so it can be mocked out in testing
func InitializeCommand(cmd *cobra.Command, args []string) error {
	// TODO: Add a new constructor to gplog to allow initializing with a custom logfile path directly
	InitializeGplog(cmd, args)
	conf = &hub.Config{}
	err := conf.Load(ConfigFilePath)
	if err != nil {
		return fmt.Errorf("Error parsing config file: %s\n", err.Error())
	}

	return nil
}

func InitializeGplog(cmd *cobra.Command, args []string) {
	// CommandPath lists the names of the called command and all of its parent commands, so this
	// turns e.g. "gp stop hub" into "gp_stop_hub" to generate a unique log file name for each command.
	logName := strings.ReplaceAll(cmd.CommandPath(), " ", "_")
	gplog.SetLogFileNameFunc(func(program string, logdir string) string {
		return fmt.Sprintf("%s/%s.log", hubLogDir, logName)
	})
	gplog.InitializeLogging(logName, "")

	timeFormat := time.Now().Format("2006-01-02 15:04:05.000000")
	hostname, _ := os.Hostname()
	gplog.SetLogPrefixFunc(func(level string) string {
		return fmt.Sprintf("%s %s  [%s] ", timeFormat, hostname, level) // TODO: decide what prefix we want, assuming we want one, but we *definitely* don't want the legacy one
	})
	if verbose {
		gplog.SetVerbosity(gplog.LOGVERBOSE)
	}
}

func connectToHub(conf *hub.Config) (idl.HubClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	credentials, err := conf.Credentials.LoadClientCredentials()
	if err != nil {
		return nil, fmt.Errorf("Could not load credentials: %w", err)
	}

	address := fmt.Sprintf("localhost:%d", conf.Port)
	var conn *grpc.ClientConn
	conn, err = DialContext(ctx, address,
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
