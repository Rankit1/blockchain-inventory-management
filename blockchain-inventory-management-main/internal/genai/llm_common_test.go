package genai

import "testing"

func TestParsePriorityScores(t *testing.T) {
	raw := "Sure, here you go:\n```json\n{\"businessCriticality\":5,\"replacementCost\":4,\"replacementLeadTime\":3,\"safetyComplianceImpact\":2,\"redundancyAvailability\":1}\n```"
	scores, err := parsePriorityScores(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := PriorityScores{BusinessCriticality: 5, ReplacementCost: 4, ReplacementLeadTime: 3, SafetyComplianceImpact: 2, RedundancyAvailability: 1}
	if scores != want {
		t.Fatalf("got %+v, want %+v", scores, want)
	}
}

func TestParsePriorityScoresClampsOutOfRange(t *testing.T) {
	scores, err := parsePriorityScores(`{"businessCriticality":9,"replacementCost":0,"replacementLeadTime":3,"safetyComplianceImpact":2,"redundancyAvailability":1}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scores.BusinessCriticality != 5 || scores.ReplacementCost != 1 {
		t.Fatalf("expected out-of-range scores to be clamped to 1-5, got %+v", scores)
	}
}

func TestParsePriorityScoresNoJSON(t *testing.T) {
	if _, err := parsePriorityScores("I cannot help with that."); err == nil {
		t.Fatalf("expected error for response with no JSON object")
	}
}

func TestDummyLLMScorePriorityFallsBackToHeuristic(t *testing.T) {
	scores, err := (&DummyLLM{}).ScorePriority("Core Switch", "network")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !scores.valid() {
		t.Fatalf("expected heuristic scores to be valid, got %+v", scores)
	}
}
