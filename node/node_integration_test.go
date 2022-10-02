package node

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ACM-Thapar/ACM-Blockchain/database"
	"github.com/ACM-Thapar/ACM-Blockchain/fs"
)

func getTestDataDirPath() string {
	return filepath.Join(os.TempDir(), ".tbb_test")
}
func TestNode_Run(t *testing.T) {
	datadir := getTestDataDirPath()
	err := fs.RemoveDir(datadir)
	if err != nil {
		t.Fatal(err)
	}

	n := New(
		datadir,
		"127.0.0.1",
		8085,
		database.NewAccount("jhnda"),
		PeerNode{},
	)

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	err = n.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNode_Mining(t *testing.T) {
	datadir := getTestDataDirPath()
	err := fs.RemoveDir(datadir)
	if err != nil {
		t.Fatal(err)
	}

	nInfo := NewPeerNode(
		"127.0.0.1",
		8085,
		false,
		database.NewAccount(""),
		true,
	)

	n := New(datadir, nInfo.IP, nInfo.Port, database.NewAccount("jhnda"), nInfo)

	ctx, closeNode := context.WithTimeout(
		context.Background(),
		time.Minute*30,
	)

	go func() {
		time.Sleep(time.Second * miningIntervalSeconds / 3)
		tx := database.NewTx("jhnda", "pimdii", 1, "")

		_ = n.AddPendingTX(tx, nInfo)
	}()

	go func() {
		time.Sleep(time.Second*miningIntervalSeconds + 2)
		tx := database.NewTx("jhnda", "pimdii", 2, "")

		_ = n.AddPendingTX(tx, nInfo)
	}()

	go func() {
		ticker := time.NewTicker(10 * time.Second)

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	_ = n.Run(ctx)

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("2 pending TX not mined into 2 under 30m")
	}
}

func TestNode_MiningStopsOnNewSyncedBlock(t *testing.T) {
	datadir := getTestDataDirPath()
	err := fs.RemoveDir(datadir)
	if err != nil {
		t.Fatal(err)
	}

	nInfo := NewPeerNode(
		"127.0.0.1",
		8085,
		false,
		database.NewAccount(""),
		true,
	)

	jhndaAcc := database.NewAccount("jhnda")
	pimdiiAcc := database.NewAccount("pimdii")

	n := New(datadir, nInfo.IP, nInfo.Port, pimdiiAcc, nInfo)

	ctx, closeNode := context.WithTimeout(context.Background(), time.Minute*30)

	tx1 := database.NewTx("jhnda", "pimdii", 1, "")
	tx2 := database.NewTx("jhndaj", "pimdii", 2, "")
	tx2Hash, _ := tx2.Hash()

	validPreMinedPb := NewPendingBlock(database.Hash{}, 0, jhndaAcc, []database.Tx{tx1})
	validSyncedBlock, err := Mine(ctx, validPreMinedPb)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds - 2))

		err := n.AddPendingTX(tx1, nInfo)
		if err != nil {
			t.Fatal(err)
		}

		err = n.AddPendingTX(tx2, nInfo)
		if err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		time.Sleep(time.Second * (miningIntervalSeconds + 2))
		if !n.isMining {
			t.Fatal("should be mining")
		}

		_, err := n.state.AddBlock(validSyncedBlock)
		if err != nil {
			t.Fatal(err)
		}

		n.newSyncedBlocks <- validSyncedBlock

		time.Sleep(time.Second * 2)
		if n.isMining {
			t.Fatal("synced block should have canceled mining")
		}

		_, onlyTX2IsPending := n.pendingTXs[tx2Hash.Hex()]

		if len(n.pendingTXs) != 1 && !onlyTX2IsPending {
			t.Fatal("synced block should have canceled mining of already mined TX")
		}

		time.Sleep(time.Second * (miningIntervalSeconds + 2))
		if !n.isMining {
			t.Fatal("should be mining again the 1 TX not included in synced block")
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second * 10)

		for {
			select {
			case <-ticker.C:
				if n.state.LatestBlock().Header.Number == 1 {
					closeNode()
					return
				}
			}
		}
	}()

	go func() {
		time.Sleep(time.Second * 2)
		startingJhndaBalance := n.state.Balances[jhndaAcc]
		startingPimdiiBalance := n.state.Balances[pimdiiAcc]

		<-ctx.Done()

		endJhndaBalance := n.state.Balances[jhndaAcc]
		endPimdiiBalance := n.state.Balances[pimdiiAcc]

		expectedEndJhndaBalance := startingJhndaBalance - tx1.Value - tx2.Value + database.BlockReward
		expectedEndPimdiiBalance := startingPimdiiBalance + tx1.Value + tx2.Value + database.BlockReward

		if endJhndaBalance != expectedEndJhndaBalance {
			t.Fatalf("Andrej expected end balance is %d not %d", expectedEndJhndaBalance, endJhndaBalance)
		}

		if endPimdiiBalance != expectedEndPimdiiBalance {
			t.Fatalf("BabaYaga expected end balance is %d not %d", expectedEndPimdiiBalance, endPimdiiBalance)
		}

		t.Logf("Starting Andrej balance: %d", startingJhndaBalance)
		t.Logf("Starting BabaYaga balance: %d", startingPimdiiBalance)
		t.Logf("Ending Andrej balance: %d", endJhndaBalance)
		t.Logf("Ending BabaYaga balance: %d", endPimdiiBalance)
	}()

	_ = n.Run(ctx)

	if n.state.LatestBlock().Header.Number != 1 {
		t.Fatal("was suppose to mine 2 pending TX into 2 valid blocks under 30m")
	}

	if len(n.pendingTXs) != 0 {
		t.Fatal("no pending TXs should be left to mine")
	}
}
