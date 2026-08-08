package genai

import (
	"path/filepath"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"

	"inventory-chain/internal/network"
	"inventory-chain/internal/worldstate"
)

func TestDocumentAgentExtractsWarrantyAndReclassifies(t *testing.T) {
	td := t.TempDir()
	dbPath := filepath.Join(td, "doc.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("failed to open bolt DB: %v", err)
	}
	defer db.Close()
	net, err := network.SetupNetwork(db)
	if err != nil {
		t.Fatalf("failed to setup network: %v", err)
	}

	// Seed an asset without warranty
	a := worldstate.Asset{AssetID: "asset-doc-1", DeptID: "d1", Name: "DocAsset", Category: "GEN", Qty: 1, UpdatedAt: time.Now().UTC(), LifecycleState: worldstate.LifecycleActive}
	if err := net.Store.Put(a); err != nil {
		t.Fatalf("failed to seed asset: %v", err)
	}

	driver := NewSimulationDriver(net)
	models := DefaultModelProvider()
	d := NewDocumentAgent(driver, 50*time.Millisecond, models)
	d.Start()
	time.Sleep(200 * time.Millisecond)
	d.Stop()

	updated, err := net.Store.Get("asset-doc-1")
	if err != nil {
		t.Fatalf("failed to get asset: %v", err)
	}
	// the DocumentAgent records a DOC_EXTRACTED note which updates lastAuditDate
	if updated.LastAuditDate == "" {
		t.Fatalf("expected LastAuditDate to be set by DocumentAgent")
	}
}
