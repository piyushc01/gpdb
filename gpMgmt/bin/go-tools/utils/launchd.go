package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
)

type darwinOS struct{}

func NewDarwinOS() *darwinOS {
	return &darwinOS{}
}

func (d *darwinOS) GenerateServiceFileContents(which string, gphome string, serviceName string) string {
	template := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%[3]s.%[1]s</string>
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

func (d *darwinOS) GetDefaultServiceDir() string {
	return "/Users/%s/Library/LaunchAgents"
}

func (d *darwinOS) CreateAndInstallHubServiceFile(gphome string, serviceDir string, serviceName string) error {
	hubServiceContents := d.GenerateServiceFileContents("hub", gphome, serviceName)
	hubServiceFilePath := fmt.Sprintf("%s/%s_hub.plist", serviceDir, serviceName)
	err := writeServiceFile(hubServiceFilePath, hubServiceContents)
	if err != nil {
		return err
	}
	
	// Unload the file first to remove any previous configuration
	err = exec.Command("launchctl", "unload", fmt.Sprintf("%s/%s_hub.plist", serviceDir, serviceName)).Run()
	if err != nil {
		return fmt.Errorf("Could not load hub service file %s: %w", hubServiceFilePath, err)
	}
	err = exec.Command("launchctl", "load", fmt.Sprintf("%s/%s_hub.plist", serviceDir, serviceName)).Run()
	if err != nil {
		return fmt.Errorf("Could not load hub service file %s: %w", hubServiceFilePath, err)
	}
	gplog.Info("Wrote hub service file to %s on coordinator host", hubServiceFilePath)
	return nil
}

func (d *darwinOS) CreateAndInstallAgentServiceFile(hostnames []string, gphome string, serviceDir string, serviceName string) error {
	agentServiceContents := d.GenerateServiceFileContents("agent", gphome, serviceName)
	localAgentServiceFilePath := fmt.Sprintf("./%s_agent.plist", serviceName)
	err := writeServiceFile(localAgentServiceFilePath, agentServiceContents)
	if err != nil {
		return err
	}
	defer os.Remove(localAgentServiceFilePath)

	remoteAgentServiceFilePath := fmt.Sprintf("%s/%s_agent.plist", serviceDir, serviceName)
	hostList := make([]string, 0)
	for _, host := range hostnames {
		hostList = append(hostList, "-h", host)
	}
	remoteCmd := append(hostList, localAgentServiceFilePath, fmt.Sprintf("=:%s", remoteAgentServiceFilePath))
	err = exec.Command("/bin/bash", "-c", fmt.Sprintf("source %s/greenplum_path.sh; gpsync %s", gphome, strings.Join(remoteCmd, " "))).Run()
	if err != nil {
		return fmt.Errorf("Could not copy agent service files to segment hosts: %w", err)
	}

	// Unload the file first to remove any previous configuration
	remoteCmd = append(hostList, "launchctl", "unload", fmt.Sprintf("%s/%s_agent.plist", serviceDir, serviceName))
	err = exec.Command("/bin/bash", "-c", fmt.Sprintf("source %s/greenplum_path.sh; gpssh %s", gphome, strings.Join(remoteCmd, " "))).Run()
	if err != nil {
		return fmt.Errorf("Could not unload agent service file %s: %w", remoteAgentServiceFilePath, err)
	}
	remoteCmd = append(hostList, "launchctl", "load", fmt.Sprintf("%s/%s_agent.plist", serviceDir, serviceName))
	err = exec.Command("/bin/bash", "-c", fmt.Sprintf("source %s/greenplum_path.sh; gpssh %s", gphome, strings.Join(remoteCmd, " "))).Run()
	if err != nil {
		return fmt.Errorf("Could not load agent service file %s: %w", remoteAgentServiceFilePath, err)
	}

	gplog.Info("Wrote agent service file to %s on segment hosts", remoteAgentServiceFilePath)
	return nil
}

func (d * darwinOS) GetStartHubCmd(serviceName string) *exec.Cmd {
	return exec.Command("launchctl", "start", fmt.Sprintf("%s.hub", serviceName))
}

func (d * darwinOS) GetStartAgentCmd(serviceName string) []string {
	return []string{"launchctl", "start", fmt.Sprintf("%s.agent", serviceName)}
}

func (d * darwinOS) GetServiceStatusMessage(serviceName string) (string, error) {
	output, err := exec.Command("launchctl", "list", serviceName).Output()
	if err != nil {
		if err.Error() != "exit status 3" { // 3 = service is stopped
			return "", err
		}
	}
	return string(output), nil
}

func (d * darwinOS) ParseServiceStatusMessage(message string) idl.ServiceStatus {
	lines := strings.Split(message, "\n")
	status := "Unknown"
	uptime := "Unknown"
	pid := 0

	pidLineRegex := regexp.MustCompile(`"PID"\s*=\s*(\d+);`)

	for _, line := range lines {
		 if pidLineRegex.MatchString(line) {
			results := pidLineRegex.FindStringSubmatch(line)
			pid, _ = strconv.Atoi(results[1])
		}
	}

	if pid > 0{
		status = "Running"
	}
	return idl.ServiceStatus{Status: status, Uptime: uptime, Pid: uint32(pid)}
}

func (d * darwinOS) DisplayServiceStatus(statuses []*idl.ServiceStatus) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	fmt.Fprintln(w, "HOST\tSTATUS\tPID\tUPTIME")
	for _, s := range statuses {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", s.Host, s.Status, s.Pid, s.Uptime)
	}
	w.Flush()
}
