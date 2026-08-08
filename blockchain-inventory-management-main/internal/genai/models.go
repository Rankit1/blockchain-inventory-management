package genai

// OCRModel extracts structured fields from documents/images.
type OCRModel interface {
	ExtractWarrantyAndAMC(doc []byte) (warranty string, amc string, err error)
}

// LLMModel provides lightweight text enrichment for classification notes.
type LLMModel interface {
	SummarizeDocument(text string) (summary string, err error)
}

// ModelProvider groups available models for agents.
type ModelProvider struct {
	OCR OCRModel
	LLM LLMModel
}

// DefaultModelProvider returns stub implementations suitable for testing.
func DefaultModelProvider() *ModelProvider {
	return &ModelProvider{OCR: &DummyOCR{}, LLM: &DummyLLM{}}
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
