package cli

import (
	"context"
	"fmt"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start processes",
	}

	startCmd.AddCommand(startHubCmd())
	startCmd.AddCommand(startAgentsCmd())

	return startCmd
}

func startHubCmd() *cobra.Command {
	startHubCmd := &cobra.Command{
		Use:    "hub",
		Short:  "Start the hub",
		PreRun: InitializeCommand,
		RunE:   RunStartHub,
	}

	return startHubCmd
}

func RunStartHub(cmd *cobra.Command, args []string) error {
	err := platform.GetStartHubCommand(conf.ServiceName).Run()
	if err != nil {
		return fmt.Errorf("Could not start hub: %w", err)
	}
	gplog.Info("Hub started successfully")
	if verbose {
		ShowHubStatus(conf)
	}
	return nil
}

func startAgentsCmd() *cobra.Command {
	startAgentsCmd := &cobra.Command{
		Use:    "agents",
		Short:  "Start the agents",
		PreRun: InitializeCommand,
		RunE:   RunStartAgent,
	}

	return startAgentsCmd
}

func RunStartAgent(cmd *cobra.Command, args []string) error {
	client, err := connectToHub(conf)
	if err != nil {
		return fmt.Errorf("Could not connect to hub: %w", err)
	}

	_, err = client.StartAgents(context.Background(), &idl.StartAgentsRequest{})
	if err != nil {
		return fmt.Errorf("Could not start agents: %w", err)
	}
	gplog.Info("Agents started successfully")
	if verbose {
		ShowAgentsStatus(client, conf)
	}
	return nil
}
