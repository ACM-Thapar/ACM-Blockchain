package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ACM-Thapar/ACM-Blockchain/database"
)

func main() {
	state, err := database.NewStateFromDisk()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	block0 := database.NewBlock(
		database.Hash{},
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("jhnda", "jogesh", 2000, ""),
			database.NewTx("jhnda", "uddu", 2000, ""),
		},
	)

	state.AddBlock(block0)
	block0hash, _ := state.Persist()

	block1 := database.NewBlock(
		block0hash,
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("jhnda", "uddu", 1000, "last night reward"),
		},
	)

	state.AddBlock(block1)
	state.Persist()
}
