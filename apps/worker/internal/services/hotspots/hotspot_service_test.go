package hotspots

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type fakeStore struct {
	repo   jobs.RepositoryForSync
	prs    []jobs.PullRequestForHotspot
	issues []jobs.IssueForHotspot

	entries map[string]uuid.UUID
	links   map[string]struct{}
}

func (f *fakeStore) GetRepositoryForSync(context.Context, uuid.UUID) (jobs.RepositoryForSync, error) {
	return f.repo, nil
}

func (f *fakeStore) ListRecentPullRequestsForHotspots(context.Context, uuid.UUID, time.Time) ([]jobs.PullRequestForHotspot, error) {
	return f.prs, nil
}

func (f *fakeStore) ListRecentIssuesForHotspots(context.Context, uuid.UUID, time.Time) ([]jobs.IssueForHotspot, error) {
	return f.issues, nil
}

func (f *fakeStore) UpsertHotspotMemoryEntry(_ context.Context, record jobs.HotspotMemoryUpsertRecord) (uuid.UUID, bool, error) {
	if f.entries == nil {
		f.entries = map[string]uuid.UUID{}
	}
	if existing, ok := f.entries[record.HotspotKey]; ok {
		return existing, false, nil
	}
	id := uuid.New()
	f.entries[record.HotspotKey] = id
	return id, true, nil
}

func (f *fakeStore) LinkMemoryEntrySource(_ context.Context, memoryEntryID uuid.UUID, sourceType string, sourceRecordID uuid.UUID) error {
	if f.links == nil {
		f.links = map[string]struct{}{}
	}
	key := memoryEntryID.String() + "|" + sourceType + "|" + sourceRecordID.String()
	f.links[key] = struct{}{}
	return nil
}

func TestRecalculateForRepositoryIdempotentUpsert(t *testing.T) {
	repoID := uuid.New()
	orgID := uuid.New()
	now := time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC)

	store := &fakeStore{
		repo: jobs.RepositoryForSync{
			ID:           repoID,
			Organization: orgID,
			OwnerLogin:   "octo",
			Name:         "repo",
		},
		prs: []jobs.PullRequestForHotspot{
			{ID: uuid.New(), RepositoryID: repoID, GitHubPrNumber: 1, Title: "Sync retry path", Body: "queue retry sync", State: "open", HTMLURL: "https://example.com/pr/1", Labels: []string{"sync", "bug"}, UpdatedAtExternal: now.Add(-2 * time.Hour)},
			{ID: uuid.New(), RepositoryID: repoID, GitHubPrNumber: 2, Title: "Sync queue refactor", Body: "sync worker path", State: "closed", HTMLURL: "https://example.com/pr/2", Labels: []string{"sync"}, UpdatedAtExternal: now.Add(-3 * time.Hour)},
		},
		issues: []jobs.IssueForHotspot{
			{ID: uuid.New(), RepositoryID: repoID, GitHubIssueNumber: 10, Title: "Sync outage", Body: "retry queue failed", State: "open", HTMLURL: "https://example.com/issues/10", Labels: []string{"sync", "incident"}, UpdatedAtExternal: now.Add(-1 * time.Hour)},
		},
	}

	svc := NewService(store)
	svc.nowFn = func() time.Time { return now }

	first, err := svc.RecalculateForRepository(context.Background(), jobs.RepoRecalculateHotspotsPayload{
		RepositoryID: repoID,
	})
	if err != nil {
		t.Fatalf("first recalc failed: %v", err)
	}
	if first.HotspotsDetected != 1 || first.MemoryEntriesCreated != 1 {
		t.Fatalf("unexpected first result: %+v", first)
	}
	firstLinks := len(store.links)
	if firstLinks == 0 {
		t.Fatal("expected source links to be written")
	}

	second, err := svc.RecalculateForRepository(context.Background(), jobs.RepoRecalculateHotspotsPayload{
		RepositoryID: repoID,
	})
	if err != nil {
		t.Fatalf("second recalc failed: %v", err)
	}
	if second.MemoryEntriesUpdated != 1 || second.MemoryEntriesCreated != 0 {
		t.Fatalf("unexpected second result: %+v", second)
	}
	if len(store.links) != firstLinks {
		t.Fatalf("expected conflict-safe links with no growth on rerun, got before=%d after=%d", firstLinks, len(store.links))
	}
}

func TestRecalculateForRepositoryNoSignalNoEntries(t *testing.T) {
	repoID := uuid.New()
	store := &fakeStore{
		repo: jobs.RepositoryForSync{
			ID:           repoID,
			Organization: uuid.New(),
		},
		prs: []jobs.PullRequestForHotspot{
			{ID: uuid.New(), RepositoryID: repoID, GitHubPrNumber: 1, Title: "Docs update", Body: "readme", State: "closed", Labels: []string{"docs"}, UpdatedAtExternal: time.Now().UTC()},
		},
	}

	svc := NewService(store)
	result, err := svc.RecalculateForRepository(context.Background(), jobs.RepoRecalculateHotspotsPayload{
		RepositoryID: repoID,
	})
	if err != nil {
		t.Fatalf("recalculate failed: %v", err)
	}
	if result.HotspotsDetected != 0 || result.MemoryEntriesCreated != 0 {
		t.Fatalf("expected no-op, got %+v", result)
	}
}
