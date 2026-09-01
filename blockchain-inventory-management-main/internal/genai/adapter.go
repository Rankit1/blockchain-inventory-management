package genai

import (
	"inventory-chain/internal/fabricclient"
)

// AutomationDriver abstracts how GenAI agents list assets and submit
// classify/schedule actions against the Fabric ledger.
type AutomationDriver interface {
	ListAssets() ([]fabricclient.Asset, error)
	ClassifyPriority(assetID string, scores PriorityScores) (txID string, tier string, score float64, err error)
	ScheduleAudit(assetID, auditDate, scope string) (txID string, err error)
	UpdatePriorityTier(assetID, tier, reason string) (txID string, err error)
	RecordAuditResult(assetID, auditDate, result, notes string) (txID string, err error)
}

// FabricDriver implements AutomationDriver using the real Fabric Gateway client.
type FabricDriver struct {
	FC *fabricclient.FabricClient
}

func NewFabricDriver(fc *fabricclient.FabricClient) *FabricDriver {
	return &FabricDriver{FC: fc}
}

func (f *FabricDriver) ListAssets() ([]fabricclient.Asset, error) {
	fa, err := f.FC.QueryAssetsByDept("")
	if err != nil {
		return nil, err
	}
	res := make([]fabricclient.Asset, 0, len(fa))
	for _, a := range fa {
		res = append(res, *a)
	}
	return res, nil
}

func (f *FabricDriver) ClassifyPriority(assetID string, scores PriorityScores) (string, string, float64, error) {
	return f.FC.ClassifyPriority(assetID, scores.BusinessCriticality, scores.ReplacementCost, scores.ReplacementLeadTime, scores.SafetyComplianceImpact, scores.RedundancyAvailability)
}

func (f *FabricDriver) ScheduleAudit(assetID, auditDate, scope string) (string, error) {
	return f.FC.ScheduleAudit(assetID, auditDate, scope)
}

func (f *FabricDriver) UpdatePriorityTier(assetID, tier, reason string) (string, error) {
	return f.FC.UpdatePriorityTier(assetID, tier, reason)
}

func (f *FabricDriver) RecordAuditResult(assetID, auditDate, result, notes string) (string, error) {
	return f.FC.RecordAuditResult(assetID, auditDate, result, notes)
}
