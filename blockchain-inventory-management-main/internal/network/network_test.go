package network_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"inventory-chain/internal/ledger"
	"inventory-chain/internal/network"
	"inventory-chain/internal/worldstate"
	bolt "go.etcd.io/bbolt"
)

func TestNetworkFlow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "network-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	net, err := network.SetupNetwork(db)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Success case: Issue transaction proposal
	assetPayload := worldstate.Asset{
		AssetID:   "asset-network-1",
		DeptID:    "Lab",
		Name:      "Thermometer",
		Category:  "Lab Equipment",
		Qty:       10,
		Threshold: 3,
		UpdatedAt: time.Now().UTC(),
	}
	payloadBytes, _ := json.Marshal(assetPayload)

	tx := ledger.Transaction{
		TxID:      "tx-net-1",
		Type:      "ISSUE",
		DeptID:    "Lab",
		Payload:   payloadBytes,
		Timestamp: time.Now().UTC(),
	}

	block, err := net.ProposeAndCommit(tx)
	if err != nil {
		t.Fatalf("expected successful issue, got error: %v", err)
	}

	if len(block.Transactions) != 1 {
		t.Errorf("expected 1 tx in block, got %d", len(block.Transactions))
	}

	// Check if asset is now in world state
	asset, err := net.Store.Get("asset-network-1")
	if err != nil {
		t.Fatalf("expected asset to be persisted: %v", err)
	}
	if asset.Name != "Thermometer" {
		t.Errorf("expected asset name Thermometer, got %s", asset.Name)
	}

	// 2. Reject: Issue transaction with negative qty
	assetPayload2 := worldstate.Asset{
		AssetID:   "asset-network-2",
		DeptID:    "Lab",
		Name:      "Glass Beaker",
		Category:  "Glassware",
		Qty:       -5,
		Threshold: 3,
		UpdatedAt: time.Now().UTC(),
	}
	payloadBytes2, _ := json.Marshal(assetPayload2)

	txInvalid := ledger.Transaction{
		TxID:      "tx-net-2",
		Type:      "ISSUE",
		DeptID:    "Lab",
		Payload:   payloadBytes2,
		Timestamp: time.Now().UTC(),
	}

	_, err = net.ProposeAndCommit(txInvalid)
	if err == nil {
		t.Error("expected error for invalid issue transaction (negative qty)")
	}
}
