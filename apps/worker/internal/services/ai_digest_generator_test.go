package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
	workersai "github.com/maab88/repomemory/apps/worker/internal/services/ai"
)

func digestInput() DigestBuildInput {
	return DigestBuildInput{
		RepositoryFullName: "octocat/repo-memory",
		PeriodStart:        time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC),
		PeriodEnd:          time.Date(2026, 3, 8, 23, 59, 59, 0, time.UTC),
		MergedPullRequests: []jobs.PullRequestForDigest{
			{ID: uuid.New(), GitHubPrNumber: 1, Title: "Sync refactor"},
		},
	}
}

func TestAIDigestGeneratorProviderErrorFallback(t *testing.T) {
	g := NewAIDigestGenerator(&fakeAIProvider{name: workersai.ProviderStub, err: workersai.ErrInvalidResponse}, NewDeterministicDigestBuilder())
	draft := g.Build(digestInput())
	if draft.GeneratedBy != "deterministic" {
		t.Fatalf("expected deterministic fallback, got %s", draft.GeneratedBy)
	}
}

func TestAIDigestGeneratorInvalidJSONFallback(t *testing.T) {
	g := NewAIDigestGenerator(&fakeAIProvider{name: workersai.ProviderStub, response: `{"title":""}`}, NewDeterministicDigestBuilder())
	draft := g.Build(digestInput())
	if draft.GeneratedBy != "deterministic" {
		t.Fatalf("expected deterministic fallback, got %s", draft.GeneratedBy)
	}
}

func TestAIDigestGeneratorValidJSONUsesAI(t *testing.T) {
	resp := `{"title":"AI Digest","summary":"AI Summary","bodyMarkdown":"## AI Highlights"}`
	g := NewAIDigestGenerator(&fakeAIProvider{name: workersai.ProviderStub, response: resp}, NewDeterministicDigestBuilder())
	draft := g.Build(digestInput())
	if draft.GeneratedBy != "ai" {
		t.Fatalf("expected ai output, got %s", draft.GeneratedBy)
	}
}
