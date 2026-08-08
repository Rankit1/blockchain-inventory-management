package network

import (
	"fmt"

	"inventory-chain/internal/ledger"
	"inventory-chain/internal/orderer"
	"inventory-chain/internal/peer"
	"inventory-chain/internal/worldstate"
	bolt "go.etcd.io/bbolt"
)

type Network struct {
	DB      *bolt.DB
	Ledger  *ledger.Ledger
	Store   *worldstate.Store
	Orderer *orderer.Orderer
	Peers   map[string]*peer.Peer
}

// SetupNetwork configures the blockchain layers, initializing bbolt database handlers.
func SetupNetwork(db *bolt.DB) (*Network, error) {
	l, err := ledger.NewLedger(db)
	if err != nil {
		return nil, fmt.Errorf("failed to setup ledger: %w", err)
	}

	s, err := worldstate.NewStore(db)
	if err != nil {
		return nil, fmt.Errorf("failed to setup worldstate: %w", err)
	}

	// Hyperledger endorsement threshold config: require >= 2 of 4 peer endorsements
	ord := orderer.NewOrderer(l, s, 2)

	peers := map[string]*peer.Peer{
		"Lab":   peer.NewPeer("Peer-Lab", "Lab", s),
		"Admin": peer.NewPeer("Peer-Admin", "Admin", s),
		"Store": peer.NewPeer("Peer-Store", "Store", s),
		"IT":    peer.NewPeer("Peer-IT", "IT", s),
	}

	return &Network{
		DB:      db,
		Ledger:  l,
		Store:   s,
		Orderer: ord,
		Peers:   peers,
	}, nil
}

// ProposeAndCommit processes a transaction through Fabric's Execute-Order-Validate flow:
// 1. Proposals are sent to peers for simulation and endorsement signatures.
// 2. Successful endorsements are attached to the transaction.
// 3. Transaction is submitted to the Orderer for sequencing, validation, state update, and block cutting.
func (n *Network) ProposeAndCommit(tx ledger.Transaction) (ledger.Block, error) {
	// 1. Request endorsements from all 4 peers in parallel or sequence
	var signatures []string
	for _, p := range n.Peers {
		sig, err := p.Endorse(tx)
		if err == nil {
			signatures = append(signatures, sig)
		}
	}

	// 2. Attach collected signatures
	tx.Endorsers = signatures

	// 3. Submit transaction with endorsements to the Orderer
	block, err := n.Orderer.Submit(tx)
	if err != nil {
		return ledger.Block{}, fmt.Errorf("ordering and commit failed: %w", err)
	}

	return block, nil
}
