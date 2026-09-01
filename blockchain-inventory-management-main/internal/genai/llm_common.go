package genai

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// priorityScorePrompt is the shared instruction given to every LLM provider
// for the asset priority classification task (see GenAI_Asset_Prioritization_Addendum.pdf).
func priorityScorePrompt(assetName, category string) string {
	return fmt.Sprintf(`You are an enterprise asset management classifier. Rate the following asset on five criteria, each on a 1-5 scale (5 = highest risk/impact):
- businessCriticality: impact on department operations if unavailable
- replacementCost: monetary/procurement cost
- replacementLeadTime: time required to source and deploy a replacement
- safetyComplianceImpact: regulatory/safety exposure
- redundancyAvailability: whether spares or alternates exist

Asset name: %s
Asset category: %s

Respond with ONLY a JSON object, no other text, in exactly this shape:
{"businessCriticality":N,"replacementCost":N,"replacementLeadTime":N,"safetyComplianceImpact":N,"redundancyAvailability":N}`, assetName, category)
}

var jsonObjectPattern = regexp.MustCompile(`\{[^{}]*\}`)

// parsePriorityScores extracts and validates a PriorityScores JSON object from
// a raw LLM completion, which may include surrounding prose or markdown fences.
func parsePriorityScores(raw string) (PriorityScores, error) {
	match := jsonObjectPattern.FindString(raw)
	if match == "" {
		return PriorityScores{}, fmt.Errorf("no JSON object found in LLM response")
	}
	var scores PriorityScores
	if err := json.Unmarshal([]byte(match), &scores); err != nil {
		return PriorityScores{}, fmt.Errorf("failed to parse priority scores: %w", err)
	}
	if !scores.valid() {
		scores = scores.clamp()
	}
	return scores, nil
}
