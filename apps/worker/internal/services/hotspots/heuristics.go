package hotspots

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type SourceRef struct {
	SourceType string
	SourceID   uuid.UUID
	Number     int32
	Title      string
	HTMLURL    string
	UpdatedAt  time.Time
}

type Candidate struct {
	Theme         string
	HitCount      int
	PRCount       int
	IssueCount    int
	BugOriented   int
	MatchedLabels []string
	Sources       []SourceRef
}

var splitPattern = regexp.MustCompile(`[^a-z0-9]+`)

var themeKeywords = map[string][]string{
	"auth":          {"auth", "oauth", "login", "permission", "permissions", "token", "sso", "rbac", "acl", "role"},
	"billing":       {"billing", "invoice", "payment", "subscription"},
	"migration":     {"migration", "migrate", "schema", "database", "postgres", "sql"},
	"notifications": {"notification", "notifications", "email", "alert"},
	"sync":          {"sync", "ingest", "import", "backfill", "queue", "job", "worker", "asynq", "retry", "retries", "backoff", "idempotent"},
	"webhooks":      {"webhook", "webhooks", "callback"},
}

var bugSignals = []string{"bug", "bugs", "fix", "hotfix", "incident", "failure", "regression", "flaky"}

func Detect(prs []jobs.PullRequestForHotspot, issues []jobs.IssueForHotspot) []Candidate {
	type themeAgg struct {
		sourceIDs  map[string]struct{}
		sourceRefs map[string]SourceRef
		prCount    int
		issueCount int
		bugCount   int
		labels     map[string]struct{}
	}

	themes := make(map[string]*themeAgg, len(themeKeywords))
	for theme := range themeKeywords {
		themes[theme] = &themeAgg{
			sourceIDs:  map[string]struct{}{},
			sourceRefs: map[string]SourceRef{},
			labels:     map[string]struct{}{},
		}
	}

	addSource := func(source SourceRef, body string, state string, labels []string) {
		text := strings.ToLower(strings.TrimSpace(source.Title + " " + body + " " + strings.Join(labels, " ")))
		tokens := tokenize(text)
		sourceKey := source.SourceType + ":" + source.SourceID.String()
		isBug := hasAny(tokens, bugSignals...)

		for theme, keywords := range themeKeywords {
			if !hasAny(tokens, keywords...) {
				continue
			}
			agg := themes[theme]
			if _, exists := agg.sourceIDs[sourceKey]; exists {
				continue
			}
			agg.sourceIDs[sourceKey] = struct{}{}
			agg.sourceRefs[sourceKey] = source
			if source.SourceType == "pull_request" {
				agg.prCount++
			} else {
				agg.issueCount++
			}
			if isBug || strings.EqualFold(strings.TrimSpace(state), "open") {
				agg.bugCount++
			}
			for _, label := range labels {
				label = cleanLabel(label)
				if label == "" {
					continue
				}
				if labelContainsTheme(label, keywords) {
					agg.labels[label] = struct{}{}
				}
			}
		}
	}

	for _, pr := range prs {
		addSource(SourceRef{
			SourceType: "pull_request",
			SourceID:   pr.ID,
			Number:     pr.GitHubPrNumber,
			Title:      pr.Title,
			HTMLURL:    pr.HTMLURL,
			UpdatedAt:  pr.UpdatedAtExternal.UTC(),
		}, pr.Body, pr.State, pr.Labels)
	}
	for _, issue := range issues {
		addSource(SourceRef{
			SourceType: "issue",
			SourceID:   issue.ID,
			Number:     issue.GitHubIssueNumber,
			Title:      issue.Title,
			HTMLURL:    issue.HTMLURL,
			UpdatedAt:  issue.UpdatedAtExternal.UTC(),
		}, issue.Body, issue.State, issue.Labels)
	}

	out := make([]Candidate, 0)
	for theme, agg := range themes {
		if len(agg.sourceIDs) < MinimumThemeSourceHits {
			continue
		}
		sources := make([]SourceRef, 0, len(agg.sourceRefs))
		for _, source := range agg.sourceRefs {
			sources = append(sources, source)
		}
		sort.Slice(sources, func(i, j int) bool {
			if !sources[i].UpdatedAt.Equal(sources[j].UpdatedAt) {
				return sources[i].UpdatedAt.After(sources[j].UpdatedAt)
			}
			if sources[i].SourceType != sources[j].SourceType {
				return sources[i].SourceType < sources[j].SourceType
			}
			if sources[i].Number != sources[j].Number {
				return sources[i].Number > sources[j].Number
			}
			return sources[i].SourceID.String() < sources[j].SourceID.String()
		})
		if len(sources) > MaxLinkedSources {
			sources = sources[:MaxLinkedSources]
		}

		matchedLabels := sortedLabels(agg.labels)
		out = append(out, Candidate{
			Theme:         theme,
			HitCount:      len(agg.sourceIDs),
			PRCount:       agg.prCount,
			IssueCount:    agg.issueCount,
			BugOriented:   agg.bugCount,
			MatchedLabels: matchedLabels,
			Sources:       sources,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].HitCount != out[j].HitCount {
			return out[i].HitCount > out[j].HitCount
		}
		if out[i].BugOriented != out[j].BugOriented {
			return out[i].BugOriented > out[j].BugOriented
		}
		return out[i].Theme < out[j].Theme
	})
	if len(out) > MaxHotspotsPerRun {
		out = out[:MaxHotspotsPerRun]
	}

	return out
}

