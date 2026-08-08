package genai

import (
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"

	"inventory-chain/internal/network"
	"inventory-chain/internal/worldstate"
)

func TestGenAIClassification(t *testing.T) {
	td := t.TempDir()
	dbPath := filepath.Join(td, "test.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("failed to open bolt DB: %v", err)
	}
	defer db.Close()

	net, err := network.SetupNetwork(db)
	if err != nil {
		t.Fatalf("failed to setup network: %v", err)
	}

	// Create an asset with empty PriorityTier so the GenAI scanner will classify it.
	a := worldstate.Asset{
		AssetID:        "asset-genai-1",
		DeptID:         "Lab",
		Name:           "TestDevice",
		Category:       "server",
		Qty:            1,
		Threshold:      1,
		LifecycleState: worldstate.LifecycleActive,
	}
	if err := net.Store.Put(a); err != nil {
		t.Fatalf("failed to put asset: %v", err)
	}

	driver := NewSimulationDriver(net)
	svc := New(driver, 100*time.Millisecond)
	svc.Start()

	// Wait for classification to occur with a timeout.
	deadline := time.Now().Add(5 * time.Second)
	var gotTier string
	for time.Now().Before(deadline) {
		as, err := net.Store.Get(a.AssetID)
		if err == nil && as.PriorityTier != "" {
			gotTier = as.PriorityTier
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	svc.Stop()

	if gotTier == "" {
		t.Fatalf("expected asset to be classified, but PriorityTier is empty")
	}
	if gotTier != "P1" && gotTier != "P2" && gotTier != "P3" {
		t.Fatalf("unexpected priority tier: %s", gotTier)
	}

	// Ensure a ledger block was appended beyond genesis (index > 0)
	blocks, err := net.Ledger.GetBlocks()
	if err != nil {
		t.Fatalf("failed to read blocks: %v", err)
	}
	if len(blocks) <= 1 {
		t.Fatalf("expected classification transaction to append a block; found %d blocks", len(blocks))
	}
}
