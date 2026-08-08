package ledger_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"inventory-chain/internal/ledger"
	bolt "go.etcd.io/bbolt"
)

func TestLedgerGenesisAndVerify(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ledger-test")
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

	l, err := ledger.NewLedger(db)
	if err != nil {
		t.Fatal(err)
	}

	// Verify chain is valid (contains just genesis block)
	valid, brokenIdx, err := l.VerifyChain()
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Errorf("expected chain to be valid, got broken at %d", brokenIdx)
	}

	// Retrieve blocks
	blocks, err := l.GetBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Errorf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Index != 0 {
		t.Errorf("expected index 0, got %d", blocks[0].Index)
	}
}

func TestLedgerAppendAndTampering(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ledger-test")
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

	l, err := ledger.NewLedger(db)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve genesis block
	blocks, err := l.GetBlocks()
	if err != nil {
		t.Fatal(err)
	}
	genesis := blocks[0]

	// Create and append block 1
	txPayload := json.RawMessage(`{"assetId": "asset-123", "qty": 10}`)
	tx := ledger.Transaction{
		TxID:      "tx-1",
		Type:      "ISSUE",
		DeptID:    "Lab",
		Payload:   txPayload,
		Endorsers: []string{"Peer-Lab", "Peer-Admin"},
		Timestamp: time.Now(),
	}

	block1 := ledger.Block{
		Index:        1,
		PrevHash:     genesis.Hash,
		Transactions: []ledger.Transaction{tx},
		Timestamp:    time.Now(),
	}
	block1.Hash = block1.CalculateHash()

	err = l.AppendBlock(block1)
	if err != nil {
		t.Fatal(err)
	}

	// Verify chain is valid
	valid, brokenIdx, err := l.VerifyChain()
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Errorf("expected chain to be valid, got broken at %d", brokenIdx)
	}

	// Tamper with Block 1 in the DB directly
	err = db.Update(func(boltTx *bolt.Tx) error {
		b := boltTx.Bucket([]byte(ledger.LedgerBucketName))
		// Fetch block1 raw data
		key := []byte("000000000001")
		v := b.Get(key)
		var blk ledger.Block
		json.Unmarshal(v, &blk)
		// Change the quantity in the transaction payload
		blk.Transactions[0].Payload = json.RawMessage(`{"assetId": "asset-123", "qty": 9999}`)
		// Put it back without recalculating hash
		tamperedData, _ := json.Marshal(blk)
		return b.Put(key, tamperedData)
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify chain is now invalid
	valid, brokenIdx, err = l.VerifyChain()
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Error("expected chain to be invalid after tampering")
	}
	if brokenIdx != 1 {
		t.Errorf("expected chain to break at index 1, got %d", brokenIdx)
	}
}
