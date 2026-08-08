package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type AssetContract struct {
	contractapi.Contract
}

type Asset struct {
	DocType          string    `json:"docType"` // Used for CouchDB indexing
	AssetID          string    `json:"assetId"`
	DeptID           string    `json:"deptId"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Qty              int       `json:"qty"`
	Threshold        int       `json:"threshold"`
	PriorityTier     string    `json:"priorityTier,omitempty"`
	CriticalityScore float64   `json:"criticalityScore,omitempty"`
	LifecycleState   string    `json:"lifecycleState,omitempty"`
	LastAuditDate    string    `json:"lastAuditDate,omitempty"`
	UtilizationRate  float64   `json:"utilizationRate,omitempty"`
	WarrantyExpiry   string    `json:"warrantyExpiry,omitempty"`
	AMCExpiry        string    `json:"amcExpiry,omitempty"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type HistoryEntry struct {
	TxID      string    `json:"txId"`
	Value     *Asset    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// IssueAsset registers a new asset in the world state.
func (c *AssetContract) IssueAsset(ctx contractapi.TransactionContextInterface, assetID, deptID, name, category string, qty, threshold int) error {
	if qty < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	if threshold < 0 {
		return fmt.Errorf("threshold cannot be negative")
	}

	// Check if already exists
	existingBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %w", err)
	}
	if existingBytes != nil {
		return fmt.Errorf("asset %s already exists", assetID)
	}

	txTime, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get tx timestamp: %w", err)
	}

	asset := Asset{
		DocType:        "asset",
		AssetID:        assetID,
		DeptID:         deptID,
		Name:           name,
		Category:       category,
		Qty:            qty,
		Threshold:      threshold,
		LifecycleState: LifecycleActive,
		UpdatedAt:      time.Unix(txTime.Seconds, int64(txTime.Nanos)).UTC(),
	}

	// Auto-classify on creation using category-based default criterion scores.
	score := computeCriticalityScore(defaultScores(category))
	asset.PriorityTier = deriveTier(score)
	asset.CriticalityScore = round2(score)

	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, assetBytes)
}

// ConsumeStock reduces the quantity of stock for an asset.
// Returns a JSON string of format {"newQty": int, "replenishTriggered": bool}.
func (c *AssetContract) ConsumeStock(ctx contractapi.TransactionContextInterface, assetID string, qty int, purpose string) (string, error) {
	if qty <= 0 {
		return "", fmt.Errorf("consume quantity must be greater than zero")
	}

	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %w", err)
	}
	if assetBytes == nil {
		return "", fmt.Errorf("asset %s not found", assetID)
	}

	var asset Asset
	if err := json.Unmarshal(assetBytes, &asset); err != nil {
		return "", err
	}

	if asset.Qty < qty {
		return "", fmt.Errorf("insufficient stock: available %d, requested %d", asset.Qty, qty)
	}

	asset.Qty -= qty
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	asset.UpdatedAt = time.Unix(txTime.Seconds, int64(txTime.Nanos)).UTC()

	updatedBytes, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	if err := ctx.GetStub().PutState(assetID, updatedBytes); err != nil {
		return "", err
	}

	replenishTriggered := asset.Qty < reorderPoint(asset)

	// If triggered, emit a REPLENISH_REQUEST event
	if replenishTriggered {
		eventPayload := map[string]interface{}{
			"assetId": assetID,
			"qty":     asset.Threshold * 2,
			"urgency": "HIGH",
			"deptId":  asset.DeptID,
		}
		eventBytes, _ := json.Marshal(eventPayload)
		_ = ctx.GetStub().SetEvent("REPLENISH_REQUEST", eventBytes)
	}

	respMap := map[string]interface{}{
		"newQty":             asset.Qty,
		"replenishTriggered": replenishTriggered,
	}
	respBytes, _ := json.Marshal(respMap)
	return string(respBytes), nil
}

