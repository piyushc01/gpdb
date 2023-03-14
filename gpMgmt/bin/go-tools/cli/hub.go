package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/utils"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	defaultHubLogDir  string = "/tmp"
	defaultHubPort    int    = 4242
	defaultAgentPort  int    = 8000
	defaultServiceDir string = "/home/%s/.config/systemd/user"

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
		PreRun: InitializeCommand,
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

	hostnames, err := getHostnames(hostnames, hostfilePath)
	if err != nil {
		return err
	}

	err = CreateConfigFile(caCertPath, caKeyPath, serverCertPath, serverKeyPath, hubPort, agentPort, hostnames, hubLogDir, serviceName)
	if err != nil {
		return err
	}

	err = CreateAndInstallHubServiceFile()
	if err != nil {
		return err
	}

	err = CreateAndInstallAgentServiceFile()
	if err != nil {
		return err
	}

	err = EnableSystemdUserServices()
	if err != nil {
		return err
	}

	return nil
}

func CreateConfigFile(caCertPath, caKeyPath, serverCertPath, serverKeyPath string, hubPort, agentPort int, hostnames []string, hubLogDir, serviceName string) error {
	creds := &utils.Credentials{caCertPath, caKeyPath, serverCertPath, serverKeyPath}
	conf = &hub.Config{hubPort, agentPort, hostnames, hubLogDir, serviceName, creds}
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

func CreateAndInstallHubServiceFile() error {
	hubServiceContents := GenerateServiceFileContents("hub", gphome, serviceName)
	hubServiceFilePath := fmt.Sprintf("%s/%s_hub.service", serviceDir, serviceName)
	err := writeServiceFile(hubServiceFilePath, hubServiceContents)
	if err != nil {
		return err
	}
	gplog.Info("Wrote hub service file to %s on coordinator host", hubServiceFilePath)
	return nil
}

func CreateAndInstallAgentServiceFile() error {
	agentServiceContents := GenerateServiceFileContents("agent", gphome, serviceName)
	localAgentServiceFilePath := fmt.Sprintf("./%s_agent.service", serviceName)
	err := writeServiceFile(localAgentServiceFilePath, agentServiceContents)
	if err != nil {
		return err
	}
	defer os.Remove(localAgentServiceFilePath)

	remoteAgentServiceFilePath := fmt.Sprintf("%s/%s_agent.service", serviceDir, serviceName)
	remoteCmd := make([]string, 0)
	for _, host := range conf.Hostnames {
		remoteCmd = append(remoteCmd, "-h", host)
	}
	remoteCmd = append(remoteCmd, localAgentServiceFilePath, fmt.Sprintf("=:%s", remoteAgentServiceFilePath))
	err = exec.Command("gpsync", remoteCmd...).Run()
	if err != nil {
		return fmt.Errorf("Could not copy agent service files to segment hosts: %w", err)
	}
	gplog.Info("Wrote agent service file to %s on segment hosts", remoteAgentServiceFilePath)
	return nil
}

func EnableSystemdUserServices() error {
	// Allow user services to run on startup and be started/stopped without root access
	err := exec.Command("loginctl", "enable-linger", serviceUser).Run()
	if err != nil {
		return fmt.Errorf("Could not enable user lingering: %w", err)
	}
	// Allow user to view the status of their services
	err = exec.Command("usermod", "-a", "-G", " systemd-journal", serviceUser).Run()
	if err != nil {
		return fmt.Errorf("Could not enable user journal access: %w", err)
	}
	return nil
}

func GenerateServiceFileContents(which string, gphome string, serviceName string) string {
	template := `[Unit]
Description=Greenplum Database management utility %[1]s

[Service]
Type=simple
Environment=GPHOME=%[2]s
ExecStart=%[2]s/bin/gp %[1]s
Restart=on-failure

[Install]
Alias=%[3]s_%[1]s.service
WantedBy=default.target
`

	return fmt.Sprintf(template, which, gphome, serviceName)
}

func writeServiceFile(filename string, contents string) error {
	handle, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	defer handle.Close()
	if err != nil {
		return fmt.Errorf("Could not create systemd service file %s: %w\n", filename, err)
	}
	_, err = handle.WriteString(contents)
	if err != nil {
		return fmt.Errorf("Could not write to systemd service file %s: %w\n", filename, err)
	}
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
