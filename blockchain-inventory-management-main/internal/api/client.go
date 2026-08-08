package api

import (
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"

	"inventory-chain/internal/chaincode"
	"inventory-chain/internal/fabricclient"
	"inventory-chain/internal/network"
	"inventory-chain/internal/worldstate"
)

// ClientAsset matches the structure of worldstate.Asset and fabricclient.Asset.
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

// toClientAsset converts a worldstate.Asset (simulation) to a ClientAsset.
func toClientAsset(a worldstate.Asset) *ClientAsset {
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
	ClassifyPriority(assetID string, scores chaincode.PriorityScores) (txID string, tier string, score float64, err error)
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

func (fa *FabricAdapter) ClassifyPriority(assetID string, scores chaincode.PriorityScores) (string, string, float64, error) {
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
	return "assistant functionality is not available in Fabric mode yet", nil
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

// SimulationAdapter wraps the in-process local prototype network.
type SimulationAdapter struct {
	net *network.Network
}

func NewSimulationAdapter(n *network.Network) *SimulationAdapter {
	return &SimulationAdapter{net: n}
}

func (sa *SimulationAdapter) IssueAsset(deptID, name, category string, qty, threshold int) (string, string, time.Time, error) {
	asset, tx, err := chaincode.IssueAsset(sa.net.Store, deptID, name, category, qty, threshold)
	if err != nil {
		return "", "", time.Time{}, err
	}
	block, err := sa.net.ProposeAndCommit(tx)
	if err != nil {
		return "", "", time.Time{}, err
	}
	var timestamp time.Time
	for _, t := range block.Transactions {
		if t.TxID == tx.TxID {
			timestamp = t.Timestamp
		}
	}
	return tx.TxID, asset.AssetID, timestamp, nil
}

func (sa *SimulationAdapter) ConsumeStock(deptID, assetID string, qty int, purpose string) (string, int, bool, error) {
	asset, tx, replenishTriggered, err := chaincode.ConsumeStock(sa.net.Store, deptID, assetID, qty, purpose)
	if err != nil {
		return "", 0, false, err
	}
	_, err = sa.net.ProposeAndCommit(tx)
	if err != nil {
		return "", 0, false, err
	}
	return tx.TxID, asset.Qty, replenishTriggered, nil
}

func (sa *SimulationAdapter) TransferAsset(fromDept, toDept, assetID string, qty int) (string, int, int, error) {
	srcAsset, targetAsset, tx, err := chaincode.TransferAsset(sa.net.Store, fromDept, toDept, assetID, qty)
	if err != nil {
		return "", 0, 0, err
	}
	_, err = sa.net.ProposeAndCommit(tx)
	if err != nil {
		return "", 0, 0, err
	}
	return tx.TxID, srcAsset.Qty, targetAsset.Qty, nil
}

func (sa *SimulationAdapter) GetAssetHistory(assetID string) ([]ClientHistoryEntry, error) {
	history, err := sa.net.Ledger.GetAssetHistory(assetID)
	if err != nil {
		return nil, err
	}
	var res []ClientHistoryEntry
	for _, tx := range history {
		isDelete := tx.Type == "DELETE"

		var value *ClientAsset
		if tx.Type == "ISSUE" {
			var asset worldstate.Asset
			if err := json.Unmarshal(tx.Payload, &asset); err == nil {
				value = toClientAsset(asset)
			}
		} else {
			if asset, err := sa.net.Store.Get(assetID); err == nil {
				value = toClientAsset(asset)
				if tx.Type == "CONSUME" {
					var payloadMap map[string]interface{}
					if err := json.Unmarshal(tx.Payload, &payloadMap); err == nil {
						if newQtyVal, ok := payloadMap["newQty"].(float64); ok {
							value.Qty = int(newQtyVal)
						}
					}
				} else if tx.Type == "TRANSFER" {
					var payloadMap map[string]interface{}
					if err := json.Unmarshal(tx.Payload, &payloadMap); err == nil {
						if fromQtyVal, ok := payloadMap["fromQty"].(float64); ok {
							value.Qty = int(fromQtyVal)
						}
					}
				}
			}
		}

		res = append(res, ClientHistoryEntry{
			TxID:      tx.TxID,
			Value:     value,
			Timestamp: tx.Timestamp,
			IsDelete:  isDelete,
		})
	}
	return res, nil
}

func (sa *SimulationAdapter) ReadAsset(assetID string) (*ClientAsset, error) {
	asset, err := sa.net.Store.Get(assetID)
	if err != nil {
		return nil, err
	}
	return toClientAsset(asset), nil
}

func (sa *SimulationAdapter) QueryAssetsByDept(deptID string) ([]*ClientAsset, error) {
	assets, err := sa.net.Store.List(deptID)
	if err != nil {
		return nil, err
	}
	var res []*ClientAsset
	for _, a := range assets {
		res = append(res, toClientAsset(a))
	}
	return res, nil
}

func (sa *SimulationAdapter) RequestReplenishment(assetID string, qty int, urgency string) (string, error) {
	tx, err := chaincode.RequestReplenishment(sa.net.Store, assetID, qty, urgency)
	if err != nil {
		return "", err
	}
	_, err = sa.net.ProposeAndCommit(tx)
	if err != nil {
		return "", err
	}
	return tx.TxID, nil
}

func (sa *SimulationAdapter) GetLedgerBlocks() ([]interface{}, error) {
	blocks, err := sa.net.Ledger.GetBlocks()
	if err != nil {
		return nil, err
	}
	var res []interface{}
	for _, b := range blocks {
		res = append(res, b)
	}
	return res, nil
}

func (sa *SimulationAdapter) VerifyLedger() (bool, *int, error) {
	valid, brokenIndex, err := sa.net.Ledger.VerifyChain()
	if err != nil {
		return false, nil, err
	}
	var broken *int
	if brokenIndex != -1 {
		idx := brokenIndex
		broken = &idx
	}
	return valid, broken, nil
}

func (sa *SimulationAdapter) ClassifyPriority(assetID string, scores chaincode.PriorityScores) (string, string, float64, error) {
	asset, tx, err := chaincode.ClassifyPriority(sa.net.Store, assetID, scores)
	if err != nil {
		return "", "", 0, err
	}
	_, err = sa.net.ProposeAndCommit(tx)
	if err != nil {
		return "", "", 0, err
	}
	return tx.TxID, asset.PriorityTier, asset.CriticalityScore, nil
}

func (sa *SimulationAdapter) UpdatePriorityTier(assetID, tier, reason string) (string, error) {
	_, tx, err := chaincode.UpdatePriorityTier(sa.net.Store, assetID, tier, reason)
	if err != nil {
		return "", err
	}
	_, err = sa.net.ProposeAndCommit(tx)
	if err != nil {
		return "", err
	}
	return tx.TxID, nil
}

func (sa *SimulationAdapter) ScheduleAudit(assetID, auditDate, scope string) (string, error) {
	tx, err := chaincode.ScheduleAudit(sa.net.Store, assetID, auditDate, scope)
	if err != nil {
		return "", err
	}
	_, err = sa.net.ProposeAndCommit(tx)
	if err != nil {
		return "", err
	}
	return tx.TxID, nil
}

func (sa *SimulationAdapter) RecordAuditResult(assetID, auditDate, result, notes string) (string, error) {
	_, tx, err := chaincode.RecordAuditResult(sa.net.Store, assetID, auditDate, result, notes)
	if err != nil {
		return "", err
	}
	_, err = sa.net.ProposeAndCommit(tx)
	if err != nil {
		return "", err
	}
	return tx.TxID, nil
}

func (sa *SimulationAdapter) RetireAsset(assetID, reason string) (string, error) {
	_, tx, err := chaincode.RetireAsset(sa.net.Store, assetID, reason)
	if err != nil {
		return "", err
	}
	_, err = sa.net.ProposeAndCommit(tx)
	if err != nil {
		return "", err
	}
	return tx.TxID, nil
}

func (sa *SimulationAdapter) AssistantQuery(userQuery string) (string, error) {
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

func (sa *SimulationAdapter) QueryAssetsByPriority(tier string) ([]*ClientAsset, error) {
	assets, err := sa.net.Store.List("")
	if err != nil {
		return nil, err
	}
	var res []*ClientAsset
	for _, a := range assets {
		if a.PriorityTier == tier {
			res = append(res, toClientAsset(a))
		}
	}
	return res, nil
}

func (sa *SimulationAdapter) ComputeUtilization(assetID string) (float64, error) {
	history, err := sa.net.Ledger.GetAssetHistory(assetID)
	if err != nil {
		return 0, err
	}
	if len(history) == 0 {
		return 0, nil
	}
	var initialQty int
	var consumed float64
	for _, tx := range history {
		switch tx.Type {
		case "ISSUE":
			var asset worldstate.Asset
			if err := json.Unmarshal(tx.Payload, &asset); err == nil {
				initialQty = asset.Qty
			}
		case "CONSUME":
			var m map[string]interface{}
			if err := json.Unmarshal(tx.Payload, &m); err == nil {
				if q, ok := m["qty"].(float64); ok {
					consumed += q
				}
			}
		}
	}
	if initialQty <= 0 {
		return 0, nil
	}
	u := consumed / float64(initialQty)
	if u > 1 {
		u = 1
	}
	return math.Round(u*100) / 100, nil
}
