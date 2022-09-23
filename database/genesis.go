package database

import (
	"encoding/json"
	"os"
)

var genesisJson = `
{
    "genesis_time": "2022-09-22T00:00:00.000000000Z",
    "chain_id": "acmTiet-blockchain-ledger",
    "balances": {
        "jhnda": 1000000
    }
}`

type genesis struct {
	Balances map[Account]uint `json:"balances"`
}

func loadGenesis(path string) (genesis, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return genesis{}, err
	}

	var loadedGenesis genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return genesis{}, err
	}

	return loadedGenesis, nil
}

func writeGenesisToDisk(path string) error {
	return os.WriteFile(path, []byte(genesisJson), 0644)
}
