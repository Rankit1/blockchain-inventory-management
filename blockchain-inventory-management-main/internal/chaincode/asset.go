package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"inventory-chain/internal/ledger"
	"inventory-chain/internal/worldstate"
	"github.com/google/uuid"
)

// Allowed departments
var validDepts = map[string]bool{
	"Lab":   true,
	"Admin": true,
	"Store": true,
	"IT":    true,
}

// IssueAsset validates and simulates creation of a new asset.
func IssueAsset(store *worldstate.Store, deptID, name, category string, qty, threshold int) (worldstate.Asset, ledger.Transaction, error) {
	if !validDepts[deptID] {
		return worldstate.Asset{}, ledger.Transaction{}, fmt.Errorf("invalid department: %s", deptID)
	}
	if qty < 0 {
		return worldstate.Asset{}, ledger.Transaction{}, errors.New("quantity cannot be negative")
	}
	if threshold < 0 {
		return worldstate.Asset{}, ledger.Transaction{}, errors.New("threshold cannot be negative")
	}
	if name == "" {
		return worldstate.Asset{}, ledger.Transaction{}, errors.New("asset name cannot be empty")
	}

	assetID := uuid.New().String()

	// Auto-classify the asset on creation using category-based default criterion scores.
	defaultScores := DefaultScores(category)
	score := round2(ComputeCriticalityScore(defaultScores))

	asset := worldstate.Asset{
		AssetID:          assetID,
		DeptID:           deptID,
		Name:             name,
		Category:         category,
		Qty:              qty,
		Threshold:        threshold,
		PriorityTier:     DeriveTier(score),
		CriticalityScore: score,
		LifecycleState:   worldstate.LifecycleActive,
		UpdatedAt:        time.Now().UTC(),
	}

	payload, err := json.Marshal(asset)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "ISSUE",
		DeptID:    deptID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	return asset, tx, nil
}

// ConsumeStock validates and simulates consumption of an asset.
// If the quantity drops below threshold, replenishTriggered is true.
func ConsumeStock(store *worldstate.Store, deptID, assetID string, qty int, purpose string) (worldstate.Asset, ledger.Transaction, bool, error) {
	if !validDepts[deptID] {
		return worldstate.Asset{}, ledger.Transaction{}, false, fmt.Errorf("invalid department: %s", deptID)
	}
	if qty <= 0 {
		return worldstate.Asset{}, ledger.Transaction{}, false, errors.New("quantity to consume must be greater than zero")
	}

	asset, err := store.Get(assetID)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, false, fmt.Errorf("asset not found: %w", err)
	}

	if asset.DeptID != deptID {
		return worldstate.Asset{}, ledger.Transaction{}, false, fmt.Errorf("asset department mismatch (asset belongs to %s, request by %s)", asset.DeptID, deptID)
	}

	if asset.Qty < qty {
		return worldstate.Asset{}, ledger.Transaction{}, false, fmt.Errorf("insufficient stock: available %d, requested %d", asset.Qty, qty)
	}

	asset.Qty -= qty
	asset.UpdatedAt = time.Now().UTC()

	payloadMap := map[string]interface{}{
		"assetId": assetID,
		"deptId":  deptID,
		"qty":     qty,
		"purpose": purpose,
		"newQty":  asset.Qty,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, false, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "CONSUME",
		DeptID:    deptID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	replenishTriggered := asset.Qty < ReorderPoint(asset)
	return asset, tx, replenishTriggered, nil
}

