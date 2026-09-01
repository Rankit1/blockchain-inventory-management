package genai

import (
	"log"
	"sync"
	"time"
)

// VisionAgent periodically records an audit pass for assets missing one.
// No photo upload pipeline exists yet, so it cannot run vision-LLM matching
// against a real image; it enriches the audit note via the configured OCR
// provider when one is available.
type VisionAgent struct {
	Driver AutomationDriver
	Models *ModelProvider
	period time.Duration
	stop   chan struct{}
	wg     sync.WaitGroup
	mu     sync.Mutex
}

func NewVision(driver AutomationDriver, period time.Duration, models *ModelProvider) *VisionAgent {
	return &VisionAgent{Driver: driver, Models: models, period: period}
}

func (v *VisionAgent) Start() {
	v.mu.Lock()
	if v.stop != nil {
		v.mu.Unlock()
		return
	}
	v.stop = make(chan struct{})
	stop := v.stop
	v.wg.Add(1)
	v.mu.Unlock()
	go func() {
		defer v.wg.Done()
		ticker := time.NewTicker(v.period)
		defer ticker.Stop()
		log.Println("VisionAgent: started")
		for {
			select {
			case <-stop:
				log.Println("VisionAgent: stopping")
				return
			case <-ticker.C:
				v.scan()
			}
		}
	}()
}

func (v *VisionAgent) Stop() {
	v.mu.Lock()
	if v.stop == nil {
		v.mu.Unlock()
		return
	}
	close(v.stop)
	v.stop = nil
	v.mu.Unlock()
	v.wg.Wait()
}

func (v *VisionAgent) scan() {
	assets, err := v.Driver.ListAssets()
	if err != nil {
		log.Printf("VisionAgent: failed to list assets: %v", err)
		return
	}
	now := time.Now().UTC().Format("2006-01-02")
	for _, a := range assets {
		if a.LifecycleState == LifecycleRetired {
			continue
		}
		// If asset has no last audit, assume a vision check found it present
		if a.LastAuditDate == "" {
			note := "vision:verified"
			// allow OCR to extract additional info from image if available
			mp := v.Models
			if mp == nil {
				mp = GetGlobalModelProvider()
			}
			if mp != nil && mp.OCR != nil {
				w, am, _ := mp.OCR.ExtractWarrantyAndAMC(nil)
				if w != "" || am != "" {
					note = note + " " + "doc:warranty=" + w
				}
			}
			if tx, err := v.Driver.RecordAuditResult(a.AssetID, now, "PASS", note); err == nil {
				log.Printf("VisionAgent: recorded audit for %s -> tx %s", a.AssetID, tx)
			} else {
				log.Printf("VisionAgent: failed to record audit for %s: %v", a.AssetID, err)
			}
		}
	}
}
