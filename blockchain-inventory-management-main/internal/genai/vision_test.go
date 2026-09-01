package genai

import (
	"testing"
	"time"

	"inventory-chain/internal/fabricclient"
)

func TestVisionRecordsAudit(t *testing.T) {
	driver := newMockDriver()
	driver.put(fabricclient.Asset{
		AssetID: "asset-vis-1", DeptID: "d1", Name: "VisAsset", Category: "GEN",
		Qty: 1, UpdatedAt: time.Now().UTC(), LifecycleState: LifecycleActive,
	})

	models := DefaultModelProvider()
	v := NewVision(driver, 50*time.Millisecond, models)
	v.Start()
	time.Sleep(150 * time.Millisecond)
	v.Stop()

	updated, ok := driver.get("asset-vis-1")
	if !ok {
		t.Fatalf("asset disappeared")
	}
	if updated.LastAuditDate == "" {
		t.Fatalf("expected LastAuditDate to be set by VisionAgent")
	}
}
