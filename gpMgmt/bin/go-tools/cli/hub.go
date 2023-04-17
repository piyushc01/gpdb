package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	platform = utils.GetOS()
	defaultHubLogDir  string = "/tmp"
	defaultHubPort    int    = 4242
	defaultAgentPort  int    = 8000
	defaultServiceDir string = platform.GetDefaultServiceDir()

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
		Use:    "hub",
		Short:  "Start a gp process in hub mode",
		Long:   "Start a gp process in hub mode",
		Hidden: true, // Should only be invoked by systemd
		PreRun: InitializeCommand,
		RunE:   RunHub,
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

	installCmd.Flags().IntVar(&agentPort, "agent-port", defaultAgentPort, `Port on which the agents should listen`)
	installCmd.Flags().StringVar(&gphome, "gphome", os.Getenv("GPHOME"), `Path to GPDB installation`)
	installCmd.Flags().IntVar(&hubPort, "hub-port", defaultHubPort, `Port on which the hub should listen`)
	installCmd.Flags().StringVar(&hubLogDir, "log-dir", defaultHubLogDir, `Path to gp hub log directory`)
	installCmd.Flags().StringVar(&serviceName, "service-name", "gp", `Name for the generated systemd service file`)
	installCmd.Flags().StringVar(&serviceDir, "service-dir", fmt.Sprintf(defaultServiceDir, os.Getenv("USER")), `Path to service file directory`)
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
		configFilePath = fmt.Sprintf(defaultConfigFilePath, gphome)
	}
	if cmd.Flags().Lookup("service-user").Changed && !cmd.Flags().Lookup("service-dir").Changed {
		serviceDir = fmt.Sprintf(defaultServiceDir, os.Getenv("USER"))
	}

	if !cmd.Flags().Lookup("host").Changed && !cmd.Flags().Lookup("hostfile").Changed {
		return errors.New("At least one hostname must be provided using either --host or --hostfile")
	}

	// Convert file/directory paths to absolute path before writing to gp.conf file
	err = updateAbsolutePath(cmd)
	if err != nil {
		gplog.Error("Error while converting file/directory paths to absolute path %w", err)
		return err
	}

	hostnames, err := getHostnames(hostnames, hostfilePath)
	if err != nil {
		return err
	}

	err = CreateConfigFile(caCertPath, caKeyPath, serverCertPath, serverKeyPath, hubPort, agentPort, hostnames, hubLogDir, serviceName, serviceDir)
	if err != nil {
		return err
	}
	err = os.MkdirAll(serviceDir, os.ModePerm)
	if err != nil {
		gplog.Error("Error creating directory for service:%s. Error:%w", serviceDir, err)
		return err
	}

	err = platform.CreateAndInstallHubServiceFile(gphome, serviceDir, serviceName)
	if err != nil {
		return err
	}

	err = platform.CreateAndInstallAgentServiceFile(hostnames, gphome, serviceDir, serviceName)
	if err != nil {
		return err
	}

	err = utils.NewLinuxOS().EnableSystemdUserServices(serviceUser)
	if err != nil {
		return err
	}

	return nil
}

func updateAbsolutePath(cmd *cobra.Command) error {
	// Convert certificate file paths to absolute path
	var err error = nil
	if cmd.Flags().Lookup("ca-certificate").Changed && !filepath.IsAbs(caCertPath) {

		caCertPath, err = filepath.Abs(caCertPath)
		if err != nil {
			gplog.Error("Error converting to absolute path for file:%s. Error: %w", caCertPath, err)
			return err
		}
		gplog.Info("Cert Path New Path:%s", caCertPath)
	}
	if cmd.Flags().Lookup("ca-key").Changed && !filepath.IsAbs(caKeyPath) {
		caKeyPath, err = filepath.Abs(caKeyPath)
		if err != nil {
			gplog.Error("Error converting to absolute path for file:%s. Error: %w", caKeyPath, err)
			return err
		}
	}
	if cmd.Flags().Lookup("server-certificate").Changed && !filepath.IsAbs(serverCertPath) {
		serverCertPath, err = filepath.Abs(serverCertPath)
		if err != nil {
			gplog.Error("Error converting to absolute path for file:%s. Error: %w", serverCertPath, err)
			return err
		}
	}
	if cmd.Flags().Lookup("server-key").Changed && !filepath.IsAbs(serverKeyPath) {
		serverKeyPath, err = filepath.Abs(serverKeyPath)
		if err != nil {
			gplog.Error("Error converting to absolute path for file:%s. Error: %w", serverKeyPath, err)
			return err
		}
	}
	gplog.Debug("Certificate Paths updated:\n CA-Cert:%s\nCA-Key:%s\nServer-Cert:%s\nServer-Key:%s", caCertPath, caKeyPath, serverCertPath, serverKeyPath)

	// Update hub data directory if not absolulte
	if cmd.Flags().Lookup("log-dir").Changed && !filepath.IsAbs(defaultHubLogDir) {
		defaultHubLogDir, err = filepath.Abs(defaultHubLogDir)
		if err != nil {
			gplog.Error("Error converting to absolute path for file:%s. Error: %w", defaultHubLogDir, err)
			return err
		}
	}

	// Update GPHOME directory path if not absolute
	if cmd.Flags().Lookup("gphome").Changed && !filepath.IsAbs(gphome) {
		gphome, err = filepath.Abs(gphome)
		if err != nil {
			gplog.Error("Error converting to absolute path for file:%s. Error: %w", defaultHubLogDir, err)
			return err
		}
	}
	return nil
}

func CreateConfigFile(caCertPath, caKeyPath, serverCertPath, serverKeyPath string, hubPort, agentPort int, hostnames []string, hubLogDir, serviceName string, serviceDir string) error {
	creds := &utils.Credentials{caCertPath, caKeyPath, serverCertPath, serverKeyPath}
	conf = &hub.Config{hubPort, agentPort, hostnames, hubLogDir, serviceName, gphome, creds}
	configContents, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		return fmt.Errorf("Could not generate configuration file: %w", err)
	}
	configHandle, err := os.OpenFile(configFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer configHandle.Close()
	if err != nil {
		return fmt.Errorf("Could not create configuration file %s: %w\n", configFilePath, err)
	}
	_, err = configHandle.Write(configContents)
	if err != nil {
		return fmt.Errorf("Could not write to configuration file %s: %w\n", configFilePath, err)
	}
	gplog.Info("Wrote configuration file to %s", configFilePath)
	return nil
}

func getHostnames(hostnames []string, hostfilePath string) ([]string, error) {
	if len(hostnames) > 0 {
		return hostnames, nil
	}

	contents, err := os.ReadFile(hostfilePath)
	if err != nil {
		return []string{}, fmt.Errorf("Could not read hostfile: %w", err)
	}
	return strings.Split(string(contents), "\n"), nil
}
