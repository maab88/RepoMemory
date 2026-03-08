package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type fakeDigestStore struct {
	repo         jobs.RepositoryForSync
	prs          []jobs.PullRequestForDigest
	issues       []jobs.IssueForDigest
	memory       []jobs.MemoryEntryForDigest
	upsertID     uuid.UUID
	upsertCreate bool
	lastUpsert   jobs.DigestUpsertRecord
}

func (f *fakeDigestStore) GetRepositoryForSync(context.Context, uuid.UUID) (jobs.RepositoryForSync, error) {
	return f.repo, nil
}
func (f *fakeDigestStore) ListPullRequestsForDigestPeriod(context.Context, uuid.UUID, time.Time, time.Time) ([]jobs.PullRequestForDigest, error) {
	return f.prs, nil
}
func (f *fakeDigestStore) ListOpenIssuesForDigestPeriod(context.Context, uuid.UUID, time.Time) ([]jobs.IssueForDigest, error) {
	return f.issues, nil
}
func (f *fakeDigestStore) ListMemoryEntriesForDigestPeriod(context.Context, uuid.UUID, time.Time, time.Time) ([]jobs.MemoryEntryForDigest, error) {
	return f.memory, nil
}
func (f *fakeDigestStore) UpsertDigestForPeriod(_ context.Context, record jobs.DigestUpsertRecord) (uuid.UUID, bool, error) {
	f.lastUpsert = record
	return f.upsertID, f.upsertCreate, nil
}

func TestDigestGenerationServiceBuildsLowActivityDigest(t *testing.T) {
	store := &fakeDigestStore{
		repo: jobs.RepositoryForSync{
			ID:           uuid.New(),
			Organization: uuid.New(),
			OwnerLogin:   "octocat",
			Name:         "repo-memory",
		},
		upsertID:     uuid.New(),
		upsertCreate: true,
	}
	svc := NewDigestGenerationService(store, NewDeterministicDigestBuilder())
	svc.nowFn = func() time.Time {
		return time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)
	}

	result, err := svc.GenerateAndPersistForRepository(context.Background(), jobs.RepoGenerateDigestPayload{
		RepositoryID:      store.repo.ID,
		OrganizationID:    store.repo.Organization,
		TriggeredByUserID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("GenerateAndPersistForRepository error: %v", err)
	}
	if result.DigestID != store.upsertID {
		t.Fatalf("expected digest id %s, got %s", store.upsertID, result.DigestID)
	}
	if store.lastUpsert.Summary == "" || store.lastUpsert.BodyMarkdown == "" {
		t.Fatal("expected summary/body to be generated for low activity")
	}
}

func TestDigestGenerationServiceReusesExistingPeriodDigest(t *testing.T) {
	repoID := uuid.New()
	store := &fakeDigestStore{
		repo: jobs.RepositoryForSync{
			ID:           repoID,
			Organization: uuid.New(),
			OwnerLogin:   "octocat",
			Name:         "repo-memory",
		},
		prs: []jobs.PullRequestForDigest{
			{ID: uuid.New(), GitHubPrNumber: 12, Title: "Refactor sync queue", HTMLURL: "https://github.com/org/repo/pull/12"},
		},
		upsertID:     uuid.New(),
		upsertCreate: false,
	}
	svc := NewDigestGenerationService(store, NewDeterministicDigestBuilder())
	svc.nowFn = func() time.Time {
		return time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)
	}

	result, err := svc.GenerateAndPersistForRepository(context.Background(), jobs.RepoGenerateDigestPayload{
		RepositoryID:      repoID,
		OrganizationID:    store.repo.Organization,
		TriggeredByUserID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("GenerateAndPersistForRepository error: %v", err)
	}
	if result.Created {
		t.Fatal("expected existing digest row to be reused")
	}
}
