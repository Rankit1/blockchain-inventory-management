package genai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// MistralClient calls the real Mistral AI chat completions API.
type MistralClient struct {
	apiKey string
}

func NewMistral(apiKey string) *MistralClient {
	return &MistralClient{apiKey: apiKey}
}

func (m *MistralClient) SummarizeDocument(text string) (string, error) {
	return m.chatCompletion("You summarize documents into a single short note.", text)
}

func (m *MistralClient) ScorePriority(assetName, category string) (PriorityScores, error) {
	content, err := m.chatCompletion("You are an enterprise asset management classifier. Respond with JSON only.", priorityScorePrompt(assetName, category))
	if err != nil {
		return PriorityScores{}, err
	}
	return parsePriorityScores(content)
}

func (m *MistralClient) chatCompletion(systemPrompt, userPrompt string) (string, error) {
	type Choice struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	reqBody := map[string]interface{}{
		"model": "mistral-small-latest",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.mistral.ai/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+m.apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mistral error: %s", string(body))
	}
	var out struct {
		Choices []Choice `json:"choices"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", fmt.Errorf("failed to parse mistral response: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("mistral returned no choices")
	}
	return out.Choices[0].Message.Content, nil
}
