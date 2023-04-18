package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpdb/gp/cli"
)

func main() {
	root := cli.RootCommand()
	root.SilenceUsage = true
	root.SilenceErrors = true

	err := root.Execute()
	if err != nil {
		if strings.HasPrefix(err.Error(), "unknown flag") {
			fmt.Println("Help text goes here!")
		} else {
			gplog.Error(err.Error())
		}
		os.Exit(1)
	}
}
