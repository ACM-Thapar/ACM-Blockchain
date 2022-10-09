package main

import (
	"fmt"
	"os"

	"github.com/ACM-Thapar/ACM-Blockchain/fs"
	"github.com/spf13/cobra"
)

const flagDataDir = "datadir"
const flagMiner = "miner"
const flagIP = "ip"
const flagPort = "port"
const flagKeystoreFile = "keystore"

func main() {
	var acmtCmd = &cobra.Command{
		Use:   "acmt",
		Short: "The CLI of ACM's own Blockchain",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
	}

	acmtCmd.AddCommand(migrateCmd())
	acmtCmd.AddCommand(versionCmd)
	acmtCmd.AddCommand(runCmd())
	acmtCmd.AddCommand(balancesCmd())
	acmtCmd.AddCommand(walletCmd())

	err := acmtCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(flagDataDir, "", "Absolute path to the node data dir where the DB will/is stored")
	cmd.MarkFlagRequired(flagDataDir)
}

func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDataDir)

	return fs.ExpandPath(dataDir)
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}

func addKeystoreFlag(cmd *cobra.Command) {
	cmd.Flags().String(flagKeystoreFile, "", "Absolute path to the encrypted keystore file")
	cmd.MarkFlagRequired(flagKeystoreFile)
}
