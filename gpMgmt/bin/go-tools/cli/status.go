package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"

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

	return statusCmd
}

func statusHubCmd() *cobra.Command {
	statusHubCmd := &cobra.Command{
		Use:    "hub",
		Short:  "Display hub status",
		PreRun: InitializeCommand,
		RunE:   RunStatusHub,
	}

	return statusHubCmd
}

func RunStatusHub(cmd *cobra.Command, args []string) error {
	err := ShowHubStatus(conf)
	if err != nil {
		return fmt.Errorf("Could not retrieve hub status: %w", err)
	}
	return nil
}

func statusAgentsCmd() *cobra.Command {
	statusAgentsCmd := &cobra.Command{
		Use:    "agents",
		Short:  "Display agents status",
		PreRun: InitializeCommand,
		RunE:   RunStatusAgent,
	}

	return statusAgentsCmd
}

func RunStatusAgent(cmd *cobra.Command, args []string) error {
	client, err := connectToHub(conf)
	if err != nil {
		return fmt.Errorf("Could not connect to hub; is the hub running?")
	}

	err = ShowAgentsStatus(client, conf)
	if err != nil {
		return fmt.Errorf("Could not retrieve agents status: %w", err)
	}
	return nil
}

func ShowHubStatus(conf *hub.Config) error {
	service := ""
	if runtime.GOOS == "linux" {
		service = fmt.Sprintf("%s_hub", conf.ServiceName)
	}else if runtime.GOOS == "darwin"{
		service = fmt.Sprintf("%s.hub", conf.ServiceName)
	}
	message, err := platform.GetServiceStatusMessage(service)
	if err != nil {
		return err
	}
	status := platform.ParseServiceStatusMessage(message)
	status.Host, _ = os.Hostname()
	platform.DisplayServiceStatus([]*idl.ServiceStatus{&status})
	return nil
}

func ShowAgentsStatus(client idl.HubClient, conf *hub.Config) error {
	reply, err := client.StatusAgents(context.Background(), &idl.StatusAgentsRequest{})
	if err != nil {
		return err
	}
	platform.DisplayServiceStatus(reply.Statuses)
	return nil
}
