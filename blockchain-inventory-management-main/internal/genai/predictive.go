package genai

import (
	"log"
	"sync"
	"time"

	"inventory-chain/internal/worldstate"
)

// PredictiveAgent runs lightweight predictive maintenance checks and schedules
// audits for high-priority assets that lack recent audits.
type PredictiveAgent struct {
	Driver AutomationDriver
	period time.Duration
	stop   chan struct{}
	wg     sync.WaitGroup
	mu     sync.Mutex
}

func NewPredictive(driver AutomationDriver, period time.Duration) *PredictiveAgent {
	return &PredictiveAgent{Driver: driver, period: period}
}

func (p *PredictiveAgent) Start() {
	p.mu.Lock()
	if p.stop != nil {
		p.mu.Unlock()
		return
	}
	p.stop = make(chan struct{})
	p.wg.Add(1)
	p.mu.Unlock()
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(p.period)
		defer ticker.Stop()
		log.Println("PredictiveAgent: started")
		for {
			select {
			case <-p.stop:
				log.Println("PredictiveAgent: stopping")
				return
			case <-ticker.C:
				p.scan()
			}
		}
	}()
}

func (p *PredictiveAgent) Stop() {
	p.mu.Lock()
	if p.stop == nil {
		p.mu.Unlock()
		return
	}
	close(p.stop)
	p.stop = nil
	p.mu.Unlock()
	p.wg.Wait()
}

func (p *PredictiveAgent) scan() {
	assets, err := p.Driver.ListAssets()
	if err != nil {
		log.Printf("PredictiveAgent: failed to list assets: %v", err)
		return
	}
	for _, a := range assets {
		if a.LifecycleState == worldstate.LifecycleRetired {
			continue
		}
		// Simple heuristic: schedule audit for P1 assets with no lastAuditDate
		if a.PriorityTier == "P1" && a.LastAuditDate == "" {
			p.scheduleAudit(a)
		}
	}
}

func (p *PredictiveAgent) scheduleAudit(a worldstate.Asset) {
	// Create an audit date for tomorrow
	auditDate := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02")
	txID, err := p.Driver.ScheduleAudit(a.AssetID, auditDate, "predictive-maintenance")
	if err != nil {
		log.Printf("PredictiveAgent: failed to schedule audit for %s: %v", a.AssetID, err)
		return
	}
	log.Printf("PredictiveAgent: scheduled audit for %s -> tx %s", a.AssetID, txID)
}
