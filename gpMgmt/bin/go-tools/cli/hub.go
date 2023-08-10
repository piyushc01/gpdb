package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	platform                  = utils.GetPlatform()
	DefaultHubLogDir   string = "/tmp"
	DefaultHubPort     int    = 4242
	DefaultAgentPort   int    = 8000
	DefaultServiceDir  string = platform.GetDefaultServiceDir()
	DefaultServiceName string = "gp"
	MasrshalIndent            = json.MarshalIndent
	OpenFile                  = os.OpenFile
	ExecCommand               = exec.Command

	agentPort      int
	caCertPath     string
	caKeyPath      string
	gphome         string
	hubLogDir      string
	hubPort        int
	hostnames      []string
	hostfilePath   string
	serverCertPath string
	serverKeyPath  string
	serviceDir     string // Provide the service file's directory and name separately so users can name different files for different clusters
	serviceName    string
	serviceUser    string
)

func hubCmd() *cobra.Command {
	hubCmd := &cobra.Command{
		Use:     "hub",
		Short:   "Start a gp process in hub mode",
		Long:    "Start a gp process in hub mode",
		Hidden:  true, // Should only be invoked by systemd
		PreRunE: InitializeCommand,
		RunE:    RunHub,
	}

	return hubCmd
}

func RunHub(cmd *cobra.Command, args []string) (err error) {
	h := hub.New(conf, grpc.DialContext)
	err = h.Start()
	if err != nil {
		return fmt.Errorf("Could not start hub: %w", err)
	}
	return nil
}

func installCmd() *cobra.Command {
	installCmd := &cobra.Command{
		Use:    "install",
		Short:  "Install gp as a systemd daemon",
		PreRun: InitializeGplog,
		RunE:   RunInstall,
	}
	installCmd.Flags().IntVar(&agentPort, "agent-port", DefaultAgentPort, `Port on which the agents should listen`)
	installCmd.Flags().StringVar(&gphome, "gphome", os.Getenv("GPHOME"), `Path to GPDB installation`)
	installCmd.Flags().IntVar(&hubPort, "hub-port", DefaultHubPort, `Port on which the hub should listen`)
	installCmd.Flags().StringVar(&hubLogDir, "log-dir", DefaultHubLogDir, `Path to gp hub log directory`)
	installCmd.Flags().StringVar(&serviceName, "service-name", DefaultServiceName, `Name for the generated systemd service file`)
	installCmd.Flags().StringVar(&serviceDir, "service-dir", fmt.Sprintf(DefaultServiceDir, os.Getenv("USER")), `Path to service file directory`)
	installCmd.Flags().StringVar(&serviceUser, "service-user", os.Getenv("USER"), `User for whom to install the service`)
	// TLS credentials are deliberately left blank if not provided, and need to be filled in by the user
	installCmd.Flags().StringVar(&caCertPath, "ca-certificate", "", `Path to SSL/TLS CA certificate`)
	installCmd.Flags().StringVar(&caKeyPath, "ca-key", "", `Path to SSL/TLS CA private key`)
	installCmd.Flags().StringVar(&serverCertPath, "server-certificate", "", `Path to hub SSL/TLS server certificate`)
	installCmd.Flags().StringVar(&serverKeyPath, "server-key", "", `Path to hub SSL/TLS server private key`)
	// Allow passing a hostfile for "real" use cases or a few host names for tests, but not both
	installCmd.Flags().StringArrayVar(&hostnames, "host", []string{}, `Segment hostname`)
	installCmd.Flags().StringVar(&hostfilePath, "hostfile", "", `Path to file containing a list of segment hostnames`)
	installCmd.MarkFlagsMutuallyExclusive("host", "hostfile")
	return installCmd
}
func RunInstall(cmd *cobra.Command, args []string) (err error) {
	if gphome == "" {
		return fmt.Errorf("GPHOME environment variable not set and --gphome flag not provided\n")
	}

	// Regenerate default flag values if a custom GPHOME or username is passed
	if cmd.Flags().Lookup("gphome").Changed && !cmd.Flags().Lookup("config-file").Changed {
		ConfigFilePath = fmt.Sprintf(defaultConfigFilePath, gphome)
	}
	if cmd.Flags().Lookup("service-user").Changed && !cmd.Flags().Lookup("service-dir").Changed {
		serviceDir = fmt.Sprintf(DefaultServiceDir, serviceUser)
	}

	if !cmd.Flags().Lookup("host").Changed && !cmd.Flags().Lookup("hostfile").Changed {
		return errors.New("At least one hostname must be provided using either --host or --hostfile")
	}

	// Convert file/directory paths to absolute path before writing to gp.conf file
	err = resolveAbsolutePaths(cmd)
	if err != nil {
		return err
	}

	hostnames, err := getHostnames(hostnames, hostfilePath)
	if err != nil {
		return fmt.Errorf("Could not get hostname from %s: %w", hostfilePath, err)
	}

	conf = &hub.Config{
		Port:        hubPort,
		AgentPort:   agentPort,
		Hostnames:   hostnames,
		LogDir:      hubLogDir,
		ServiceName: serviceName,
		GpHome:      gphome,
		Credentials: &utils.GpCredentials{
			CACertPath:     caCertPath,
			CAKeyPath:      caKeyPath,
			ServerCertPath: serverCertPath,
			ServerKeyPath:  serverKeyPath,
		},
	}
	err = conf.Write(ConfigFilePath)
	if err != nil {
		return fmt.Errorf("Could not create config file %s: %w", ConfigFilePath, err)
	}

	err = platform.CreateServiceDir(hostnames, serviceDir, gphome)
	if err != nil {
		return err
	}

	err = platform.CreateAndInstallHubServiceFile(gphome, serviceDir, serviceName)
	if err != nil {
		return fmt.Errorf("Could not install hub service file: %w", err)
	}

	err = platform.CreateAndInstallAgentServiceFile(hostnames, gphome, serviceDir, serviceName)
	if err != nil {
		return fmt.Errorf("Could not install agent service file: %w", err)
	}

	err = platform.EnableUserLingering(hostnames, gphome, serviceUser)
	if err != nil {
		return fmt.Errorf("Could not enable user lingering: %w", err)
	}
	return nil
}

func resolveAbsolutePaths(cmd *cobra.Command) error {
	paths := []*string{&caCertPath, &caKeyPath, &serverCertPath, &serverKeyPath, &DefaultHubLogDir, &gphome}
	for _, path := range paths {
		p, err := filepath.Abs(*path)
		if err != nil {
			return fmt.Errorf("Error resolving absolute path for %s: %w", *path, err)
		}
		*path = p
	}
	return nil
}

func getHostnames(hostnames []string, hostfilePath string) ([]string, error) {
	if len(hostnames) > 0 {
		return hostnames, nil
	}

	contents, err := ReadFile(hostfilePath)
	if err != nil {
		return []string{}, fmt.Errorf("Could not read hostfile: %w", err)
	}
	return strings.Fields(string(contents)), nil
}
