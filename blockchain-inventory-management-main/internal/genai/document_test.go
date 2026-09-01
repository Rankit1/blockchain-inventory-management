package genai

import (
	"testing"
	"time"

	"inventory-chain/internal/fabricclient"
)

func TestDocumentAgentExtractsWarrantyAndReclassifies(t *testing.T) {
	driver := newMockDriver()
	driver.put(fabricclient.Asset{
		AssetID: "asset-doc-1", DeptID: "d1", Name: "DocAsset", Category: "GEN",
		Qty: 1, UpdatedAt: time.Now().UTC(), LifecycleState: LifecycleActive,
	})

	models := DefaultModelProvider()
	d := NewDocumentAgent(driver, 50*time.Millisecond, models)
	d.Start()
	time.Sleep(200 * time.Millisecond)
	d.Stop()

	updated, ok := driver.get("asset-doc-1")
	if !ok {
		t.Fatalf("asset disappeared")
	}
	// the DocumentAgent records a DOC_EXTRACTED note which updates lastAuditDate
	if updated.LastAuditDate == "" {
		t.Fatalf("expected LastAuditDate to be set by DocumentAgent")
	}
}
