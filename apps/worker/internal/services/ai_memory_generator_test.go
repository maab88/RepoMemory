package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
	workersai "github.com/maab88/repomemory/apps/worker/internal/services/ai"
)

type fakeAIProvider struct {
	name     string
	response string
	err      error
}

func (f *fakeAIProvider) Name() string { return f.name }

func (f *fakeAIProvider) CompleteJSON(context.Context, workersai.CompletionRequest) (string, error) {
	return f.response, f.err
}

func samplePR() jobs.PullRequestForMemory {
	return jobs.PullRequestForMemory{
		ID:             uuid.New(),
		RepositoryID:   uuid.New(),
		GitHubPrNumber: 7,
		Title:          "Queue retry fix",
		Body:           "Improve retry behavior",
		State:          "merged",
		AuthorLogin:    "octocat",
		HTMLURL:        "https://github.com/octo/repo/pull/7",
		Labels:         []string{"sync"},
	}
}

func TestAIMemoryGeneratorDisabledUsesDeterministic(t *testing.T) {
	g := NewAIMemoryGenerator(&fakeAIProvider{name: workersai.ProviderDisabled}, NewDeterministicMemoryGenerator())
	draft, ok := g.GenerateFromPullRequest(samplePR())
	if !ok {
		t.Fatal("expected draft")
	}
	if draft.GeneratedBy != "deterministic" {
		t.Fatalf("expected deterministic fallback, got %s", draft.GeneratedBy)
	}
}

func TestAIMemoryGeneratorProviderErrorFallback(t *testing.T) {
	g := NewAIMemoryGenerator(&fakeAIProvider{name: workersai.ProviderStub, err: errors.New("boom")}, NewDeterministicMemoryGenerator())
	draft, ok := g.GenerateFromPullRequest(samplePR())
	if !ok {
		t.Fatal("expected draft")
	}
	if draft.GeneratedBy != "deterministic" {
		t.Fatalf("expected fallback, got %s", draft.GeneratedBy)
	}
}

func TestAIMemoryGeneratorInvalidJSONFallback(t *testing.T) {
	g := NewAIMemoryGenerator(&fakeAIProvider{name: workersai.ProviderStub, response: `{"title":"x"}`}, NewDeterministicMemoryGenerator())
	draft, ok := g.GenerateFromPullRequest(samplePR())
	if !ok {
		t.Fatal("expected draft")
	}
	if draft.GeneratedBy != "deterministic" {
		t.Fatalf("expected fallback, got %s", draft.GeneratedBy)
	}
}

func TestAIMemoryGeneratorValidJSONUsesAI(t *testing.T) {
	resp := `{"type":"pr_summary","title":"AI title","summary":"AI summary","whyItMatters":"AI matters","impactedAreas":["sync"],"risks":["retry drift"],"followUps":["monitor"]}`
	g := NewAIMemoryGenerator(&fakeAIProvider{name: workersai.ProviderStub, response: resp}, NewDeterministicMemoryGenerator())
	draft, ok := g.GenerateFromPullRequest(samplePR())
	if !ok {
		t.Fatal("expected draft")
	}
	if draft.GeneratedBy != "ai" {
		t.Fatalf("expected ai draft, got %s", draft.GeneratedBy)
	}
}
