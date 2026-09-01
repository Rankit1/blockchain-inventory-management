package genai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GeminiClient calls the real Google Generative Language API.
type GeminiClient struct {
	apiKey string
}

func NewGemini(apiKey string) *GeminiClient {
	return &GeminiClient{apiKey: apiKey}
}

func (g *GeminiClient) SummarizeDocument(text string) (string, error) {
	return g.generateContent("Summarize the following document into a single short note:\n\n" + text)
}

func (g *GeminiClient) ScorePriority(assetName, category string) (PriorityScores, error) {
	content, err := g.generateContent(priorityScorePrompt(assetName, category))
	if err != nil {
		return PriorityScores{}, err
	}
	return parsePriorityScores(content)
}

func (g *GeminiClient) generateContent(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}
	b, _ := json.Marshal(reqBody)
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + g.apiKey
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini error: %s", string(body))
	}
	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("failed to parse gemini response: %w", err)
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned no content")
	}
	return out.Candidates[0].Content.Parts[0].Text, nil
}
