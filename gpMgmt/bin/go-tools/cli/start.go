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

var (
	StartHubService        = StartHubServiceFn
	RunStartHub            = RunStartHubFn
	RunStartAgent          = RunStartAgentFn
	StartAgentsAll         = StartAgentsAllFn
	RunStartService        = RunStartServiceFn
	WaitAndRetryHubConnect = WaitAndRetryHubConnectFn
)

func startHubCmd() *cobra.Command {
	startHubCmd := &cobra.Command{
		Use:     "hub",
		Short:   "Start the hub",
		PreRunE: InitializeCommand,
		RunE:    RunStartHub,
	}

	return startHubCmd
}

func StartHubServiceFn(serviceName string) error {
	err := Platform.GetStartHubCommand(serviceName).Run()
	if err != nil {
		return fmt.Errorf("Failed to start hub service: %s Error: %w", serviceName, err)
	}

	return nil
}
func RunStartHubFn(cmd *cobra.Command, args []string) error {
	err := StartHubService(Conf.ServiceName)
	if err != nil {
		return fmt.Errorf("hub service %s. Error: %w", Conf.ServiceName, err)
	}
	gplog.Info("Hub %s started successfully", Conf.ServiceName)
	if Verbose {
		_, err = ShowHubStatus(Conf, true)
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

func RunStartAgentFn(cmd *cobra.Command, args []string) error {
	_, err := StartAgentsAll(Conf)
	if err != nil {
		return err
	}
	gplog.Info("Agents started successfully")
	if Verbose {
		err = ShowAgentsStatus(Conf, true)
		if err != nil {
			return fmt.Errorf("Could not retrieve agent status: %w", err)
		}
	}

	return nil
}

func StartAgentsAllFn(hubConfig *hub.Config) (idl.HubClient, error) {
	client, err := ConnectToHub(hubConfig)
	if err != nil {
		return client, err
	}

	_, err = client.StartAgents(context.Background(), &idl.StartAgentsRequest{})
	if err != nil {
		return client, fmt.Errorf("Could not start agents: %w", err)
	}

	return client, nil
}

func RunStartServiceFn(cmd *cobra.Command, args []string) error {
	//Starts  Hub service followed by Agent service
	err := StartHubService(Conf.ServiceName)
	if err != nil {
		return err
	}
	err = WaitAndRetryHubConnect()
	if err != nil {
		return fmt.Errorf("Error while connecting hb service:%w", err)
	}
	gplog.Info("Hub %s started successfully", Conf.ServiceName)

	// Start agents service
	_, err = StartAgentsAll(Conf)
	if err != nil {
		return fmt.Errorf("Failed to start agents. Error: %w", err)
	}
	gplog.Info("Agents %s started successfully", Conf.ServiceName)
	if Verbose {
		err = PrintServicesStatus()
		if err != nil {
			return err
		}
	}

	return nil
}

func WaitAndRetryHubConnectFn() error {
	count := 10
	success := false
	var err error
	for count > 0 {
		_, err = ConnectToHub(Conf)
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
