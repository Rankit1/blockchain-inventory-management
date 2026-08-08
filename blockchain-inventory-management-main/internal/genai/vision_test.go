package genai

import (
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"

	"inventory-chain/internal/network"
	"inventory-chain/internal/worldstate"
)

func TestVisionRecordsAudit(t *testing.T) {
	td := t.TempDir()
	dbPath := filepath.Join(td, "vis.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("failed to open bolt DB: %v", err)
	}
	defer db.Close()

	net, err := network.SetupNetwork(db)
	if err != nil {
		t.Fatalf("failed to setup network: %v", err)
	}

	// Seed an asset without lastAuditDate
	a := worldstate.Asset{AssetID: "asset-vis-1", DeptID: "d1", Name: "VisAsset", Category: "GEN", Qty: 1, UpdatedAt: time.Now().UTC(), LifecycleState: worldstate.LifecycleActive}
	if err := net.Store.Put(a); err != nil {
		t.Fatalf("failed to seed asset: %v", err)
	}

	driver := NewSimulationDriver(net)
	models := DefaultModelProvider()
	v := NewVision(driver, 50*time.Millisecond, models)
	v.Start()
	time.Sleep(150 * time.Millisecond)
	v.Stop()

	updated, err := net.Store.Get("asset-vis-1")
	if err != nil {
		t.Fatalf("failed to get asset: %v", err)
	}
	if updated.LastAuditDate == "" {
		t.Fatalf("expected LastAuditDate to be set by VisionAgent")
	}
}
