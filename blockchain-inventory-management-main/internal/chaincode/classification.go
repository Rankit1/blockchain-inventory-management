package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"inventory-chain/internal/ledger"
	"inventory-chain/internal/worldstate"
	"github.com/google/uuid"
)

// Priority tiers.
const (
	P1Tier = "P1"
	P2Tier = "P2"
	P3Tier = "P3"
)

// Weighted criteria (decimals sum to 1.0).
const (
	WeightBusinessCriticality    = 0.30
	WeightReplacementCost        = 0.20
	WeightReplacementLeadTime    = 0.20
	WeightSafetyComplianceImpact = 0.15
	WeightRedundancyAvailability = 0.15
)

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

// ComputeCriticalityScore calculates the weighted 1-5 criticality score.
func ComputeCriticalityScore(s PriorityScores) float64 {
	return WeightBusinessCriticality*float64(s.BusinessCriticality) +
		WeightReplacementCost*float64(s.ReplacementCost) +
		WeightReplacementLeadTime*float64(s.ReplacementLeadTime) +
		WeightSafetyComplianceImpact*float64(s.SafetyComplianceImpact) +
		WeightRedundancyAvailability*float64(s.RedundancyAvailability)
}

// DeriveTier maps a criticality score to a priority tier.
func DeriveTier(score float64) string {
	switch {
	case score >= 4.0:
		return P1Tier
	case score >= 2.5:
		return P2Tier
	default:
		return P3Tier
	}
}

// ValidTier returns true if the tier is one of P1/P2/P3.
func ValidTier(tier string) bool {
	return tier == P1Tier || tier == P2Tier || tier == P3Tier
}

// ReorderPoint returns the tier-aware stock level that triggers replenishment,
// replacing the fixed threshold with a priority-based dynamic reorder point.
func ReorderPoint(a worldstate.Asset) int {
	switch a.PriorityTier {
	case P1Tier:
		return 2 * a.Threshold
	case P2Tier:
		return a.Threshold + a.Threshold/2
	default:
		return a.Threshold
	}
}

// DefaultScores returns heuristic criterion ratings based on asset category
// so classification can run immediately on asset creation.
func DefaultScores(category string) PriorityScores {
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

// ClassifyPriority computes and persists the priority tier and criticality score for an asset.
func ClassifyPriority(store *worldstate.Store, assetID string, scores PriorityScores) (worldstate.Asset, ledger.Transaction, error) {
	if !scores.valid() {
		return worldstate.Asset{}, ledger.Transaction{}, errors.New("all criterion scores must be between 1 and 5")
	}

	asset, err := store.Get(assetID)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, fmt.Errorf("asset not found: %w", err)
	}

	score := round2(ComputeCriticalityScore(scores))
	tier := DeriveTier(score)

	asset.PriorityTier = tier
	asset.CriticalityScore = score
	asset.UpdatedAt = time.Now().UTC()

	payloadMap := map[string]interface{}{
		"assetId":          assetID,
		"priorityTier":     tier,
		"criticalityScore": score,
		"scores":           scores,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "PRIORITY_CLASSIFY",
		DeptID:    asset.DeptID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	return asset, tx, nil
}

// UpdatePriorityTier applies a manager-approved manual tier override with justification.
func UpdatePriorityTier(store *worldstate.Store, assetID, tier, reason string) (worldstate.Asset, ledger.Transaction, error) {
	if !ValidTier(tier) {
		return worldstate.Asset{}, ledger.Transaction{}, fmt.Errorf("invalid priority tier: %s", tier)
	}
	if strings.TrimSpace(reason) == "" {
		return worldstate.Asset{}, ledger.Transaction{}, errors.New("a justification is required for manual priority override")
	}

	asset, err := store.Get(assetID)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, fmt.Errorf("asset not found: %w", err)
	}

	asset.PriorityTier = tier
	asset.UpdatedAt = time.Now().UTC()

	payloadMap := map[string]interface{}{
		"assetId":      assetID,
		"priorityTier": tier,
		"reason":       reason,
		"deptId":       asset.DeptID,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "PRIORITY_UPDATE",
		DeptID:    asset.DeptID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}

	return asset, tx, nil
}

// ScheduleAudit records a planned audit for an asset.
func ScheduleAudit(store *worldstate.Store, assetID, auditDate, scope string) (ledger.Transaction, error) {
	asset, err := store.Get(assetID)
	if err != nil {
		return ledger.Transaction{}, fmt.Errorf("asset not found: %w", err)
	}
	if strings.TrimSpace(auditDate) == "" {
		return ledger.Transaction{}, errors.New("audit date is required")
	}

	payloadMap := map[string]interface{}{
		"assetId":   assetID,
		"auditDate": auditDate,
		"scope":     scope,
		"deptId":    asset.DeptID,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return ledger.Transaction{}, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "AUDIT_SCHEDULE",
		DeptID:    asset.DeptID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
	return tx, nil
}

// RecordAuditResult records the outcome of an audit and updates lastAuditDate.
func RecordAuditResult(store *worldstate.Store, assetID, auditDate, result, notes string) (worldstate.Asset, ledger.Transaction, error) {
	asset, err := store.Get(assetID)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, fmt.Errorf("asset not found: %w", err)
	}
	if strings.TrimSpace(result) == "" {
		return worldstate.Asset{}, ledger.Transaction{}, errors.New("audit result is required")
	}
	if strings.TrimSpace(auditDate) == "" {
		return worldstate.Asset{}, ledger.Transaction{}, errors.New("audit date is required")
	}

	asset.LastAuditDate = auditDate
	asset.UpdatedAt = time.Now().UTC()

	payloadMap := map[string]interface{}{
		"assetId":   assetID,
		"auditDate": auditDate,
		"result":    result,
		"notes":     notes,
		"deptId":    asset.DeptID,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "AUDIT_RESULT",
		DeptID:    asset.DeptID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
	return asset, tx, nil
}

// RetireAsset marks an asset as RETIRED with a recorded reason (dual-signature enforced upstream).
func RetireAsset(store *worldstate.Store, assetID, reason string) (worldstate.Asset, ledger.Transaction, error) {
	asset, err := store.Get(assetID)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, fmt.Errorf("asset not found: %w", err)
	}
	if strings.TrimSpace(reason) == "" {
		return worldstate.Asset{}, ledger.Transaction{}, errors.New("a retirement reason is required")
	}
	if asset.LifecycleState == worldstate.LifecycleRetired {
		return worldstate.Asset{}, ledger.Transaction{}, errors.New("asset is already retired")
	}

	asset.LifecycleState = worldstate.LifecycleRetired
	asset.UpdatedAt = time.Now().UTC()

	payloadMap := map[string]interface{}{
		"assetId": assetID,
		"reason":  reason,
		"deptId":  asset.DeptID,
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return worldstate.Asset{}, ledger.Transaction{}, err
	}

	tx := ledger.Transaction{
		TxID:      uuid.New().String(),
		Type:      "RETIRE",
		DeptID:    asset.DeptID,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
	return asset, tx, nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
