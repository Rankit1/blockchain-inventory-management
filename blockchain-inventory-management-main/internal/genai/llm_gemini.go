package genai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Gemini client placeholder (Google PaLM/ Gemini-ish). Real integration
// requires proper endpoint and auth.
type GeminiClient struct {
	apiKey string
}

func NewGemini(apiKey string) *GeminiClient {
	return &GeminiClient{apiKey: apiKey}
}

func (g *GeminiClient) SummarizeDocument(text string) (string, error) {
	reqBody := map[string]interface{}{"input": text}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://gemini.api.google/v1/generate", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini error: %s", string(body))
	}
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body), nil
}
