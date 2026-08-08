package genai

import (
	"fmt"
	"log"
	"sync"
	"time"

	"inventory-chain/internal/chaincode"
	"inventory-chain/internal/worldstate"
)

// DocumentAgent simulates document intelligence that extracts warranty/AMC
// metadata from uploaded invoices and enriches the asset record via an
// audit note and a re-classification.
type DocumentAgent struct {
	Driver AutomationDriver
	Models *ModelProvider
	period time.Duration
	stop   chan struct{}
	wg     sync.WaitGroup
	mu     sync.Mutex
}

func NewDocumentAgent(driver AutomationDriver, period time.Duration, models *ModelProvider) *DocumentAgent {
	return &DocumentAgent{Driver: driver, Models: models, period: period}
}

func (d *DocumentAgent) Start() {
	d.mu.Lock()
	if d.stop != nil {
		d.mu.Unlock()
		return
	}
	d.stop = make(chan struct{})
	d.wg.Add(1)
	d.mu.Unlock()
	go func() {
		defer d.wg.Done()
		ticker := time.NewTicker(d.period)
		defer ticker.Stop()
		log.Println("DocumentAgent: started")
		for {
			select {
			case <-d.stop:
				log.Println("DocumentAgent: stopping")
				return
			case <-ticker.C:
				d.scan()
			}
		}
	}()
}

func (d *DocumentAgent) Stop() {
	d.mu.Lock()
	if d.stop == nil {
		d.mu.Unlock()
		return
	}
	close(d.stop)
	d.stop = nil
	d.mu.Unlock()
	d.wg.Wait()
}

func (d *DocumentAgent) scan() {
	assets, err := d.Driver.ListAssets()
	if err != nil {
		log.Printf("DocumentAgent: failed to list assets: %v", err)
		return
	}
	for _, a := range assets {
		if a.LifecycleState == worldstate.LifecycleRetired {
			continue
		}
		// Process each asset once per agent lifetime after the document extraction
		// has produced a concrete audit signal.
		if a.WarrantyExpiry == "" && a.LastAuditDate == "" {
			// Use OCR model from global provider if available to extract warranty/AMC
			warranty := ""
			amc := ""
			mp := d.Models
			if mp == nil {
				mp = GetGlobalModelProvider()
			}
			if mp != nil && mp.OCR != nil {
				w, am, _ := mp.OCR.ExtractWarrantyAndAMC(nil)
				warranty = w
				amc = am
			}
			if warranty == "" {
				warranty = time.Now().UTC().AddDate(1, 0, 0).Format("2006-01-02")
			}
			note := fmt.Sprintf("doc:warranty=%s amc=%s", warranty, amc)
			if tx, err := d.Driver.RecordAuditResult(a.AssetID, time.Now().UTC().Format("2006-01-02"), "DOC_EXTRACTED", note); err == nil {
				log.Printf("DocumentAgent: attached warranty for %s -> tx %s", a.AssetID, tx)
			} else {
				log.Printf("DocumentAgent: failed to attach warranty for %s: %v", a.AssetID, err)
			}
			// Optionally use LLM to summarize and enrich reclassification notes
			scores := chaincode.DefaultScores(a.Category)
			mp2 := d.Models
			if mp2 == nil {
				mp2 = GetGlobalModelProvider()
			}
			if mp2 != nil && mp2.LLM != nil {
				summary, _ := mp2.LLM.SummarizeDocument(note)
				// Record the summary as an audit result as well
				_, _ = d.Driver.RecordAuditResult(a.AssetID, time.Now().UTC().Format("2006-01-02"), "DOC_SUMMARY", summary)
			}
			if tx2, _, _, err := d.Driver.ClassifyPriority(a.AssetID, scores); err == nil {
				log.Printf("DocumentAgent: reclassified %s -> tx %s", a.AssetID, tx2)
			} else {
				log.Printf("DocumentAgent: failed to reclassify %s: %v", a.AssetID, err)
			}
		}
	}
}
