package cli

import (
	"context"
	"fmt"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/idl"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

func stopCmd() *cobra.Command {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop processes",
	}

	stopCmd.AddCommand(stopHubCmd())
	stopCmd.AddCommand(stopAgentsCmd())
	stopCmd.AddCommand(StopServicesCmd())

	return stopCmd
}

func stopHubCmd() *cobra.Command {
	stopHubCmd := &cobra.Command{
		Use:     "hub",
		Short:   "Stop hub",
		PreRunE: InitializeCommand,
		RunE:    RunStopHub,
	}

	return stopHubCmd
}

func RunStopHub(cmd *cobra.Command, args []string) error {
	err := StopHubService()
	if err != nil {
		return fmt.Errorf("Error stopping hub service:%w", err)
	}
	gplog.Info("Hub stopped successfully")
	if verbose {
		_ = ShowHubStatus(conf, false)
	}
	return nil
}

var StopHubService = func() error {
	client, err := connectToHub(conf)
	if err != nil {
		return fmt.Errorf("Could not connect to hub; is the hub running?")
	}
	_, err = client.Stop(context.Background(), &idl.StopHubRequest{})
	// Ignore a "hub already stopped" error
	if err != nil {
		errCode := grpcStatus.Code(err)
		errMsg := grpcStatus.Convert(err).Message()
		// XXX: "transport is closing" is not documented but is needed to uniquely interpret codes.Unavailable
		// https://github.com/grpc/grpc/blob/v1.24.0/doc/statuscodes.md
		if errCode != codes.Unavailable || errMsg != "transport is closing" {
			return fmt.Errorf("Could not stop hub: %w", err)
		}
	}
	return nil
}

func stopAgentsCmd() *cobra.Command {
	stopAgentsCmd := &cobra.Command{
		Use:     "agents",
		Short:   "Stop agents",
		PreRunE: InitializeCommand,
		RunE:    RunStopAgents,
	}

	return stopAgentsCmd
}

func RunStopAgents(cmd *cobra.Command, args []string) error {
	err := StopAgentService()
	if err != nil {
		return fmt.Errorf("Error stopping agent service:%w", err)
	}
	gplog.Info("Agents stopped successfully")
	if verbose {
		_ = ShowAgentsStatus(conf, false)
	}
	return nil
}

var StopAgentService = func() error {
	client, err := connectToHub(conf)
	if err != nil {
		return fmt.Errorf("Could not connect to hub; is the hub running?")
	}

	_, err = client.StopAgents(context.Background(), &idl.StopAgentsRequest{})
	if err != nil {
		return fmt.Errorf("Could not stop agents: %w", err)
	}
	return nil
}

func StopServicesCmd() *cobra.Command {
	stopServicesCmd := &cobra.Command{
		Use:     "services",
		Short:   "Stop agents and hub services",
		PreRunE: InitializeCommand,
		RunE:    RunStopServices,
	}

	return stopServicesCmd
}

func RunStopServices(cmd *cobra.Command, args []string) error {
	err := StopAgentService()
	if err != nil {
		return fmt.Errorf("Error while stopping Agent Service:%w", err)
	}
	gplog.Info("Agents stopped successfully")
	err = StopHubService()
	if err != nil {
		return fmt.Errorf("Error while stopping Hub Service:%w", err)
	}
	gplog.Info("Hub stopped successfully")
	return nil
}
