package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/spf13/cobra"
)

func statusCmd() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Display status",
	}

	statusCmd.AddCommand(statusHubCmd())
	statusCmd.AddCommand(statusAgentsCmd())
	statusCmd.AddCommand(statusServicesCmd())

	return statusCmd
}

func statusHubCmd() *cobra.Command {
	statusHubCmd := &cobra.Command{
		Use:     "hub",
		Short:   "Display hub status",
		PreRunE: InitializeCommand,
		RunE:    RunStatusHub,
	}

	return statusHubCmd
}

func RunStatusHub(cmd *cobra.Command, args []string) error {
	_, err := ShowHubStatus(conf, false)
	if err != nil {
		return fmt.Errorf("Could not retrieve hub status: %w", err)
	}
	return nil
}

func statusAgentsCmd() *cobra.Command {
	statusAgentsCmd := &cobra.Command{
		Use:     "agents",
		Short:   "Display agents status",
		PreRunE: InitializeCommand,
		RunE:    RunStatusAgent,
	}

	return statusAgentsCmd
}

func statusServicesCmd() *cobra.Command {
	statusServicesCmd := &cobra.Command{
		Use:     "services",
		Short:   "Display Hub and Agent services status",
		PreRunE: InitializeCommand,
		RunE:    RunServiceStatus,
	}
	return statusServicesCmd
}

func RunStatusAgent(cmd *cobra.Command, args []string) error {
	err := ShowAgentsStatus(conf, false)
	if err != nil {
		return fmt.Errorf("Could not retrieve agents status: %w", err)
	}
	return nil
}

var ShowHubStatus = func(conf *hub.Config, skipHeader bool) (bool, error) {
	message, err := platform.GetServiceStatusMessage(fmt.Sprintf("%s_hub", conf.ServiceName))
	if err != nil {
		return false, err
	}
	status := platform.ParseServiceStatusMessage(message)
	status.Host, _ = os.Hostname()
	platform.DisplayServiceStatus("Hub", []*idl.ServiceStatus{&status}, skipHeader)
	if status.Status == "Unknown" {
		return false, nil
	}
	return true, nil
}

var ShowAgentsStatus = func(conf *hub.Config, skipHeader bool) error {
	client, err := connectToHub(conf)
	if err != nil {
		return fmt.Errorf("Could not connect to hub; is the hub running?")
	}

	reply, err := client.StatusAgents(context.Background(), &idl.StatusAgentsRequest{})
	if err != nil {
		return err
	}
	platform.DisplayServiceStatus("Agent", reply.Statuses, skipHeader)
	return nil
}

func RunServiceStatus(cmd *cobra.Command, args []string) error {
	err := PrintServicesStatus()
	if err != nil {
		return fmt.Errorf("Error while getting the services status:%w", err)
	}
	return nil
}

var PrintServicesStatus = func() error {
	// TODO: Check if Hub is down, do not check for Agents
	hubRunning, err := ShowHubStatus(conf, false)
	if err != nil {
		return fmt.Errorf("Error while showing the Hub status:%w", err)
	}
	if !hubRunning {
		fmt.Println("Hub service not running, not able to fetch agent status.")
		return nil
	}
	err = ShowAgentsStatus(conf, true)
	if err != nil {
		return fmt.Errorf("Error while showing the Agent status:%w", err)
	}
	return nil
}
