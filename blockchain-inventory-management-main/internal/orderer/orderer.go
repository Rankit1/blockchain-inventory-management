package orderer

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"inventory-chain/internal/chaincode"
	"inventory-chain/internal/ledger"
	"inventory-chain/internal/worldstate"
	"github.com/google/uuid"
)

type BlockEvent struct {
	Block ledger.Block
}

type Orderer struct {
	mu                   sync.Mutex
	Ledger               *ledger.Ledger
	Store                *worldstate.Store
	EndorsementThreshold int
	EventChan            chan BlockEvent
}

func NewOrderer(l *ledger.Ledger, store *worldstate.Store, threshold int) *Orderer {
	return &Orderer{
		Ledger:               l,
		Store:                store,
		EndorsementThreshold: threshold,
		EventChan:            make(chan BlockEvent, 100),
	}
}

// Submit validates endorsement threshold, commits state changes to WorldState,
// cuts a block, appends to the ledger, and sends block event notifications.
func (o *Orderer) Submit(tx ledger.Transaction) (ledger.Block, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// 1. Verify endorsement threshold policy
	if len(tx.Endorsers) < o.EndorsementThreshold {
		return ledger.Block{}, fmt.Errorf("endorsement threshold not met: got %d, need %d", len(tx.Endorsers), o.EndorsementThreshold)
	}

	// 2. Prepare transactions to store in the block
	transactions := []ledger.Transaction{tx}

	// 3. Commit changes to WorldState based on transaction type
	switch tx.Type {
	case "ISSUE":
		var asset worldstate.Asset
		if err := json.Unmarshal(tx.Payload, &asset); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to decode issue payload: %w", err)
		}
		if err := o.Store.Put(asset); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to put asset in world state: %w", err)
		}
	case "CONSUME":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to decode consume payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		qtyVal, _ := m["qty"].(float64)
		qty := int(qtyVal)

		asset, err := o.Store.Get(assetID)
		if err != nil {
			return ledger.Block{}, fmt.Errorf("failed to read asset from world state: %w", err)
		}
		asset.Qty -= qty
		if err := o.Store.Put(asset); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to update asset in world state: %w", err)
		}

		// Automatic replenishment trigger if stock falls below the tier-aware reorder point
		if asset.Qty < chaincode.ReorderPoint(asset) {
			replenishPayloadMap := map[string]interface{}{
				"assetId": assetID,
				"qty":     asset.Threshold * 2, // Replenish double the threshold qty
				"urgency": "HIGH",
				"deptId":  asset.DeptID,
			}
			replenishPayload, _ := json.Marshal(replenishPayloadMap)
			replenishTx := ledger.Transaction{
				TxID:      uuid.New().String(),
				Type:      "REPLENISH_REQUEST",
				DeptID:    asset.DeptID,
				Payload:   replenishPayload,
				Endorsers: tx.Endorsers, // Inherit same endorsing peer signatures
				Timestamp: time.Now().UTC(),
			}
			transactions = append(transactions, replenishTx)
		}
	case "TRANSFER":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to decode transfer payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		qtyVal, _ := m["qty"].(float64)
		qty := int(qtyVal)
		fromDept, _ := m["fromDept"].(string)
		toDept, _ := m["toDept"].(string)
		targetAssetID, _ := m["targetAssetId"].(string)

		// Deduct from source
		srcAsset, err := o.Store.Get(assetID)
		if err != nil {
			return ledger.Block{}, fmt.Errorf("failed to read source asset from world state: %w", err)
		}
		if srcAsset.DeptID != fromDept {
			return ledger.Block{}, fmt.Errorf("transfer failed: source asset belongs to %s, expected %s", srcAsset.DeptID, fromDept)
		}
		srcAsset.Qty -= qty
		if err := o.Store.Put(srcAsset); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to update source asset: %w", err)
		}

		// Add to target copy
		var targetAsset worldstate.Asset
		targetAsset, err = o.Store.Get(targetAssetID)
		if err != nil {
			// If target asset doesn't exist, create it in target department
			targetAsset = worldstate.Asset{
				AssetID:   targetAssetID,
				DeptID:    toDept,
				Name:      srcAsset.Name,
				Category:  srcAsset.Category,
				Qty:       qty,
				Threshold: srcAsset.Threshold,
			}
		} else {
			targetAsset.Qty += qty
		}
		if err := o.Store.Put(targetAsset); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to update target asset: %w", err)
		}
	case "REPLENISH_REQUEST":
		// No world state changes needed, just record on ledger
	case "ANCHOR_HASH":
		// No world state changes needed, just record on ledger
	case "PRIORITY_CLASSIFY":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to decode classify payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		tier, _ := m["priorityTier"].(string)
		scoreVal, _ := m["criticalityScore"].(float64)

		asset, err := o.Store.Get(assetID)
		if err != nil {
			return ledger.Block{}, fmt.Errorf("failed to read asset from world state: %w", err)
		}
		asset.PriorityTier = tier
		asset.CriticalityScore = scoreVal
		if err := o.Store.Put(asset); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to update asset in world state: %w", err)
		}
	case "PRIORITY_UPDATE":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to decode priority update payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		tier, _ := m["priorityTier"].(string)

		asset, err := o.Store.Get(assetID)
		if err != nil {
			return ledger.Block{}, fmt.Errorf("failed to read asset from world state: %w", err)
		}
		asset.PriorityTier = tier
		if err := o.Store.Put(asset); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to update asset in world state: %w", err)
		}
	case "AUDIT_SCHEDULE":
		// No world state changes needed, just record on ledger
	case "AUDIT_RESULT":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to decode audit result payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		auditDate, _ := m["auditDate"].(string)

		asset, err := o.Store.Get(assetID)
		if err != nil {
			return ledger.Block{}, fmt.Errorf("failed to read asset from world state: %w", err)
		}
		asset.LastAuditDate = auditDate
		if err := o.Store.Put(asset); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to update asset in world state: %w", err)
		}
	case "RETIRE":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to decode retire payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)

		asset, err := o.Store.Get(assetID)
		if err != nil {
			return ledger.Block{}, fmt.Errorf("failed to read asset from world state: %w", err)
		}
		asset.LifecycleState = worldstate.LifecycleRetired
		if err := o.Store.Put(asset); err != nil {
			return ledger.Block{}, fmt.Errorf("failed to update asset in world state: %w", err)
		}
	}

	// 4. Retrieve current blockchain blocks to find previous block hash
	blocks, err := o.Ledger.GetBlocks()
	if err != nil {
		return ledger.Block{}, fmt.Errorf("failed to fetch blocks: %w", err)
	}
	if len(blocks) == 0 {
		return ledger.Block{}, fmt.Errorf("ledger is empty; genesis block missing")
	}
	prevBlock := blocks[len(blocks)-1]

	// 5. Cut a new block
	newBlock := ledger.Block{
		Index:        prevBlock.Index + 1,
		PrevHash:     prevBlock.Hash,
		Transactions: transactions,
		Timestamp:    time.Now().UTC(),
	}
	newBlock.Hash = newBlock.CalculateHash()

	// 6. Commit block to ledger persistence
	if err := o.Ledger.AppendBlock(newBlock); err != nil {
		return ledger.Block{}, fmt.Errorf("failed to append block: %w", err)
	}

	// 7. Publish block event asynchronously to subscribers
	select {
	case o.EventChan <- BlockEvent{Block: newBlock}:
	default:
		// Drop event if queue is full to prevent blocking
	}

	return newBlock, nil
}
