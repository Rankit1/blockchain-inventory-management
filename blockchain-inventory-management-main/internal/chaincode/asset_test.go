package chaincode_test

import (
	"os"
	"path/filepath"
	"testing"

	"inventory-chain/internal/chaincode"
	"inventory-chain/internal/worldstate"
	bolt "go.etcd.io/bbolt"
)

func TestChaincodeIssueAsset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chaincode-test")
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

	// 1. Success case
	asset, tx, err := chaincode.IssueAsset(store, "Lab", "Beaker", "Glassware", 10, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if asset.Name != "Beaker" || asset.Qty != 10 {
		t.Errorf("asset fields mismatch: %+v", asset)
	}
	if tx.Type != "ISSUE" || tx.DeptID != "Lab" {
		t.Errorf("tx fields mismatch: %+v", tx)
	}

	// 2. Failure: Invalid department
	_, _, err = chaincode.IssueAsset(store, "InvalidDept", "Beaker", "Glassware", 10, 3)
	if err == nil {
		t.Error("expected error for invalid department")
	}

	// 3. Failure: Negative Qty
	_, _, err = chaincode.IssueAsset(store, "Lab", "Beaker", "Glassware", -1, 3)
	if err == nil {
		t.Error("expected error for negative quantity")
	}
}

func TestChaincodeConsumeStock(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chaincode-test")
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

	// Setup initial asset
	origAsset := worldstate.Asset{
		AssetID:   "asset-123",
		DeptID:    "Lab",
		Name:      "Beaker",
		Category:  "Glassware",
		Qty:       10,
		Threshold: 5,
	}
	err = store.Put(origAsset)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Happy path consume (no replenish trigger)
	updated, tx, replenish, err := chaincode.ConsumeStock(store, "Lab", "asset-123", 4, "experiments")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Qty != 6 {
		t.Errorf("expected updated qty 6, got %d", updated.Qty)
	}
	if replenish {
		t.Error("replenish should not be triggered (6 >= 5)")
	}
	if tx.Type != "CONSUME" {
		t.Errorf("tx type incorrect: %s", tx.Type)
	}

	// 2. Consume that triggers replenishment
	// We save the updated asset to store first to simulate state progression
	err = store.Put(updated)
	if err != nil {
		t.Fatal(err)
	}
	updated2, _, replenish2, err := chaincode.ConsumeStock(store, "Lab", "asset-123", 2, "more experiments")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated2.Qty != 4 {
		t.Errorf("expected updated qty 4, got %d", updated2.Qty)
	}
	if !replenish2 {
		t.Error("expected replenish to be triggered (4 < 5)")
	}

	// 3. Reject: insufficient stock
	err = store.Put(updated2)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = chaincode.ConsumeStock(store, "Lab", "asset-123", 10, "too much")
	if err == nil {
		t.Error("expected rejection for insufficient stock")
	}

	// 4. Reject: wrong department
	_, _, _, err = chaincode.ConsumeStock(store, "Store", "asset-123", 1, "wrong dept")
	if err == nil {
		t.Error("expected rejection for department mismatch")
	}
}

func TestChaincodeTransferAsset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "chaincode-test")
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

	// Setup source asset in Lab
	srcAsset := worldstate.Asset{
		AssetID:   "asset-src",
		DeptID:    "Lab",
		Name:      "Laptop",
		Category:  "Electronics",
		Qty:       5,
		Threshold: 1,
	}
	err = store.Put(srcAsset)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Transfer to destination IT (dest asset doesn't exist yet)
	updatedSrc, updatedDest, tx, err := chaincode.TransferAsset(store, "Lab", "IT", "asset-src", 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updatedSrc.Qty != 3 {
		t.Errorf("expected source qty 3, got %d", updatedSrc.Qty)
	}
	if updatedDest.Qty != 2 {
		t.Errorf("expected destination qty 2, got %d", updatedDest.Qty)
	}
	if updatedDest.DeptID != "IT" || updatedDest.Name != "Laptop" {
		t.Errorf("destination asset fields incorrect: %+v", updatedDest)
	}
	if tx.Type != "TRANSFER" {
		t.Errorf("tx type incorrect: %s", tx.Type)
	}

	// Put assets to store to simulate commit
	err = store.Put(updatedSrc)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Put(*updatedDest)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Transfer again to IT (dest asset exists now, should update existing copy)
	_, updatedDest2, _, err := chaincode.TransferAsset(store, "Lab", "IT", "asset-src", 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updatedDest2.Qty != 3 {
		t.Errorf("expected destination qty 3 after second transfer, got %d", updatedDest2.Qty)
	}
}
