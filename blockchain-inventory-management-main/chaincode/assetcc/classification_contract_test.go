package main

import (
	"encoding/json"
	"testing"
)

func TestIssueAssetAutoClassify(t *testing.T) {
	contract := new(AssetContract)
	stub := &MockStub{
		state:  make(map[string][]byte),
		events: make(map[string][]byte),
	}
	ctx := &MockContext{stub: stub}

	err := contract.IssueAsset(ctx, "srv1", "IT", "AppServer", "server", 2, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	val, _ := stub.GetState("srv1")
	var asset Asset
	if err := json.Unmarshal(val, &asset); err != nil {
		t.Fatal(err)
	}
	if asset.PriorityTier != "P1" {
		t.Errorf("expected auto-tier P1 for server category, got %s", asset.PriorityTier)
	}
	if asset.LifecycleState != "ACTIVE" {
		t.Errorf("expected lifecycleState ACTIVE, got %s", asset.LifecycleState)
	}
	if asset.CriticalityScore != 4.05 {
		t.Errorf("expected criticality score 4.05, got %.2f", asset.CriticalityScore)
	}
}

func TestClassifyPriority(t *testing.T) {
	contract := new(AssetContract)
	stub := &MockStub{
		state:  make(map[string][]byte),
		events: make(map[string][]byte),
	}
	ctx := &MockContext{stub: stub}

	asset := Asset{DocType: "asset", AssetID: "a1", DeptID: "Lab", Name: "Beaker", Category: "Glassware", Qty: 10, Threshold: 3}
	assetBytes, _ := json.Marshal(asset)
	stub.PutState("a1", assetBytes)

	resp, err := contract.ClassifyPriority(ctx, "a1", 5, 5, 5, 5, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	var res struct {
		PriorityTier    string  `json:"priorityTier"`
		CriticalityScore float64 `json:"criticalityScore"`
	}
	if err := json.Unmarshal([]byte(resp), &res); err != nil {
		t.Fatal(err)
	}
	if res.PriorityTier != "P1" || res.CriticalityScore != 5.0 {
		t.Errorf("unexpected classification result: %+v", res)
	}

	if _, err := contract.ClassifyPriority(ctx, "missing", 5, 5, 5, 5, 5); err == nil {
		t.Error("expected error for missing asset")
	}
	if _, err := contract.ClassifyPriority(ctx, "a1", 6, 5, 5, 5, 5); err == nil {
		t.Error("expected error for out-of-range score")
	}
}

func TestUpdatePriorityTier(t *testing.T) {
	contract := new(AssetContract)
	stub := &MockStub{state: make(map[string][]byte), events: make(map[string][]byte)}
	ctx := &MockContext{stub: stub}

	asset := Asset{DocType: "asset", AssetID: "a1", DeptID: "IT", Name: "Switch", Category: "Network", Qty: 1, Threshold: 1}
	assetBytes, _ := json.Marshal(asset)
	stub.PutState("a1", assetBytes)

	if err := contract.UpdatePriorityTier(ctx, "a1", "P1", "Project window"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	val, _ := stub.GetState("a1")
	var updated Asset
	json.Unmarshal(val, &updated)
	if updated.PriorityTier != "P1" {
		t.Errorf("expected P1, got %s", updated.PriorityTier)
	}

	if err := contract.UpdatePriorityTier(ctx, "a1", "P5", "reason"); err == nil {
		t.Error("expected error for invalid tier")
	}
	if err := contract.UpdatePriorityTier(ctx, "a1", "P2", ""); err == nil {
		t.Error("expected error for missing justification")
	}
}

func TestScheduleAudit(t *testing.T) {
	contract := new(AssetContract)
	stub := &MockStub{state: make(map[string][]byte), events: make(map[string][]byte)}
	ctx := &MockContext{stub: stub}

	asset := Asset{DocType: "asset", AssetID: "a1", DeptID: "Lab", Name: "Microscope", Category: "Diagnostic", Qty: 1, Threshold: 1}
	assetBytes, _ := json.Marshal(asset)
	stub.PutState("a1", assetBytes)

	if err := contract.ScheduleAudit(ctx, "a1", "2026-08-15", "calibration"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	found := false
	for k := range stub.state {
		if len(k) > 6 && k[:6] == "audit:" {
			found = true
		}
	}
	if !found {
		t.Error("expected an audit schedule state record")
	}
	if err := contract.ScheduleAudit(ctx, "a1", "", ""); err == nil {
		t.Error("expected error for missing audit date")
	}
}

func TestRecordAuditResult(t *testing.T) {
	contract := new(AssetContract)
	stub := &MockStub{state: make(map[string][]byte), events: make(map[string][]byte)}
	ctx := &MockContext{stub: stub}

	asset := Asset{DocType: "asset", AssetID: "a1", DeptID: "Lab", Name: "Microscope", Category: "Diagnostic", Qty: 1, Threshold: 1}
	assetBytes, _ := json.Marshal(asset)
	stub.PutState("a1", assetBytes)

	if err := contract.RecordAuditResult(ctx, "a1", "2026-08-15", "PASS", "ok"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	val, _ := stub.GetState("a1")
	var updated Asset
	json.Unmarshal(val, &updated)
	if updated.LastAuditDate != "2026-08-15" {
		t.Errorf("expected lastAuditDate, got %s", updated.LastAuditDate)
	}
}

func TestRetireAsset(t *testing.T) {
	contract := new(AssetContract)
	stub := &MockStub{state: make(map[string][]byte), events: make(map[string][]byte)}
	ctx := &MockContext{stub: stub}

	asset := Asset{DocType: "asset", AssetID: "a1", DeptID: "Store", Name: "Chair", Category: "Furniture", Qty: 4, Threshold: 1}
	assetBytes, _ := json.Marshal(asset)
	stub.PutState("a1", assetBytes)

	if err := contract.RetireAsset(ctx, "a1", "Depreciated"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	val, _ := stub.GetState("a1")
	var updated Asset
	json.Unmarshal(val, &updated)
	if updated.LifecycleState != "RETIRED" {
		t.Errorf("expected RETIRED, got %s", updated.LifecycleState)
	}

	if err := contract.RetireAsset(ctx, "a1", "again"); err == nil {
		t.Error("expected error for already retired asset")
	}
	if err := contract.RetireAsset(ctx, "a1", ""); err == nil {
		t.Error("expected error for missing reason")
	}
}
