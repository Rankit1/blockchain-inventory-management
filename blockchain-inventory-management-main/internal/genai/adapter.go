package genai

import (
	"inventory-chain/internal/chaincode"
	"inventory-chain/internal/fabricclient"
	"inventory-chain/internal/network"
	"inventory-chain/internal/worldstate"
)

// AutomationDriver abstracts away the environment (simulation or Fabric)
// so GenAI agents can list assets and submit classify/schedule actions.
type AutomationDriver interface {
	ListAssets() ([]worldstate.Asset, error)
	ClassifyPriority(assetID string, scores chaincode.PriorityScores) (txID string, tier string, score float64, err error)
	ScheduleAudit(assetID, auditDate, scope string) (txID string, err error)
	UpdatePriorityTier(assetID, tier, reason string) (txID string, err error)
	RecordAuditResult(assetID, auditDate, result, notes string) (txID string, err error)
}

// SimulationDriver implements AutomationDriver using the in-memory Network.
type SimulationDriver struct {
	Net *network.Network
}

func NewSimulationDriver(n *network.Network) *SimulationDriver {
	return &SimulationDriver{Net: n}
}

func (s *SimulationDriver) ListAssets() ([]worldstate.Asset, error) {
	return s.Net.Store.List("")
}

func (s *SimulationDriver) ClassifyPriority(assetID string, scores chaincode.PriorityScores) (string, string, float64, error) {
	asset, tx, err := chaincode.ClassifyPriority(s.Net.Store, assetID, scores)
	if err != nil {
		return "", "", 0, err
	}
	if _, err := s.Net.ProposeAndCommit(tx); err != nil {
		return "", "", 0, err
	}
	return tx.TxID, asset.PriorityTier, asset.CriticalityScore, nil
}

func (s *SimulationDriver) ScheduleAudit(assetID, auditDate, scope string) (string, error) {
	tx, err := chaincode.ScheduleAudit(s.Net.Store, assetID, auditDate, scope)
	if err != nil {
		return "", err
	}
	if _, err := s.Net.ProposeAndCommit(tx); err != nil {
		return "", err
	}
	return tx.TxID, nil
}

func (s *SimulationDriver) UpdatePriorityTier(assetID, tier, reason string) (string, error) {
	_, tx, err := chaincode.UpdatePriorityTier(s.Net.Store, assetID, tier, reason)
	if err != nil {
		return "", err
	}
	if _, err := s.Net.ProposeAndCommit(tx); err != nil {
		return "", err
	}
	return tx.TxID, nil
}

func (s *SimulationDriver) RecordAuditResult(assetID, auditDate, result, notes string) (string, error) {
	_, tx, err := chaincode.RecordAuditResult(s.Net.Store, assetID, auditDate, result, notes)
	if err != nil {
		return "", err
	}
	if _, err := s.Net.ProposeAndCommit(tx); err != nil {
		return "", err
	}
	return tx.TxID, nil
}

// FabricDriver implements AutomationDriver using the Fabric client.
type FabricDriver struct {
	FC *fabricclient.FabricClient
}

func NewFabricDriver(fc *fabricclient.FabricClient) *FabricDriver {
	return &FabricDriver{FC: fc}
}

func (f *FabricDriver) ListAssets() ([]worldstate.Asset, error) {
	// Query fabric for all assets and convert to worldstate.Asset shape.
	fa, err := f.FC.QueryAssetsByDept("")
	if err != nil {
		return nil, err
	}
	var res []worldstate.Asset
	for _, a := range fa {
		res = append(res, worldstate.Asset{
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
		})
	}
	return res, nil
}

func (f *FabricDriver) ClassifyPriority(assetID string, scores chaincode.PriorityScores) (string, string, float64, error) {
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
