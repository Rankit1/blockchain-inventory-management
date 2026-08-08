package genai

import (
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"

	"inventory-chain/internal/network"
	"inventory-chain/internal/worldstate"
)

func TestPredictiveSchedulesAudit(t *testing.T) {
	td := t.TempDir()
	dbPath := filepath.Join(td, "pred.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("failed to open bolt DB: %v", err)
	}
	defer db.Close()

	net, err := network.SetupNetwork(db)
	if err != nil {
		t.Fatalf("failed to setup network: %v", err)
	}

	// Create a P1 asset without lastAuditDate
	a := worldstate.Asset{
		AssetID:        "asset-pred-1",
		DeptID:         "IT",
		Name:           "CriticalRouter",
		Category:       "network",
		Qty:            1,
		Threshold:      1,
		PriorityTier:   "P1",
		LifecycleState: worldstate.LifecycleActive,
	}
	if err := net.Store.Put(a); err != nil {
		t.Fatalf("failed to put asset: %v", err)
	}

	driver := NewSimulationDriver(net)
	pa := NewPredictive(driver, 100*time.Millisecond)
	pa.Start()

	// wait for scheduled audit (should create audit state entry and append ledger)
	deadline := time.Now().Add(5 * time.Second)
	var found bool
	for time.Now().Before(deadline) {
		// Check ledger blocks for AUDIT_SCHEDULE tx
		blocks, err := net.Ledger.GetBlocks()
		if err == nil {
			for _, b := range blocks {
				for _, tx := range b.Transactions {
					if tx.Type == "AUDIT_SCHEDULE" {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}
		if found {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	pa.Stop()

	if !found {
		t.Fatalf("expected predictive agent to schedule an audit; none found")
	}
}
