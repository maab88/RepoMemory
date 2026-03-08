package ai

import (
	"fmt"
	"strings"
	"time"
)

const (
	MemoryPromptVersion = "memory-v1"
	DigestPromptVersion = "digest-v1"
)

func MemorySystemPrompt() string {
	return "You write concise engineering memory summaries. Return strict JSON only. No markdown."
}

func DigestSystemPrompt() string {
	return "You write concise weekly engineering digests. Return strict JSON only."
}

type MemoryPromptInput struct {
	SourceType       string
	Number           int
	Title            string
	Body             string
	State            string
	AuthorLogin      string
	Labels           []string
	WhyItMattersHint string
}

func BuildMemoryUserPrompt(input MemoryPromptInput) string {
	return fmt.Sprintf(
		"PromptVersion: %s\nSourceType: %s\nNumber: %d\nTitle: %s\nBody: %s\nState: %s\nAuthor: %s\nLabels: %s\nHint: %s\n\nReturn JSON with fields: type,title,summary,whyItMatters,impactedAreas,risks,followUps.\nUse only arrays of strings for impactedAreas/risks/followUps.",
		MemoryPromptVersion,
		input.SourceType,
		input.Number,
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Body),
		strings.TrimSpace(input.State),
		strings.TrimSpace(input.AuthorLogin),
		strings.Join(input.Labels, ", "),
		strings.TrimSpace(input.WhyItMattersHint),
	)
}

type DigestPromptInput struct {
	RepositoryFullName string
	PeriodStart        time.Time
	PeriodEnd          time.Time
	MergedPRLines      []string
	OpenIssueLines     []string
	Hotspots           []string
	OnboardingNotes    []string
}

func BuildDigestUserPrompt(input DigestPromptInput) string {
	return fmt.Sprintf(
		"PromptVersion: %s\nRepository: %s\nPeriodStart: %s\nPeriodEnd: %s\nMergedPRs:\n%s\nOpenIssues:\n%s\nHotspots: %s\nOnboardingNotes: %s\n\nReturn JSON with fields: title,summary,bodyMarkdown.",
		DigestPromptVersion,
		input.RepositoryFullName,
		input.PeriodStart.UTC().Format(time.RFC3339),
		input.PeriodEnd.UTC().Format(time.RFC3339),
		strings.Join(input.MergedPRLines, "\n"),
		strings.Join(input.OpenIssueLines, "\n"),
		strings.Join(input.Hotspots, ", "),
		strings.Join(input.OnboardingNotes, "; "),
	)
}
