package ledger

import (
	"encoding/json"
	"fmt"
	bolt "go.etcd.io/bbolt"
)

const LedgerBucketName = "ledger"

type Ledger struct {
	db *bolt.DB
}

func NewLedger(db *bolt.DB) (*Ledger, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(LedgerBucketName))
		return err
	})
	if err != nil {
		return nil, err
	}
	l := &Ledger{db: db}
	// If the ledger is empty, append the genesis block
	blocks, err := l.GetBlocks()
	if err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		genesis := NewGenesisBlock()
		if err := l.AppendBlock(genesis); err != nil {
			return nil, err
		}
	}
	return l, nil
}

// AppendBlock appends a block to the bbolt ledger bucket.
func (l *Ledger) AppendBlock(block Block) error {
	return l.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LedgerBucketName))
		if b == nil {
			return fmt.Errorf("ledger bucket not found")
		}
		data, err := json.Marshal(block)
		if err != nil {
			return err
		}
		key := []byte(fmt.Sprintf("%012d", block.Index))
		return b.Put(key, data)
	})
}

// GetBlocks retrieves all blocks in order.
func (l *Ledger) GetBlocks() ([]Block, error) {
	var blocks []Block
	err := l.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LedgerBucketName))
		if b == nil {
			return nil
		}
		return b.ForEach(func(k, v []byte) error {
			var block Block
			if err := json.Unmarshal(v, &block); err != nil {
				return err
			}
			blocks = append(blocks, block)
			return nil
		})
	})
	return blocks, err
}

// GetBlock retrieves a single block by its index.
func (l *Ledger) GetBlock(index int) (Block, error) {
	var block Block
	err := l.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(LedgerBucketName))
		if b == nil {
			return fmt.Errorf("ledger bucket not found")
		}
		v := b.Get([]byte(fmt.Sprintf("%012d", index)))
		if v == nil {
			return fmt.Errorf("block not found at index %d", index)
		}
		return json.Unmarshal(v, &block)
	})
	return block, err
}

// VerifyChain walks the hash chain to check for any modifications.
// If valid, returns true, -1, nil.
// If invalid, returns false, the index of the first broken block, nil.
func (l *Ledger) VerifyChain() (bool, int, error) {
	blocks, err := l.GetBlocks()
	if err != nil {
		return false, -1, err
	}
	if len(blocks) == 0 {
		return true, -1, nil
	}
	for i := 0; i < len(blocks); i++ {
		// 1. Verify self hash
		calculated := blocks[i].CalculateHash()
		if blocks[i].Hash != calculated {
			return false, i, nil
		}
		// 2. Verify link to previous hash
		if i > 0 {
			if blocks[i].PrevHash != blocks[i-1].Hash {
				return false, i, nil
			}
		}
	}
	return true, -1, nil
}

// GetAssetID parses the payload JSON to look for "assetId" or "assetID".
func (tx *Transaction) GetAssetID() string {
	var m map[string]interface{}
	if err := json.Unmarshal(tx.Payload, &m); err != nil {
		return ""
	}
	if val, ok := m["assetId"].(string); ok {
		return val
	}
	if val, ok := m["assetID"].(string); ok {
		return val
	}
	return ""
}

// GetAssetHistory scans the ledger blocks for transactions involving the given asset ID.
func (l *Ledger) GetAssetHistory(assetID string) ([]Transaction, error) {
	var history []Transaction
	blocks, err := l.GetBlocks()
	if err != nil {
		return nil, err
	}
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			if tx.GetAssetID() == assetID {
				history = append(history, tx)
			}
		}
	}
	return history, nil
}