// TransferAsset moves a specified quantity of an asset from one department to another.
// Decision: We manage department-scoped copies of assets.
// - If the destination department already has an asset with the same name and category, we add the qty to it.
// - Otherwise, we create a new asset in the destination department with the transferred qty.
func TransferAsset(store *worldstate.Store, fromDept, toDept, assetID string, qty int) (worldstate.Asset, *worldstate.Asset, ledger.Transaction, error) {
	if !validDepts[fromDept] || !validDepts[toDept] {
		return worldstate.Asset{}, nil, ledger.Transaction{}, errors.New("invalid from/to department")
	}
	if fromDept == toDept {
		return worldstate.Asset{}, nil, ledger.Transaction{}, errors.New("source and destination departments must be different")
	}
	if qty <= 0 {
		return worldstate.Asset{}, nil, ledger.Transaction{}, errors.New("transfer quantity must be greater than zero")
	}

	srcAsset, err := store.Get(assetID)
	if err != nil {
		return worldstate.Asset{}, nil, ledger.Transaction{}, fmt.Errorf("source asset not found: %w", err)
	}

	if srcAsset.DeptID != fromDept {
		return worldstate.Asset{}, nil, ledger.Transaction{}, fmt.Errorf("source asset department mismatch (expected %s, got %s)", fromDept, srcAsset.DeptID)
	}

	if srcAsset.Qty < qty {
		return worldstate.Asset{}, nil, ledger.Transaction{}, fmt.Errorf("insufficient stock to transfer: available %d, requested %d", srcAsset.Qty, qty)
	}

	// Update source asset
	srcAsset.Qty -= qty
	srcAsset.UpdatedAt = time.Now().UTC()

	// Find or create destination asset
	destAssets, err := store.List(toDept)
	if err != nil {
		return worldstate.Asset{}, nil, ledger.Transaction{}, err
	}

	var targetAsset *worldstate.Asset
	for _, asset := range destAssets {
		if asset.Name == srcAsset.Name && asset.Category == srcAsset.Category {
			targetAsset = &asset
			break
		}
	}

	if targetAsset != nil {
		targetAsset.Qty += qty
		targetAsset.UpdatedAt = time.Now().UTC()
	} else {
		targetAsset = &worldstate.Asset{
			AssetID:   uuid.New().String(),
			DeptID:    toDept,
			Name:      srcAsset.Name,
			Category:  srcAsset.Category,
			Qty:       qty,
			Threshold: srcAsset.Threshold,
			UpdatedAt: time.Now().UTC(),
		}
	}

	payloadMap := map[string]interface{}{
		"assetId":       assetID,
		"fromDept":      fromDept,
		"toDept":        toDept,
		"qty":           qty,
		"targetAssetId": targetAsset.AssetID,
		"fromQty":       srcAsset.Qty,
		"toQty":         targetAsset.Qty,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return worldstate.Asset{}, nil, ledger.Transaction{}, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "TRANSFER",
		DeptID:    fromDept,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	return srcAsset, targetAsset, tx, nil
}

// RequestReplenishment simulates a replenishment request.
func RequestReplenishment(store *worldstate.Store, assetID string, qty int, urgency string) (ledger.Transaction, error) {
	asset, err := store.Get(assetID)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("asset not found: %w", err)
	}

	if qty <= 0 {
		return ledger.Transaction{}, errors.New("replenish quantity must be greater than zero")
	}

	payloadMap := map[string]interface{}{
		"assetId": assetID,
		"qty":     qty,
		"urgency": urgency,
		"deptId":  asset.DeptID,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return ledger.Transaction{}, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "REPLENISH_REQUEST",
		DeptID:    asset.DeptID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	return tx, nil
}

// AnchorHash stores a hash against an asset for off-chain file integrity checks.
func AnchorHash(store *worldstate.Store, assetID string, sha256Hash string) (ledger.Transaction, error) {
	_, err := store.Get(assetID)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("asset not found: %w", err)
	}

	if len(sha256Hash) != 64 {
		return ledger.Transaction{}, errors.New("invalid sha256 hash length")
	}

	payloadMap := map[string]interface{}{
		"assetId":    assetID,
		"sha256Hash": sha256Hash,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return ledger.Transaction{}, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "ANCHOR_HASH",
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	return tx, nil
}
