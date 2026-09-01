package genai

import "fmt"

// DummyOCR is a simple deterministic OCR stub.
type DummyOCR struct{}

func (d *DummyOCR) ExtractWarrantyAndAMC(doc []byte) (string, string, error) {
	// Pretend we extracted a warranty 1 year from now and a fixed AMC expiry
	return "2027-01-01", "2026-12-31", nil
}

// DummyLLM is a trivial summarizer used when no real LLM provider is
// configured (no OPENAI_API_KEY/MISTRAL_API_KEY/GEMINI_API_KEY set).
type DummyLLM struct{}

func (d *DummyLLM) SummarizeDocument(text string) (string, error) {
	return fmt.Sprintf("summary:%s", text), nil
}

func (d *DummyLLM) ScorePriority(assetName, category string) (PriorityScores, error) {
	return heuristicScores(category), nil
}
