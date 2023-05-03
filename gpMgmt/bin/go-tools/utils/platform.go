package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
)

var (
	platform Platform
)

type GpPlatform struct {
	OS         string
	ServiceCmd string // Binary for managing services
	UserArg    string // systemd always needs a "--user" flag passed, launchctl does not
	ServiceExt string // Extension for service files
	StatusArg  string // Argument passed to ServiceCmd to get status of a service
}

func NewPlatform(os string) Platform {
	switch os {
	case "darwin":
		return GpPlatform{
			OS:         "darwin",
			ServiceCmd: "launchctl",
			UserArg:    "",
			ServiceExt: "plist",
			StatusArg:  "list",
		}
	case "linux":
		return GpPlatform{
			OS:         "linux",
			ServiceCmd: "systemctl",
			UserArg:    "--user",
			ServiceExt: "service",
			StatusArg:  "status",
		}
	default:
		panic("Unsupported OS")
	}
}

type Platform interface {
	CreateServiceDir(hostnames []string, serviceDir string, gphome string) error
	GenerateServiceFileContents(which string, gphome string, serviceName string) string
	GetDefaultServiceDir() string
	CreateAndInstallHubServiceFile(gphome string, serviceDir string, serviceName string) error
	CreateAndInstallAgentServiceFile(hostnames []string, gphome string, serviceDir string, serviceName string) error
	GetStartHubCommand(serviceName string) *exec.Cmd
	GetStartAgentCommandString(serviceName string) []string
	GetServiceStatusMessage(serviceName string) (string, error)
	ParseServiceStatusMessage(message string) idl.ServiceStatus
	DisplayServiceStatus(statuses []*idl.ServiceStatus)
	EnableUserLingering(hostnames []string, gphome string, serviceUser string) error
}

func GetPlatform() Platform {
	if platform == nil {
		platform = NewPlatform(runtime.GOOS)
	}
	return platform
}

func (p GpPlatform) CreateServiceDir(hostnames []string, serviceDir string, gphome string) error {
	hostList := make([]string, 0)
	for _, host := range hostnames {
		hostList = append(hostList, "-h", host)
	}

	// Create service directory if it does not exist
	args := append(hostList, "mkdir", "-p", serviceDir)
	err := exec.Command(fmt.Sprintf("%s/bin/gpssh", gphome), args...).Run()
	if err != nil {
		return fmt.Errorf("Could not create service directory %s on hosts: %w", serviceDir, err)
	}

	gplog.Info("Created service file directory %s on all hosts", serviceDir)
	return nil
}

func writeServiceFile(filename string, contents string) error {
	handle, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("Could not create service file %s: %w\n", filename, err)
	}
	defer handle.Close()

	_, err = handle.WriteString(contents)
	if err != nil {
		return fmt.Errorf("Could not write to service file %s: %w\n", filename, err)
	}
	return nil
}

func (p GpPlatform) GenerateServiceFileContents(which string, gphome string, serviceName string) string {
	if p.OS == "darwin" {
		return GenerateDarwinServiceFileContents(which, gphome, serviceName)
	}
	return GenerateLinuxServiceFileContents(which, gphome, serviceName)
}

