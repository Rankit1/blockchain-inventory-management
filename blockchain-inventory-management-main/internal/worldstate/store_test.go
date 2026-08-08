package worldstate_test

import (
	"os"
	"path/filepath"
	"testing"

	"inventory-chain/internal/worldstate"
	bolt "go.etcd.io/bbolt"
)

func TestStorePutGetListDelete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "worldstate-test")
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

	store, err := worldstate.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}

	// Verify not found
	_, err = store.Get("asset-1")
	if err == nil {
		t.Error("expected error retrieving non-existent asset")
	}

	// Put asset
	asset1 := worldstate.Asset{
		AssetID:   "asset-1",
		DeptID:    "Lab",
		Name:      "Beaker",
		Category:  "Glassware",
		Qty:       10,
		Threshold: 3,
	}
	err = store.Put(asset1)
	if err != nil {
		t.Fatal(err)
	}

	// Get asset
	retrieved, err := store.Get("asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.Name != "Beaker" || retrieved.Qty != 10 {
		t.Errorf("incorrect asset retrieved: %+v", retrieved)
	}

	// Put another asset
	asset2 := worldstate.Asset{
		AssetID:   "asset-2",
		DeptID:    "Admin",
		Name:      "Stapler",
		Category:  "Stationery",
		Qty:       5,
		Threshold: 1,
	}
	err = store.Put(asset2)
	if err != nil {
		t.Fatal(err)
	}

	// List all
	all, err := store.List("")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 assets, got %d", len(all))
	}

	// List filter by dept
	labAssets, err := store.List("Lab")
	if err != nil {
		t.Fatal(err)
	}
	if len(labAssets) != 1 || labAssets[0].AssetID != "asset-1" {
		t.Errorf("expected 1 lab asset, got: %+v", labAssets)
	}

	// Delete
	err = store.Delete("asset-1")
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.Get("asset-1")
	if err == nil {
		t.Error("expected error after deleting asset")
	}
}
