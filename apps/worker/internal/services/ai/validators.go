package ai

import (
	"encoding/json"
	"strings"
)

type AIMemorySummary struct {
	Type          string   `json:"type"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	WhyItMatters  string   `json:"whyItMatters"`
	ImpactedAreas []string `json:"impactedAreas"`
	Risks         []string `json:"risks"`
	FollowUps     []string `json:"followUps"`
}

type AIDigestSummary struct {
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	BodyMarkdown string `json:"bodyMarkdown"`
}

func ParseAndValidateMemorySummary(raw string) (AIMemorySummary, bool) {
	var payload AIMemorySummary
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return AIMemorySummary{}, false
	}
	if strings.TrimSpace(payload.Type) == "" ||
		strings.TrimSpace(payload.Title) == "" ||
		strings.TrimSpace(payload.Summary) == "" ||
		strings.TrimSpace(payload.WhyItMatters) == "" {
		return AIMemorySummary{}, false
	}
	if !isStringArray(payload.ImpactedAreas) || !isStringArray(payload.Risks) || !isStringArray(payload.FollowUps) {
		return AIMemorySummary{}, false
	}
	payload.Type = strings.TrimSpace(payload.Type)
	payload.Title = strings.TrimSpace(payload.Title)
	payload.Summary = strings.TrimSpace(payload.Summary)
	payload.WhyItMatters = strings.TrimSpace(payload.WhyItMatters)
	return payload, true
}

func ParseAndValidateDigestSummary(raw string) (AIDigestSummary, bool) {
	var payload AIDigestSummary
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return AIDigestSummary{}, false
	}
	if strings.TrimSpace(payload.Title) == "" ||
		strings.TrimSpace(payload.Summary) == "" ||
		strings.TrimSpace(payload.BodyMarkdown) == "" {
		return AIDigestSummary{}, false
	}
	payload.Title = strings.TrimSpace(payload.Title)
	payload.Summary = strings.TrimSpace(payload.Summary)
	payload.BodyMarkdown = strings.TrimSpace(payload.BodyMarkdown)
	return payload, true
}

func isStringArray(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return false
		}
	}
	return true
}
