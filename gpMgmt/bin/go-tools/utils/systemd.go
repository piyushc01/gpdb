package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/greenplum-db/gpdb/gp/idl"
)

func GetServiceStatusMessage(serviceName string) (string, error) {
	output, err := exec.Command("service", serviceName, "status").Output()
	if err != nil {
		if err.Error() != "exit status 3" { // 3 = service is stopped
			return "", err
		}
	}
	return string(output), nil
}

func ParseServiceStatusMessage(message string) idl.ServiceStatus {
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

func DisplayServiceStatus(statuses []*idl.ServiceStatus) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', 0)
	fmt.Fprintln(w, "HOST\tSTATUS\tPID\tUPTIME")
	for _, s := range statuses {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", s.Host, s.Status, s.Pid, s.Uptime)
	}
	w.Flush()
}
