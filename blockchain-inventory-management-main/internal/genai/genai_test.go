package genai

import (
	"testing"
	"time"

	"inventory-chain/internal/fabricclient"
)

func TestGenAIClassification(t *testing.T) {
	driver := newMockDriver()
	driver.put(fabricclient.Asset{
		AssetID:        "asset-genai-1",
		DeptID:         "Lab",
		Name:           "TestDevice",
		Category:       "server",
		Qty:            1,
		Threshold:      1,
		LifecycleState: LifecycleActive,
	})

	svc := New(driver, 50*time.Millisecond)
	svc.Start()

	deadline := time.Now().Add(5 * time.Second)
	var gotTier string
	for time.Now().Before(deadline) {
		if a, ok := driver.get("asset-genai-1"); ok && a.PriorityTier != "" {
			gotTier = a.PriorityTier
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	svc.Stop()

	if gotTier == "" {
		t.Fatalf("expected asset to be classified, but PriorityTier is empty")
	}
	if gotTier != "P1" && gotTier != "P2" && gotTier != "P3" {
		t.Fatalf("unexpected priority tier: %s", gotTier)
	}
}
