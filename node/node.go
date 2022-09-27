package node

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
	info    PeerNode
	ip      string
	port    uint64
	state   *database.State

	knownPeers      map[string]PeerNode
	pendingTXs      map[string]database.Tx
	archivedTXs     map[string]database.Tx
	newSyncedBlocks chan database.Block
	newPendingTXs   chan database.Tx
}

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`
	connected   bool
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

func (n *Node) Run() error {
	ctx := context.Background()
	fmt.Println(fmt.Sprintf("Listening on: %s:%d", n.ip, n.port))

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	defer state.Close()
	n.state = state

	go n.sync(ctx)

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	http.HandleFunc("/tx/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, n)
	})

	http.HandleFunc(endpointStatus, func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})

	http.HandleFunc(endpointSync, func(w http.ResponseWriter, r *http.Request) {
		syncHandler(w, r, n)
	})

	http.HandleFunc(endpointAddPeer, func(w http.ResponseWriter, r *http.Request) {
		addPeerHandler(w, r, n)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", n.port), nil)
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, connected bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, connected}
}

func (n *Node) RemovePeer(peer PeerNode) {
	delete(n.knownPeers, peer.TcpAddress())
}

func (n *Node) IsKnownPeer(peer PeerNode) bool {
	if peer.IP == n.ip && peer.Port == n.port {
		return true
	}

	_, isKnownPeer := n.knownPeers[peer.TcpAddress()]

	return isKnownPeer
}
func (n *Node) AddPeer(peer PeerNode) {
	n.knownPeers[peer.TcpAddress()] = peer
}

func (n *Node) AddPendingTX(tx database.Tx, fromPeer PeerNode) error {
	txHash, err := tx.Hash()
	if err != nil {
		return err
	}

	txJson, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	_, isAlreadyPending := n.pendingTXs[txHash.Hex()]
	_, isArchived := n.archivedTXs[txHash.Hex()]

	if !isAlreadyPending && !isArchived {
		fmt.Printf("Added Pending TX %s from Peer %s\n", txJson, fromPeer.TcpAddress())
		n.pendingTXs[txHash.Hex()] = tx
		n.newPendingTXs <- tx
	}

	return nil
}
