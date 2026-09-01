package genai

import "strings"

// Asset lifecycle states (mirrors chaincode/assetcc's on-chain values).
const (
	LifecycleActive      = "ACTIVE"
	LifecycleMaintenance = "MAINTENANCE"
	LifecycleIdle        = "IDLE"
	LifecycleRetired     = "RETIRED"
)

// PriorityScores holds the five 1-5 criterion ratings used for classification.
// The weighted sum and tier mapping are computed on-chain by classifyPriority();
// this struct is just the transport for the agent's ratings.
type PriorityScores struct {
	BusinessCriticality    int `json:"businessCriticality"`
	ReplacementCost        int `json:"replacementCost"`
	ReplacementLeadTime    int `json:"replacementLeadTime"`
	SafetyComplianceImpact int `json:"safetyComplianceImpact"`
	RedundancyAvailability int `json:"redundancyAvailability"`
}

func (s PriorityScores) valid() bool {
	inRange := func(v int) bool { return v >= 1 && v <= 5 }
	return inRange(s.BusinessCriticality) && inRange(s.ReplacementCost) &&
		inRange(s.ReplacementLeadTime) && inRange(s.SafetyComplianceImpact) &&
		inRange(s.RedundancyAvailability)
}

func (s PriorityScores) clamp() PriorityScores {
	clamp := func(v int) int {
		if v < 1 {
			return 1
		}
		if v > 5 {
			return 5
		}
		return v
	}
	return PriorityScores{
		BusinessCriticality:    clamp(s.BusinessCriticality),
		ReplacementCost:        clamp(s.ReplacementCost),
		ReplacementLeadTime:    clamp(s.ReplacementLeadTime),
		SafetyComplianceImpact: clamp(s.SafetyComplianceImpact),
		RedundancyAvailability: clamp(s.RedundancyAvailability),
	}
}

// heuristicScores is the deterministic fallback used when no LLM provider is
// configured, or when a live LLM call fails. Category keywords are matched
// against the same bands the chaincode's own default scorer uses.
func heuristicScores(category string) PriorityScores {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "server", "servers", "network", "network device", "network devices", "storage", "secure storage", "infrastructure":
		return PriorityScores{BusinessCriticality: 5, ReplacementCost: 5, ReplacementLeadTime: 4, SafetyComplianceImpact: 3, RedundancyAvailability: 2}
	case "diagnostic", "medical", "diagnostic equipment", "medical equipment":
		return PriorityScores{BusinessCriticality: 5, ReplacementCost: 4, ReplacementLeadTime: 4, SafetyComplianceImpact: 4, RedundancyAvailability: 2}
	case "laptop", "computer", "desktop", "workstation", "electronics":
		return PriorityScores{BusinessCriticality: 3, ReplacementCost: 3, ReplacementLeadTime: 2, SafetyComplianceImpact: 2, RedundancyAvailability: 3}
	default:
		return PriorityScores{BusinessCriticality: 2, ReplacementCost: 2, ReplacementLeadTime: 2, SafetyComplianceImpact: 2, RedundancyAvailability: 3}
	}
}
