package genai

import (
	"log"
	"sync"
	"time"

	"inventory-chain/internal/chaincode"
	"inventory-chain/internal/worldstate"
)

// GenAIService is a lightweight automation layer that scans assets via an
// AutomationDriver and applies AI-driven decisions.
type GenAIService struct {
	Driver AutomationDriver
	period time.Duration
	stop   chan struct{}
	wg     sync.WaitGroup
	mu     sync.Mutex
}

// New returns a configured GenAIService. period controls how often the
// scanner runs.
func New(driver AutomationDriver, period time.Duration) *GenAIService {
	return &GenAIService{Driver: driver, period: period}
}

// Start begins background processing.
func (g *GenAIService) Start() {
	g.mu.Lock()
	if g.stop != nil {
		g.mu.Unlock()
		return
	}
	g.stop = make(chan struct{})
	g.wg.Add(1)
	g.mu.Unlock()
	go func() {
		defer g.wg.Done()
		ticker := time.NewTicker(g.period)
		defer ticker.Stop()
		log.Println("GenAIService: started classification scanner")
		for {
			select {
			case <-g.stop:
				log.Println("GenAIService: stopping")
				return
			case <-ticker.C:
				g.scanAndClassify()
			}
		}
	}()
}

// Stop signals the service to stop and waits for background work to finish.
func (g *GenAIService) Stop() {
	g.mu.Lock()
	if g.stop == nil {
		g.mu.Unlock()
		return
	}
	close(g.stop)
	g.stop = nil
	g.mu.Unlock()
	g.wg.Wait()
}

// scanAndClassify walks all assets and triggers classification for assets
// that are missing a priority tier or whose classification is stale.
func (g *GenAIService) scanAndClassify() {
	assets, err := g.Driver.ListAssets()
	if err != nil {
		log.Printf("GenAIService: failed to list assets: %v", err)
		return
	}
	now := time.Now().UTC()
	for _, a := range assets {
		// Skip retired assets
		if a.LifecycleState == worldstate.LifecycleRetired {
			continue
		}
		needs := false
		if a.PriorityTier == "" {
			needs = true
		} else if now.Sub(a.UpdatedAt) > 24*time.Hour {
			// re-evaluate stale classifications (simple heuristic)
			needs = true
		}
		if needs {
			g.classify(a)
		}
	}
}

// classify runs the lightweight (stub) GenAI classifier via the AutomationDriver.
func (g *GenAIService) classify(a worldstate.Asset) {
	// For now we use the heuristic defaultScores from chaincode as a stand-in
	// for an AI model. This keeps behavior deterministic and testable.
	scores := chaincode.DefaultScores(a.Category)

	txID, tier, score, err := g.Driver.ClassifyPriority(a.AssetID, chaincode.PriorityScores{
		BusinessCriticality:    scores.BusinessCriticality,
		ReplacementCost:        scores.ReplacementCost,
		ReplacementLeadTime:    scores.ReplacementLeadTime,
		SafetyComplianceImpact: scores.SafetyComplianceImpact,
		RedundancyAvailability: scores.RedundancyAvailability,
	})
	if err != nil {
		log.Printf("GenAIService: classification failed for %s: %v", a.AssetID, err)
		return
	}
	log.Printf("GenAIService: classified asset %s -> tx %s (tier=%s score=%.2f)", a.AssetID, txID, tier, score)
}
