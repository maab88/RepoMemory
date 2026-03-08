package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type DigestGenerationStore interface {
	GetRepositoryForSync(ctx context.Context, repositoryID uuid.UUID) (jobs.RepositoryForSync, error)
	ListPullRequestsForDigestPeriod(ctx context.Context, repositoryID uuid.UUID, periodStart, periodEnd time.Time) ([]jobs.PullRequestForDigest, error)
	ListOpenIssuesForDigestPeriod(ctx context.Context, repositoryID uuid.UUID, periodEnd time.Time) ([]jobs.IssueForDigest, error)
	ListMemoryEntriesForDigestPeriod(ctx context.Context, repositoryID uuid.UUID, periodStart, periodEnd time.Time) ([]jobs.MemoryEntryForDigest, error)
	UpsertDigestForPeriod(ctx context.Context, record jobs.DigestUpsertRecord) (uuid.UUID, bool, error)
}

type DigestGenerationResult struct {
	RepositoryID uuid.UUID
	DigestID     uuid.UUID
	Created      bool
	PeriodStart  time.Time
	PeriodEnd    time.Time
}

type DigestBuilder interface {
	Build(input DigestBuildInput) DigestDraft
}

type DigestGenerationService struct {
	store   DigestGenerationStore
	builder DigestBuilder
	nowFn   func() time.Time
}

func NewDigestGenerationService(store DigestGenerationStore, builder DigestBuilder) *DigestGenerationService {
	return &DigestGenerationService{
		store:   store,
		builder: builder,
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (s *DigestGenerationService) GenerateAndPersistForRepository(ctx context.Context, payload jobs.RepoGenerateDigestPayload) (DigestGenerationResult, error) {
	repo, err := s.store.GetRepositoryForSync(ctx, payload.RepositoryID)
	if err != nil {
		return DigestGenerationResult{}, fmt.Errorf("load repository for digest generation: %w", err)
	}

	periodStart, periodEnd := calendarWeekWindow(s.nowFn())

	mergedPRs, err := s.store.ListPullRequestsForDigestPeriod(ctx, payload.RepositoryID, periodStart, periodEnd)
	if err != nil {
		return DigestGenerationResult{}, fmt.Errorf("list merged pull requests: %w", err)
	}
	openIssues, err := s.store.ListOpenIssuesForDigestPeriod(ctx, payload.RepositoryID, periodEnd)
	if err != nil {
		return DigestGenerationResult{}, fmt.Errorf("list open issues: %w", err)
	}
	memoryEntries, err := s.store.ListMemoryEntriesForDigestPeriod(ctx, payload.RepositoryID, periodStart, periodEnd)
	if err != nil {
		return DigestGenerationResult{}, fmt.Errorf("list memory entries: %w", err)
	}

	draft := s.builder.Build(DigestBuildInput{
		RepositoryFullName: fmt.Sprintf("%s/%s", repo.OwnerLogin, repo.Name),
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		MergedPullRequests: mergedPRs,
		OpenIssues:         openIssues,
		MemoryEntries:      memoryEntries,
	})

	digestID, created, err := s.store.UpsertDigestForPeriod(ctx, jobs.DigestUpsertRecord{
		OrganizationID: repo.Organization,
		RepositoryID:   payload.RepositoryID,
		PeriodStart:    periodStart,
		PeriodEnd:      periodEnd,
		Title:          draft.Title,
		Summary:        draft.Summary,
		BodyMarkdown:   draft.BodyMarkdown,
		GeneratedBy:    draft.GeneratedBy,
	})
	if err != nil {
		return DigestGenerationResult{}, fmt.Errorf("upsert digest: %w", err)
	}

	return DigestGenerationResult{
		RepositoryID: payload.RepositoryID,
		DigestID:     digestID,
		Created:      created,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	}, nil
}

func calendarWeekWindow(now time.Time) (time.Time, time.Time) {
	now = now.UTC()
	weekday := int(now.Weekday())
	// Make Monday the week start.
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(weekday - 1))
	end := start.AddDate(0, 0, 7).Add(-time.Second)
	return start, end
}
