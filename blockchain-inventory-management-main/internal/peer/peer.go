package peer

import (
	"encoding/json"
	"errors"
	"fmt"

	"inventory-chain/internal/ledger"
	"inventory-chain/internal/worldstate"
)

type Peer struct {
	ID    string
	Dept  string
	Store *worldstate.Store
}

func NewPeer(id string, dept string, store *worldstate.Store) *Peer {
	return &Peer{
		ID:    id,
		Dept:  dept,
		Store: store,
	}
}

// Endorse simulates a transaction proposal and returns a signature if valid.
func (p *Peer) Endorse(tx ledger.Transaction) (string, error) {
	// Re-run validations of chaincode based on transaction type.
	// In Hyperledger Fabric, endorsement is simulation against the world state.
	switch tx.Type {
	case "ISSUE":
		var asset worldstate.Asset
		if err := json.Unmarshal(tx.Payload, &asset); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		if asset.Qty < 0 {
			return "", errors.New("endorsement failed: negative quantity")
		}
		if asset.Threshold < 0 {
			return "", errors.New("endorsement failed: negative threshold")
		}
		if asset.DeptID != tx.DeptID {
			return "", errors.New("endorsement failed: dept mismatch")
		}
	case "CONSUME":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		qtyVal, _ := m["qty"].(float64)
		qty := int(qtyVal)

		asset, err := p.Store.Get(assetID)
		if err != nil {
			return "", fmt.Errorf("endorsement failed: asset not found: %w", err)
		}
		if asset.DeptID != tx.DeptID {
			return "", fmt.Errorf("endorsement failed: dept mismatch (asset belongs to %s, tx by %s)", asset.DeptID, tx.DeptID)
		}
		if asset.Qty < qty {
			return "", fmt.Errorf("endorsement failed: insufficient stock (available %d, requested %d)", asset.Qty, qty)
		}
	case "TRANSFER":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		qtyVal, _ := m["qty"].(float64)
		qty := int(qtyVal)
		fromDept, _ := m["fromDept"].(string)

		asset, err := p.Store.Get(assetID)
		if err != nil {
			return "", fmt.Errorf("endorsement failed: source asset not found: %w", err)
		}
		if asset.DeptID != fromDept {
			return "", fmt.Errorf("endorsement failed: source asset belongs to %s, expected %s", asset.DeptID, fromDept)
		}
		if asset.Qty < qty {
			return "", fmt.Errorf("endorsement failed: insufficient stock to transfer: %d < %d", asset.Qty, qty)
		}
	case "REPLENISH_REQUEST":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		qtyVal, _ := m["qty"].(float64)
		qty := int(qtyVal)
		if qty <= 0 {
			return "", errors.New("endorsement failed: replenish qty must be > 0")
		}
		_, err := p.Store.Get(assetID)
		if err != nil {
			return "", fmt.Errorf("endorsement failed: asset not found: %w", err)
		}
	case "ANCHOR_HASH":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		sha256Hash, _ := m["sha256Hash"].(string)
		if len(sha256Hash) != 64 {
			return "", errors.New("endorsement failed: invalid sha256 hash length")
		}
		_, err := p.Store.Get(assetID)
		if err != nil {
			return "", fmt.Errorf("endorsement failed: asset not found: %w", err)
		}
	case "PRIORITY_CLASSIFY":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		tier, _ := m["priorityTier"].(string)
		scoreVal, _ := m["criticalityScore"].(float64)
		if tier != "P1" && tier != "P2" && tier != "P3" {
			return "", errors.New("endorsement failed: invalid priority tier")
		}
		if scoreVal < 1.0 || scoreVal > 5.0 {
			return "", errors.New("endorsement failed: criticality score out of range")
		}
		_, err := p.Store.Get(assetID)
		if err != nil {
			return "", fmt.Errorf("endorsement failed: asset not found: %w", err)
		}
	case "PRIORITY_UPDATE":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		tier, _ := m["priorityTier"].(string)
		reason, _ := m["reason"].(string)
		if tier != "P1" && tier != "P2" && tier != "P3" {
			return "", errors.New("endorsement failed: invalid priority tier")
		}
		if reason == "" {
			return "", errors.New("endorsement failed: priority override requires justification")
		}
		_, err := p.Store.Get(assetID)
		if err != nil {
			return "", fmt.Errorf("endorsement failed: asset not found: %w", err)
		}
	case "AUDIT_SCHEDULE":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		auditDate, _ := m["auditDate"].(string)
		if auditDate == "" {
			return "", errors.New("endorsement failed: audit date required")
		}
		_, err := p.Store.Get(assetID)
		if err != nil {
			return "", fmt.Errorf("endorsement failed: asset not found: %w", err)
		}
	case "AUDIT_RESULT":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		result, _ := m["result"].(string)
		if result == "" {
			return "", errors.New("endorsement failed: audit result required")
		}
		_, err := p.Store.Get(assetID)
		if err != nil {
			return "", fmt.Errorf("endorsement failed: asset not found: %w", err)
		}
	case "RETIRE":
		var m map[string]interface{}
		if err := json.Unmarshal(tx.Payload, &m); err != nil {
			return "", fmt.Errorf("endorsement failed to parse payload: %w", err)
		}
		assetID, _ := m["assetId"].(string)
		reason, _ := m["reason"].(string)
		if reason == "" {
			return "", errors.New("endorsement failed: retirement reason required")
		}
		asset, err := p.Store.Get(assetID)
		if err != nil {
			return "", fmt.Errorf("endorsement failed: asset not found: %w", err)
		}
		if asset.LifecycleState == worldstate.LifecycleRetired {
			return "", errors.New("endorsement failed: asset already retired")
		}
	default:
		return "", fmt.Errorf("endorsement failed: unknown transaction type: %s", tx.Type)
	}

	// Simulated signature is the peer's identifier
	return p.ID, nil
}
