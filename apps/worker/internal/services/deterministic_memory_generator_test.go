package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

func TestGenerateFromPullRequest(t *testing.T) {
	now := time.Now().UTC()
	g := NewDeterministicMemoryGenerator()
	draft, ok := g.GenerateFromPullRequest(jobs.PullRequestForMemory{
		ID:             uuid.New(),
		RepositoryID:   uuid.New(),
		GitHubPrNumber: 42,
		Title:          "Add queue retry guard for OAuth sync",
		Body:           "This updates worker retry behavior for auth token refresh in billing sync.",
		State:          "closed",
		AuthorLogin:    "octocat",
		HTMLURL:        "https://github.com/octo/repo/pull/42",
		MergedAt:       &now,
		Labels:         []string{"sync", "billing"},
	})
	if !ok {
		t.Fatal("expected PR draft to generate")
	}
	if draft.Type != MemoryTypePRSummary {
		t.Fatalf("unexpected type: %s", draft.Type)
	}
	if draft.GeneratedBy != "deterministic" {
		t.Fatalf("unexpected generatedBy: %s", draft.GeneratedBy)
	}
	if draft.Title == "" || draft.Summary == "" || draft.WhyItMatters == "" {
		t.Fatal("expected non-empty memory fields")
	}
	if len(draft.ImpactedAreas) == 0 {
		t.Fatal("expected impacted areas")
	}
}

func TestGenerateFromIssue(t *testing.T) {
	g := NewDeterministicMemoryGenerator()
	draft, ok := g.GenerateFromIssue(jobs.IssueForMemory{
		ID:                uuid.New(),
		RepositoryID:      uuid.New(),
		GitHubIssueNumber: 8,
		Title:             "Webhook permissions bug",
		Body:              "Users without permission can trigger webhook retries.",
		State:             "open",
		AuthorLogin:       "hubot",
		HTMLURL:           "https://github.com/octo/repo/issues/8",
		Labels:            []string{"auth"},
	})
	if !ok {
		t.Fatal("expected issue draft to generate")
	}
	if draft.Type != MemoryTypeIssueSummary {
		t.Fatalf("unexpected type: %s", draft.Type)
	}
	if len(draft.Risks) == 0 {
		t.Fatal("expected issue risks")
	}
}
