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

	return stopCmd
}

func stopHubCmd() *cobra.Command {
	stopHubCmd := &cobra.Command{
		Use:    "hub",
		Short:  "Stop hub",
		PreRun: InitializeCommand,
		RunE:   RunStopHub,
	}

	return stopHubCmd
}

func RunStopHub(cmd *cobra.Command, args []string) error {
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
	gplog.Info("Hub stopped successfully")
	if verbose {
		_ = ShowHubStatus(conf)
	}
	return nil
}

func stopAgentsCmd() *cobra.Command {
	stopAgentsCmd := &cobra.Command{
		Use:    "agents",
		Short:  "Stop agents",
		PreRun: InitializeCommand,
		RunE:   RunStopAgents,
	}

	return stopAgentsCmd
}

func RunStopAgents(cmd *cobra.Command, args []string) error {
	client, err := connectToHub(conf)
	if err != nil {
		return fmt.Errorf("Could not connect to hub; is the hub running?")
	}

	_, err = client.StopAgents(context.Background(), &idl.StopAgentsRequest{})
	if err != nil {
		return fmt.Errorf("Could not stop agents: %w", err)
	}
	gplog.Info("Agents stopped successfully")
	if verbose {
		_ = ShowAgentsStatus(client, conf)
	}
	return nil
}
