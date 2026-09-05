package genai

// OCRModel extracts structured fields from documents/images.
type OCRModel interface {
	ExtractWarrantyAndAMC(doc []byte) (warranty string, amc string, err error)
}

// LLMModel provides lightweight text enrichment for classification notes and
// asset priority scoring.
type LLMModel interface {
	SummarizeDocument(text string) (summary string, err error)
	// ScorePriority asks the model to rate an asset 1-5 on each of the five
	// priority criteria (business criticality, replacement cost/lead time,
	// safety/compliance impact, redundancy) based on its name and category.
	ScorePriority(assetName, category string) (PriorityScores, error)
}

// ModelProvider groups available models for agents.
type ModelProvider struct {
	OCRName string
	LLMName string
	OCR     OCRModel
	LLM     LLMModel
}

// DefaultModelProvider returns stub implementations suitable for testing.
func DefaultModelProvider() *ModelProvider {
	return &ModelProvider{OCRName: "dummy", LLMName: "dummy", OCR: &DummyOCR{}, LLM: &DummyLLM{}}
}

var globalModelProvider = DefaultModelProvider()

func SetGlobalModelProvider(mp *ModelProvider) {
	if mp == nil {
		return
	}
	globalModelProvider = mp
}

func GetGlobalModelProvider() *ModelProvider {
	return globalModelProvider
}
