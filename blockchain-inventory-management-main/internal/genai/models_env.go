package genai

import (
	"os"
)

// NewModelProviderFromEnv inspects environment variables to construct
// appropriate OCR/LLM adapters. Supported values:
// OCR_PROVIDER: "tesseract", "azure", "dummy" (default)
// LLM_PROVIDER: "openai", "mistral", "gemini", "dummy" (default)
func NewModelProviderFromEnv() *ModelProvider {
	ocr := os.Getenv("OCR_PROVIDER")
	llm := os.Getenv("LLM_PROVIDER")

	mp := &ModelProvider{}

	switch ocr {
	case "tesseract":
		mp.OCR = &TesseractOCR{}
	case "azure":
		// expects AZURE_OCR_ENDPOINT and AZURE_OCR_KEY
		endpoint := os.Getenv("AZURE_OCR_ENDPOINT")
		key := os.Getenv("AZURE_OCR_KEY")
		if endpoint != "" && key != "" {
			mp.OCR = NewAzureOCR(endpoint, key)
		}
	default:
		mp.OCR = &DummyOCR{}
	}

	switch llm {
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key != "" {
			mp.LLM = NewOpenAI(key)
		}
	case "mistral":
		key := os.Getenv("MISTRAL_API_KEY")
		if key != "" {
			mp.LLM = NewMistral(key)
		}
	case "gemini":
		key := os.Getenv("GEMINI_API_KEY")
		if key != "" {
			mp.LLM = NewGemini(key)
		}
	default:
		mp.LLM = &DummyLLM{}
	}

	return mp
}

// NewModelProviderFromConfig constructs a ModelProvider from explicit config.
// ocrProvider/llmProvider are provider names ("tesseract", "azure", "openai", "dummy").
// params can include keys like OPENAI_API_KEY, AZURE_OCR_ENDPOINT, AZURE_OCR_KEY.
func NewModelProviderFromConfig(ocrProvider, llmProvider string, params map[string]string) *ModelProvider {
	mp := &ModelProvider{}
	switch ocrProvider {
	case "tesseract":
		mp.OCR = &TesseractOCR{}
	case "azure":
		ep := params["AZURE_OCR_ENDPOINT"]
		key := params["AZURE_OCR_KEY"]
		if ep != "" && key != "" {
			mp.OCR = NewAzureOCR(ep, key)
		} else {
			mp.OCR = &DummyOCR{}
		}
	default:
		mp.OCR = &DummyOCR{}
	}

	switch llmProvider {
	case "openai":
		key := params["OPENAI_API_KEY"]
		if key != "" {
			mp.LLM = NewOpenAI(key)
		} else {
			mp.LLM = &DummyLLM{}
		}
	case "mistral":
		// Placeholder: mistral client requires endpoint/key
		key := params["MISTRAL_API_KEY"]
		if key != "" {
			mp.LLM = NewMistral(key)
		} else {
			mp.LLM = &DummyLLM{}
		}
	case "gemini":
		key := params["GEMINI_API_KEY"]
		if key != "" {
			mp.LLM = NewGemini(key)
		} else {
			mp.LLM = &DummyLLM{}
		}
	default:
		mp.LLM = &DummyLLM{}
	}
	return mp
}
