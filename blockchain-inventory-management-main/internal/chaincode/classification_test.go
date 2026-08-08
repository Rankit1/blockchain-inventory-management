package chaincode_test

import (
	"os"
	"path/filepath"
	"testing"

	"inventory-chain/internal/chaincode"
	"inventory-chain/internal/worldstate"
	bolt "go.etcd.io/bbolt"
)

func TestComputeCriticalityScore(t *testing.T) {
	scores := chaincode.PriorityScores{
		BusinessCriticality:    5,
		ReplacementCost:        5,
		ReplacementLeadTime:    4,
		SafetyComplianceImpact: 3,
		RedundancyAvailability: 2,
	}
	got := chaincode.ComputeCriticalityScore(scores)
	want := 0.30*5 + 0.20*5 + 0.20*4 + 0.15*3 + 0.15*2
	if got != want {
		t.Errorf("expected score %.2f, got %.2f", want, got)
	}
}

func TestDeriveTier(t *testing.T) {
	cases := []struct {
		score float64
		tier  string
	}{
		{4.0, "P1"},
		{4.5, "P1"},
		{3.9, "P2"},
		{2.5, "P2"},
		{2.4, "P3"},
		{1.0, "P3"},
	}
	for _, c := range cases {
		if got := chaincode.DeriveTier(c.score); got != c.tier {
			t.Errorf("DeriveTier(%.1f) = %s, want %s", c.score, got, c.tier)
		}
	}
}

func TestDefaultScores(t *testing.T) {
	server := chaincode.DefaultScores("server")
	if server.BusinessCriticality != 5 {
		t.Errorf("expected server criticality 5, got %d", server.BusinessCriticality)
	}
	if chaincode.DeriveTier(chaincode.ComputeCriticalityScore(server)) != "P1" {
		t.Error("expected server category to default to P1")
	}
}

func newTestStore(t *testing.T) *worldstate.Store {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "classification-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	db, err := bolt.Open(filepath.Join(tmpDir, "test.db"), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	store, err := worldstate.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestClassifyPriority(t *testing.T) {
	store := newTestStore(t)

	asset := worldstate.Asset{AssetID: "a1", DeptID: "Lab", Name: "Beaker", Category: "Glassware", Qty: 10, Threshold: 3}
	if err := store.Put(asset); err != nil {
		t.Fatal(err)
	}

	updated, tx, err := chaincode.ClassifyPriority(store, "a1", chaincode.PriorityScores{
		BusinessCriticality: 5, ReplacementCost: 5, ReplacementLeadTime: 5,
		SafetyComplianceImpact: 5, RedundancyAvailability: 5,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.PriorityTier != "P1" {
		t.Errorf("expected P1, got %s", updated.PriorityTier)
	}
	if updated.CriticalityScore != 5.0 {
		t.Errorf("expected score 5.0, got %.2f", updated.CriticalityScore)
	}
	if tx.Type != "PRIORITY_CLASSIFY" {
		t.Errorf("expected PRIORITY_CLASSIFY tx type, got %s", tx.Type)
	}

	if _, _, err := chaincode.ClassifyPriority(store, "missing", chaincode.PriorityScores{
		BusinessCriticality: 5, ReplacementCost: 5, ReplacementLeadTime: 5,
		SafetyComplianceImpact: 5, RedundancyAvailability: 5,
	}); err == nil {
		t.Error("expected error for missing asset")
	}

	if _, _, err := chaincode.ClassifyPriority(store, "a1", chaincode.PriorityScores{
		BusinessCriticality: 6, ReplacementCost: 5, ReplacementLeadTime: 5,
		SafetyComplianceImpact: 5, RedundancyAvailability: 5,
	}); err == nil {
		t.Error("expected error for out-of-range score")
	}
}

func TestUpdatePriorityTier(t *testing.T) {
	store := newTestStore(t)
	asset := worldstate.Asset{AssetID: "a1", DeptID: "IT", Name: "Switch", Category: "Network", Qty: 1, Threshold: 1}
	if err := store.Put(asset); err != nil {
		t.Fatal(err)
	}

	updated, tx, err := chaincode.UpdatePriorityTier(store, "a1", "P1", "Critical during campus upgrade window")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.PriorityTier != "P1" {
		t.Errorf("expected P1, got %s", updated.PriorityTier)
	}
	if tx.Type != "PRIORITY_UPDATE" {
		t.Errorf("expected PRIORITY_UPDATE tx type, got %s", tx.Type)
	}

	if _, _, err := chaincode.UpdatePriorityTier(store, "a1", "P5", "reason"); err == nil {
		t.Error("expected error for invalid tier")
	}
	if _, _, err := chaincode.UpdatePriorityTier(store, "a1", "P2", ""); err == nil {
		t.Error("expected error for missing justification")
	}
}

func TestScheduleAudit(t *testing.T) {
	store := newTestStore(t)
	asset := worldstate.Asset{AssetID: "a1", DeptID: "Lab", Name: "Microscope", Category: "Diagnostic", Qty: 1, Threshold: 1}
	if err := store.Put(asset); err != nil {
		t.Fatal(err)
	}

	tx, err := chaincode.ScheduleAudit(store, "a1", "2026-08-15", "Physical count + calibration check")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.Type != "AUDIT_SCHEDULE" {
		t.Errorf("expected AUDIT_SCHEDULE tx type, got %s", tx.Type)
	}
	if _, err := chaincode.ScheduleAudit(store, "a1", "", ""); err == nil {
		t.Error("expected error for missing audit date")
	}
}

func TestRecordAuditResult(t *testing.T) {
	store := newTestStore(t)
	asset := worldstate.Asset{AssetID: "a1", DeptID: "Lab", Name: "Microscope", Category: "Diagnostic", Qty: 1, Threshold: 1}
	if err := store.Put(asset); err != nil {
		t.Fatal(err)
	}

	updated, tx, err := chaincode.RecordAuditResult(store, "a1", "2026-08-15", "PASS", "All calibration marks valid")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.LastAuditDate != "2026-08-15" {
		t.Errorf("expected lastAuditDate 2026-08-15, got %s", updated.LastAuditDate)
	}
	if tx.Type != "AUDIT_RESULT" {
		t.Errorf("expected AUDIT_RESULT tx type, got %s", tx.Type)
	}
}

func TestRetireAsset(t *testing.T) {
	store := newTestStore(t)
	asset := worldstate.Asset{AssetID: "a1", DeptID: "Store", Name: "Chair", Category: "Furniture", Qty: 4, Threshold: 1}
	if err := store.Put(asset); err != nil {
		t.Fatal(err)
	}

	updated, tx, err := chaincode.RetireAsset(store, "a1", "Depreciated beyond useful life")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.LifecycleState != worldstate.LifecycleRetired {
		t.Errorf("expected lifecycleState RETIRED, got %s", updated.LifecycleState)
	}
	if tx.Type != "RETIRE" {
		t.Errorf("expected RETIRE tx type, got %s", tx.Type)
	}

	// Persist the state change the same way the orderer commits it before re-checking.
	if err := store.Put(updated); err != nil {
		t.Fatal(err)
	}

	if _, _, err := chaincode.RetireAsset(store, "a1", "again"); err == nil {
		t.Error("expected error for already retired asset")
	}
	if _, _, err := chaincode.RetireAsset(store, "a1", ""); err == nil {
		t.Error("expected error for missing reason")
	}
}
