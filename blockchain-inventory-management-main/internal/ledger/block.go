package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Transaction struct {
	TxID      string          `json:"txId"`
	Type      string          `json:"type"` // ISSUE, CONSUME, TRANSFER, REPLENISH_REQUEST, ANCHOR_HASH
	DeptID    string          `json:"deptId"`
	Payload   json.RawMessage `json:"payload"`
	Endorsers []string        `json:"endorsers"`
	Timestamp time.Time       `json:"timestamp"`
}

type Block struct {
	Index        int           `json:"index"`
	PrevHash     string        `json:"prevHash"`
	Hash         string        `json:"hash"`
	Transactions []Transaction `json:"transactions"`
	Timestamp    time.Time     `json:"timestamp"`
}

// CalculateHash computes the SHA-256 hash of the block.
func (b *Block) CalculateHash() string {
	txsJSON, _ := json.Marshal(b.Transactions)
	// Using RFC3339Nano in UTC to ensure deterministic formatting
	data := fmt.Sprintf("%d%s%s%s", b.Index, b.PrevHash, string(txsJSON), b.Timestamp.UTC().Format(time.RFC3339Nano))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// NewGenesisBlock creates the first block of the chain.
func NewGenesisBlock() Block {
	b := Block{
		Index:        0,
		PrevHash:     "0000000000000000000000000000000000000000000000000000000000000000",
		Transactions: []Transaction{},
		Timestamp:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	b.Hash = b.CalculateHash()
	return b
}
