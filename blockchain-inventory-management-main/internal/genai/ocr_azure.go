package genai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

type AzureOCR struct {
	endpoint string
	key      string
}

func NewAzureOCR(endpoint, key string) *AzureOCR {
	return &AzureOCR{endpoint: endpoint, key: key}
}

// ExtractWarrantyAndAMC posts the document to Azure Read API and polls the
// Operation-Location until the analysis completes, then extracts a date.
func (a *AzureOCR) ExtractWarrantyAndAMC(doc []byte) (string, string, error) {
	if len(doc) == 0 {
		return "", "", nil
	}
	req, err := http.NewRequest("POST", a.endpoint+"/vision/v3.2/read/analyze", bytes.NewReader(doc))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", a.key)
	req.Header.Set("Content-Type", "application/octet-stream")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// If Azure accepted the job, it returns 202 and an Operation-Location header
	opURL := resp.Header.Get("Operation-Location")
	if opURL == "" {
		// fallback: try to parse body synchronously
		body, _ := ioutil.ReadAll(resp.Body)
		raw := string(body)
		re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
		m := re.FindString(raw)
		if m != "" {
			return m, "", nil
		}
		return "", "", nil
	}

	// Poll operation URL until succeeded or timeout
	timeout := time.After(10 * time.Second)
	ticker := time.Tick(500 * time.Millisecond)
	var lastBody []byte
	for {
		select {
		case <-timeout:
			// give up and try to parse last body if available
			if len(lastBody) > 0 {
				raw := string(lastBody)
				re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
				m := re.FindString(raw)
				if m != "" {
					return m, "", nil
				}
			}
			return "", "", fmt.Errorf("azure ocr: operation timed out")
		case <-ticker:
			r2, err := http.NewRequest("GET", opURL, nil)
			if err != nil {
				return "", "", err
			}
			r2.Header.Set("Ocp-Apim-Subscription-Key", a.key)
			res2, err := client.Do(r2)
			if err != nil {
				return "", "", err
			}
			body, _ := ioutil.ReadAll(res2.Body)
			res2.Body.Close()
			lastBody = body
			var js map[string]interface{}
			_ = json.Unmarshal(body, &js)
			// status may be under "status" or nested
			status := ""
			if s, ok := js["status"].(string); ok {
				status = s
			} else if v, ok := js["analyzeResult"]; ok {
				// if analyzeResult present assume succeeded
				status = "succeeded"
				_ = v
			}
			if status == "succeeded" {
				// extract text from analyzeResult if present
				raw := string(body)
				re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
				m := re.FindString(raw)
				if m != "" {
					return m, "", nil
				}
				return "", "", nil
			}
			// loop until timeout
		}
	}
}
