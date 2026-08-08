package genai

import (
	"fmt"
	"sync"
)

// Manager tracks running agents and allows toggling them at runtime.
type Manager struct {
	mu         sync.Mutex
	GenAI      *GenAIService
	Predictive *PredictiveAgent
	Vision     *VisionAgent
	Document   *DocumentAgent
}

var GlobalManager = &Manager{}

func (m *Manager) Register(genaiSvc *GenAIService, pred *PredictiveAgent, vis *VisionAgent, doc *DocumentAgent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GenAI = genaiSvc
	m.Predictive = pred
	m.Vision = vis
	m.Document = doc
}

func (m *Manager) ToggleAgent(name string, enable bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch name {
	case "genai", "classification":
		if m.GenAI == nil {
			return fmt.Errorf("agent not registered: %s", name)
		}
		if enable {
			m.GenAI.Start()
		} else {
			m.GenAI.Stop()
		}
	case "predictive":
		if m.Predictive == nil {
			return fmt.Errorf("agent not registered: %s", name)
		}
		if enable {
			m.Predictive.Start()
		} else {
			m.Predictive.Stop()
		}
	case "vision":
		if m.Vision == nil {
			return fmt.Errorf("agent not registered: %s", name)
		}
		if enable {
			m.Vision.Start()
		} else {
			m.Vision.Stop()
		}
	case "document":
		if m.Document == nil {
			return fmt.Errorf("agent not registered: %s", name)
		}
		if enable {
			m.Document.Start()
		} else {
			m.Document.Stop()
		}
	default:
		return fmt.Errorf("unknown agent: %s", name)
	}
	return nil
}

func (m *Manager) SetModels(mp *ModelProvider) {
	// Update global provider used by agents.
	SetGlobalModelProvider(mp)
	if m == nil || mp == nil {
		return
	}
	if m.Vision != nil {
		m.Vision.Models = mp
	}
	if m.Document != nil {
		m.Document.Models = mp
	}
}
