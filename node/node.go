package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ACM-Thapar/ACM-Blockchain/database"
)

const DefaultIP = "127.0.0.1"
const DefaultHTTPort = 8080
const endpointStatus = "/node/status"

const endpointSync = "/node/sync"
const endpointSyncQueryKeyFromBlock = "fromBlock"

const endpointAddPeer = "/node/peer"
const endpointAddPeerQueryKeyIP = "ip"
const endpointAddPeerQueryKeyPort = "port"

func (pn PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}

type Node struct {
	dataDir string
	ip      string
	port    uint64
	state   *database.State

	knownPeers map[string]PeerNode
}

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	IsActive    bool   `json:"is_active"`
}

func New(dataDir string, ip string, port uint64, bootstrap PeerNode) *Node {
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TcpAddress()] = bootstrap

	return &Node{
		dataDir:    dataDir,
		ip:         ip,
		port:       port,
		knownPeers: knownPeers,
	}
}

type ErrRes struct {
	Error string `json:"error"`
}

type BalancesRes struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type TxAddReq struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type TxAddRes struct {
	Hash database.Hash `json:"block_hash"`
}

type StatusRes struct {
	Hash       database.Hash `json:"block_hash"`
	Number     uint64        `json:"block_number"`
	KnownPeers []PeerNode    `json:"peers_known"`
}

type SyncRes struct {
	Blocks []database.Block `json:"blocks"`
}

func (n *Node) Run() error {
	ctx := context.Background()
	fmt.Println(fmt.Sprintf("Listening on HTTP port: %d", n.port))

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	go n.sync(ctx)

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, state)
	})

	http.HandleFunc("/node/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, state)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, connected bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, connected}
}

func listBalancesHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	writeRes(w, BalancesRes{state.LatestBlockHash(), state.Balances})
}

func txAddHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	req := TxAddReq{}
	err := readReq(r, &req)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	tx := database.NewTx(database.NewAccount(req.From), database.NewAccount(req.To), req.Value, req.Data)

	block := database.NewBlock(
		state.LatestBlockHash(),
		state.NextBlockNumber(),
		uint64(time.Now().Unix()),
		[]database.Tx{tx},
	)

	hash, err := state.AddBlock(block)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	writeRes(w, TxAddRes{hash})
}

func writeErrRes(w http.ResponseWriter, err error) {
	jsonErrRes, _ := json.Marshal(ErrRes{err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsonErrRes)
}

func writeRes(w http.ResponseWriter, content interface{}) {
	contentJson, err := json.Marshal(content)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(contentJson)
}

func readReq(r *http.Request, reqBody interface{}) error {
	reqBodyJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read request body. %s", err.Error())
	}
	defer r.Body.Close()

	err = json.Unmarshal(reqBodyJson, reqBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal request body. %s", err.Error())
	}

	return nil
}

func statusHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	res := StatusRes{
		Hash:   state.LatestBlockHash(),
		Number: state.LatestBlock().Header.Number,
	}
	writeRes(w, res)
}
func syncHandler(w http.ResponseWriter, r *http.Request, dataDir string) {
	// What's your latest block?
	// I will check my state, if I have newer blocks
	reqHash := r.URL.Query().Get(endpointSyncQueryKeyFromBlock)
	hash := database.Hash{}
	err := hash.UnmarshalText([]byte(reqHash))
	if err != nil {
		writeErrRes(w, err)
		return
	}
	// Read newer blocks from the DB
	blocks, err := database.GetBlocksAfter(hash, dataDir)
	if err != nil {
		writeErrRes(w, err)
		return
	}
	// JSON encode the blocks and return them in the response
	writeRes(w, SyncRes{Blocks: blocks})
}

func readRes(r *http.Response, reqBody interface{}) error {
	reqBodyJson, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body. %s", err.Error())
	}
	defer r.Body.Close()

	err = json.Unmarshal(reqBodyJson, reqBody)
	if err != nil {
		return fmt.Errorf("unable to unmarshal response body. %s", err.Error())
	}

	return nil
}
