package api

import (
	"math"
	"sort"
	"strings"
	"time"

	"inventory-chain/internal/fabricclient"
	"inventory-chain/internal/genai"
)

// ClientAsset matches the structure of fabricclient.Asset.
type ClientAsset struct {
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

// ClientHistoryEntry matches the structure of fabricclient.HistoryEntry.
type ClientHistoryEntry struct {
	TxID      string       `json:"txId"`
	Value     *ClientAsset `json:"value"`
	Timestamp time.Time    `json:"timestamp"`
	IsDelete  bool         `json:"isDelete"`
}

// toClientAssetFabric converts a fabricclient.Asset to a ClientAsset.
func toClientAssetFabric(a fabricclient.Asset) *ClientAsset {
	return &ClientAsset{
		AssetID:          a.AssetID,
		DeptID:           a.DeptID,
		Name:             a.Name,
		Category:         a.Category,
		Qty:              a.Qty,
		Threshold:        a.Threshold,
		PriorityTier:     a.PriorityTier,
		CriticalityScore: a.CriticalityScore,
		LifecycleState:   a.LifecycleState,
		LastAuditDate:    a.LastAuditDate,
		UtilizationRate:  a.UtilizationRate,
		WarrantyExpiry:   a.WarrantyExpiry,
		AMCExpiry:        a.AMCExpiry,
		UpdatedAt:        a.UpdatedAt,
	}
}

// BlockchainClient defines standard operations for the web server handlers.
type BlockchainClient interface {
	IssueAsset(deptID, name, category string, qty, threshold int) (txID string, assetID string, timestamp time.Time, err error)
	ConsumeStock(deptID, assetID string, qty int, purpose string) (txID string, newQty int, replenishTriggered bool, err error)
	TransferAsset(fromDept, toDept, assetID string, qty int) (txID string, fromQty int, toQty int, err error)
	GetAssetHistory(assetID string) ([]ClientHistoryEntry, error)
	ReadAsset(assetID string) (*ClientAsset, error)
	QueryAssetsByDept(deptID string) ([]*ClientAsset, error)
	RequestReplenishment(assetID string, qty int, urgency string) (string, error)
	GetLedgerBlocks() ([]interface{}, error)
	VerifyLedger() (bool, *int, error)

	// GenAI-augmented asset prioritization & automation
	ClassifyPriority(assetID string, scores genai.PriorityScores) (txID string, tier string, score float64, err error)
	UpdatePriorityTier(assetID, tier, reason string) (txID string, err error)
	ScheduleAudit(assetID, auditDate, scope string) (txID string, err error)
	RecordAuditResult(assetID, auditDate, result, notes string) (txID string, err error)
	RetireAsset(assetID, reason string) (txID string, err error)
	AssistantQuery(userQuery string) (string, error)
	QueryAssetsByPriority(tier string) ([]*ClientAsset, error)
	ComputeUtilization(assetID string) (float64, error)
}

// FabricAdapter wraps the real Fabric Gateway SDK client.
type FabricAdapter struct {
	client *fabricclient.FabricClient
}

func NewFabricAdapter(c *fabricclient.FabricClient) *FabricAdapter {
	return &FabricAdapter{client: c}
}

func (fa *FabricAdapter) IssueAsset(deptID, name, category string, qty, threshold int) (string, string, time.Time, error) {
	return fa.client.IssueAsset(deptID, name, category, qty, threshold)
}

func (fa *FabricAdapter) ConsumeStock(deptID, assetID string, qty int, purpose string) (string, int, bool, error) {
	return fa.client.ConsumeStock(deptID, assetID, qty, purpose)
}

func (fa *FabricAdapter) TransferAsset(fromDept, toDept, assetID string, qty int) (string, int, int, error) {
	return fa.client.TransferAsset(fromDept, toDept, assetID, qty)
}

func (fa *FabricAdapter) GetAssetHistory(assetID string) ([]ClientHistoryEntry, error) {
	history, err := fa.client.GetAssetHistory(assetID)
	if err != nil {
		return nil, err
	}
	var res []ClientHistoryEntry
	for _, h := range history {
		var val *ClientAsset
		if h.Value != nil {
			val = toClientAssetFabric(*h.Value)
		}
		res = append(res, ClientHistoryEntry{
			TxID:      h.TxID,
			Value:     val,
			Timestamp: h.Timestamp,
			IsDelete:  h.IsDelete,
		})
	}
	return res, nil
}

func (fa *FabricAdapter) ReadAsset(assetID string) (*ClientAsset, error) {
	asset, err := fa.client.ReadAsset(assetID)
	if err != nil {
		return nil, err
	}
	return toClientAssetFabric(*asset), nil
}

func (fa *FabricAdapter) QueryAssetsByDept(deptID string) ([]*ClientAsset, error) {
	assets, err := fa.client.QueryAssetsByDept(deptID)
	if err != nil {
		return nil, err
	}
	var res []*ClientAsset
	for _, a := range assets {
		res = append(res, toClientAssetFabric(*a))
	}
	return res, nil
}

func (fa *FabricAdapter) RequestReplenishment(assetID string, qty int, urgency string) (string, error) {
	return fa.client.RequestReplenishment(assetID, qty, urgency)
}

func (fa *FabricAdapter) GetLedgerBlocks() ([]interface{}, error) {
	return []interface{}{}, nil
}

func (fa *FabricAdapter) VerifyLedger() (bool, *int, error) {
	return true, nil, nil
}

func (fa *FabricAdapter) ClassifyPriority(assetID string, scores genai.PriorityScores) (string, string, float64, error) {
	return fa.client.ClassifyPriority(assetID, scores.BusinessCriticality, scores.ReplacementCost, scores.ReplacementLeadTime, scores.SafetyComplianceImpact, scores.RedundancyAvailability)
}

func (fa *FabricAdapter) UpdatePriorityTier(assetID, tier, reason string) (string, error) {
	return fa.client.UpdatePriorityTier(assetID, tier, reason)
}

func (fa *FabricAdapter) ScheduleAudit(assetID, auditDate, scope string) (string, error) {
	return fa.client.ScheduleAudit(assetID, auditDate, scope)
}

func (fa *FabricAdapter) RecordAuditResult(assetID, auditDate, result, notes string) (string, error) {
	return fa.client.RecordAuditResult(assetID, auditDate, result, notes)
}

func (fa *FabricAdapter) RetireAsset(assetID, reason string) (string, error) {
	return fa.client.RetireAsset(assetID, reason)
}

func (fa *FabricAdapter) AssistantQuery(userQuery string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(userQuery))
	if normalized == "" {
		return "Please provide a query about assets, audits, or priorities.", nil
	}
	if strings.Contains(normalized, "classify") || strings.Contains(normalized, "priority") {
		return "To classify an asset, call POST /api/assets/classify with the five priority criteria. For manual adjustments, use POST /api/assets/update-priority.", nil
	}
	if strings.Contains(normalized, "audit") {
		return "Audit operations are supported via POST /api/assets/schedule-audit and POST /api/assets/record-audit. Use an asset ID, date, scope, and result details.", nil
	}
	if strings.Contains(normalized, "retire") {
		return "Asset retirement is done with POST /api/assets/retire and must include a retirement reason for auditability.", nil
	}
	return "This conversational assistant is a stub. Please use the available asset endpoints for classification, auditing, and retirement.", nil
}

