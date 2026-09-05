package api

import (
	"encoding/json"
	"net/http"

	"inventory-chain/internal/genai"
)

// Admin endpoints for runtime agent control and model swapping.
func (h *Handlers) AgentControl(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Agent       string            `json:"agent,omitempty"`
		Enabled     *bool             `json:"enabled,omitempty"`
		Enable      *bool             `json:"enable,omitempty"`
		Models      map[string]string `json:"models,omitempty"`
		LLMProvider string            `json:"llmProvider,omitempty"`
		OCRProvider string            `json:"ocrProvider,omitempty"`
		LlmProvider string            `json:"llm_provider,omitempty"`
		OcrProvider string            `json:"ocr_provider,omitempty"`
		Providers   map[string]string `json:"providers,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var enabled *bool
	if req.Enabled != nil {
		enabled = req.Enabled
	} else if req.Enable != nil {
		enabled = req.Enable
	}
	if enabled != nil && req.Agent != "" {
		if err := genai.GlobalManager.ToggleAgent(req.Agent, *enabled); err != nil {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	models := req.Models
	if models == nil && req.Providers != nil {
		models = req.Providers
	}
	if models == nil {
		if req.OCRProvider != "" || req.LLMProvider != "" || req.OcrProvider != "" || req.LlmProvider != "" {
			models = map[string]string{"ocr": req.OCRProvider, "llm": req.LLMProvider}
		}
	}
	if models != nil {
		ocr := models["ocr"]
		llm := models["llm"]
		if ocr == "" {
			ocr = req.OCRProvider
		}
		if llm == "" {
			llm = req.LLMProvider
		}
		if ocr == "" {
			ocr = req.OcrProvider
		}
		if llm == "" {
			llm = req.LlmProvider
		}
		mp := genai.NewModelProviderFromConfig(ocr, llm, models)
		genai.GlobalManager.SetModels(mp)
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetAgentControl returns the current state of runtime agents and active providers.
func (h *Handlers) GetAgentControl(w http.ResponseWriter, r *http.Request) {
	mp := genai.GetGlobalModelProvider()
	llmName := "dummy"
	ocrName := "dummy"
	if mp != nil {
		if mp.LLMName != "" {
			llmName = mp.LLMName
		}
		if mp.OCRName != "" {
			ocrName = mp.OCRName
		}
	}
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"llmProvider": llmName,
		"ocrProvider": ocrName,
		"agents": map[string]bool{
			"genai":      true,
			"predictive": true,
			"vision":     true,
			"document":   true,
		},
	})
}
