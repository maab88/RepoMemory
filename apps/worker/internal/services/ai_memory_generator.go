package services

import (
	"context"
	"fmt"

	"github.com/maab88/repomemory/apps/worker/internal/jobs"
	workersai "github.com/maab88/repomemory/apps/worker/internal/services/ai"
	"github.com/rs/zerolog/log"
)

type AIMemoryGenerator struct {
	provider workersai.Provider
	fallback MemoryDraftGenerator
}

func NewAIMemoryGenerator(provider workersai.Provider, fallback MemoryDraftGenerator) *AIMemoryGenerator {
	return &AIMemoryGenerator{
		provider: provider,
		fallback: fallback,
	}
}

func (g *AIMemoryGenerator) GenerateFromPullRequest(pr jobs.PullRequestForMemory) (MemoryEntryDraft, bool) {
	fallbackDraft, ok := g.fallback.GenerateFromPullRequest(pr)
	if !ok {
		return MemoryEntryDraft{}, false
	}
	if g.provider == nil || g.provider.Name() == workersai.ProviderDisabled {
		return fallbackDraft, true
	}

	raw, err := g.provider.CompleteJSON(context.Background(), workersai.CompletionRequest{
		SystemPrompt: workersai.MemorySystemPrompt(),
		UserPrompt: workersai.BuildMemoryUserPrompt(workersai.MemoryPromptInput{
			SourceType:       "pull_request",
			Number:           int(pr.GitHubPrNumber),
			Title:            pr.Title,
			Body:             pr.Body,
			State:            pr.State,
			AuthorLogin:      pr.AuthorLogin,
			Labels:           pr.Labels,
			WhyItMattersHint: fallbackDraft.WhyItMatters,
		}),
	})
	if err != nil {
		log.Warn().Str("provider", g.provider.Name()).Msg("ai memory generation failed, using deterministic fallback")
		return fallbackDraft, true
	}

	parsed, valid := workersai.ParseAndValidateMemorySummary(raw)
	if !valid || parsed.Type != MemoryTypePRSummary {
		log.Warn().Str("provider", g.provider.Name()).Msg("ai memory summary invalid, using deterministic fallback")
		return fallbackDraft, true
	}

	return MemoryEntryDraft{
		Type:          parsed.Type,
		Title:         parsed.Title,
		Summary:       parsed.Summary,
		WhyItMatters:  parsed.WhyItMatters,
		ImpactedAreas: parsed.ImpactedAreas,
		Risks:         parsed.Risks,
		FollowUps:     parsed.FollowUps,
		GeneratedBy:   "ai",
	}, true
}

func (g *AIMemoryGenerator) GenerateFromIssue(issue jobs.IssueForMemory) (MemoryEntryDraft, bool) {
	fallbackDraft, ok := g.fallback.GenerateFromIssue(issue)
	if !ok {
		return MemoryEntryDraft{}, false
	}
	if g.provider == nil || g.provider.Name() == workersai.ProviderDisabled {
		return fallbackDraft, true
	}

	raw, err := g.provider.CompleteJSON(context.Background(), workersai.CompletionRequest{
		SystemPrompt: workersai.MemorySystemPrompt(),
		UserPrompt: workersai.BuildMemoryUserPrompt(workersai.MemoryPromptInput{
			SourceType:       "issue",
			Number:           int(issue.GitHubIssueNumber),
			Title:            issue.Title,
			Body:             issue.Body,
			State:            issue.State,
			AuthorLogin:      issue.AuthorLogin,
			Labels:           issue.Labels,
			WhyItMattersHint: fallbackDraft.WhyItMatters,
		}),
	})
	if err != nil {
		log.Warn().Str("provider", g.provider.Name()).Msg("ai memory generation failed, using deterministic fallback")
		return fallbackDraft, true
	}

	parsed, valid := workersai.ParseAndValidateMemorySummary(raw)
	if !valid || parsed.Type != MemoryTypeIssueSummary {
		log.Warn().Str("provider", g.provider.Name()).Msg("ai memory summary invalid, using deterministic fallback")
		return fallbackDraft, true
	}

	return MemoryEntryDraft{
		Type:          parsed.Type,
		Title:         parsed.Title,
		Summary:       parsed.Summary,
		WhyItMatters:  parsed.WhyItMatters,
		ImpactedAreas: parsed.ImpactedAreas,
		Risks:         parsed.Risks,
		FollowUps:     parsed.FollowUps,
		GeneratedBy:   "ai",
	}, true
}

var _ MemoryDraftGenerator = (*AIMemoryGenerator)(nil)

func (g *AIMemoryGenerator) String() string {
	if g.provider == nil {
		return "ai-memory-generator(nil-provider)"
	}
	return fmt.Sprintf("ai-memory-generator(%s)", g.provider.Name())
}
