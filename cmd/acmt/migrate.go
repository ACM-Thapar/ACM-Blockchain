package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ACM-Thapar/ACM-Blockchain/database"
	"github.com/ACM-Thapar/ACM-Blockchain/node"
	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			miner, _ := cmd.Flags().GetString(flagMiner)
			ip, _ := cmd.Flags().GetString(flagIP)
			port, _ := cmd.Flags().GetUint64(flagPort)

			peer := node.NewPeerNode(
				"127.0.0.1",
				8080,
				true,
				database.NewAccount("jhnda"),
				false,
			)

			n := node.New(getDataDirFromCmd(cmd), ip, port, database.NewAccount(miner), peer)

			n.AddPendingTX(database.NewTx("jhnda", "uddu", 1000, "last night reward"), peer)
			n.AddPendingTX(database.NewTx("jhnda", "jogesh", 100, "reward"), peer)
			n.AddPendingTX(database.NewTx("jhnda", "pimdii", 10000, ""), peer)
			n.AddPendingTX(database.NewTx("jhnda", "pimdii", 100000, ""), peer)
			n.AddPendingTX(database.NewTx("jhnda", "haris", 50, ""), peer)
			n.AddPendingTX(database.NewTx("haris", "jhnda", 50, "money back"), peer)

			ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*15)

			go func() {
				ticker := time.NewTicker(time.Second * 10)

				for {
					select {
					case <-ticker.C:
						if !n.LatestBlockHash().IsEmpty() {
							closeNode()
							return
						}
					}
				}
			}()

			err := n.Run(ctx)
			if err != nil {
				fmt.Println(err)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)
	migrateCmd.Flags().String(flagMiner, node.DefaultMiner, "miner account of this node to receive block rewards")
	migrateCmd.Flags().String(flagIP, node.DefaultIP, "exposed IP for communication with peers")
	migrateCmd.Flags().Uint64(flagPort, node.DefaultHTTPort, "exposed HTTP port for communication with peers")

	return migrateCmd
}
