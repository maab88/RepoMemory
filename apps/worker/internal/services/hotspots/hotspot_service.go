package hotspots

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type Store interface {
	GetRepositoryForSync(ctx context.Context, repositoryID uuid.UUID) (jobs.RepositoryForSync, error)
	ListRecentPullRequestsForHotspots(ctx context.Context, repositoryID uuid.UUID, since time.Time) ([]jobs.PullRequestForHotspot, error)
	ListRecentIssuesForHotspots(ctx context.Context, repositoryID uuid.UUID, since time.Time) ([]jobs.IssueForHotspot, error)
	UpsertHotspotMemoryEntry(ctx context.Context, record jobs.HotspotMemoryUpsertRecord) (uuid.UUID, bool, error)
	LinkMemoryEntrySource(ctx context.Context, memoryEntryID uuid.UUID, sourceType string, sourceRecordID uuid.UUID) error
}

type RecalculationResult struct {
	RepositoryID         uuid.UUID
	WindowStart          time.Time
	WindowEnd            time.Time
	SourcesAnalyzed      int
	HotspotsDetected     int
	MemoryEntriesCreated int
	MemoryEntriesUpdated int
}

type Service struct {
	store Store
	nowFn func() time.Time
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
		nowFn: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) RecalculateForRepository(ctx context.Context, payload jobs.RepoRecalculateHotspotsPayload) (RecalculationResult, error) {
	windowEnd := s.nowFn().UTC()
	windowStart := windowEnd.AddDate(0, 0, -AnalysisWindowDays)
	result := RecalculationResult{
		RepositoryID: payload.RepositoryID,
		WindowStart:  windowStart,
		WindowEnd:    windowEnd,
	}

	repo, err := s.store.GetRepositoryForSync(ctx, payload.RepositoryID)
	if err != nil {
		return result, fmt.Errorf("load repository: %w", err)
	}

	prs, err := s.store.ListRecentPullRequestsForHotspots(ctx, payload.RepositoryID, windowStart)
	if err != nil {
		return result, fmt.Errorf("list recent pull requests: %w", err)
	}
	issues, err := s.store.ListRecentIssuesForHotspots(ctx, payload.RepositoryID, windowStart)
	if err != nil {
		return result, fmt.Errorf("list recent issues: %w", err)
	}
	result.SourcesAnalyzed = len(prs) + len(issues)

	candidates := Detect(prs, issues)
	result.HotspotsDetected = len(candidates)
	if len(candidates) == 0 {
		return result, nil
	}

	for _, candidate := range candidates {
		primaryURL := ""
		if len(candidate.Sources) > 0 {
			primaryURL = candidate.Sources[0].HTMLURL
		}
		entryID, created, err := s.store.UpsertHotspotMemoryEntry(ctx, jobs.HotspotMemoryUpsertRecord{
			OrganizationID: repo.Organization,
			RepositoryID:   payload.RepositoryID,
			HotspotKey:     HotspotKey(candidate.Theme),
			Title:          BuildTitle(candidate.Theme),
			Summary:        BuildSummary(candidate),
			WhyItMatters:   BuildWhyItMatters(candidate.Theme, candidate.BugOriented),
			ImpactedAreas:  BuildImpactedAreas(candidate.Theme),
			Risks:          BuildRisks(candidate.Theme, candidate.BugOriented),
			FollowUps:      BuildFollowUps(candidate.Theme),
			SourceURL:      primaryURL,
			GeneratedBy:    "deterministic",
		})
		if err != nil {
			return result, fmt.Errorf("upsert hotspot memory for %s: %w", candidate.Theme, err)
		}
		if created {
			result.MemoryEntriesCreated++
		} else {
			result.MemoryEntriesUpdated++
		}

		for _, source := range candidate.Sources {
			if err := s.store.LinkMemoryEntrySource(ctx, entryID, source.SourceType, source.SourceID); err != nil {
				return result, fmt.Errorf("link hotspot source (%s:%s): %w", source.SourceType, source.SourceID, err)
			}
		}
	}

	return result, nil
}
