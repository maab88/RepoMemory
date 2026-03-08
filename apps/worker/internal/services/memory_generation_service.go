package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type MemoryGenerationStore interface {
	GetRepositoryForSync(ctx context.Context, repositoryID uuid.UUID) (jobs.RepositoryForSync, error)
	ListPullRequestsForRepository(ctx context.Context, repositoryID uuid.UUID) ([]jobs.PullRequestForMemory, error)
	ListIssuesForRepository(ctx context.Context, repositoryID uuid.UUID) ([]jobs.IssueForMemory, error)
	UpsertMemoryEntryForSource(ctx context.Context, record jobs.MemoryEntryUpsertRecord) (uuid.UUID, bool, error)
}

type MemoryGenerationResult struct {
	RepositoryID          uuid.UUID
	PullRequestsProcessed int
	IssuesProcessed       int
	MemoryEntriesCreated  int
	MemoryEntriesUpdated  int
	Skipped               int
}

type MemoryDraftGenerator interface {
	GenerateFromPullRequest(pr jobs.PullRequestForMemory) (MemoryEntryDraft, bool)
	GenerateFromIssue(issue jobs.IssueForMemory) (MemoryEntryDraft, bool)
}

type MemoryGenerationService struct {
	store      MemoryGenerationStore
	generator  MemoryDraftGenerator
	sourceType struct {
		pullRequest string
		issue       string
	}
}

func NewMemoryGenerationService(store MemoryGenerationStore, generator MemoryDraftGenerator) *MemoryGenerationService {
	return &MemoryGenerationService{
		store:     store,
		generator: generator,
		sourceType: struct {
			pullRequest string
			issue       string
		}{
			pullRequest: "pull_request",
			issue:       "issue",
		},
	}
}

func (s *MemoryGenerationService) GenerateAndPersistForRepository(ctx context.Context, payload jobs.RepoGenerateMemoryPayload) (MemoryGenerationResult, error) {
	repo, err := s.store.GetRepositoryForSync(ctx, payload.RepositoryID)
	if err != nil {
		return MemoryGenerationResult{}, fmt.Errorf("load repository for memory generation: %w", err)
	}

	result := MemoryGenerationResult{
		RepositoryID: payload.RepositoryID,
	}

	prs, err := s.store.ListPullRequestsForRepository(ctx, payload.RepositoryID)
	if err != nil {
		return result, fmt.Errorf("list pull requests: %w", err)
	}
	for _, pr := range prs {
		result.PullRequestsProcessed++
		draft, ok := s.generator.GenerateFromPullRequest(pr)
		if !ok {
			result.Skipped++
			continue
		}
		_, created, err := s.store.UpsertMemoryEntryForSource(ctx, jobs.MemoryEntryUpsertRecord{
			OrganizationID: repo.Organization,
			RepositoryID:   payload.RepositoryID,
			Type:           draft.Type,
			Title:          draft.Title,
			Summary:        draft.Summary,
			WhyItMatters:   draft.WhyItMatters,
			ImpactedAreas:  draft.ImpactedAreas,
			Risks:          draft.Risks,
			FollowUps:      draft.FollowUps,
			SourceKind:     s.sourceType.pullRequest,
			SourceID:       pr.ID,
			SourceURL:      pr.HTMLURL,
			GeneratedBy:    draft.GeneratedBy,
			SourceType:     s.sourceType.pullRequest,
			SourceRecordID: pr.ID,
		})
		if err != nil {
			return result, fmt.Errorf("persist pull request memory (pr id %s): %w", pr.ID, err)
		}
		if created {
			result.MemoryEntriesCreated++
		} else {
			result.MemoryEntriesUpdated++
		}
	}

	issues, err := s.store.ListIssuesForRepository(ctx, payload.RepositoryID)
	if err != nil {
		return result, fmt.Errorf("list issues: %w", err)
	}
	for _, issue := range issues {
		result.IssuesProcessed++
		draft, ok := s.generator.GenerateFromIssue(issue)
		if !ok {
			result.Skipped++
			continue
		}
		_, created, err := s.store.UpsertMemoryEntryForSource(ctx, jobs.MemoryEntryUpsertRecord{
			OrganizationID: repo.Organization,
			RepositoryID:   payload.RepositoryID,
			Type:           draft.Type,
			Title:          draft.Title,
			Summary:        draft.Summary,
			WhyItMatters:   draft.WhyItMatters,
			ImpactedAreas:  draft.ImpactedAreas,
			Risks:          draft.Risks,
			FollowUps:      draft.FollowUps,
			SourceKind:     s.sourceType.issue,
			SourceID:       issue.ID,
			SourceURL:      issue.HTMLURL,
			GeneratedBy:    draft.GeneratedBy,
			SourceType:     s.sourceType.issue,
			SourceRecordID: issue.ID,
		})
		if err != nil {
			return result, fmt.Errorf("persist issue memory (issue id %s): %w", issue.ID, err)
		}
		if created {
			result.MemoryEntriesCreated++
		} else {
			result.MemoryEntriesUpdated++
		}
	}

	return result, nil
}
