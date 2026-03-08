package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type fakeMemoryStore struct {
	repo      jobs.RepositoryForSync
	prs       []jobs.PullRequestForMemory
	issues    []jobs.IssueForMemory
	upsertErr error

	upsertCalls int
	seen        map[string]uuid.UUID
}

func (f *fakeMemoryStore) GetRepositoryForSync(context.Context, uuid.UUID) (jobs.RepositoryForSync, error) {
	return f.repo, nil
}

func (f *fakeMemoryStore) ListPullRequestsForRepository(context.Context, uuid.UUID) ([]jobs.PullRequestForMemory, error) {
	return f.prs, nil
}

func (f *fakeMemoryStore) ListIssuesForRepository(context.Context, uuid.UUID) ([]jobs.IssueForMemory, error) {
	return f.issues, nil
}

func (f *fakeMemoryStore) UpsertMemoryEntryForSource(_ context.Context, record jobs.MemoryEntryUpsertRecord) (uuid.UUID, bool, error) {
	if f.upsertErr != nil {
		return uuid.Nil, false, f.upsertErr
	}
	if f.seen == nil {
		f.seen = map[string]uuid.UUID{}
	}
	f.upsertCalls++
	key := record.Type + "|" + record.SourceType + "|" + record.SourceRecordID.String()
	if id, ok := f.seen[key]; ok {
		return id, false, nil
	}
	id := uuid.New()
	f.seen[key] = id
	return id, true, nil
}

func TestGenerateAndPersistForRepositoryNoOp(t *testing.T) {
	repositoryID := uuid.New()
	store := &fakeMemoryStore{
		repo: jobs.RepositoryForSync{
			ID:           repositoryID,
			Organization: uuid.New(),
			OwnerLogin:   "octo",
			Name:         "repo",
		},
	}

	svc := NewMemoryGenerationService(store, NewDeterministicMemoryGenerator())
	result, err := svc.GenerateAndPersistForRepository(context.Background(), jobs.RepoGenerateMemoryPayload{
		RepositoryID:      repositoryID,
		OrganizationID:    store.repo.Organization,
		TriggeredByUserID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.MemoryEntriesCreated != 0 || result.MemoryEntriesUpdated != 0 {
		t.Fatalf("expected no-op result, got %+v", result)
	}
}

func TestGenerateAndPersistForRepositoryIdempotentUpsert(t *testing.T) {
	repositoryID := uuid.New()
	prID := uuid.New()
	issueID := uuid.New()
	store := &fakeMemoryStore{
		repo: jobs.RepositoryForSync{
			ID:           repositoryID,
			Organization: uuid.New(),
			OwnerLogin:   "octo",
			Name:         "repo",
		},
		prs: []jobs.PullRequestForMemory{
			{ID: prID, RepositoryID: repositoryID, GitHubPrNumber: 1, Title: "Sync queue auth fix", Body: "retry behavior changed", State: "open", HTMLURL: "https://github.com/octo/repo/pull/1"},
		},
		issues: []jobs.IssueForMemory{
			{ID: issueID, RepositoryID: repositoryID, GitHubIssueNumber: 2, Title: "Billing issue", Body: "payment retries", State: "open", HTMLURL: "https://github.com/octo/repo/issues/2"},
		},
	}

	svc := NewMemoryGenerationService(store, NewDeterministicMemoryGenerator())
	first, err := svc.GenerateAndPersistForRepository(context.Background(), jobs.RepoGenerateMemoryPayload{
		RepositoryID: repositoryID,
	})
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if first.MemoryEntriesCreated != 2 {
		t.Fatalf("expected 2 created on first run, got %+v", first)
	}

	second, err := svc.GenerateAndPersistForRepository(context.Background(), jobs.RepoGenerateMemoryPayload{
		RepositoryID: repositoryID,
	})
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if second.MemoryEntriesUpdated != 2 {
		t.Fatalf("expected 2 updated on second run, got %+v", second)
	}
}

func TestGenerateAndPersistForRepositoryReturnsPersistenceError(t *testing.T) {
	repositoryID := uuid.New()
	store := &fakeMemoryStore{
		repo: jobs.RepositoryForSync{ID: repositoryID, Organization: uuid.New(), OwnerLogin: "octo", Name: "repo"},
		prs: []jobs.PullRequestForMemory{
			{ID: uuid.New(), RepositoryID: repositoryID, GitHubPrNumber: 1, Title: "PR title", State: "open", HTMLURL: "https://github.com/octo/repo/pull/1"},
		},
		upsertErr: errors.New("insert failed"),
	}

	svc := NewMemoryGenerationService(store, NewDeterministicMemoryGenerator())
	if _, err := svc.GenerateAndPersistForRepository(context.Background(), jobs.RepoGenerateMemoryPayload{RepositoryID: repositoryID}); err == nil {
		t.Fatal("expected error")
	}
}