// TransferAsset moves a specified quantity of an asset to another department.
func (c *AssetContract) TransferAsset(ctx contractapi.TransactionContextInterface, fromDept, toDept, assetID string, qty int) error {
	if fromDept == toDept {
		return fmt.Errorf("source and destination departments must be different")
	}
	if qty <= 0 {
		return fmt.Errorf("transfer quantity must be greater than zero")
	}

	srcBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return err
	}
	if srcBytes == nil {
		return fmt.Errorf("source asset %s not found", assetID)
	}

	var srcAsset Asset
	if err := json.Unmarshal(srcBytes, &srcAsset); err != nil {
		return err
	}

	if srcAsset.DeptID != fromDept {
		return fmt.Errorf("source asset belongs to %s, expected %s", srcAsset.DeptID, fromDept)
	}

	if srcAsset.Qty < qty {
		return fmt.Errorf("insufficient stock to transfer: available %d, requested %d", srcAsset.Qty, qty)
	}

	// Update source
	srcAsset.Qty -= qty
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	srcAsset.UpdatedAt = time.Unix(txTime.Seconds, int64(txTime.Nanos)).UTC()

	srcUpdatedBytes, _ := json.Marshal(srcAsset)
	if err := ctx.GetStub().PutState(assetID, srcUpdatedBytes); err != nil {
		return err
	}

	// Perform CouchDB Rich Query to check if destination department already has a matching asset
	queryString := fmt.Sprintf(`{"selector":{"docType":"asset","deptId":"%s","name":"%s","category":"%s"}}`, toDept, srcAsset.Name, srcAsset.Category)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return fmt.Errorf("failed to query destination assets: %w", err)
	}
	defer resultsIterator.Close()

	var targetAsset Asset
	hasTarget := false

	if resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err == nil {
			if err := json.Unmarshal(queryResponse.Value, &targetAsset); err == nil {
				hasTarget = true
			}
		}
	}

	if hasTarget {
		targetAsset.Qty += qty
		targetAsset.UpdatedAt = srcAsset.UpdatedAt
		targetBytes, _ := json.Marshal(targetAsset)
		return ctx.GetStub().PutState(targetAsset.AssetID, targetBytes)
	}

	// Create new asset copy in destination department
	newDestAssetID := fmt.Sprintf("asset-%s", ctx.GetStub().GetTxID()[:8]) // unique ID prefix
	newDestAsset := Asset{
		DocType:   "asset",
		AssetID:   newDestAssetID,
		DeptID:    toDept,
		Name:      srcAsset.Name,
		Category:  srcAsset.Category,
		Qty:       qty,
		Threshold: srcAsset.Threshold,
		UpdatedAt: srcAsset.UpdatedAt,
	}
	newDestBytes, _ := json.Marshal(newDestAsset)
	return ctx.GetStub().PutState(newDestAssetID, newDestBytes)
}

// RequestReplenishment emits/registers a replenishment state.
func (c *AssetContract) RequestReplenishment(ctx contractapi.TransactionContextInterface, assetID string, qty int, urgency string) error {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return err
	}
	if assetBytes == nil {
		return fmt.Errorf("asset %s not found", assetID)
	}

	txID := ctx.GetStub().GetTxID()
	key := fmt.Sprintf("replenish:%s", txID)

	payload := map[string]interface{}{
		"docType": "replenish_request",
		"assetId": assetID,
		"qty":     qty,
		"urgency": urgency,
	}
	payloadBytes, _ := json.Marshal(payload)
	return ctx.GetStub().PutState(key, payloadBytes)
}

// GetAssetHistory returns transaction history records for a specific key.
func (c *AssetContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, assetID string) ([]HistoryEntry, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(assetID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var history []HistoryEntry
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if !response.IsDelete {
			if err := json.Unmarshal(response.Value, &asset); err != nil {
				return nil, err
			}
		}

		entry := HistoryEntry{
			TxID:      response.TxId,
			Timestamp: time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).UTC(),
			IsDelete:  response.IsDelete,
		}
		if !response.IsDelete {
			entry.Value = &asset
		}

		history = append(history, entry)
	}

	return history, nil
}

// AnchorHash links an off-chain hash to an asset.
func (c *AssetContract) AnchorHash(ctx contractapi.TransactionContextInterface, assetID, sha256Hash string) error {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return err
	}
	if assetBytes == nil {
		return fmt.Errorf("asset %s not found", assetID)
	}

	txID := ctx.GetStub().GetTxID()
	key := fmt.Sprintf("anchor:%s:%s", assetID, txID)

	payload := map[string]interface{}{
		"docType":    "anchor_hash",
		"assetId":    assetID,
		"sha256Hash": sha256Hash,
	}
	payloadBytes, _ := json.Marshal(payload)
	return ctx.GetStub().PutState(key, payloadBytes)
}

// ReadAsset retrieves a single asset.
func (c *AssetContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, err
	}
	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s not found", assetID)
	}

	var asset Asset
	if err := json.Unmarshal(assetBytes, &asset); err != nil {
		return nil, err
	}

	return &asset, nil
}

