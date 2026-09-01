package genai

import (
	"testing"
	"time"

	"inventory-chain/internal/fabricclient"
)

func TestPredictiveSchedulesAudit(t *testing.T) {
	driver := newMockDriver()
	driver.put(fabricclient.Asset{
		AssetID:        "asset-pred-1",
		DeptID:         "IT",
		Name:           "CriticalRouter",
		Category:       "network",
		Qty:            1,
		Threshold:      1,
		PriorityTier:   "P1",
		LifecycleState: LifecycleActive,
	})

	pa := NewPredictive(driver, 50*time.Millisecond)
	pa.Start()

	deadline := time.Now().Add(5 * time.Second)
	var found bool
	for time.Now().Before(deadline) {
		driver.mu.Lock()
		found = len(driver.scheduledAudits) > 0
		driver.mu.Unlock()
		if found {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	pa.Stop()

	if !found {
		t.Fatalf("expected predictive agent to schedule an audit; none found")
	}
}
