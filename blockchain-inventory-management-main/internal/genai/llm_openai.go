package genai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type OpenAIClient struct {
	apiKey string
}

func NewOpenAI(apiKey string) *OpenAIClient {
	return &OpenAIClient{apiKey: apiKey}
}

func (o *OpenAIClient) SummarizeDocument(text string) (string, error) {
	return o.chatCompletion("You summarize documents into a single short note.", text)
}

func (o *OpenAIClient) ScorePriority(assetName, category string) (PriorityScores, error) {
	content, err := o.chatCompletion("You are an enterprise asset management classifier. Respond with JSON only.", priorityScorePrompt(assetName, category))
	if err != nil {
		return PriorityScores{}, err
	}
	return parsePriorityScores(content)
}

func (o *OpenAIClient) chatCompletion(systemPrompt, userPrompt string) (string, error) {
	type Choice struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	}
	reqBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens": 200,
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error: %s", string(body))
	}
	var out struct {
		Choices []Choice `json:"choices"`
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", nil
	}
	return out.Choices[0].Message.Content, nil
}
