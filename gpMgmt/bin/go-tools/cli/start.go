package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/hub"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/spf13/cobra"
)

func startCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start hub, agents services",
	}

	startCmd.AddCommand(startHubCmd())
	startCmd.AddCommand(startAgentsCmd())
	startCmd.AddCommand(startServiceCmd())

	return startCmd
}

func startHubCmd() *cobra.Command {
	startHubCmd := &cobra.Command{
		Use:     "hub",
		Short:   "Start the hub",
		PreRunE: InitializeCommand,
		RunE:    RunStartHub,
	}

	return startHubCmd
}

var startHubService = func(serviceName string) error {
	err := platform.GetStartHubCommand(serviceName).Run()
	if err != nil {
		return fmt.Errorf("Could not start hub: %w", err)
	}

	return nil
}
var RunStartHub = func(cmd *cobra.Command, args []string) error {
	err := startHubService(conf.ServiceName)
	if err != nil {
		return fmt.Errorf("Failed to start hub service %s. Error: %w", conf.ServiceName, err)
	}
	gplog.Info("Hub %s started successfully", conf.ServiceName)
	if verbose {
		_, err = ShowHubStatus(conf, true)
		if err != nil {
			return fmt.Errorf("Could not retrieve hub status: %w", err)
		}
	}

	return nil
}

func startAgentsCmd() *cobra.Command {
	startAgentsCmd := &cobra.Command{
		Use:     "agents",
		Short:   "Start the agents",
		PreRunE: InitializeCommand,
		RunE:    RunStartAgent,
	}

	return startAgentsCmd
}

func startServiceCmd() *cobra.Command {
	startServicesCmd := &cobra.Command{
		Use:     "services",
		Short:   "Starts hub, agent services",
		PreRunE: InitializeCommand,
		RunE:    RunStartService,
	}

	return startServicesCmd
}

var RunStartAgent = func(cmd *cobra.Command, args []string) error {
	_, err := startAgentsAll(conf)
	if err != nil {
		return fmt.Errorf("Could not start agents: %w", err)
	}
	gplog.Info("Agents started successfully")
	if verbose {
		err = ShowAgentsStatus(conf, true)
		if err != nil {
			return fmt.Errorf("Could not retrieve agent status: %w", err)
		}
	}

	return nil
}

var startAgentsAll = func(hubConfig *hub.Config) (idl.HubClient, error) {
	client, err := connectToHub(hubConfig)
	if err != nil {
		return client, fmt.Errorf("Could not connect to hub on port%d Error: %w", hubConfig.Port, err)
	}

	_, err = client.StartAgents(context.Background(), &idl.StartAgentsRequest{})
	if err != nil {
		return client, fmt.Errorf("Could not start agents: %w", err)
	}

	return client, nil
}

var RunStartService = func(cmd *cobra.Command, args []string) error {
	//Starts  Hub service followed by Agent service
	err := startHubService(conf.ServiceName)
	if err != nil {
		return fmt.Errorf("Failed to start hub service %s. Error: %w", conf.ServiceName, err)
	}
	err = WaitAndRetryHubConnect()
	if err != nil {
		return fmt.Errorf("Error while connecting hb service:%w", err)
	}
	gplog.Info("Hub %s started successfully", conf.ServiceName)

	// Start agents service
	_, err = startAgentsAll(conf)
	if err != nil {
		return fmt.Errorf("Failed to start agents. Error: %w", err)
	}
	gplog.Info("Agents %s started successfully", conf.ServiceName)
	if verbose {
		err = PrintServicesStatus()
		if err != nil {
			return fmt.Errorf("Could not retrieve hub status: %w", err)
		}
	}

	return nil
}

var WaitAndRetryHubConnect = func() error {
	count := 10
	success := false

	var err error
	for count > 0 {
		_, err = connectToHub(conf)
		if err == nil {
			success = true
			break
		}
		// Wait for half second before next retry
		time.Sleep(time.Second / 2)
		count--
	}
	if !success {
		return fmt.Errorf("Hub service started but failed to connect. Bailing out. Error:%w", err)
	}

	return nil
}
