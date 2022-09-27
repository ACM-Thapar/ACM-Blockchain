package node

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/ACM-Thapar/ACM-Blockchain/database"
)

func TestValidBlockHash(t *testing.T) {
	hexHash := "000000fa04f816039...a4db586086168edfa"
	var hash = database.Hash{}
	hex.Decode(hash[:], []byte(hexHash))

	isValid := database.IsBlockHashValid(hash)
	if !isValid {
		t.Fatalf("hash '%s' with 6 zeroes should be valid", hexHash)
	}
}

func TestMine(t *testing.T) {
	miner := database.NewAccount("jhnda")
	pendingBlock := createRandomPendingBlock(miner)
	ctx := context.Background()
	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil {
		t.Fatal(err)
	}
	minedBlockHash, err := minedBlock.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if !database.IsBlockHashValid(minedBlockHash) {
		t.Fatal()
	}
}