func HotspotKey(theme string) string {
	return "hotspot:" + strings.TrimSpace(strings.ToLower(theme))
}

func BuildTitle(theme string) string {
	return fmt.Sprintf("Recurring %s-related activity", theme)
}

func BuildSummary(candidate Candidate) string {
	base := fmt.Sprintf(
		"Detected %d related items in the last %d days (%d PRs, %d issues) tied to %s.",
		candidate.HitCount,
		AnalysisWindowDays,
		candidate.PRCount,
		candidate.IssueCount,
		candidate.Theme,
	)
	if len(candidate.MatchedLabels) > 0 {
		return base + " Repeated labels: " + strings.Join(candidate.MatchedLabels, ", ") + "."
	}
	return base
}

func BuildWhyItMatters(theme string, bugSignals int) string {
	if bugSignals >= 2 {
		return fmt.Sprintf("Repeated %s work includes bug-oriented churn, signaling a reliability hotspot worth proactive ownership.", theme)
	}
	return fmt.Sprintf("Repeated %s work across multiple sources suggests concentrated engineering effort and recurring complexity.", theme)
}

func BuildImpactedAreas(theme string) []string {
	switch theme {
	case "sync":
		return []string{"sync", "workers", "reliability"}
	case "auth":
		return []string{"auth", "permissions", "security"}
	default:
		return []string{theme}
	}
}

func BuildRisks(theme string, bugSignals int) []string {
	out := []string{
		fmt.Sprintf("Recurring %s-related churn can cause repeated regressions if root causes stay unresolved.", theme),
	}
	if bugSignals >= 2 {
		out = append(out, "Bug-oriented repetition indicates operational instability may persist without systemic fixes.")
	}
	return out
}

func BuildFollowUps(theme string) []string {
	return []string{
		fmt.Sprintf("Assign an owner to track %s trends over the next sprint.", theme),
		fmt.Sprintf("Review recent %s incidents/changes and document shared root causes.", theme),
	}
}

func tokenize(text string) map[string]struct{} {
	parts := splitPattern.Split(strings.ToLower(text), -1)
	out := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out[part] = struct{}{}
	}
	return out
}

func hasAny(tokens map[string]struct{}, words ...string) bool {
	for _, word := range words {
		if _, ok := tokens[word]; ok {
			return true
		}
	}
	return false
}

func cleanLabel(label string) string {
	label = strings.TrimSpace(strings.ToLower(label))
	return strings.Join(strings.Fields(label), " ")
}

func labelContainsTheme(label string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(label, keyword) {
			return true
		}
	}
	return false
}

func sortedLabels(m map[string]struct{}) []string {
	labels := make([]string, 0, len(m))
	for label := range m {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	if len(labels) > 3 {
		labels = labels[:3]
	}
	return labels
}