// QueryAssetsByDept filters assets by department ID using CouchDB.
func (c *AssetContract) QueryAssetsByDept(ctx contractapi.TransactionContextInterface, deptID string) ([]*Asset, error) {
	var queryString string
	if deptID == "" {
		queryString = `{"selector":{"docType":"asset"}}`
	} else {
		queryString = fmt.Sprintf(`{"selector":{"docType":"asset","deptId":"%s"}}`, deptID)
	}
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if err := json.Unmarshal(queryResponse.Value, &asset); err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// ClassifyPriority computes and persists the priority tier and criticality score for an asset.
// Returns JSON string {"priorityTier": "P1", "criticalityScore": 4.5}.
func (c *AssetContract) ClassifyPriority(ctx contractapi.TransactionContextInterface, assetID string, businessCriticality, replacementCost, replacementLeadTime, safetyComplianceImpact, redundancyAvailability int) (string, error) {
	scores := PriorityScores{
		BusinessCriticality:    businessCriticality,
		ReplacementCost:        replacementCost,
		ReplacementLeadTime:    replacementLeadTime,
		SafetyComplianceImpact: safetyComplianceImpact,
		RedundancyAvailability: redundancyAvailability,
	}
	if !scores.valid() {
		return "", fmt.Errorf("all criterion scores must be between 1 and 5")
	}

	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %w", err)
	}
	if assetBytes == nil {
		return "", fmt.Errorf("asset %s not found", assetID)
	}

	var asset Asset
	if err := json.Unmarshal(assetBytes, &asset); err != nil {
		return "", err
	}

	score := round2(computeCriticalityScore(scores))
	tier := deriveTier(score)

	asset.PriorityTier = tier
	asset.CriticalityScore = score
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	asset.UpdatedAt = time.Unix(txTime.Seconds, int64(txTime.Nanos)).UTC()

	updatedBytes, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}
	if err := ctx.GetStub().PutState(assetID, updatedBytes); err != nil {
		return "", err
	}

	respBytes, _ := json.Marshal(map[string]interface{}{
		"priorityTier":     tier,
		"criticalityScore": score,
	})
	return string(respBytes), nil
}

// UpdatePriorityTier applies a manager-approved manual tier override with justification.
func (c *AssetContract) UpdatePriorityTier(ctx contractapi.TransactionContextInterface, assetID, tier, reason string) error {
	if tier != "P1" && tier != "P2" && tier != "P3" {
		return fmt.Errorf("invalid priority tier: %s", tier)
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("a justification is required for manual priority override")
	}

	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %w", err)
	}
	if assetBytes == nil {
		return fmt.Errorf("asset %s not found", assetID)
	}

	var asset Asset
	if err := json.Unmarshal(assetBytes, &asset); err != nil {
		return err
	}

	asset.PriorityTier = tier
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	asset.UpdatedAt = time.Unix(txTime.Seconds, int64(txTime.Nanos)).UTC()

	updatedBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(assetID, updatedBytes)
}

// ScheduleAudit records a planned audit for an asset.
func (c *AssetContract) ScheduleAudit(ctx contractapi.TransactionContextInterface, assetID, auditDate, scope string) error {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %w", err)
	}
	if assetBytes == nil {
		return fmt.Errorf("asset %s not found", assetID)
	}
	if auditDate == "" {
		return fmt.Errorf("audit date is required")
	}

	txID := ctx.GetStub().GetTxID()
	key := fmt.Sprintf("audit:%s:%s", assetID, txID)
	payload := map[string]interface{}{
		"docType":   "audit_schedule",
		"assetId":   assetID,
		"auditDate": auditDate,
		"scope":     scope,
	}
	payloadBytes, _ := json.Marshal(payload)
	return ctx.GetStub().PutState(key, payloadBytes)
}

// RecordAuditResult records the outcome of an audit and updates lastAuditDate.
func (c *AssetContract) RecordAuditResult(ctx contractapi.TransactionContextInterface, assetID, auditDate, result, notes string) error {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %w", err)
	}
	if assetBytes == nil {
		return fmt.Errorf("asset %s not found", assetID)
	}
	if auditDate == "" {
		return fmt.Errorf("audit date is required")
	}
	if result == "" {
		return fmt.Errorf("audit result is required")
	}

	var asset Asset
	if err := json.Unmarshal(assetBytes, &asset); err != nil {
		return err
	}

	asset.LastAuditDate = auditDate
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	asset.UpdatedAt = time.Unix(txTime.Seconds, int64(txTime.Nanos)).UTC()

	updatedBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(assetID, updatedBytes)
}

