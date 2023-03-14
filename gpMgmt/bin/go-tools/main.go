package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/greenplum-db/gpdb/gp/cli"
)

func main() {
	root := cli.RootCommand()
	root.SilenceUsage = true

	err := root.Execute()
	if err != nil {
		if strings.HasPrefix(err.Error(), "unknown flag") {
			fmt.Println("Help text goes here!")
		} else {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}
