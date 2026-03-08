package ai

import "strings"

const (
	ProviderDisabled = "disabled"
	ProviderStub     = "stub"
	ProviderOpenAI   = "openai"
)

type Config struct {
	Provider    string
	OpenAIKey   string
	OpenAIURL   string
	OpenAIModel string
}

func NormalizeProvider(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case ProviderStub, ProviderOpenAI:
		return normalized
	default:
		return ProviderDisabled
	}
}
