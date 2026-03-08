package services

import (
	"fmt"
	"strings"

	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

const (
	MemoryTypePRSummary    = "pr_summary"
	MemoryTypeIssueSummary = "issue_summary"
)

type MemoryEntryDraft struct {
	Type          string
	Title         string
	Summary       string
	WhyItMatters  string
	ImpactedAreas []string
	Risks         []string
	FollowUps     []string
	GeneratedBy   string
}

type DeterministicMemoryGenerator struct{}

func NewDeterministicMemoryGenerator() *DeterministicMemoryGenerator {
	return &DeterministicMemoryGenerator{}
}

func (g *DeterministicMemoryGenerator) GenerateFromPullRequest(pr jobs.PullRequestForMemory) (MemoryEntryDraft, bool) {
	title := cleanOneLine(pr.Title)
	if title == "" {
		return MemoryEntryDraft{}, false
	}

	body := cleanOneLine(pr.Body)
	combined := strings.TrimSpace(title + " " + body + " " + strings.Join(pr.Labels, " "))
	merged := pr.MergedAt != nil
	stateSummary := strings.ToLower(strings.TrimSpace(pr.State))
	if merged {
		stateSummary = "merged"
	}

	summary := fmt.Sprintf("PR #%d %s. %s", pr.GitHubPrNumber, stateSummary, truncate(nonEmpty(body, title), 220))
	impacted := extractImpactedAreas(combined, pr.Labels)
	risks := extractRisks(combined, pr.Labels)
	followUps := extractFollowUps(pr.State, merged, combined)

	return MemoryEntryDraft{
		Type:          MemoryTypePRSummary,
		Title:         fmt.Sprintf("PR #%d: %s", pr.GitHubPrNumber, title),
		Summary:       summary,
		WhyItMatters:  composeWhyItMatters(impacted, risks, "Pull request changes can affect repository behavior and team workflows."),
		ImpactedAreas: impacted,
		Risks:         risks,
		FollowUps:     followUps,
		GeneratedBy:   "deterministic",
	}, true
}

func (g *DeterministicMemoryGenerator) GenerateFromIssue(issue jobs.IssueForMemory) (MemoryEntryDraft, bool) {
	title := cleanOneLine(issue.Title)
	if title == "" {
		return MemoryEntryDraft{}, false
	}

	body := cleanOneLine(issue.Body)
	combined := strings.TrimSpace(title + " " + body + " " + strings.Join(issue.Labels, " "))
	stateSummary := strings.ToLower(strings.TrimSpace(issue.State))

	summary := fmt.Sprintf("Issue #%d %s. %s", issue.GitHubIssueNumber, stateSummary, truncate(nonEmpty(body, title), 220))
	impacted := extractImpactedAreas(combined, issue.Labels)
	risks := extractRisks(combined, issue.Labels)
	followUps := extractFollowUps(issue.State, false, combined)

	return MemoryEntryDraft{
		Type:          MemoryTypeIssueSummary,
		Title:         fmt.Sprintf("Issue #%d: %s", issue.GitHubIssueNumber, title),
		Summary:       summary,
		WhyItMatters:  composeWhyItMatters(impacted, risks, "Issue context helps preserve known problems and intended follow-through."),
		ImpactedAreas: impacted,
		Risks:         risks,
		FollowUps:     followUps,
		GeneratedBy:   "deterministic",
	}, true
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
