package gpctl_cli

import (
	"github.com/greenplum-db/gpdb/gp/gpservice/hub"
	"github.com/spf13/cobra"
)

var (
	ConfigFilePath string
	Conf           *hub.Config
	Verbose        bool
)

func RootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "gpctl command",
	}
	//root.PersistentFlags().StringVar(&ConfigFilePath, "config-file", filepath.Join(os.Getenv("GPHOME"), constants.ConfigFileName), `Path to gpsesrvice configuration file`)
	//root.PersistentFlags().BoolVar(&Verbose, "verbose", false, `Provide verbose output`)

	root.AddCommand(initClusterCmd())
	return root
}
