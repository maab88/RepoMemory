package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type DigestDraft struct {
	Title        string
	Summary      string
	BodyMarkdown string
	GeneratedBy  string
}

type DigestBuildInput struct {
	RepositoryFullName string
	PeriodStart        time.Time
	PeriodEnd          time.Time
	MergedPullRequests []jobs.PullRequestForDigest
	OpenIssues         []jobs.IssueForDigest
	MemoryEntries      []jobs.MemoryEntryForDigest
}

type DeterministicDigestBuilder struct{}

func NewDeterministicDigestBuilder() *DeterministicDigestBuilder {
	return &DeterministicDigestBuilder{}
}

func (b *DeterministicDigestBuilder) Build(input DigestBuildInput) DigestDraft {
	title := fmt.Sprintf("Weekly Digest: %s - %s", input.PeriodStart.Format("Jan 2"), input.PeriodEnd.Format("Jan 2"))
	hotspots := topHotspots(input.MemoryEntries, 3)
	onboarding := onboardingNotes(input.MemoryEntries, 3)

	summaryParts := []string{
		fmt.Sprintf("%d merged PRs", len(input.MergedPullRequests)),
		fmt.Sprintf("%d open issues", len(input.OpenIssues)),
	}
	if len(hotspots) > 0 {
		summaryParts = append(summaryParts, "hotspots: "+strings.Join(hotspots, ", "))
	} else {
		summaryParts = append(summaryParts, "no strong hotspots detected")
	}
	summary := strings.Join(summaryParts, ", ") + "."

	lines := []string{
		fmt.Sprintf("# %s", title),
		"",
		"## Highlights",
		fmt.Sprintf("- %d pull requests were merged this week.", len(input.MergedPullRequests)),
		fmt.Sprintf("- %d issues remain open with recent activity.", len(input.OpenIssues)),
	}
	if len(hotspots) > 0 {
		lines = append(lines, "- Primary hotspots: "+strings.Join(hotspots, ", ")+".")
	} else {
		lines = append(lines, "- Primary hotspots: low activity this week.")
	}

	lines = append(lines, "", "## Merged Pull Requests")
	if len(input.MergedPullRequests) == 0 {
		lines = append(lines, "- No merged pull requests in this period.")
	} else {
		for _, pr := range limitPullRequests(input.MergedPullRequests, 5) {
			lines = append(lines, fmt.Sprintf("- PR #%d %s (%s)", pr.GitHubPrNumber, cleanLine(pr.Title), pr.HTMLURL))
		}
	}

	lines = append(lines, "", "## Significant Open Issues")
	if len(input.OpenIssues) == 0 {
		lines = append(lines, "- No open issues with notable recent activity.")
	} else {
		for _, issue := range limitIssues(input.OpenIssues, 5) {
			lines = append(lines, fmt.Sprintf("- Issue #%d %s (%s)", issue.GitHubIssueNumber, cleanLine(issue.Title), issue.HTMLURL))
		}
	}

	lines = append(lines, "", "## Onboarding Notes")
	if len(onboarding) == 0 {
		lines = append(lines, "- Repository activity was light; no specific onboarding notes captured this week.")
	} else {
		for _, note := range onboarding {
			lines = append(lines, "- "+note)
		}
	}

	return DigestDraft{
		Title:        title,
		Summary:      summary,
		BodyMarkdown: strings.Join(lines, "\n"),
		GeneratedBy:  "deterministic",
	}
}

func topHotspots(entries []jobs.MemoryEntryForDigest, limit int) []string {
	if limit <= 0 {
		return []string{}
	}
	counts := map[string]int{}
	for _, entry := range entries {
		for _, area := range entry.ImpactedAreas {
			key := strings.TrimSpace(strings.ToLower(area))
			if key == "" {
				continue
			}
			counts[key]++
		}
	}
	type pair struct {
		name  string
		count int
	}
	ordered := make([]pair, 0, len(counts))
	for k, v := range counts {
		ordered = append(ordered, pair{name: k, count: v})
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].count == ordered[j].count {
			return ordered[i].name < ordered[j].name
		}
		return ordered[i].count > ordered[j].count
	})
	result := make([]string, 0, limit)
	for i := 0; i < len(ordered) && i < limit; i++ {
		result = append(result, ordered[i].name)
	}
	return result
}

func onboardingNotes(entries []jobs.MemoryEntryForDigest, limit int) []string {
	if limit <= 0 {
		return []string{}
	}
	notes := make([]string, 0, limit)
	seen := map[string]struct{}{}
	for _, entry := range entries {
		for _, candidate := range entry.FollowUps {
			note := cleanLine(candidate)
			if note == "" {
				continue
			}
			key := strings.ToLower(note)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			notes = append(notes, note)
			if len(notes) >= limit {
				return notes
			}
		}
	}
	return notes
}

func limitPullRequests(items []jobs.PullRequestForDigest, limit int) []jobs.PullRequestForDigest {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func limitIssues(items []jobs.IssueForDigest, limit int) []jobs.IssueForDigest {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func cleanLine(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if value == "" {
		return "Untitled"
	}
	return value
}
