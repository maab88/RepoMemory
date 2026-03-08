package services

import (
	"context"
	"fmt"

	workersai "github.com/maab88/repomemory/apps/worker/internal/services/ai"
	"github.com/rs/zerolog/log"
)

type AIDigestGenerator struct {
	provider workersai.Provider
	fallback DigestBuilder
}

func NewAIDigestGenerator(provider workersai.Provider, fallback DigestBuilder) *AIDigestGenerator {
	return &AIDigestGenerator{
		provider: provider,
		fallback: fallback,
	}
}

func (g *AIDigestGenerator) Build(input DigestBuildInput) DigestDraft {
	fallbackDraft := g.fallback.Build(input)
	if g.provider == nil || g.provider.Name() == workersai.ProviderDisabled {
		return fallbackDraft
	}

	mergedPRLines := make([]string, 0, len(input.MergedPullRequests))
	for _, pr := range input.MergedPullRequests {
		mergedPRLines = append(mergedPRLines, fmt.Sprintf("PR #%d %s", pr.GitHubPrNumber, cleanLine(pr.Title)))
	}
	openIssueLines := make([]string, 0, len(input.OpenIssues))
	for _, issue := range input.OpenIssues {
		openIssueLines = append(openIssueLines, fmt.Sprintf("Issue #%d %s", issue.GitHubIssueNumber, cleanLine(issue.Title)))
	}

	raw, err := g.provider.CompleteJSON(context.Background(), workersai.CompletionRequest{
		SystemPrompt: workersai.DigestSystemPrompt(),
		UserPrompt: workersai.BuildDigestUserPrompt(workersai.DigestPromptInput{
			RepositoryFullName: input.RepositoryFullName,
			PeriodStart:        input.PeriodStart,
			PeriodEnd:          input.PeriodEnd,
			MergedPRLines:      mergedPRLines,
			OpenIssueLines:     openIssueLines,
			Hotspots:           topHotspots(input.MemoryEntries, 5),
			OnboardingNotes:    onboardingNotes(input.MemoryEntries, 5),
		}),
	})
	if err != nil {
		log.Warn().Str("provider", g.provider.Name()).Msg("ai digest generation failed, using deterministic fallback")
		return fallbackDraft
	}

	parsed, valid := workersai.ParseAndValidateDigestSummary(raw)
	if !valid {
		log.Warn().Str("provider", g.provider.Name()).Msg("ai digest summary invalid, using deterministic fallback")
		return fallbackDraft
	}

	return DigestDraft{
		Title:        parsed.Title,
		Summary:      parsed.Summary,
		BodyMarkdown: parsed.BodyMarkdown,
		GeneratedBy:  "ai",
	}
}

var _ DigestBuilder = (*AIDigestGenerator)(nil)