// RetireAsset marks an asset as RETIRED with a recorded reason (dual-signature enforced upstream).
func (c *AssetContract) RetireAsset(ctx contractapi.TransactionContextInterface, assetID, reason string) error {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %w", err)
	}
	if assetBytes == nil {
		return fmt.Errorf("asset %s not found", assetID)
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("a retirement reason is required")
	}

	var asset Asset
	if err := json.Unmarshal(assetBytes, &asset); err != nil {
		return err
	}
	if asset.LifecycleState == "RETIRED" {
		return fmt.Errorf("asset %s is already retired", assetID)
	}

	asset.LifecycleState = "RETIRED"
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	asset.UpdatedAt = time.Unix(txTime.Seconds, int64(txTime.Nanos)).UTC()

	updatedBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(assetID, updatedBytes)
}

// QueryAssetsByPriority filters assets by priority tier using CouchDB.
func (c *AssetContract) QueryAssetsByPriority(ctx contractapi.TransactionContextInterface, tier string) ([]*Asset, error) {
	if tier != "P1" && tier != "P2" && tier != "P3" {
		return nil, fmt.Errorf("invalid priority tier: %s", tier)
	}
	queryString := fmt.Sprintf(`{"selector":{"docType":"asset","priorityTier":"%s"}}`, tier)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if err := json.Unmarshal(queryResponse.Value, &asset); err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// PriorityScores holds the five 1-5 criterion ratings used for classification.
type PriorityScores struct {
	BusinessCriticality    int `json:"businessCriticality"`
	ReplacementCost        int `json:"replacementCost"`
	ReplacementLeadTime    int `json:"replacementLeadTime"`
	SafetyComplianceImpact int `json:"safetyComplianceImpact"`
	RedundancyAvailability int `json:"redundancyAvailability"`
}

func (s PriorityScores) valid() bool {
	return s.BusinessCriticality >= 1 && s.BusinessCriticality <= 5 &&
		s.ReplacementCost >= 1 && s.ReplacementCost <= 5 &&
		s.ReplacementLeadTime >= 1 && s.ReplacementLeadTime <= 5 &&
		s.SafetyComplianceImpact >= 1 && s.SafetyComplianceImpact <= 5 &&
		s.RedundancyAvailability >= 1 && s.RedundancyAvailability <= 5
}

func computeCriticalityScore(s PriorityScores) float64 {
	return 0.30*float64(s.BusinessCriticality) +
		0.20*float64(s.ReplacementCost) +
		0.20*float64(s.ReplacementLeadTime) +
		0.15*float64(s.SafetyComplianceImpact) +
		0.15*float64(s.RedundancyAvailability)
}

func deriveTier(score float64) string {
	switch {
	case score >= 4.0:
		return "P1"
	case score >= 2.5:
		return "P2"
	default:
		return "P3"
	}
}

func defaultScores(category string) PriorityScores {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "server", "servers", "network", "network device", "network devices", "storage", "secure storage", "infrastructure":
		return PriorityScores{BusinessCriticality: 5, ReplacementCost: 5, ReplacementLeadTime: 4, SafetyComplianceImpact: 3, RedundancyAvailability: 2}
	case "diagnostic", "medical", "diagnostic equipment", "medical equipment":
		return PriorityScores{BusinessCriticality: 5, ReplacementCost: 4, ReplacementLeadTime: 4, SafetyComplianceImpact: 4, RedundancyAvailability: 2}
	case "laptop", "computer", "desktop", "workstation", "electronics":
		return PriorityScores{BusinessCriticality: 3, ReplacementCost: 3, ReplacementLeadTime: 2, SafetyComplianceImpact: 2, RedundancyAvailability: 3}
	default:
		return PriorityScores{BusinessCriticality: 2, ReplacementCost: 2, ReplacementLeadTime: 2, SafetyComplianceImpact: 2, RedundancyAvailability: 3}
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// reorderPoint returns the tier-aware stock level that triggers replenishment.
func reorderPoint(a Asset) int {
	switch a.PriorityTier {
	case "P1":
		return 2 * a.Threshold
	case "P2":
		return a.Threshold + a.Threshold/2
	default:
		return a.Threshold
	}
}

// Lifecycle state constants for assets.
const (
	LifecycleActive = "ACTIVE"
)
