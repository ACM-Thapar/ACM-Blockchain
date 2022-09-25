package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ACM-Thapar/ACM-Blockchain/database"
	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			block0 := database.NewBlock(
				database.Hash{},
				state.NextBlockNumber(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("jhnda", "uddu", 1000, ""),
					database.NewTx("jhnda", "jogesh", 1000, ""),
				},
			)

			block0hash, err := state.AddBlock(block0)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block1 := database.NewBlock(
				block0hash,
				state.NextBlockNumber(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("jhnda", "uddu", 1000, "last night reward"),
					database.NewTx("jhnda", "jogesh", 100, "reward"),
					database.NewTx("jhnda", "pimdii", 10000, ""),
					database.NewTx("jhnda", "pimdii", 100000, ""),
					database.NewTx("jhnda", "haris", 50, ""),
					database.NewTx("haris", "jhnda", 50, "money back"),
				},
			)

			block1hash, err := state.AddBlock(block1)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			block2 := database.NewBlock(
				block1hash,
				state.NextBlockNumber(),
				uint64(time.Now().Unix()),
				[]database.Tx{
					database.NewTx("jhnda", "jhnda", 24700, "reward"),
				},
			)

			_, err = state.AddBlock(block2)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	return migrateCmd
}