func GenerateDarwinServiceFileContents(which string, gphome string, serviceName string) string {
	template := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%[3]s_%[1]s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%[2]s/bin/gp</string>
        <string>%[1]s</string>
    </array>
    <key>StandardOutPath</key>
    <string>/tmp/grpc_%[1]s.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/grpc_%[1]s.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
        <key>GPHOME</key>
        <string>%[2]s</string>
    </dict>
</dict>
</plist>
`
	return fmt.Sprintf(template, which, gphome, serviceName)
}

func GenerateLinuxServiceFileContents(which string, gphome string, serviceName string) string {
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

func (p GpPlatform) GetDefaultServiceDir() string {
	if p.OS == "darwin" {
		return "/Users/%s/Library/LaunchAgents"
	}
	return "/home/%s/.config/systemd/user"
}

func (p GpPlatform) CreateAndInstallHubServiceFile(gphome string, serviceDir string, serviceName string) error {
	hubServiceContents := p.GenerateServiceFileContents("hub", gphome, serviceName)
	hubServiceFilePath := fmt.Sprintf("%s/%s_hub.%s", serviceDir, serviceName, p.ServiceExt)
	err := writeServiceFile(hubServiceFilePath, hubServiceContents)
	if err != nil {
		return err
	}

	err = p.reloadHubService(hubServiceFilePath)
	if err != nil {
		return err
	}

	gplog.Info("Wrote hub service file to %s on coordinator host", hubServiceFilePath)
	return nil
}

func (p GpPlatform) reloadHubService(servicePath string) error {
	if p.OS == "darwin" {

		// launchctl does not have a single reload command. Hence unload and load the file to update the configuration.
		err := exec.Command(p.ServiceCmd, "unload", servicePath).Run()
		if err != nil {
			return fmt.Errorf("Could not unload hub service file %s: %w", servicePath, err)
		}
		err = exec.Command(p.ServiceCmd, "load", servicePath).Run()
		if err != nil {
			return fmt.Errorf("Could not load hub service file %s: %w", servicePath, err)
		}
		return nil
	}

	err := exec.Command(p.ServiceCmd, p.UserArg, "daemon-reload").Run()
	if err != nil {
		return fmt.Errorf("Could not reload hub service file %s: %w", servicePath, err)
	}
	return nil
}

func (p GpPlatform) reloadAgentService(gphome string, hostList []string, servicePath string) error {
	args := append(hostList, p.ServiceCmd)

	if p.OS == "darwin" { // launchctl reloads a specific service, not all of them

		// launchctl does not have a single reload command. Hence unload and load the file to update the configuration.
		err := exec.Command(fmt.Sprintf("%s/bin/gpssh", gphome), append(args, "unload", servicePath)...).Run()
		if err != nil {
			return fmt.Errorf("Could not unload agent service file %s on segment hosts: %w", servicePath, err)
		}
		err = exec.Command(fmt.Sprintf("%s/bin/gpssh", gphome), append(args, "load", servicePath)...).Run()
		if err != nil {
			return fmt.Errorf("Could not load agent service file %s on segment hosts: %w", servicePath, err)
		}
		return nil
	}

	err := exec.Command(fmt.Sprintf("%s/bin/gpssh", gphome), append(args, p.UserArg, "daemon-reload")...).Run()
	if err != nil {
		return fmt.Errorf("Could not reload agent service %s on segment hosts: %w", servicePath, err)
	}
	return nil
}

func (p GpPlatform) CreateAndInstallAgentServiceFile(hostnames []string, gphome string, serviceDir string, serviceName string) error {
	agentServiceContents := p.GenerateServiceFileContents("agent", gphome, serviceName)
	localAgentServiceFilePath := fmt.Sprintf("./%s_agent.%s", serviceName, p.ServiceExt)
	err := writeServiceFile(localAgentServiceFilePath, agentServiceContents)
	if err != nil {
		return err
	}
	defer os.Remove(localAgentServiceFilePath)

	remoteAgentServiceFilePath := fmt.Sprintf("%s/%s_agent.%s", serviceDir, serviceName, p.ServiceExt)
	hostList := make([]string, 0)
	for _, host := range hostnames {
		hostList = append(hostList, "-h", host)
	}

	// Copy the file to segment host service directories
	args := append(hostList, localAgentServiceFilePath, fmt.Sprintf("=:%s", remoteAgentServiceFilePath))
	err = exec.Command(fmt.Sprintf("%s/bin/gpsync", gphome), args...).Run()
	if err != nil {
		return fmt.Errorf("Could not copy agent service files to segment hosts: %w", err)
	}

	err = p.reloadAgentService(gphome, hostList, remoteAgentServiceFilePath)
	if err != nil {
		return err
	}

	gplog.Info("Wrote agent service file to %s on segment hosts", remoteAgentServiceFilePath)
	return nil
}

func (p GpPlatform) GetStartHubCommand(serviceName string) *exec.Cmd {
	args := []string{p.UserArg, "start", fmt.Sprintf("%s_hub", serviceName)}

	if p.OS == "darwin" { // empty strings are also treated as arguments
		args = args[1:]
	}

	return exec.Command(p.ServiceCmd, args...)
}

func (p GpPlatform) GetStartAgentCommandString(serviceName string) []string {
	return []string{p.ServiceCmd, p.UserArg, "start", fmt.Sprintf("%s_agent", serviceName)}
}

func (p GpPlatform) GetServiceStatusMessage(serviceName string) (string, error) {
	args := []string{p.UserArg, p.StatusArg, serviceName}

	if p.OS == "darwin" { // empty strings are also treated as arguments
		args = args[1:]
	}
	output, err := exec.Command(p.ServiceCmd, args...).Output()
	if err != nil {
		if err.Error() != "exit status 3" { // 3 = service is stopped
			return "", err
		}
	}
	return string(output), nil
}

func (p GpPlatform) ParseServiceStatusMessage(message string) idl.ServiceStatus {
	lines := strings.Split(message, "\n")
	status := "Unknown"
	uptime := "Unknown"
	pid := 0

	var statusLineRegex, pidLineRegex *regexp.Regexp
	if p.OS == "darwin" {
		// launchctl doesn't provide status and uptime information, so we
		// leave statusLineRegex set to nil and check for that below.
		pidLineRegex = regexp.MustCompile(`"PID"\s*=\s*(\d+);`)
	} else {
		statusLineRegex = regexp.MustCompile(`Active: (.+) (since .+)`)
		pidLineRegex = regexp.MustCompile(`Main PID: (\d+) `)
	}

	for _, line := range lines {
		if statusLineRegex != nil && statusLineRegex.MatchString(line) {
			results := statusLineRegex.FindStringSubmatch(line)
			status = results[1]
			uptime = results[2]
		} else if pidLineRegex.MatchString(line) {
			results := pidLineRegex.FindStringSubmatch(line)
			pid, _ = strconv.Atoi(results[1])
		}
	}

	if status == "Unknown" && pid > 0 {
		status = "Running"
	}
	return idl.ServiceStatus{Status: status, Uptime: uptime, Pid: uint32(pid)}
}

func (p GpPlatform) DisplayServiceStatus(statuses []*idl.ServiceStatus) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	fmt.Fprintln(w, "HOST\tSTATUS\tPID\tUPTIME")
	for _, s := range statuses {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", s.Host, s.Status, s.Pid, s.Uptime)
	}
	w.Flush()
}

// Allow systemd services to run on startup and be started/stopped without root access
// This is a no-op on Mac, as launchctl lacks the concept of user lingering
func (p GpPlatform) EnableUserLingering(hostnames []string, gphome string, serviceUser string) error {
	if p.OS != "linux" {
		return nil
	}

	hostList := make([]string, 0)
	for _, host := range hostnames {
		hostList = append(hostList, "-h", host)
	}
	remoteCmd := append(hostList, "loginctl enable-linger ", serviceUser)
	err := exec.Command(fmt.Sprintf("%s/bin/gpssh", gphome), remoteCmd...).Run()
	if err != nil {
		return fmt.Errorf("Could not enable user lingering: %w", err)
	}
	return nil
}
