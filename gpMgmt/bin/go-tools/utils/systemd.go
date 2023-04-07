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

type linuxOS struct{}

func NewLinuxOS() *linuxOS {
	return &linuxOS{}
}

func (l *linuxOS) GenerateServiceFileContents(which string, gphome string, serviceName string) string {
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

func (l *linuxOS) GetDefaultServiceDir() string {
	return "/home/%s/.config/systemd/user"
}

func (l *linuxOS) CreateAndInstallHubServiceFile(gphome string, serviceDir string, serviceName string) error {
	hubServiceContents := l.GenerateServiceFileContents("hub", gphome, serviceName)
	hubServiceFilePath := fmt.Sprintf("%s/%s_hub.service", serviceDir, serviceName)
	err := writeServiceFile(hubServiceFilePath, hubServiceContents)
	if err != nil {
		return err
	}
	gplog.Info("Wrote hub service file to %s on coordinator host", hubServiceFilePath)
	return nil
}


func (l *linuxOS) CreateAndInstallAgentServiceFile(hostnames []string, gphome string, serviceDir string, serviceName string) error {
	agentServiceContents := l.GenerateServiceFileContents("agent", gphome, serviceName)
	localAgentServiceFilePath := fmt.Sprintf("./%s_agent.service", serviceName)
	err := writeServiceFile(localAgentServiceFilePath, agentServiceContents)
	if err != nil {
		return err
	}
	defer os.Remove(localAgentServiceFilePath)

	remoteAgentServiceFilePath := fmt.Sprintf("%s/%s_agent.service", serviceDir, serviceName)
	remoteCmd := make([]string, 0)
	for _, host := range hostnames {
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

func (l *linuxOS) EnableSystemdUserServices(serviceUser string) error {
	if runtime.GOOS != "linux" {
		return nil
	}
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

func (l * linuxOS) GetStartHubCmd(serviceName string) *exec.Cmd {
	return exec.Command("service", fmt.Sprintf("%s_hub", serviceName), "start")
}

func (l * linuxOS) GetStartAgentCmd(serviceName string) []string {
	return []string{"service", fmt.Sprintf("%s_agent", serviceName), "start"}
}

func (l * linuxOS) GetServiceStatusMessage(serviceName string) (string, error) {
	output, err := exec.Command("service", serviceName, "status").Output()
	if err != nil {
		if err.Error() != "exit status 3" { // 3 = service is stopped
			return "", err
		}
	}
	return string(output), nil
}

func (l * linuxOS) ParseServiceStatusMessage(message string) idl.ServiceStatus {
	lines := strings.Split(message, "\n")
	status := "Unknown"
	uptime := "unknown"
	pid := 0
	statusLineRegex := regexp.MustCompile(`Active: (.+) (since .+)`)
	pidLineRegex := regexp.MustCompile(`Main PID: (\d+) `)

	for _, line := range lines {
		if statusLineRegex.MatchString(line) {
			results := statusLineRegex.FindStringSubmatch(line)
			status = results[1]
			uptime = results[2]
		} else if pidLineRegex.MatchString(line) {
			results := pidLineRegex.FindStringSubmatch(line)
			pid, _ = strconv.Atoi(results[1])
		}
	}
	return idl.ServiceStatus{Status: status, Uptime: uptime, Pid: uint32(pid)}
}

func (l * linuxOS) DisplayServiceStatus(statuses []*idl.ServiceStatus) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	fmt.Fprintln(w, "HOST\tSTATUS\tPID\tUPTIME")
	for _, s := range statuses {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", s.Host, s.Status, s.Pid, s.Uptime)
	}
	w.Flush()
}
