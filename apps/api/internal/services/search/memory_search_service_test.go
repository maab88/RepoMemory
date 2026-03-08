package search

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/maab88/repomemory/apps/api/internal/db"
	"github.com/maab88/repomemory/apps/api/internal/repositories"
)

type fakeMembershipChecker struct {
	allowed bool
	err     error
}

func (f *fakeMembershipChecker) UserHasMembership(_ context.Context, _ db.UserHasMembershipParams) (bool, error) {
	return f.allowed, f.err
}

type fakeRepositoryReader struct {
	repo db.Repository
	err  error
}

func (f *fakeRepositoryReader) GetByID(_ context.Context, _ uuid.UUID) (db.Repository, error) {
	return f.repo, f.err
}

type fakeMemorySearcher struct {
	rows []repositories.MemorySearchResultRow
	err  error
	got  repositories.MemorySearchInput
}

func (f *fakeMemorySearcher) Search(_ context.Context, input repositories.MemorySearchInput) ([]repositories.MemorySearchResultRow, error) {
	f.got = input
	return f.rows, f.err
}

func TestSearchMemoryReturnsEmptyOnBlankQuery(t *testing.T) {
	svc := NewService(
		&fakeMembershipChecker{allowed: true},
		&fakeRepositoryReader{},
		&fakeMemorySearcher{rows: []repositories.MemorySearchResultRow{{ID: uuid.New()}}},
	)

	result, err := svc.SearchMemory(context.Background(), MemorySearchInput{
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		Query:          "   ",
		Page:           0,
		PageSize:       0,
	})
	if err != nil {
		t.Fatalf("SearchMemory error: %v", err)
	}
	if result.Total != 0 || len(result.Results) != 0 {
		t.Fatalf("expected empty result, got total=%d len=%d", result.Total, len(result.Results))
	}
	if result.Page != DefaultPage || result.PageSize != DefaultPageSize {
		t.Fatalf("expected default pagination, got page=%d pageSize=%d", result.Page, result.PageSize)
	}
}

func TestSearchMemoryForbiddenWithoutMembership(t *testing.T) {
	svc := NewService(
		&fakeMembershipChecker{allowed: false},
		&fakeRepositoryReader{},
		&fakeMemorySearcher{},
	)

	_, err := svc.SearchMemory(context.Background(), MemorySearchInput{
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		Query:          "retry",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestSearchMemoryRejectsRepositoryOutsideOrg(t *testing.T) {
	orgID := uuid.New()
	svc := NewService(
		&fakeMembershipChecker{allowed: true},
		&fakeRepositoryReader{repo: db.Repository{ID: uuid.New(), OrganizationID: uuid.New()}},
		&fakeMemorySearcher{},
	)
	repoID := uuid.New()

	_, err := svc.SearchMemory(context.Background(), MemorySearchInput{
		UserID:         uuid.New(),
		OrganizationID: orgID,
		RepositoryID:   &repoID,
		Query:          "retry",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestSearchMemoryRepositoryNotFound(t *testing.T) {
	repoID := uuid.New()
	svc := NewService(
		&fakeMembershipChecker{allowed: true},
		&fakeRepositoryReader{err: pgx.ErrNoRows},
		&fakeMemorySearcher{},
	)

	_, err := svc.SearchMemory(context.Background(), MemorySearchInput{
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		RepositoryID:   &repoID,
		Query:          "retry",
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestSearchMemoryReturnsRowsWithSnippetAndTotal(t *testing.T) {
	now := time.Now().UTC()
	orgID := uuid.New()
	repoID := uuid.New()
	filterRepoID := repoID
	searcher := &fakeMemorySearcher{
		rows: []repositories.MemorySearchResultRow{
			{
				ID:             uuid.New(),
				RepositoryID:   repoID,
				RepositoryName: "RepoMemory",
				Type:           "pr_summary",
				Title:          "Retry refactor",
				Summary:        "Moved retry scheduling into worker queue and updated ownership boundaries for sync orchestration.",
				SourceURL:      "https://github.com/org/repo/pull/1",
				CreatedAt:      now,
				TotalCount:     2,
			},
		},
	}
	svc := NewService(
		&fakeMembershipChecker{allowed: true},
		&fakeRepositoryReader{repo: db.Repository{ID: repoID, OrganizationID: orgID}},
		searcher,
	)

	result, err := svc.SearchMemory(context.Background(), MemorySearchInput{
		UserID:         uuid.New(),
		OrganizationID: orgID,
		RepositoryID:   &filterRepoID,
		Query:          "retry",
		Page:           2,
		PageSize:       15,
	})
	if err != nil {
		t.Fatalf("SearchMemory error: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if result.Results[0].SummarySnippet == "" {
		t.Fatal("expected non-empty summary snippet")
	}
	if searcher.got.RepositoryID == nil || *searcher.got.RepositoryID != filterRepoID {
		t.Fatalf("expected repository filter to be forwarded")
	}
	if searcher.got.Offset != 15 || searcher.got.Limit != 15 {
		t.Fatalf("expected pagination offset/limit, got offset=%d limit=%d", searcher.got.Offset, searcher.got.Limit)
	}
}