func (fa *FabricAdapter) QueryAssetsByPriority(tier string) ([]*ClientAsset, error) {
	assets, err := fa.client.QueryAssetsByPriority(tier)
	if err != nil {
		return nil, err
	}
	var res []*ClientAsset
	for _, a := range assets {
		res = append(res, toClientAssetFabric(*a))
	}
	return res, nil
}

func (fa *FabricAdapter) ComputeUtilization(assetID string) (float64, error) {
	history, err := fa.client.GetAssetHistory(assetID)
	if err != nil {
		return 0, err
	}
	return utilizationFromHistory(history), nil
}

// utilizationFromHistory computes utilization as the ratio of quantity consumed
// versus the initial issued quantity, diffing consecutive asset snapshots.
func utilizationFromHistory(history []fabricclient.HistoryEntry) float64 {
	type snap struct {
		t time.Time
		q int
	}
	var snaps []snap
	for _, h := range history {
		if h.Value == nil {
			continue
		}
		snaps = append(snaps, snap{t: h.Timestamp, q: h.Value.Qty})
	}
	if len(snaps) < 2 {
		return 0
	}
	sort.SliceStable(snaps, func(i, j int) bool { return snaps[i].t.Before(snaps[j].t) })
	initialQty := snaps[0].q
	if initialQty <= 0 {
		return 0
	}
	var consumed float64
	for i := 1; i < len(snaps); i++ {
		if snaps[i].q < snaps[i-1].q {
			consumed += float64(snaps[i-1].q - snaps[i].q)
		}
	}
	u := consumed / float64(initialQty)
	if u > 1 {
		u = 1
	}
	return math.Round(u*100) / 100
}
