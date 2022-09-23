package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var acmtCmd = &cobra.Command{
		Use:   "acmt",
		Short: "The CLI of ACM's own Blockchain",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	acmtCmd.AddCommand(versionCmd)
	acmtCmd.AddCommand(balancesCmd())
	acmtCmd.AddCommand(txCmd())

	err := acmtCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
