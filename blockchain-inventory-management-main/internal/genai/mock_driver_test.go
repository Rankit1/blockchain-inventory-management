package genai

import (
	"fmt"
	"sync"
	"time"

	"inventory-chain/internal/fabricclient"
)

// mockDriver is a minimal in-memory AutomationDriver used to exercise agent
// behavior without a live Fabric network.
type mockDriver struct {
	mu     sync.Mutex
	assets map[string]fabricclient.Asset
	txSeq  int

	// auditResults records every RecordAuditResult call, in order.
	auditResults []recordedAudit
	// scheduledAudits records every ScheduleAudit call, in order.
	scheduledAudits []scheduledAudit
}

type recordedAudit struct {
	assetID, auditDate, result, notes string
}

type scheduledAudit struct {
	assetID, auditDate, scope string
}

func newMockDriver() *mockDriver {
	return &mockDriver{assets: make(map[string]fabricclient.Asset)}
}

func (m *mockDriver) put(a fabricclient.Asset) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.assets[a.AssetID] = a
}

func (m *mockDriver) get(assetID string) (fabricclient.Asset, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.assets[assetID]
	return a, ok
}

func (m *mockDriver) nextTxID() string {
	m.txSeq++
	return fmt.Sprintf("tx-%d", m.txSeq)
}

func (m *mockDriver) ListAssets() ([]fabricclient.Asset, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]fabricclient.Asset, 0, len(m.assets))
	for _, a := range m.assets {
		res = append(res, a)
	}
	return res, nil
}

func (m *mockDriver) ClassifyPriority(assetID string, scores PriorityScores) (string, string, float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.assets[assetID]
	if !ok {
		return "", "", 0, fmt.Errorf("asset not found: %s", assetID)
	}
	score := 0.30*float64(scores.BusinessCriticality) + 0.20*float64(scores.ReplacementCost) +
		0.20*float64(scores.ReplacementLeadTime) + 0.15*float64(scores.SafetyComplianceImpact) +
		0.15*float64(scores.RedundancyAvailability)
	tier := "P3"
	if score >= 4.0 {
		tier = "P1"
	} else if score >= 2.5 {
		tier = "P2"
	}
	a.PriorityTier = tier
	a.CriticalityScore = score
	a.UpdatedAt = time.Now().UTC()
	m.assets[assetID] = a
	return m.nextTxID(), tier, score, nil
}

func (m *mockDriver) ScheduleAudit(assetID, auditDate, scope string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.assets[assetID]; !ok {
		return "", fmt.Errorf("asset not found: %s", assetID)
	}
	m.scheduledAudits = append(m.scheduledAudits, scheduledAudit{assetID, auditDate, scope})
	return m.nextTxID(), nil
}

func (m *mockDriver) UpdatePriorityTier(assetID, tier, reason string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.assets[assetID]
	if !ok {
		return "", fmt.Errorf("asset not found: %s", assetID)
	}
	a.PriorityTier = tier
	m.assets[assetID] = a
	return m.nextTxID(), nil
}

func (m *mockDriver) RecordAuditResult(assetID, auditDate, result, notes string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.assets[assetID]
	if !ok {
		return "", fmt.Errorf("asset not found: %s", assetID)
	}
	a.LastAuditDate = auditDate
	m.assets[assetID] = a
	m.auditResults = append(m.auditResults, recordedAudit{assetID, auditDate, result, notes})
	return m.nextTxID(), nil
}
